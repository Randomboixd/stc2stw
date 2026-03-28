package lorebook

import (
	"strings"
	"testing"

	"github.com/deck/stc2stw/internal/card"
)

func TestBuildIncludesOnlyNonEmptySections(t *testing.T) {
	t.Parallel()

	book := Build(card.Card{
		Name:            "Alice",
		Description:     "Desc",
		FirstMessage:    "Hello",
		MessageExamples: "Ally: hi\n{{user}}: hello",
	})

	entry := book.Entries["0"]
	if strings.Contains(entry.Content, "## Personality") {
		t.Fatal("expected empty sections to be omitted")
	}
	if !strings.Contains(entry.Content, "## Description\nDesc") {
		t.Fatal("expected description section to be rendered")
	}
	if len(entry.Key) != 2 {
		t.Fatalf("expected 2 keys (name + alias), got %d", len(entry.Key))
	}
	if entry.Key[1] != "Ally" {
		t.Fatalf("expected alias Ally, got %q", entry.Key[1])
	}
}

func TestMarshalIsPrettyPrinted(t *testing.T) {
	t.Parallel()

	data, err := Marshal(Build(card.Card{Name: "Alice"}))
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	text := string(data)
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("expected trailing newline")
	}
	if !strings.Contains(text, "\n  \"entries\":") {
		t.Fatal("expected indented json output")
	}
}

func TestBuildManyAssignsSequentialEntries(t *testing.T) {
	t.Parallel()

	book := BuildMany([]card.Card{
		{Name: "Alice"},
		{Name: "Bob"},
	})

	if len(book.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(book.Entries))
	}
	if book.Entries["0"].UID != 0 || book.Entries["0"].Comment != "Alice" {
		t.Fatalf("unexpected first entry: %+v", book.Entries["0"])
	}
	if book.Entries["1"].UID != 1 || book.Entries["1"].Comment != "Bob" {
		t.Fatalf("unexpected second entry: %+v", book.Entries["1"])
	}
}
