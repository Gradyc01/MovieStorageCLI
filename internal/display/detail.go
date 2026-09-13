// Package display: see table.go for the package-level doc comment.
package display

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"movie-tracker/internal/movie"
)

// detailField describes one label/value row of the detail view — the
// vertical-layout counterpart to table.go's column. get() must return
// PLAIN text (no ANSI codes); that's what the label:value colon
// alignment is computed from. color(), if non-nil, is applied to the
// plain value AFTER alignment has already been decided, same
// "color is a pure post-processing step" rule column.color follows.
type detailField struct {
	label string
	get   func(m *movie.Movie) string
	color func(m *movie.Movie, value string) string
}

func movieDetailFields() []detailField {
	return []detailField{
		{
			label: "ID",
			get:   func(m *movie.Movie) string { return m.ID },
			color: func(m *movie.Movie, value string) string { return colorize(dim, value) },
		},
		{
			label: "FilmType",
			get:   func(m *movie.Movie) string { return m.FilmType },
			color: func(m *movie.Movie, value string) string { return colorize(dim, value) },
		},
		{
			label: "Title",
			get:   func(m *movie.Movie) string { return m.Title },
			color: func(m *movie.Movie, value string) string { return colorize(bold, colorizeTitle(m, value)) },
		},
		{label: "Release Date", get: func(m *movie.Movie) string { return m.ReleaseDate }},
		{
			label: "Rating",
			get:   func(m *movie.Movie) string { return m.RatingOrWatched() },
			color: func(m *movie.Movie, value string) string { return applyRatingColor(m, value) },
		},
		{label: "Genres", get: func(m *movie.Movie) string { return m.ListDisplay(m.Genres) }},
		{
			label: "Tags",
			get:   func(m *movie.Movie) string { return m.ListDisplay(m.Tags) },
			color: func(m *movie.Movie, value string) string { return colorize(darkOlive, value) },
		},
		{
			label: "Directors",
			get:   func(m *movie.Movie) string { return m.ListDisplay(m.Directors) },
			color: func(m *movie.Movie, value string) string { return colorize(cyan, value) },
		},
		{
			label: "Known Actors",
			get:   func(m *movie.Movie) string { return m.ListDisplay(m.KnownActors) },
		},
		{
			label: "IMDB-ID",
			get:   func(m *movie.Movie) string { return m.ImdbID },
			color: func(m *movie.Movie, value string) string { return colorize(white, value) },
		},
		{
			label: "Production Companies",
			get:   func(m *movie.Movie) string { return m.ListDisplay(m.ProductionCompanies) },
			color: func(m *movie.Movie, value string) string { return colorize(dim, value) },
		},
		{
			label: "Production Countries",
			get:   func(m *movie.Movie) string { return m.ListDisplay(m.ProductionCountries) },
			color: func(m *movie.Movie, value string) string { return colorize(dim, value) },
		},
		{
			label: "Spoken Languages",
			get:   func(m *movie.Movie) string { return m.ListDisplay(m.SpokenLanguages) },
			color: func(m *movie.Movie, value string) string { return colorize(dim, value) },
		},
		{label: "Note", get: func(m *movie.Movie) string { return m.Notes }},
	}
}

// PrintMovieDetail renders a single movie as a vertical, labeled
// block — decorative separator lines, a centered title, then one
// "Label:  Value" line per field, with colons aligned. This is
// intentionally a general-purpose renderer: it's used right now for
// the confirmation after `add`, but nothing about it assumes that —
// it would work identically as the output of a future
// `movie-tracker show <id>` command.
func PrintMovieDetail(m *movie.Movie) {
	fields := movieDetailFields()

	// Pre-compute every field's PLAIN value once — same reasoning as
	// PrintMovies: we need it to measure the label column, and get()
	// may not be cheap (ListDisplay joins a slice).
	values := make([]string, len(fields))
	for i, f := range fields {
		values[i] = f.get(m)
	}

	// Label width = widest "Label:" in the field list. Computed from
	// plain label text only, so nothing about coloring the value can
	// ever throw off colon alignment.
	labelWidth := 0
	for _, f := range fields {
		if w := utf8.RuneCountInString(f.label) + 1; w > labelWidth { // +1 for the colon
			labelWidth = w
		}
	}

	printSeparator()
	fmt.Println(colorize(bold+cyan, centerText("MOVIE DETAILS", separatorWidth)))
	printSeparator()

	for i, f := range fields {
		label := padRight(f.label+":", labelWidth)
		value := values[i]
		if f.color != nil {
			// Coloring happens AFTER the label has been padded and
			// the value text finalized, so color never has to be
			// stripped out or re-derived to fix alignment — same
			// rule table.go's per-cell coloring follows.
			value = f.color(m, value)
		}
		fmt.Println(label + strings.Repeat(" ", columnGap) + value)
	}

	printSeparator()
}
