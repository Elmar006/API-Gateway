// Package entity defines domain models used across the gateway.
package entity

// Metrics represents aggregated statistics for the gateway's request history.
type Metrics struct {
	TotalRequests   int64         `json:"total_requests"`
	RPS             float64       `json:"rps"`
	AvgResponseTime float64       `json:"avg_response_time_ms"`
	StatusCounts    map[int]int64 `json:"status_counts"`
	SlowestRoutes   []RouteStat   `json:"slowest_routes"`
}

type RouteStat struct {
	Path      string  `json:"path"`
	AvgTimeMs float64 `json:"avg_time_ms"`
	Count     int64   `json:"count"`
}
