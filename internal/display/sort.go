package display

import (
	"cmp"
	"slices"

	"movie-tracker/internal/movie"
)

// sortState.col is an index into movieColumns(); -1 means unsorted.
type sortState struct {
	col  int
	desc bool
}

func noSort() sortState { return sortState{col: -1} }

// cycle advances a column through three states:
//
//	unsorted -> ascending -> descending -> unsorted
//
// If a different column is currently sorted, that column counts as
// "unsorted" here, so cycling it starts at ascending and replaces the
// previous sort.
func (s sortState) cycle(col int) sortState {
	switch {
	case s.col != col:
		return sortState{col: col} // ascending
	case !s.desc:
		return sortState{col: col, desc: true}
	default:
		return noSort()
	}
}

// statusRank orders statuses by "actionability" rather than alphabetically.
// Reorder to taste.
var statusRank = map[string]int{
	movie.WATCHING:  0,
	movie.SHORTLIST: 1,
	movie.UNWATCHED: 2,
	movie.UNRATED:   3,
	movie.WATCHED:   4,
}

// sortableIndexes returns the indexes of columns that define compare().
func sortableIndexes(cols []column) []int {
	var out []int
	for i, c := range cols {
		if c.compare != nil {
			out = append(out, i)
		}
	}
	return out
}

// sortMovies returns a sorted COPY; the original slice is never mutated,
// so clearing the sort restores the original order.
func sortMovies(movies []*movie.Movie, cols []column, st sortState) []*movie.Movie {
	out := slices.Clone(movies)
	if st.col < 0 {
		return out
	}
	col := cols[st.col]
	isMissing := func(m *movie.Movie) bool { return col.missing != nil && col.missing(m) }

	slices.SortStableFunc(out, func(a, b *movie.Movie) int {
		am, bm := isMissing(a), isMissing(b)
		switch {
		case am && bm:
			return 0
		case am:
			return 1 // missing values stay last in both directions
		case bm:
			return -1
		}
		c := col.compare(a, b)
		if st.desc {
			c = -c
		}
		return c
	})
	return out
}

// compareStatusThenRating is used by the RATING column.
func compareStatusThenRating(a, b *movie.Movie) int {
	if c := cmp.Compare(statusRank[a.Status], statusRank[b.Status]); c != 0 {
		return c
	}
	return cmp.Compare(a.Rating, b.Rating)
}
