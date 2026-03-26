package parser

import (
	"os"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "envcheck-*.env")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestParseFile_BasicKeys(t *testing.T) {
	path := writeTempFile(t, `
DB_HOST=localhost
DB_PORT=5432
APP_NAME=myapp
`)
	defer os.Remove(path)

	result, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if result["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %s", result["DB_HOST"])
	}
	if result["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432, got %s", result["DB_PORT"])
	}
}

func TestParseFile_SkipsCommentsAndBlankLines(t *testing.T) {
	path := writeTempFile(t, `
# This is a comment
DB_HOST=localhost

# Another comment
APP_NAME=myapp
`)
	defer os.Remove(path)

	result, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
}

func TestParseFile_QuotedValues(t *testing.T) {
	path := writeTempFile(t, `
SECRET="hello world"
TOKEN='my-token-value'
`)
	defer os.Remove(path)

	result, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if result["SECRET"] != "hello world" {
		t.Errorf("expected 'hello world', got %s", result["SECRET"])
	}
	if result["TOKEN"] != "my-token-value" {
		t.Errorf("expected 'my-token-value', got %s", result["TOKEN"])
	}
}

func TestParseFile_ValueWithEquals(t *testing.T) {
	path := writeTempFile(t, `
DATABASE_URL=postgres://user:pass@localhost/db?sslmode=disable
`)
	defer os.Remove(path)

	result, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := "postgres://user:pass@localhost/db?sslmode=disable"
	if result["DATABASE_URL"] != expected {
		t.Errorf("expected %s, got %s", expected, result["DATABASE_URL"])
	}
}

func TestDiff_FindsMissingKeys(t *testing.T) {
	actual := map[string]string{
		"DB_HOST": "localhost",
	}
	example := map[string]string{
		"DB_HOST": "",
		"DB_PORT": "",
	}

	missing, undocumented := Diff(actual, example)

	if len(missing) != 1 || missing[0] != "DB_PORT" {
		t.Errorf("expected DB_PORT to be missing, got %v", missing)
	}
	if len(undocumented) != 0 {
		t.Errorf("expected no undocumented keys, got %v", undocumented)
	}
}

func TestDiff_FindsUndocumentedKeys(t *testing.T) {
	actual := map[string]string{
		"DB_HOST":    "localhost",
		"SECRET_KEY": "abc123",
	}
	example := map[string]string{
		"DB_HOST": "",
	}

	missing, undocumented := Diff(actual, example)

	if len(missing) != 0 {
		t.Errorf("expected no missing keys, got %v", missing)
	}
	if len(undocumented) != 1 || undocumented[0] != "SECRET_KEY" {
		t.Errorf("expected SECRET_KEY to be undocumented, got %v", undocumented)
	}
}
