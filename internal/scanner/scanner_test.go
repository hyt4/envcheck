package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindUsedKeys_Python(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "app.py", `
import os
db = os.getenv("DB_HOST")
secret = os.environ["SECRET_KEY"]
port = os.environ.get("DB_PORT")
`)

	used, err := FindUsedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"DB_HOST", "SECRET_KEY", "DB_PORT"} {
		if !used[key] {
			t.Errorf("expected %s to be found in python file", key)
		}
	}
}

func TestFindUsedKeys_JavaScript(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "app.js", `
const host = process.env.DB_HOST
const port = process.env.DB_PORT
`)

	used, err := FindUsedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"DB_HOST", "DB_PORT"} {
		if !used[key] {
			t.Errorf("expected %s to be found in js file", key)
		}
	}
}

func TestFindUsedKeys_Go(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "main.go", `
package main
import "os"
func main() {
	host := os.Getenv("DB_HOST")
	_ = host
}
`)

	used, err := FindUsedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !used["DB_HOST"] {
		t.Error("expected DB_HOST to be found in go file")
	}
}

func TestFindUsedKeys_SkipsNodeModules(t *testing.T) {
	dir := t.TempDir()

	// This file should be skipped
	nmDir := filepath.Join(dir, "node_modules")
	os.Mkdir(nmDir, 0755)
	writeTempFile(t, nmDir, "app.js", `
const secret = process.env.SECRET_KEY
`)

	// This file should be scanned
	writeTempFile(t, dir, "app.js", `
const host = process.env.DB_HOST
`)

	used, err := FindUsedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !used["DB_HOST"] {
		t.Error("expected DB_HOST to be found")
	}
	if used["SECRET_KEY"] {
		t.Error("expected SECRET_KEY to be skipped — it's in node_modules")
	}
}

func TestFindUnused(t *testing.T) {
	defined := map[string]string{
		"DB_HOST":    "localhost",
		"SECRET_KEY": "abc123",
		"DB_PORT":    "5432",
	}
	used := map[string]bool{
		"DB_HOST": true,
	}

	unused := FindUnused(defined, used)

	if len(unused) != 2 {
		t.Errorf("expected 2 unused keys, got %d: %v", len(unused), unused)
	}
}
