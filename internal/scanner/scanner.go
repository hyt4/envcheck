package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// patterns maps file extensions to their env access patterns
var patterns = map[string]*regexp.Regexp{
	".js": regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`),
	".ts": regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`),
	".py": regexp.MustCompile(`os\.environ\.get\(['"']([A-Z_][A-Z0-9_]*)['"']\)|os\.environ\[['"']([A-Z_][A-Z0-9_]*)['"']\]|os\.getenv\(['"']([A-Z_][A-Z0-9_]*)['"']\)`),
	".go": regexp.MustCompile(`os\.Getenv\(['"']([A-Z_][A-Z0-9_]*)['"']\)`),
}

// skipDirs are directories we never want to scan
var skipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
}

// FindUsedKeys walks the project directory and returns all env keys
// that are actually referenced in source files
func FindUsedKeys(root string) (map[string]bool, error) {
	used := make(map[string]bool)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories we don't want to scan
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this file extension has a pattern
		ext := strings.ToLower(filepath.Ext(path))
		pattern, ok := patterns[ext]
		if !ok {
			return nil
		}

		// Scan the file for env key references
		keys, err := scanFile(path, pattern)
		if err != nil {
			return err
		}

		for _, key := range keys {
			used[key] = true
		}

		return nil
	})

	return used, err
}

// scanFile reads a single file and returns all env keys referenced in it
func scanFile(path string, pattern *regexp.Regexp) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keys []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		matches := pattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			// Some patterns have multiple capture groups — find the non-empty one
			for _, group := range match[1:] {
				if group != "" {
					keys = append(keys, group)
				}
			}
		}
	}

	return keys, scanner.Err()
}

// FindUnused takes the keys defined in .env and the keys used in source
// and returns any that are defined but never referenced
func FindUnused(defined map[string]string, used map[string]bool) []string {
	var unused []string
	for key := range defined {
		if !used[key] {
			unused = append(unused, key)
		}
	}
	return unused
}
