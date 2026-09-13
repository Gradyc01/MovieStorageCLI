package environment

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	PROPS_FILE        = "environment.properties"
	SYSTEM_PROPS_FILE = "system.properties"
)

// Set at build time via -ldflags "-X environment.tmdbAPIKey=..."
var tmdbAPIKey string

// buildTimeVars maps variable names to their compile-time injected values.
// Add more entries here as more variables get baked in this way.
var buildTimeVars = map[string]string{
	"TMDB_API_KEY": tmdbAPIKey,
}

func GetVariable(variable string) (string, error) {
	if props, err := loadProperties(PROPS_FILE); err == nil {
		if key, ok := props[variable]; ok && key != "" {
			return key, nil
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("reading %s: %w", PROPS_FILE, err)
	}

	if key := os.Getenv(variable); key != "" {
		return key, nil
	}

	if key, ok := buildTimeVars[variable]; ok && key != "" {
		return key, nil
	}

	return "", fmt.Errorf(
		"%s not found — add it to %s (%s=your_variable_here) to set it as an environment variable",
		variable,
		PROPS_FILE,
		variable,
	)
}

// Every variable GetVariableFromOutside might be asked for that's allowed
// to live in system.properties. Add to this list as you add new callers.
var systemPropsKeys = []string{
	"GITHUB_TOKEN",
	"MOVIE_TRACKER_PAGE_SIZE",
	"MOVIE_FILE_PATH",
	"MOVIE_REPO",
}

func GetVariableFromOutside(variable string) (string, error) {
	val, err := getVariableOutside(variable)
	if err == nil {
		return val, err
	}
	val2, err2 := GetVariable(variable)
	if err2 == nil {
		return val2, err2
	}
	return "", fmt.Errorf("failed to get variable %s from outside of %s and internal environment variable", variable, SYSTEM_PROPS_FILE)
}

func getVariableOutside(variable string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating executable path: %w", err)
	}
	systemPropsPath := filepath.Join(filepath.Dir(exePath), SYSTEM_PROPS_FILE)

	props, err := loadProperties(systemPropsPath)
	if err != nil {
		if os.IsNotExist(err) {
			if writeErr := createTemplatePropertiesFile(systemPropsPath); writeErr != nil {
				return "", fmt.Errorf(
					"%s not found, and creating a template failed: %w",
					SYSTEM_PROPS_FILE, writeErr,
				)
			}
			return "", fmt.Errorf(
				"%s was missing, so a template was created at %s — fill in %s and restart",
				SYSTEM_PROPS_FILE, systemPropsPath, variable,
			)
		}
		return "", fmt.Errorf("reading %s: %w", systemPropsPath, err)
	}

	if key, ok := props[variable]; ok && key != "" {
		return key, nil
	}

	return "", fmt.Errorf(
		"%s not found — add it to %s or %s (%s=your_variable_here)",
		variable, PROPS_FILE, systemPropsPath, variable,
	)
}

// createTemplatePropertiesFile writes an empty-valued properties file
// listing every known key, so the user has a starting point to fill in.
func createTemplatePropertiesFile(path string) error {
	keys := make([]string, len(systemPropsKeys))
	copy(keys, systemPropsKeys)
	sort.Strings(keys)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, k := range keys {
		if _, err := fmt.Fprintf(f, "%s=\n", k); err != nil {
			return err
		}
	}
	return nil
}

// loadProperties reads a simple "key=value" properties file (e.g.
// environment.properties in the project root) and returns it as a map.
// Blank lines and lines starting with "#" or "!" are treated as comments.
// Values may optionally be wrapped in double or single quotes.
func loadProperties(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	props := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		var key, value string
		if idx := strings.IndexAny(line, "="); idx != -1 {
			key = strings.TrimSpace(line[:idx])
			value = strings.TrimSpace(line[idx+1:])
		} else {
			continue // skip malformed lines
		}

		// Strip surrounding quotes, if present.
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		props[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return props, nil
}
