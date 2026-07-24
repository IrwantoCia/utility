// Package sqlite provides a pure-Go (CGO-free) SQLite database helper using
// modernc.org/sqlite. It wraps database/sql to offer common SQLite operations:
// opening databases, listing tables, introspecting columns and indexes,
// counting rows, and executing paginated queries.
package sqlite

import (
	"bytes"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// FilterOp represents a comparison operator for a filter condition.
type FilterOp string

const (
	OpEQ      FilterOp = "="
	OpNE      FilterOp = "!="
	OpGT      FilterOp = ">"
	OpLT      FilterOp = "<"
	OpGTE     FilterOp = ">="
	OpLTE     FilterOp = "<="
	OpLIKE    FilterOp = "LIKE"
	OpNOTLIKE FilterOp = "NOT LIKE"
)

// Filter represents a single WHERE condition: column operator value.
type Filter struct {
	Column string
	Op     FilterOp
	Value  any
}

// DB wraps a *sql.DB connection to a SQLite database.
type DB struct {
	db   *sql.DB
	path string
}

// Column describes a single column from PRAGMA table_info.
type Column struct {
	Name    string
	Type    string
	NotNull bool
	PK      bool
	Default sql.NullString
}

// Index describes a single index from PRAGMA index_list.
type Index struct {
	Name    string
	Unique  bool
	Columns []string
}

// Row is a single row of query results.
type Row []any

// Open opens (or creates) a SQLite database at path and verifies connectivity.
func Open(path string) (*DB, error) {
	d, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenDB, err)
	}

	// Verify the connection is alive.
	if err := d.Ping(); err != nil {
		d.Close()
		return nil, fmt.Errorf("%w: ping failed: %w", ErrOpenDB, err)
	}

	return &DB{db: d, path: path}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if err := db.db.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrCloseDB, err)
	}
	return nil
}

// Tables returns the list of user table names in the database.
func (db *DB) Tables() ([]string, error) {
	rows, err := db.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing tables: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning table name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tables: %w", err)
	}

	return names, nil
}

// Columns returns column metadata for the given table.
func (db *DB) Columns(table string) ([]Column, error) {
	query := fmt.Sprintf(`PRAGMA table_info("%s")`, table)
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("getting columns for %q: %w", table, err)
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var c Column
		var notNull int
		var pk int
		var dflt sql.NullString

		// PRAGMA table_info returns: cid, name, type, notnull, dflt_value, pk
		var cid int
		var name, typ string
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("scanning column for %q: %w", table, err)
		}

		c.Name = name
		c.Type = typ
		c.NotNull = notNull != 0
		c.PK = pk != 0
		c.Default = dflt

		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating columns for %q: %w", table, err)
	}

	return cols, nil
}

// RowCount returns the number of rows in the given table.
func (db *DB) RowCount(table string) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, table)
	var count int64
	if err := db.db.QueryRow(query).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting rows in %q: %w", table, err)
	}
	return count, nil
}

// Indexes returns index metadata for the given table, including column names.
func (db *DB) Indexes(table string) ([]Index, error) {
	// Get the list of indexes.
	listQuery := fmt.Sprintf(`PRAGMA index_list("%s")`, table)
	listRows, err := db.db.Query(listQuery)
	if err != nil {
		return nil, fmt.Errorf("listing indexes for %q: %w", table, err)
	}
	defer listRows.Close()

	var indexes []Index
	for listRows.Next() {
		var idx Index
		var seq int
		var unique int
		var origin string
		var partial int

		if err := listRows.Scan(&seq, &idx.Name, &unique, &origin, &partial); err != nil {
			return nil, fmt.Errorf("scanning index for %q: %w", table, err)
		}
		idx.Unique = unique != 0

		// Get columns for this index.
		infoQuery := fmt.Sprintf(`PRAGMA index_info("%s")`, idx.Name)
		infoRows, err := db.db.Query(infoQuery)
		if err != nil {
			return nil, fmt.Errorf("getting index info for %q on %q: %w", idx.Name, table, err)
		}

		for infoRows.Next() {
			var seqno int
			var cid int
			var colName string
			if err := infoRows.Scan(&seqno, &cid, &colName); err != nil {
				infoRows.Close()
				return nil, fmt.Errorf("scanning index_info for %q on %q: %w", idx.Name, table, err)
			}
			idx.Columns = append(idx.Columns, colName)
		}
		infoRows.Close()
		if err := infoRows.Err(); err != nil {
			return nil, fmt.Errorf("iterating index_info for %q on %q: %w", idx.Name, table, err)
		}

		indexes = append(indexes, idx)
	}
	if err := listRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating index_list for %q: %w", table, err)
	}

	return indexes, nil
}

// Query runs a SELECT * on the given table with optional limit/offset.
// Returns the rows as []Row, the column names as []string.
func (db *DB) Query(table string, limit, offset int) ([]Row, []string, error) {
	return db.QueryFiltered(table, nil, limit, offset)
}

// QueryFiltered queries a table with optional filters and returns rows + column names.
// The filters are applied using parameterized queries (no SQL injection).
// limit and offset control pagination. If limit <= 0 it defaults to 100.
func (db *DB) QueryFiltered(table string, filters []Filter, limit, offset int) ([]Row, []string, error) {
	if limit <= 0 {
		limit = 100
	}

	var clause bytes.Buffer
	clause.WriteString(fmt.Sprintf(`SELECT * FROM "%s"`, table))

	args := make([]any, 0, len(filters)+2)

	if len(filters) > 0 {
		clause.WriteString(" WHERE ")
		for i, f := range filters {
			if i > 0 {
				clause.WriteString(" AND ")
			}
			clause.WriteString(f.Column)
			clause.WriteString(" ")
			clause.WriteString(string(f.Op))
			clause.WriteString(" ?")
			args = append(args, f.Value)
		}
	}

	clause.WriteString(" ORDER BY rowid LIMIT ? OFFSET ?")
	args = append(args, limit, offset)

	query := clause.String()
	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: query %q: %w", ErrQuery, query, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: getting columns: %w", ErrQuery, err)
	}

	var result []Row
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, fmt.Errorf("%w: scanning row: %w", ErrQuery, err)
		}

		row := make(Row, len(columns))
		for i, v := range values {
			// Convert []byte to string for cleaner output.
			if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = v
			}
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("%w: iterating rows: %w", ErrQuery, err)
	}

	return result, columns, nil
}
