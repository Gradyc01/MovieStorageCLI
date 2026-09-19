# Movie Storage CLI

Movie Storage CLI is a Go command-line application for maintaining a personal
movie and TV show watch list. It stores records in a local JSON file, enriches
new records with metadata from [TMDB](https://www.themoviedb.org/) using an
IMDb URL or ID, and provides commands for listing, searching, viewing,
updating, and removing titles.

When started without a subcommand, the application opens an interactive prompt
and can synchronize the local movie list with a JSON file in a GitHub
repository. It also includes a natural-language assistant command powered by
Gemini for interacting with the tracker without writing raw CLI flags.

## Requirements

- Go 1.27 or later
- A TMDB API key for the `add` command and TMDB-backed metadata lookups
- A Gemini API key for the `clank` command
- A GitHub personal access token and repository for interactive
  synchronization

## Configuration

The application reads configuration from `environment.properties` in the
project root. If a value is not present there, it falls back to the
`system.properties` file created next to the executable and then the process
environment.

Create or update `environment.properties` with values appropriate for your
setup:

```properties
TMDB_API_KEY=your_tmdb_api_key
GEMINI_API_KEY=your_gemini_api_key
GITHUB_TOKEN=your_github_token
MOVIE_REPO=your-github-user/your-movie-repository
MOVIE_FILE_PATH=your movie file path (default: movies.json)
MOVIE_TRACKER_PAGE_SIZE=tracker page size (default: 12)
```

`TMDB_API_KEY` is required when adding a movie or TV show. The API key can be
obtained from the TMDB account API settings.

`GEMINI_API_KEY` is required for the `clank` command, which lets the CLI
interpret natural-language requests such as "find my highest-rated sci-fi
movies" or "add The Matrix with a score of 9".

`GITHUB_TOKEN` and `MOVIE_REPO` are required only for interactive mode. The
repository reference can be either `owner/repository` or a GitHub URL.
`MOVIE_FILE_PATH` is the path to the JSON file inside that repository and
defaults to `movies.json` when it is not set. The local cache is also
`movies.json` unless `MOVIE_LOCAL_PATH` is set as a process environment
variable.

Do not commit real API keys or access tokens to source control.

## Running the project

From the repository root, download dependencies and run the CLI with:

```bash
go mod download
go run . --help
```

To build a standalone executable:

```bash
go build -o movie-tracker .
```

Then use `movie-tracker` in place of `go run .` in the examples below. On
Windows, the generated executable is typically `movie-tracker.exe`.

The local data file is created automatically as commands write data:

```text
movies.json
```

## Commands

### Start the interactive prompt

Running the application without a subcommand opens the REPL.

```bash
movie-tracker
```

In interactive mode you can type commands directly in the terminal, and the
application will pull and save the tracked list from GitHub when the required
configuration is present.

### Add a movie or TV season

`add` accepts an IMDb URL or IMDb ID. The command queries TMDB and stores the
resulting metadata.

```bash
add "https://www.imdb.com/title/tt1160419/"
add tt1160419 --rating 8.5
add tt0903747 --season 1
add tt0903747 --season 1 --rating 9
```

Options:

- `--rating`: Set a watched rating from `0.0` to `10.0`. The default `-1`
  leaves the title unwatched.
- `--season`: Treat the IMDb reference as a TV show and add the specified
  season.

### List tracked titles

```bash
list
```

The list view supports arrow-key navigation. Press Enter to view a selected
title, `q` to return from the detail view to the list, and `q` from the list
to exit. The page size defaults to 15 and can be changed with
`MOVIE_TRACKER_PAGE_SIZE`.

### View a title

Use the ID shown by `list`:

```bash
get the-matrix-1999
```

### Search

Search expressions use the form `FIELD OPERATOR VALUE`. Multiple expressions
can be combined with a comma followed by a space (`, `), and all expressions
must match.

Supported fields and operators:

| Field | Operators | Value |
| --- | --- | --- |
| `TITLE` | `contains`, `equals` | Text |
| `DIRECTOR` | `contains`, `equals` | Text |
| `TAGS` | `contains`, `equals` | Text |
| `YEAR` | `>`, `<`, `==` | Number |
| `RATING` | `>`, `<`, `==` | Number |
| `RELEASE` | `>`, `<`, `==` | `YYYY-MM-DD` |
| `WATCHED` | `==` | `true` or `false` |

Examples:

```bash
search "TITLE contains matrix"
search "DIRECTOR equals Christopher Nolan"
search "YEAR > 2000, RATING > 8"
search "WATCHED == false"
```

For numeric and date fields, the supported comparison operators are `>`, `<`,
and `==`. Text fields support `contains` and `equals`.

### Ask the assistant

`clank` lets you issue natural-language requests that the CLI translates into
tracker actions.

```bash
clank "show me my ten highest-rated movies"
clank "add the movie The Matrix with a rating of 9"
clank "find the latest Christopher Nolan film I have not watched"
```

This command requires `GEMINI_API_KEY` to be configured.

### Update a title

Updates are applied to the title ID supplied as the positional argument. At
least one update option must be provided.

```bash
update the-matrix-1999 --updateTitle "The Matrix"
update the-matrix-1999 --updateDirectors "Lana Wachowski, Lilly Wachowski"
update the-matrix-1999 --updateTags "science fiction, favorites"
update the-matrix-1999 --updateRating 9
```

Options:

- `--updateTitle`: Replace the title.
- `--updateDirectors`: Replace the director list. Separate multiple directors
  with commas.
- `--updateTags`: Replace the tag list. Separate multiple tags with commas.
- `--updateRating`: Use `0.0` through `10.0` for a watched rating, `-1` to
  mark the title unwatched, `-2` to mark it watched but unrated. `-3` to mark it unwatched but shortlisted.
- `--updateActors`: Replace the actor list. Separate multiple actors with
  commas.
- `--updateNote`: Replace the note text associated with the title.

### Remove a title

```bash
remove the-matrix-1999
```

### Get help

```bash
help
add --help
update --help
clank --help
```

## Dependencies

The CLI uses [Cobra](https://github.com/spf13/cobra) for command parsing,
[Google GenAI](https://github.com/googleapis/go-genai) for the natural-language
assistant, `go-github` for optional GitHub synchronization, and TMDB's API for
movie and TV metadata.
