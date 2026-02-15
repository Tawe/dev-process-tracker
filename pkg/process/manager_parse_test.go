package process

import (
	"reflect"
	"testing"
)

func TestParseCommandArgs(t *testing.T) {
	t.Parallel()

	got, err := parseCommandArgs(`python3 -m uvicorn "app.main:app" --reload`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	want := []string{"python3", "-m", "uvicorn", "app.main:app", "--reload"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected argv: got %#v want %#v", got, want)
	}
}

func TestParseCommandArgs_UnterminatedQuote(t *testing.T) {
	t.Parallel()

	if _, err := parseCommandArgs(`npm run "dev`); err == nil {
		t.Fatal("expected unterminated quote error")
	}
}
