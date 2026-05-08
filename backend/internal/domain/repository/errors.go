// Package repository defines interfaces for data persistence and access patterns.
package repository

import "errors"

// ErrNotFound is returned when a requested entity is not found in the database.
var ErrNotFound = errors.New("not found")
