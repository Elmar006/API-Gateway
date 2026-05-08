package entity

import "time"

// Environment groups routes by deployment target (production, staging, …).
// Routes may belong to one environment or none (legacy / shared).
type Environment struct {
	ID          int       `json:"id"           db:"id"`
	Name        string    `json:"name"         db:"name"`
	BaseDomain  string    `json:"base_domain"  db:"base_domain"`
	Color       string    `json:"color"        db:"color"`
	Description string    `json:"description"  db:"description"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"   db:"updated_at"`
}
