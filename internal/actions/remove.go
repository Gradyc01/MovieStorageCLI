package actions

import (
	"fmt"
	"movie-tracker/internal/display"
)

func (store *Store) RemoveMovie(id string) error {
	// Look the movie up first, purely so the messages below can
	// reference its title, not just its ID — "Removed 'Dune'" reads
	// a lot better than "Removed movie with id dune-2021".
	m, err := store.storage.GetByID(id)
	if err != nil {
		display.PrintError(fmt.Sprintf("No movie found with id %q", id))
		return err
	}

	if err := store.storage.Remove(id); err != nil {
		display.PrintError(fmt.Sprintf("Could not remove %q: %v", m.Title, err))
		return err
	}

	display.PrintSuccess(fmt.Sprintf("Removed %q (%s)", m.Title, m.ID))
	return nil
}
