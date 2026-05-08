// Package repos contains PostgreSQL implementations of domain repositories.
package repos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LogRepo persists and queries request_logs.
type LogRepo struct {
	pool *pgxpool.Pool
}

// NewLogRepo constructs a LogRepo for the given pool.
func NewLogRepo(pool *pgxpool.Pool) *LogRepo {
	return &LogRepo{pool: pool}
}

// Insert writes a single log row. Prefer InsertBatch in hot paths.
func (r *LogRepo) Insert(ctx context.Context, log entity.RequestLog) error {
	const query = `
        INSERT INTO request_logs
            (request_id, method, path, query, route_id, target_url, user_id, client_ip,
             status_code, response_time_ms, trace_id, created_at)
        VALUES
            ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
    `
	_, err := r.pool.Exec(ctx, query,
		log.RequestID,
		log.Method,
		log.Path,
		log.Query,
		log.RouteID,
		log.TargetURL,
		log.UserID,
		nullableIP(log.ClientIP),
		log.StatusCode,
		log.ResponseTimeMs,
		log.TraceID,
		log.CreatedAt,
	)
	return err
}

// InsertBatch performs a single COPY FROM round-trip, far cheaper than per-row
// inserts. Useful for the AsyncLogger's batched path.
func (r *LogRepo) InsertBatch(ctx context.Context, logs []entity.RequestLog) error {
	if len(logs) == 0 {
		return nil
	}
	cols := []string{
		"request_id", "method", "path", "query", "route_id", "target_url",
		"user_id", "client_ip", "status_code", "response_time_ms", "trace_id", "created_at",
	}
	rows := pgx.CopyFromSlice(len(logs), func(i int) ([]any, error) {
		l := logs[i]
		return []any{
			l.RequestID,
			l.Method,
			l.Path,
			l.Query,
			l.RouteID,
			l.TargetURL,
			l.UserID,
			nullableIP(l.ClientIP),
			l.StatusCode,
			l.ResponseTimeMs,
			l.TraceID,
			l.CreatedAt,
		}, nil
	})
	_, err := r.pool.CopyFrom(ctx, pgx.Identifier{"request_logs"}, cols, rows)
	return err
}

// GetWithFilters returns paginated, filtered request logs and the total count.
func (r *LogRepo) GetWithFilters(ctx context.Context, filters repository.LogFilters, limit, offset int) ([]entity.RequestLog, int, error) {
	whereParts := []string{}
	args := []any{}
	idx := 1

	if filters.Path != "" {
		whereParts = append(whereParts, fmt.Sprintf("path ILIKE $%d", idx))
		args = append(args, "%"+filters.Path+"%")
		idx++
	}
	if filters.StatusCode != 0 {
		whereParts = append(whereParts, fmt.Sprintf("status_code = $%d", idx))
		args = append(args, filters.StatusCode)
		idx++
	}
	if filters.From != nil {
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, *filters.From)
		idx++
	}
	if filters.To != nil {
		whereParts = append(whereParts, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, *filters.To)
		idx++
	}

	where := ""
	if len(whereParts) > 0 {
		where = "WHERE " + strings.Join(whereParts, " AND ")
	}

	dataQuery := fmt.Sprintf(`
        SELECT id, request_id, method, path, query, route_id, user_id, host(client_ip) AS client_ip,
               status_code, response_time_ms, target_url, trace_id, created_at
        FROM request_logs
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, idx, idx+1)

	dataArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []entity.RequestLog
	for rows.Next() {
		var l entity.RequestLog
		if err := rows.Scan(
			&l.ID,
			&l.RequestID,
			&l.Method,
			&l.Path,
			&l.Query,
			&l.RouteID,
			&l.UserID,
			&l.ClientIP,
			&l.StatusCode,
			&l.ResponseTimeMs,
			&l.TargetURL,
			&l.TraceID,
			&l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM request_logs %s`, where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// GetMetrics returns aggregated stats for the [from, to] window.
func (r *LogRepo) GetMetrics(ctx context.Context, from, to time.Time) (entity.Metrics, error) {
	const queryStats = `
        SELECT COUNT(*) AS total_requests,
               COALESCE(AVG(response_time_ms), 0) AS avg_response_time
        FROM request_logs
        WHERE created_at BETWEEN $1 AND $2
    `
	var total int64
	var avgTime float64
	if err := r.pool.QueryRow(ctx, queryStats, from, to).Scan(&total, &avgTime); err != nil {
		return entity.Metrics{}, err
	}

	const queryStatus = `
        SELECT status_code, COUNT(*)
        FROM request_logs
        WHERE created_at BETWEEN $1 AND $2
        GROUP BY status_code
        ORDER BY status_code
    `
	statusRows, err := r.pool.Query(ctx, queryStatus, from, to)
	if err != nil {
		return entity.Metrics{}, err
	}
	defer statusRows.Close()
	statusCounts := make(map[int]int64)
	for statusRows.Next() {
		var code int
		var cnt int64
		if err := statusRows.Scan(&code, &cnt); err != nil {
			return entity.Metrics{}, err
		}
		statusCounts[code] = cnt
	}

	const querySlow = `
        SELECT path, AVG(response_time_ms) AS avg_time, COUNT(*) AS cnt
        FROM request_logs
        WHERE created_at BETWEEN $1 AND $2
        GROUP BY path
        ORDER BY avg_time DESC
        LIMIT 5
    `
	slowRows, err := r.pool.Query(ctx, querySlow, from, to)
	if err != nil {
		return entity.Metrics{}, err
	}
	defer slowRows.Close()
	slowest := []entity.RouteStat{}
	for slowRows.Next() {
		var path string
		var avgT float64
		var cnt int64
		if err := slowRows.Scan(&path, &avgT, &cnt); err != nil {
			return entity.Metrics{}, err
		}
		slowest = append(slowest, entity.RouteStat{Path: path, AvgTimeMs: avgT, Count: cnt})
	}

	duration := to.Sub(from).Seconds()
	rps := 0.0
	if duration > 0 {
		rps = float64(total) / duration
	}

	return entity.Metrics{
		TotalRequests:   total,
		RPS:             rps,
		AvgResponseTime: avgTime,
		StatusCounts:    statusCounts,
		SlowestRoutes:   slowest,
	}, nil
}

// nullableIP turns an empty string into nil so PostgreSQL stores NULL in the
// INET column rather than failing on a parse error.
func nullableIP(ip string) any {
	if strings.TrimSpace(ip) == "" {
		return nil
	}
	return ip
}
