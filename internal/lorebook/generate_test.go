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
		CreatorNotes:    "Hidden",
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
	if strings.Contains(entry.Content, "## Creator Notes") {
		t.Fatal("expected creator notes to be omitted by default")
	}
	if len(entry.Key) != 2 {
		t.Fatalf("expected 2 keys (name + alias), got %d", len(entry.Key))
	}
	if entry.Key[1] != "Ally" {
		t.Fatalf("expected alias Ally, got %q", entry.Key[1])
	}
}

func TestBuildManyCanIncludeCreatorNotes(t *testing.T) {
	t.Parallel()

	book := BuildManyWithOptions([]card.Card{{
		Name:         "Alice",
		CreatorNotes: "Private note",
	}}, DefaultPreset(), BuildOptions{
		Compact:             true,
		IncludeCreatorNotes: true,
	})

	if !strings.Contains(book.Entries["0"].Content, "## Creator Notes\nPrivate note") {
		t.Fatalf("expected creator notes in output, got %q", book.Entries["0"].Content)
	}
}

func TestBuildSeparatesAlternateGreetings(t *testing.T) {
	t.Parallel()

	book := Build(card.Card{
		Name:               "Alice",
		AlternateGreetings: []string{"Hello there", "Good evening"},
	})

	content := book.Entries["0"].Content
	if !strings.Contains(content, "## Alternate Greetings\n### Greeting 1\nHello there\n\n### Greeting 2\nGood evening") {
		t.Fatalf("expected numbered alternate greetings, got %q", content)
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
		{
			Name: "Alice",
			EmbeddedLorebookEntries: []card.EmbeddedLorebookEntry{
				{Key: []string{"Town"}, Comment: "Town", Content: "Lore"},
			},
		},
		{Name: "Bob"},
	}, DefaultPreset())

	if len(book.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(book.Entries))
	}
	if book.Entries["0"].UID != 0 || book.Entries["0"].Comment != "Alice" {
		t.Fatalf("unexpected first entry: %+v", book.Entries["0"])
	}
	if book.Entries["1"].UID != 1 || book.Entries["1"].Comment != "(src: Alice) -> Town" {
		t.Fatalf("unexpected second entry: %+v", book.Entries["1"])
	}
	if book.Entries["2"].UID != 2 || book.Entries["2"].Comment != "Bob" {
		t.Fatalf("unexpected third entry: %+v", book.Entries["2"])
	}
}

func TestDefaultPresetIsAtDepthUser(t *testing.T) {
	t.Parallel()

	book := Build(card.Card{Name: "Alice"})
	entry := book.Entries["0"]
	if entry.Position != positionAtDepth {
		t.Fatalf("expected at-depth position, got %d", entry.Position)
	}
	if entry.Role != roleUser {
		t.Fatalf("expected user role, got %d", entry.Role)
	}
}

func TestResolvePositionPresetOutlet(t *testing.T) {
	t.Parallel()

	preset, err := ResolvePositionPreset("outlet")
	if err != nil {
		t.Fatalf("ResolvePositionPreset returned error: %v", err)
	}
	if preset.Position != positionOutlet {
		t.Fatalf("expected outlet position, got %d", preset.Position)
	}
	if preset.OutletName != defaultOutletName {
		t.Fatalf("expected outlet name %q, got %q", defaultOutletName, preset.OutletName)
	}
}

func TestBuildManyCompactsEmbeddedEntries(t *testing.T) {
	t.Parallel()

	book := BuildMany([]card.Card{{
		Name: "Alice",
		EmbeddedLorebookEntries: []card.EmbeddedLorebookEntry{{
			Key:            []string{"Town"},
			KeySecondary:   []string{"Urban"},
			Comment:        "Capital",
			Content:        "Big city",
			Selective:      true,
			SelectiveLogic: 3,
			Position:       positionBefore,
		}},
	}}, DefaultPreset())

	embedded := book.Entries["1"]
	if embedded.Comment != "(src: Alice) -> Capital" {
		t.Fatalf("unexpected compacted comment: %q", embedded.Comment)
	}
	if len(embedded.KeySecondary) != 2 || embedded.KeySecondary[1] != "Alice" {
		t.Fatalf("expected source gate appended to secondary keys, got %+v", embedded.KeySecondary)
	}
	if !embedded.Selective || embedded.SelectiveLogic != 3 {
		t.Fatalf("expected selective settings preserved, got %+v", embedded)
	}
	if embedded.Position != positionBefore {
		t.Fatalf("expected embedded position preserved, got %d", embedded.Position)
	}
}

func TestBuildManyCanDisableCompacting(t *testing.T) {
	t.Parallel()

	book := BuildManyWithOptions([]card.Card{{
		Name: "Alice",
		EmbeddedLorebookEntries: []card.EmbeddedLorebookEntry{{
			Key:     []string{"Town"},
			Comment: "Capital",
			Content: "Big city",
		}},
	}}, DefaultPreset(), BuildOptions{
		Compact: false,
	})

	if len(book.Entries) != 1 {
		t.Fatalf("expected only primary entry when compacting is off, got %d", len(book.Entries))
	}
}

func TestBuildManyCompactingEnablesSelectiveForNewSourceGate(t *testing.T) {
	t.Parallel()

	book := BuildMany([]card.Card{{
		Name: "Alice",
		EmbeddedLorebookEntries: []card.EmbeddedLorebookEntry{{
			Key:     []string{"Town"},
			Comment: "Capital",
			Content: "Big city",
		}},
	}}, DefaultPreset())

	embedded := book.Entries["1"]
	if !embedded.Selective {
		t.Fatalf("expected source-gated entry to become selective, got %+v", embedded)
	}
	if len(embedded.KeySecondary) != 1 || embedded.KeySecondary[0] != "Alice" {
		t.Fatalf("expected source-only secondary key, got %+v", embedded.KeySecondary)
	}
}
