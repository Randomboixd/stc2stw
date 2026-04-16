package persona

import "testing"

func TestParseJSONFindsPersonaInKeyedExport(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"personas": {
			"1740236705309-Alice.png": "Alice",
			"1740236705310-Bob.png": "Bob"
		},
		"persona_descriptions": {
			"1740236705309-Alice.png": {"description":"Analyst"},
			"1740236705310-Bob.png": {"description":"Builder"}
		},
		"default_persona": null
	}`)

	persona, err := ParseJSON(input, "bob")
	if err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}
	if persona.Name != "Bob" {
		t.Fatalf("expected Bob, got %q", persona.Name)
	}
	if persona.Description != "Builder" {
		t.Fatalf("expected Builder, got %q", persona.Description)
	}
}

func TestParseJSONCarriesPersonaLorebook(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"personas": {
			"1740236705310-Bob.png": "Bob"
		},
		"persona_descriptions": {
			"1740236705310-Bob.png": {
				"description":"Builder",
				"lorebook":{
					"entries":{
						"0":{
							"key":["Hammer"],
							"comment":"Tool",
							"content":"Heavy"
						}
					}
				}
			}
		}
	}`)

	persona, err := ParseJSON(input, "Bob")
	if err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}
	if len(persona.EmbeddedLorebookEntries) != 1 {
		t.Fatalf("expected embedded lorebook entry, got %d", len(persona.EmbeddedLorebookEntries))
	}
	if persona.EmbeddedLorebookEntries[0].Comment != "Tool" {
		t.Fatalf("unexpected embedded entry: %+v", persona.EmbeddedLorebookEntries[0])
	}
}

func TestParseJSONAcceptsMissingDescriptionRecord(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"personas": {
			"1740236705309-Alice.png": "Alice"
		},
		"persona_descriptions": {},
		"default_persona": null
	}`)

	persona, err := ParseJSON(input, "Alice")
	if err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}
	if persona.Description != "" {
		t.Fatalf("expected empty description, got %q", persona.Description)
	}
}

func TestParseJSONRejectsAmbiguousPersonaNames(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"personas": {
			"1740236705309-Alice.png": "Alice",
			"1740236705310-alice.png": "alice"
		},
		"persona_descriptions": {
			"1740236705309-Alice.png": {"description":"One"},
			"1740236705310-alice.png": {"description":"Two"}
		}
	}`)

	if _, err := ParseJSON(input, "Alice"); err == nil {
		t.Fatal("expected ambiguous persona error")
	}
}

func TestLooksLikeExportJSON(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"personas": {"1740236705309-Alice.png":"Alice"},
		"persona_descriptions": {"1740236705309-Alice.png":{"description":"One"}}
	}`)
	if !LooksLikeExportJSON(input) {
		t.Fatal("expected persona export detection")
	}
}
