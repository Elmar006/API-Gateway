package entity

import "time"

// Cluster strategies for selecting a target on each request.
const (
	ClusterStrategyRoundRobin = "round_robin"
	ClusterStrategyWeighted   = "weighted"
	ClusterStrategyLeastConn  = "least_conn"
)

// Cluster groups multiple upstream targets behind a single logical endpoint.
// Routes may either keep their own target_url (single backend) or reference
// a cluster (load-balanced over multiple backends).
type Cluster struct {
	ID          int       `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Strategy    string    `json:"strategy"    db:"strategy"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`

	// Targets is populated by repository.GetByID / GetAll for convenience;
	// it is not persisted on Cluster directly.
	Targets []ClusterTarget `json:"targets,omitempty"`
}

// ClusterTarget is one upstream backend belonging to a cluster.
type ClusterTarget struct {
	ID         int        `json:"id"          db:"id"`
	ClusterID  int        `json:"cluster_id"  db:"cluster_id"`
	URL        string     `json:"url"         db:"url"`
	Weight     int        `json:"weight"      db:"weight"`
	IsHealthy  bool       `json:"is_healthy"  db:"is_healthy"`
	LastCheck  *time.Time `json:"last_check,omitempty" db:"last_check"`
	CreatedAt  time.Time  `json:"created_at"  db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"  db:"updated_at"`
}
