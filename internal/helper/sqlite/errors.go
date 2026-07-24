// Package sqlite provides sentinel errors for the SQLite database helper.
package sqlite

import "errors"

// Sentinel errors for the sqlite package.
var (
	ErrOpenDB   = errors.New("sqlite: failed to open database")
	ErrCloseDB  = errors.New("sqlite: failed to close database")
	ErrQuery    = errors.New("sqlite: query failed")
	ErrNotFound = errors.New("sqlite: not found")
)
