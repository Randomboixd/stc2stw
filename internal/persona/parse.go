package persona

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/deck/stc2stw/internal/card"
)

type exportFile struct {
	Personas            map[string]string            `json:"personas"`
	PersonaDescriptions map[string]descriptionRecord `json:"persona_descriptions"`
	DefaultPersona      *string                      `json:"default_persona"`
}

type descriptionRecord struct {
	Description string `json:"description"`
}

type record struct {
	StorageKey  string
	Name        string
	Description string
}

// ParseFile loads a persona export JSON and returns the selected persona as a normalized card.
func ParseFile(path, personaName string) (card.Card, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return card.Card{}, fmt.Errorf("read input: %w", err)
	}

	return ParseJSON(data, personaName)
}

// ParseJSON extracts a single persona from a SillyTavern persona export JSON blob.
func ParseJSON(data []byte, personaName string) (card.Card, error) {
	personaName = strings.TrimSpace(personaName)
	if personaName == "" {
		return card.Card{}, errors.New("persona name is required")
	}

	records, err := parseRecords(data)
	if err != nil {
		return card.Card{}, err
	}

	lowerTarget := strings.ToLower(personaName)
	var matches []record
	for _, record := range records {
		if strings.ToLower(record.Name) == lowerTarget {
			matches = append(matches, record)
		}
	}

	switch len(matches) {
	case 0:
		return card.Card{}, fmt.Errorf("persona %q was not found in the export", personaName)
	case 1:
		return card.Card{
			Name:        matches[0].Name,
			Description: matches[0].Description,
		}, nil
	default:
		return card.Card{}, fmt.Errorf("persona %q is ambiguous in the export", personaName)
	}
}

// LooksLikeExportJSON reports whether the JSON appears to be a SillyTavern persona export.
func LooksLikeExportJSON(data []byte) bool {
	records, err := parseRecords(data)
	return err == nil && len(records) > 0
}

func parseRecords(data []byte) ([]record, error) {
	var payload exportFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse persona export json: %w", err)
	}

	if len(payload.Personas) == 0 {
		return nil, errors.New("json does not contain a SillyTavern persona export")
	}

	records := make([]record, 0, len(payload.Personas))
	for storageKey, name := range payload.Personas {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		record := record{
			StorageKey: storageKey,
			Name:       name,
		}

		if description, ok := payload.PersonaDescriptions[storageKey]; ok {
			record.Description = description.Description
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, errors.New("json does not contain any named personas")
	}

	return records, nil
}
