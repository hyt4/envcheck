package parser

import (
	"bufio"
	"os"
	"strings"
)

func ParseFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = stripQuotes(value)

		result[key] = value
	}

	return result, scanner.Err()
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func Diff(actual, example map[string]string) (missing, undocumented []string) {
	for key := range example {
		if _, exists := actual[key]; !exists {
			missing = append(missing, key)
		}
	}

	for key := range actual {
		if _, exists := example[key]; !exists {
			undocumented = append(undocumented, key)
		}
	}

	return missing, undocumented
}
