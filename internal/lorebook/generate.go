package lorebook

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/deck/stc2stw/internal/card"
)

const (
	defaultOrder    = 100
	defaultPosition = 4
)

var speakerPrefixPattern = regexp.MustCompile(`(?m)^(?:<START>\s*)?(?:<([^>\n]+)>|([A-Za-z0-9][A-Za-z0-9 ._'’-]{0,62}[A-Za-z0-9])):`)

type Lorebook struct {
	Entries map[string]Entry `json:"entries"`
}

type Entry struct {
	UID                 int      `json:"uid"`
	Key                 []string `json:"key"`
	KeySecondary        []string `json:"keysecondary"`
	Comment             string   `json:"comment"`
	Content             string   `json:"content"`
	Constant            bool     `json:"constant"`
	Vectorized          bool     `json:"vectorized"`
	Selective           bool     `json:"selective"`
	SelectiveLogic      int      `json:"selectiveLogic"`
	AddMemo             bool     `json:"addMemo"`
	Order               int      `json:"order"`
	Position            int      `json:"position"`
	Disable             bool     `json:"disable"`
	ExcludeRecursion    bool     `json:"excludeRecursion"`
	PreventRecursion    bool     `json:"preventRecursion"`
	DelayUntilRecursion bool     `json:"delayUntilRecursion"`
	Probability         int      `json:"probability"`
	UseProbability      bool     `json:"useProbability"`
	Depth               int      `json:"depth"`
	Group               string   `json:"group"`
	GroupOverride       bool     `json:"groupOverride"`
	GroupWeight         int      `json:"groupWeight"`
	ScanDepth           *int     `json:"scanDepth"`
	CaseSensitive       *bool    `json:"caseSensitive"`
	MatchWholeWords     *bool    `json:"matchWholeWords"`
	UseGroupScoring     *bool    `json:"useGroupScoring"`
	AutomationID        string   `json:"automationId"`
	Role                int      `json:"role"`
	Sticky              int      `json:"sticky"`
	Cooldown            int      `json:"cooldown"`
	Delay               int      `json:"delay"`
	DisplayIndex        int      `json:"displayIndex"`
}

// Build converts a normalized card into a standalone SillyTavern lorebook.
func Build(c card.Card) Lorebook {
	return BuildMany([]card.Card{c})
}

// BuildMany converts a list of normalized cards into a standalone SillyTavern lorebook.
func BuildMany(cards []card.Card) Lorebook {
	entries := make(map[string]Entry, len(cards))
	for i, c := range cards {
		entry := Entry{
			UID:                 0,
			Key:                 buildKeys(c),
			KeySecondary:        []string{},
			Comment:             c.Name,
			Content:             buildMarkdown(c),
			Constant:            false,
			Vectorized:          false,
			Selective:           false,
			SelectiveLogic:      0,
			AddMemo:             true,
			Order:               defaultOrder,
			Position:            defaultPosition,
			Disable:             false,
			ExcludeRecursion:    false,
			PreventRecursion:    false,
			DelayUntilRecursion: false,
			Probability:         100,
			UseProbability:      true,
			Depth:               0,
			Group:               "",
			GroupOverride:       false,
			GroupWeight:         100,
			ScanDepth:           nil,
			CaseSensitive:       nil,
			MatchWholeWords:     nil,
			UseGroupScoring:     nil,
			AutomationID:        "",
			Role:                0,
			Sticky:              0,
			Cooldown:            0,
			Delay:               0,
			DisplayIndex:        0,
		}
		entry.UID = i
		entry.DisplayIndex = i
		entries[strconv.Itoa(i)] = entry
	}

	return Lorebook{
		Entries: entries,
	}
}

// Marshal returns deterministic pretty-printed JSON.
func Marshal(book Lorebook) ([]byte, error) {
	data, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal lorebook json: %w", err)
	}

	data = append(data, '\n')
	return data, nil
}

func buildKeys(c card.Card) []string {
	keys := []string{strings.TrimSpace(c.Name)}
	keys = append(keys, extractAliases(c)...)

	seen := make(map[string]struct{}, len(keys))
	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		canonical := strings.ToLower(trimmed)
		if _, ok := seen[canonical]; ok {
			continue
		}
		seen[canonical] = struct{}{}
		filtered = append(filtered, trimmed)
	}

	return filtered
}

func extractAliases(c card.Card) []string {
	var aliases []string
	baseName := strings.ToLower(strings.TrimSpace(c.Name))
	texts := []string{c.FirstMessage, c.MessageExamples}

	for _, text := range texts {
		matches := speakerPrefixPattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			candidate := strings.TrimSpace(match[1])
			if candidate == "" {
				candidate = strings.TrimSpace(match[2])
			}
			if candidate == "" {
				continue
			}

			lower := strings.ToLower(candidate)
			if lower == baseName {
				continue
			}
			if strings.Contains(lower, "{{char}}") || strings.Contains(lower, "{{user}}") {
				continue
			}
			if lower == "user" || lower == "char" || lower == "system" {
				continue
			}

			aliases = append(aliases, candidate)
		}
	}

	return slices.Clip(aliases)
}

func buildMarkdown(c card.Card) string {
	var sections []string
	sections = append(sections, "# "+c.Name)

	appendSection := func(title, content string) {
		content = strings.TrimSpace(content)
		if content == "" {
			return
		}
		sections = append(sections, "## "+title+"\n"+content)
	}

	appendSection("Description", c.Description)
	appendSection("Personality", c.Personality)
	appendSection("Scenario", c.Scenario)
	appendSection("First Message", c.FirstMessage)
	appendSection("Example Messages", c.MessageExamples)
	appendSection("Creator Notes", c.CreatorNotes)
	appendSection("System Prompt", c.SystemPrompt)
	appendSection("Post-History Instructions", c.PostHistoryInstructions)
	if len(c.AlternateGreetings) > 0 {
		appendSection("Alternate Greetings", strings.Join(nonEmpty(c.AlternateGreetings), "\n\n"))
	}
	if len(c.Tags) > 0 {
		appendSection("Tags", strings.Join(nonEmpty(c.Tags), ", "))
	}
	appendSection("Creator", c.Creator)
	appendSection("Character Version", c.CharacterVersion)

	return strings.Join(sections, "\n\n")
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
