// Package main implements the API Gateway binary entrypoint.
// It validates configuration, runs database migrations, wires the dependency
// graph, starts the public proxy and admin API in parallel, and gracefully
// shuts everything down on SIGINT/SIGTERM.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prodlich/internal/application"
	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
	"prodlich/internal/infrastructure/auth"
	"prodlich/internal/infrastructure/config"
	"prodlich/internal/infrastructure/db"
	"prodlich/internal/infrastructure/limiter"
	"prodlich/internal/infrastructure/log"
	"prodlich/internal/infrastructure/metrics"
	repos "prodlich/internal/infrastructure/repo"
	"prodlich/internal/infrastructure/tracing"
	"prodlich/internal/model"
	"prodlich/internal/presentation/admin"
	"prodlich/internal/presentation/public"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"
)

const (
	defaultShutdownTimeout = 30 * time.Second
	httpReadTimeout        = 15 * time.Second
	httpWriteTimeout       = 15 * time.Second
	httpIdleTimeout        = 60 * time.Second
)

func main() {
	migrateFlag := flag.String("migrate", "", "run migrations and exit: up, down, or downN where N is the number of steps (e.g. down1)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	log.Init(cfg.App.LogLvl, cfg.App.LogFormat)
	log.L().WithFields(map[string]any{
		"public_port": cfg.App.Port,
		"admin_port":  cfg.App.AdminPort,
		"log_level":   cfg.App.LogLvl,
		"tls":         cfg.App.TLSEnabled,
	}).Info("starting API Gateway")

	if *migrateFlag != "" {
		if err := runMigrationsCLI(*migrateFlag, &cfg.DB); err != nil {
			log.L().WithError(err).Fatal("migrate failed")
		}
		return
	}

	if err := run(cfg); err != nil {
		log.L().WithError(err).Error("gateway exited with error")
		os.Exit(1)
	}
	log.L().Info("API Gateway stopped")
}

func run(cfg config.AppConfig) error {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	if err := db.Connect(rootCtx, &cfg.DB); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	if err := db.MigrateUp(&cfg.DB); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	traceShutdown, err := tracing.Init(rootCtx, cfg.App.TracingServiceName, cfg.App.TracingEnabled)
	if err != nil {
		log.L().WithError(err).Warn("tracing init failed; continuing without traces")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = traceShutdown(ctx)
	}()

	// Repositories.
	routeRepo := repos.NewRouteRepo(db.Pool)
	logRepo := repos.NewLogRepo(db.Pool)
	adminRepo := repos.NewAdminRepo(db.Pool)
	userRepo := repos.NewUserRepoPG(db.Pool)
	envRepo := repos.NewEnvironmentRepo(db.Pool)
	clusterRepo := repos.NewClusterRepo(db.Pool)
	middlewareRepo := repos.NewMiddlewareRepo(db.Pool)
	apiTokenRepo := repos.NewAPITokenRepo(db.Pool)

	if err := bootstrapAdmin(rootCtx, adminRepo, cfg.App.AdminUsername, cfg.App.AdminPassword); err != nil {
		return fmt.Errorf("bootstrap admin: %w", err)
	}

	// Use cases.
	jwtService := auth.NewJWTService(cfg.App.JWTSecret)
	adminAuthUC := application.NewAdminAuthUseCase(adminRepo, jwtService)
	routeUC := application.NewRouterService(routeRepo, true)
	logUC := application.NewLogUseCase(logRepo)
	metricUC := application.NewMetricUseCase(logRepo)
	authUC := application.NewAuthUseCase(userRepo, cfg.App.JWTSecret)
	environmentUC := application.NewEnvironmentUseCase(envRepo)
	clusterUC := application.NewClusterUseCase(clusterRepo)
	middlewareUC := application.NewMiddlewareUseCase(middlewareRepo)
	userMgmtUC := application.NewUserUseCase(adminRepo)
	apiTokenUC := application.NewAPITokenUseCase(apiTokenRepo)

	if err := routeUC.ReloadCache(rootCtx); err != nil {
		return fmt.Errorf("load routes cache: %w", err)
	}

	// Cross-cutting infrastructure.
	asyncLogger := log.NewAsyncLogger(
		logRepo,
		cfg.App.LogBufferSize,
		cfg.App.LogBatchSize,
		cfg.App.LogFlushPeriod,
	)
	ipLimiter := limiter.NewIPLimiter(cfg.App.RateLimitTTL)

	metricReg := metrics.New()
	metricReg.StartPoolWatcher(rootCtx, db.Pool, 5*time.Second)
	routeUC.OnReload(func(routes []entity.Route) {
		metricReg.ActiveRoutes.Set(float64(len(routes)))
	})
	metricReg.ActiveRoutes.Set(float64(len(routeUC.GetActiveRoutes())))

	// Servers.
	publicHandler := public.Handler(public.ProxyDeps{
		Routes:        routeUC,
		IPLimiter:     ipLimiter,
		AdminAuth:     adminAuthUC,
		UserAuth:      authUC,
		AsyncLog:      asyncLogger,
		DBHealth:      dbHealth{},
		MaxBodyBytes:  cfg.App.MaxBodyBytes,
		RateLimitHits: metricReg.RateLimitHits,
	})
	publicHandler = metricReg.HTTPMiddleware("public")(publicHandler)

	adminHandler := admin.NewRouter(admin.RouterConfig{
		RouteUC:       routeUC,
		AuthUC:        adminAuthUC,
		LogUC:         logUC,
		MetricUC:      metricUC,
		EnvironmentUC: environmentUC,
		ClusterUC:     clusterUC,
		MiddlewareUC:  middlewareUC,
		UserUC:        userMgmtUC,
		TokenUC:       apiTokenUC,
		Metrics:       metricReg,
		ReadyFn:       readinessHandler(),
	})

	publicServer := newServer(":"+cfg.App.Port, publicHandler)
	adminServer := newServer(":"+cfg.App.AdminPort, adminHandler)

	signalCtx, signalStop := signal.NotifyContext(rootCtx, syscall.SIGINT, syscall.SIGTERM)
	defer signalStop()

	g, gctx := errgroup.WithContext(signalCtx)

	g.Go(func() error {
		return startServer(publicServer, "public", cfg.App)
	})
	g.Go(func() error {
		return startServer(adminServer, "admin", cfg.App)
	})
	g.Go(func() error {
		<-gctx.Done()
		return shutdownAll(publicServer, adminServer, asyncLogger, ipLimiter)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadTimeout:       httpReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       httpIdleTimeout,
	}
}

// startServer launches HTTP or HTTPS depending on TLS config and treats
// http.ErrServerClosed as a clean exit.
func startServer(srv *http.Server, name string, app model.Config) error {
	log.L().WithFields(map[string]any{
		"component": name,
		"addr":      srv.Addr,
		"tls":       app.TLSEnabled,
	}).Info("server listening")
	var err error
	if app.TLSEnabled {
		err = srv.ListenAndServeTLS(app.CertFile, app.KeyFile)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s server: %w", name, err)
	}
	return nil
}

func shutdownAll(publicSrv, adminSrv *http.Server, alog *log.AsyncLogger, lim *limiter.IPLimiter) error {
	log.L().Info("graceful shutdown initiated")
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return publicSrv.Shutdown(gctx) })
	g.Go(func() error { return adminSrv.Shutdown(gctx) })
	if err := g.Wait(); err != nil {
		log.L().WithError(err).Warn("server shutdown error")
	}
	if alog != nil {
		alog.Close()
	}
	if lim != nil {
		lim.Stop()
	}
	return nil
}

// dbHealth adapts db.HealthCheck to the public.HealthChecker interface.
type dbHealth struct{}

func (dbHealth) HealthCheck(ctx context.Context) error { return db.HealthCheck(ctx) }

func readinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")
		if err := db.HealthCheck(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","reason":"db ping failed"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}

func bootstrapAdmin(ctx context.Context, repo repository.AdminRepo, username, password string) error {
	if username == "" || password == "" {
		return nil
	}
	if _, err := repo.GetByUsername(ctx, username); err == nil {
		return nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}
	if _, err := repo.CreateAdmin(ctx, entity.AdminUser{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
		CreatedAt:    time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("create bootstrap admin: %w", err)
	}
	log.L().WithField("username", username).Info("bootstrap admin created")
	return nil
}

func runMigrationsCLI(arg string, dbCfg *model.DB) error {
	switch arg {
	case "up":
		if err := db.MigrateUp(dbCfg); err != nil {
			return err
		}
		log.L().Info("migrations up applied")
	case "down":
		if err := db.MigrateDown(dbCfg, -1); err != nil {
			return err
		}
		log.L().Info("all migrations rolled back")
	default:
		return fmt.Errorf("unsupported -migrate value %q (use 'up' or 'down')", arg)
	}
	return nil
}
