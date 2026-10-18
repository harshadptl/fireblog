// Package store is the data-access layer. Handlers talk to the database only
// through these methods, which keeps SQL in one place.
package store

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/uptrace/bun"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// Store wraps the Bun database handle.
type Store struct {
	DB *bun.DB
}

// New returns a Store.
func New(db *bun.DB) *Store { return &Store{DB: db} }

// Page describes a slice of a result set.
type Page struct {
	Number  int // 1-based
	PerPage int
	Total   int
}

// Limit returns the SQL limit for the page.
func (p Page) Limit() int64 { return int64(p.PerPage) }

// Offset returns the SQL offset for the page.
func (p Page) Offset() int64 { return int64((p.Number - 1) * p.PerPage) }

// Pages returns the total number of pages (at least 1).
func (p Page) Pages() int {
	if p.PerPage <= 0 || p.Total == 0 {
		return 1
	}
	return (p.Total + p.PerPage - 1) / p.PerPage
}

// HasPrev reports whether there is a previous page.
func (p Page) HasPrev() bool { return p.Number > 1 }

// HasNext reports whether there is a next page.
func (p Page) HasNext() bool { return p.Number < p.Pages() }

// NewPage normalises user-supplied paging input.
func NewPage(number, perPage int) Page {
	if number < 1 {
		number = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	return Page{Number: number, PerPage: perPage}
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// likePattern escapes LIKE wildcards in user input and wraps it in %...%.
// Use with: LIKE ? ESCAPE '\'
func likePattern(q string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(strings.ToLower(q)) + "%"
}
