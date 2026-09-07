package environment

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const PROPS_FILE = "environment.properties"

func GetVariable(variable string) (string, error) {
	if props, err := loadProperties(PROPS_FILE); err == nil {
		if key, ok := props[variable]; ok && key != "" {
			return key, nil
		}
	} else if !os.IsNotExist(err) {
		// File exists but couldn't be read for some other reason.
		return "", fmt.Errorf("reading %s: %w", PROPS_FILE, err)
	}

	// Fall back to a real environment variable, e.g. for CI/CD.
	if key := os.Getenv(variable); key != "" {
		return key, nil
	}

	return "", fmt.Errorf(
		"%s not found — add it to %s (%s=your_variable_here) to set it as an environment variable",
		variable,
		PROPS_FILE,
		variable,
	)
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
