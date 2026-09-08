// Package storage defines how movies get persisted. The key idea here
// is the Storage interface below: your CLI commands (add/remove/list)
// will depend only on this interface, never on a concrete JSON or
// SQLite type. That's the same principle as coding against a Java
// interface (e.g. java.util.List) instead of a concrete class
// (ArrayList) — it lets you swap implementations without touching
// the calling code.
package storage

import (
	"movie-tracker/internal/movie"
)

// Storage is the contract every backend must satisfy. In Go, you do
// NOT write "class JSONStorage implements Storage". Instead, any type
// that happens to have methods matching this exact signature list
// automatically satisfies the interface. This is called "structural
// typing" or "duck typing at compile time" — if it walks like a Storage
// and quacks like a Storage, it IS a Storage, no declaration needed.
//
// Note the return types: Go has no exceptions. Instead, functions that
// can fail return an `error` as their last return value. Callers are
// expected to check `if err != nil` rather than try/catch. It's more
// verbose than Java's exceptions, but it makes every failure point
// visible in the function signature.
type Storage interface {
	// Add saves a new movie. Returns an error if something went wrong
	// (e.g. write failure, duplicate ID) — nil error means success.
	Add(m *movie.Movie) error

	// Remove deletes a movie by ID. Returns an error if not found or
	// if the underlying write fails.
	Remove(id string) error

	// List returns every stored movie. Slices ([]*movie.Movie) are
	// Go's equivalent of a Java List<Movie> — a resizable sequence.
	List() ([]*movie.Movie, error)

	// Search does a simple substring/title match. Real fuzzy search
	// can come later; keep the interface simple for now.
	Search(query string) ([]*movie.Movie, error)

	// GetByID returns the single movie with the given ID, or an error
	// if none exists. This backs both the standalone `get` command and
	// arrow-key selection inside `list`.
	GetByID(id string) (*movie.Movie, error)

	// Update persists changes to an existing movie, matched by its ID.
	// GetByID hands back a pointer into memory that's discarded the
	// moment that call returns — mutating fields on it does nothing
	// durable by itself. Update is what actually writes those changes
	// back to storage.
	Update(m *movie.Movie) error
}
