package gemini

import (
	"fmt"
	"movie-tracker/internal/actions"
	"movie-tracker/internal/display"
)

// Executor wires Gemini's function calls to the actual movie-store logic.
// Swap the yourmodule/internal/movies import for wherever your Cobra
// commands' underlying functions live - the point is to call the same
// functions Cobra calls, not to shell out to your own binary.
type Executor struct {
	store *actions.Store
}

func NewExecutor(store *actions.Store) *Executor {
	return &Executor{store: store}
}

// Run dispatches a single named function call with its arguments (as decoded
// from Gemini's response) and returns a result to feed back to the model.
func (e *Executor) Run(name string, args map[string]any) (any, error) {
	switch name {
	case "search_tmdb":
		query, _ := args["query"].(string)
		year, _ := args["year"].(string)
		filmType, _ := args["film_type"].(string)
		return e.store.FindMovie(query, year, filmType)

	case "add_title":
		imdbRef, _ := args["imdb_ref"].(string)
		rating := -1.0
		if r, ok := args["rating"].(float64); ok {
			rating = r
		}
		season := 0
		if s, ok := args["season"].(float64); ok {
			season = int(s)
		}
		return e.store.AddMovie(imdbRef, rating, season)

	case "list_titles":
		return e.store.ListMovie()

	case "get_title":
		id, _ := args["id"].(string)
		return e.store.GetMovie(id)

	case "search_titles":
		expr, _ := args["expression"].(string)
		movies, err := e.store.SearchMovie(expr)

		shouldDisplay, _ := args["displaySearchResults"].(bool)
		if shouldDisplay {
			if display.PrintMovieTable(movies) != nil {
				return nil, fmt.Errorf("error displaying movies")
			}
		}
		return movies, err

	case "update_title":
		id, _ := args["id"].(string)
		fields := actions.UpdateFields{}

		if v, ok := args["title"].(string); ok {
			fields.Title = &v
		}
		if v, ok := args["directors"].(string); ok {
			d := actions.SplitAndTrim(v, ",")
			fields.Directors = &d
		}
		if v, ok := args["tags"].(string); ok {
			t := actions.SplitAndTrim(v, ",")
			fields.Tags = &t
		}
		if v, ok := args["actors"].(string); ok {
			a := actions.SplitAndTrim(v, ",")
			fields.Actors = &a
		}
		if v, ok := args["note"].(string); ok {
			fields.Note = &v
		}
		if v, ok := args["rating"].(float64); ok {
			fields.Rating = &v
		}
		m, err := e.store.UpdateMovie(id, fields)
		if err == nil {
			display.PrintMovieDetail(m)
		}
		return m, err
	case "remove_title":
		id, _ := args["id"].(string)
		return nil, e.store.RemoveMovie(id)

	default:
		return nil, fmt.Errorf("unknown function call from model: %s", name)
	}
}
