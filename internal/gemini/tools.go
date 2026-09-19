package gemini

import "google.golang.org/genai"

// These declarations mirror the Movie Storage CLI's Cobra commands.
// Descriptions double as the "documentation" Gemini uses to decide
// which function to call and how to fill in arguments, so keep them
// in sync with README.md when the CLI's syntax changes.

var searchTMDBFn = &genai.FunctionDeclaration{
	Name: "search_tmdb",
	Description: "Search TMDB by title to find candidate movies/shows and their IMDb IDs. " +
		"Call this FIRST whenever the user names a title instead of giving an IMDb ID or URL directly.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"query": {
				Type:        genai.TypeString,
				Description: "The movie or show title to search for, e.g. '12 Years a Slave'.",
			},
			"year": {
				Type:        genai.TypeInteger,
				Description: "Optional release year to narrow the search, if the user mentioned one.",
			},
			"film_type": {
				Type: genai.TypeString,
				Description: "Determines whether to search in the movie database or the show database. " +
					"If unclear which one it is prompt the user to ask. String must be either 'tv' or 'movie'",
			},
		},
		Required: []string{"query"},
	},
}

var addTitleFn = &genai.FunctionDeclaration{
	Name: "add_title",
	Description: "Add a movie or TV season to the watch list. Requires an IMDb ID or URL - " +
		"if you only have a title, call search_tmdb first to resolve it.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"imdb_ref": {
				Type:        genai.TypeString,
				Description: "IMDb URL (https://www.imdb.com/title/ttXXXXXXX/) or bare IMDb ID (ttXXXXXXX).",
			},
			"rating": {
				Type:        genai.TypeNumber,
				Description: "Watched rating from 0.0 to 10.0. Omit entirely if the user hasn't watched it yet.",
			},
			"season": {
				Type:        genai.TypeInteger,
				Description: "Season number. Only set this if the title is a TV show and a season was mentioned.",
			},
		},
		Required: []string{"imdb_ref"},
	},
}

var listTitlesFn = &genai.FunctionDeclaration{
	Name:        "list_titles",
	Description: "List all tracked titles. Use when the user wants to browse or see everything in their list.",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var getTitleFn = &genai.FunctionDeclaration{
	Name:        "get_title",
	Description: "View full details of one tracked title by its ID (the slug shown in list output, e.g. 'the-matrix-1999').",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"id": {Type: genai.TypeString, Description: "The title's ID, e.g. 'the-matrix-1999'."},
		},
		Required: []string{"id"},
	},
}

var searchTitlesFn = &genai.FunctionDeclaration{
	Name: "search_titles",
	Description: "Search the local watch list using field/operator/value expressions. " +
		"Fields: TITLE, DIRECTOR, TAGS, GENRE, FILM_TYPE, PROD_COMPANY, PROD_COUNTRY, LANGUAGE, STATUS, ACTOR (operators: contains, equals); " +
		"YEAR, RATING, RELEASE (operators: >, <, ==; RELEASE values are YYYY-MM-DD); " +
		"WATCHED (operator: ==, value true/false). " +
		"Combine multiple expressions with ', ' (comma-space) - all must match. " +
		"Example: 'YEAR > 2000, RATING > 8'.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"expression": {
				Type:        genai.TypeString,
				Description: "One or more comma-space-separated FIELD OPERATOR VALUE expressions.",
			},
			"displaySearchResults": {
				Type: genai.TypeBoolean,
				Description: "If true, displays the movie table for the user containing all of the movies from the search." +
					"This movie table is interactive therefore this should only ever be used as the very last function. " +
					"(For example: displaySearchResults should be false when a search is required to find a movie to delete, " +
					"but should be true when a user wants to find all movies with a director called Tim)",
			},
		},
		Required: []string{"expression"},
	},
}

var updateTitleFn = &genai.FunctionDeclaration{
	Name:        "update_title",
	Description: "Update one or more fields on an existing title, identified by its ID. At least one field besides id must be set.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"id":        {Type: genai.TypeString, Description: "The title's existing ID."},
			"title":     {Type: genai.TypeString, Description: "New title text, if renaming."},
			"directors": {Type: genai.TypeString, Description: "Comma-separated list of directors, replaces the existing list. "},
			"tags":      {Type: genai.TypeString, Description: "Comma-separated list of tags, replaces the existing list."},
			"actors":    {Type: genai.TypeString, Description: "Comma-separated list of actors, replaces the existing list."},
			"note":      {Type: genai.TypeString, Description: "new note text, replaces the existing note."},
			"rating": {
				Type: genai.TypeNumber,
				Description: "0.0-10.0 for a watched rating; -1 = unwatched; " +
					"-2 = watched but unrated; -3 = unwatched but shortlisted.",
			},
		},
		Required: []string{"id"},
	},
}

var removeTitleFn = &genai.FunctionDeclaration{
	Name:        "remove_title",
	Description: "Permanently remove a title from the watch list by its ID. Destructive - only call after the user has confirmed.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"id": {Type: genai.TypeString},
		},
		Required: []string{"id"},
	},
}

// AskClarification is not a real CLI command. It's a signal Gemini uses to
// pause and get more information from the user before acting, instead of
// guessing at ambiguous input.
var askClarificationFn = &genai.FunctionDeclaration{
	Name: "ask_clarification",
	Description: "Ask the user a multiple-choice question when their request is ambiguous, " +
		"underspecified, or a destructive action (like remove_title) needs confirmation. " +
		"Use this instead of guessing at missing information.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"question": {Type: genai.TypeString, Description: "The question to show the user."},
			"options": {
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "2-6 short answer choices for the user to pick from.",
			},
		},
		Required: []string{"question", "options"},
	},
}

// Tools bundles every declaration into the single genai.Tool the chat loop sends.
func Tools() []*genai.Tool {
	return []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				searchTMDBFn,
				addTitleFn,
				listTitlesFn,
				getTitleFn,
				searchTitlesFn,
				updateTitleFn,
				removeTitleFn,
				askClarificationFn,
			},
		},
	}
}

// SystemPrompt grounds Gemini in the CLI's exact conventions. Keep this in
// sync with README.md - it is the main lever for avoiding malformed commands.
const SystemPrompt = `You are a natural-language interface to a personal movie/TV watch-list CLI.

Translate the user's request into exactly one function call per turn from the
available tools. Never invent an IMDb ID - if the user gives a title instead
of an IMDb ID/URL, call search_tmdb first and wait for the result.

If a request is ambiguous (multiple plausible matches, a missing required
detail, or a destructive action like remove_title), call ask_clarification
with a short question and 2-6 concrete options rather than guessing.

Once a function's result comes back, summarize the outcome for the user in
one or two plain sentences - don't dump raw JSON at them.`
