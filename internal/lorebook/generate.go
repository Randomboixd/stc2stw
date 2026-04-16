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
	defaultOrder     = 100
	positionBefore   = 0
	positionAfter    = 1
	positionANTop    = 2
	positionANBottom = 3
	positionAtDepth  = 4
	positionEMTop    = 5
	positionEMBottom = 6
	positionOutlet   = 7

	roleSystem    = 0
	roleUser      = 1
	roleAssistant = 2

	defaultOutletName = "stc2stw"
)

var speakerPrefixPattern = regexp.MustCompile(`(?m)^(?:<START>\s*)?(?:<([^>\n]+)>|([A-Za-z0-9][A-Za-z0-9 ._'’-]{0,62}[A-Za-z0-9])):`)

type Lorebook struct {
	Entries map[string]Entry `json:"entries"`
}

type Entry struct {
	UID                       int      `json:"uid"`
	Key                       []string `json:"key"`
	KeySecondary              []string `json:"keysecondary"`
	Comment                   string   `json:"comment"`
	Content                   string   `json:"content"`
	Constant                  bool     `json:"constant"`
	Vectorized                bool     `json:"vectorized"`
	Selective                 bool     `json:"selective"`
	SelectiveLogic            int      `json:"selectiveLogic"`
	AddMemo                   bool     `json:"addMemo"`
	Order                     int      `json:"order"`
	Position                  int      `json:"position"`
	Disable                   bool     `json:"disable"`
	ExcludeRecursion          bool     `json:"excludeRecursion"`
	PreventRecursion          bool     `json:"preventRecursion"`
	DelayUntilRecursion       bool     `json:"delayUntilRecursion"`
	Probability               int      `json:"probability"`
	UseProbability            bool     `json:"useProbability"`
	Depth                     int      `json:"depth"`
	Group                     string   `json:"group"`
	GroupOverride             bool     `json:"groupOverride"`
	GroupWeight               int      `json:"groupWeight"`
	ScanDepth                 *int     `json:"scanDepth"`
	CaseSensitive             *bool    `json:"caseSensitive"`
	MatchWholeWords           *bool    `json:"matchWholeWords"`
	UseGroupScoring           *bool    `json:"useGroupScoring"`
	AutomationID              string   `json:"automationId"`
	Role                      int      `json:"role"`
	OutletName                string   `json:"outletName"`
	Sticky                    int      `json:"sticky"`
	Cooldown                  int      `json:"cooldown"`
	Delay                     int      `json:"delay"`
	DisplayIndex              int      `json:"displayIndex"`
	CharacterFilterExclude    *bool    `json:"characterFilterExclude"`
	CharacterFilterNames      []string `json:"characterFilterNames"`
	CharacterFilterTags       []string `json:"characterFilterTags"`
	MatchCharacterDepthPrompt *bool    `json:"matchCharacterDepthPrompt"`
	MatchCharacterDescription *bool    `json:"matchCharacterDescription"`
	MatchCharacterPersonality *bool    `json:"matchCharacterPersonality"`
	MatchCreatorNotes         *bool    `json:"matchCreatorNotes"`
	MatchPersonaDescription   *bool    `json:"matchPersonaDescription"`
	MatchScenario             *bool    `json:"matchScenario"`
}

type InsertionPreset struct {
	Position   int
	Role       int
	OutletName string
}

type BuildOptions struct {
	Compact             bool
	IncludeCreatorNotes bool
}

var defaultPreset = InsertionPreset{
	Position: positionAtDepth,
	Role:     roleUser,
}

var defaultBuildOptions = BuildOptions{
	Compact:             true,
	IncludeCreatorNotes: false,
}

// Build converts a normalized card into a standalone SillyTavern lorebook.
func Build(c card.Card) Lorebook {
	return BuildMany([]card.Card{c}, defaultPreset)
}

// BuildMany converts a list of normalized cards into a standalone SillyTavern lorebook.
func BuildMany(cards []card.Card, preset InsertionPreset) Lorebook {
	return BuildManyWithOptions(cards, preset, defaultBuildOptions)
}

func BuildManyWithOptions(cards []card.Card, preset InsertionPreset, options BuildOptions) Lorebook {
	flattened := make([]Entry, 0, len(cards))
	for _, c := range cards {
		flattened = append(flattened, buildPrimaryEntry(c, preset, options))
		if options.Compact {
			flattened = append(flattened, compactEmbeddedEntries(c)...)
		}
	}

	entries := make(map[string]Entry, len(flattened))
	for i, entry := range flattened {
		entry.UID = i
		entry.DisplayIndex = i
		entries[strconv.Itoa(i)] = entry
	}

	return Lorebook{Entries: entries}
}

func DefaultPreset() InsertionPreset {
	return defaultPreset
}

func ResolvePositionPreset(value string) (InsertionPreset, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "@duser":
		return defaultPreset, nil
	case "bchar":
		return InsertionPreset{Position: positionBefore, Role: roleSystem}, nil
	case "achar":
		return InsertionPreset{Position: positionAfter, Role: roleSystem}, nil
	case "bex":
		return InsertionPreset{Position: positionEMTop, Role: roleSystem}, nil
	case "aex":
		return InsertionPreset{Position: positionEMBottom, Role: roleSystem}, nil
	case "tan":
		return InsertionPreset{Position: positionANTop, Role: roleSystem}, nil
	case "ban":
		return InsertionPreset{Position: positionANBottom, Role: roleSystem}, nil
	case "@dsys":
		return InsertionPreset{Position: positionAtDepth, Role: roleSystem}, nil
	case "@dass":
		return InsertionPreset{Position: positionAtDepth, Role: roleAssistant}, nil
	case "outlet":
		return InsertionPreset{Position: positionOutlet, Role: roleSystem, OutletName: defaultOutletName}, nil
	default:
		return InsertionPreset{}, fmt.Errorf("invalid --position %q; expected one of: bchar, achar, bex, aex, tan, ban, @dsys, @duser, @dass, outlet", value)
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

func buildPrimaryEntry(c card.Card, preset InsertionPreset, options BuildOptions) Entry {
	return Entry{
		UID:                 0,
		Key:                 buildKeys(c),
		KeySecondary:        []string{},
		Comment:             c.Name,
		Content:             buildMarkdown(c, options),
		Constant:            false,
		Vectorized:          false,
		Selective:           false,
		SelectiveLogic:      0,
		AddMemo:             true,
		Order:               defaultOrder,
		Position:            preset.Position,
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
		Role:                preset.Role,
		OutletName:          preset.OutletName,
		Sticky:              0,
		Cooldown:            0,
		Delay:               0,
		DisplayIndex:        0,
	}
}

func compactEmbeddedEntries(c card.Card) []Entry {
	if len(c.EmbeddedLorebookEntries) == 0 {
		return nil
	}

	entries := make([]Entry, 0, len(c.EmbeddedLorebookEntries))
	for _, embedded := range c.EmbeddedLorebookEntries {
		keySecondary := dedupeStrings(append(slices.Clone(embedded.KeySecondary), c.Name))
		selective := embedded.Selective
		if len(keySecondary) > 0 && !selective {
			selective = true
		}

		commentSuffix := firstNonEmpty(embedded.Comment, embedded.Name)
		entry := Entry{
			UID:                       0,
			Key:                       dedupeStrings(slices.Clone(embedded.Key)),
			KeySecondary:              keySecondary,
			Comment:                   fmt.Sprintf("(src: %s) -> %s", c.Name, commentSuffix),
			Content:                   embedded.Content,
			Constant:                  embedded.Constant,
			Vectorized:                embedded.Vectorized,
			Selective:                 selective,
			SelectiveLogic:            embedded.SelectiveLogic,
			AddMemo:                   embedded.AddMemo,
			Order:                     embedded.Order,
			Position:                  embedded.Position,
			Disable:                   embedded.Disable,
			ExcludeRecursion:          embedded.ExcludeRecursion,
			PreventRecursion:          embedded.PreventRecursion,
			DelayUntilRecursion:       embedded.DelayUntilRecursion,
			Probability:               embedded.Probability,
			UseProbability:            embedded.UseProbability,
			Depth:                     embedded.Depth,
			Group:                     embedded.Group,
			GroupOverride:             embedded.GroupOverride,
			GroupWeight:               embedded.GroupWeight,
			ScanDepth:                 cloneIntPointer(embedded.ScanDepth),
			CaseSensitive:             cloneBoolPointer(embedded.CaseSensitive),
			MatchWholeWords:           cloneBoolPointer(embedded.MatchWholeWords),
			UseGroupScoring:           cloneBoolPointer(embedded.UseGroupScoring),
			AutomationID:              embedded.AutomationID,
			Role:                      embedded.Role,
			OutletName:                embedded.OutletName,
			Sticky:                    embedded.Sticky,
			Cooldown:                  embedded.Cooldown,
			Delay:                     embedded.Delay,
			DisplayIndex:              0,
			CharacterFilterExclude:    cloneBoolPointer(embedded.CharacterFilterExclude),
			CharacterFilterNames:      slices.Clone(embedded.CharacterFilterNames),
			CharacterFilterTags:       slices.Clone(embedded.CharacterFilterTags),
			MatchCharacterDepthPrompt: cloneBoolPointer(embedded.MatchCharacterDepthPrompt),
			MatchCharacterDescription: cloneBoolPointer(embedded.MatchCharacterDescription),
			MatchCharacterPersonality: cloneBoolPointer(embedded.MatchCharacterPersonality),
			MatchCreatorNotes:         cloneBoolPointer(embedded.MatchCreatorNotes),
			MatchPersonaDescription:   cloneBoolPointer(embedded.MatchPersonaDescription),
			MatchScenario:             cloneBoolPointer(embedded.MatchScenario),
		}
		entries = append(entries, entry)
	}

	return entries
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

func buildMarkdown(c card.Card, options BuildOptions) string {
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
	if options.IncludeCreatorNotes {
		appendSection("Creator Notes", c.CreatorNotes)
	}
	appendSection("System Prompt", c.SystemPrompt)
	appendSection("Post-History Instructions", c.PostHistoryInstructions)
	if len(c.AlternateGreetings) > 0 {
		appendSection("Alternate Greetings", formatAlternateGreetings(c.AlternateGreetings))
	}
	if len(c.Tags) > 0 {
		appendSection("Tags", strings.Join(nonEmpty(c.Tags), ", "))
	}
	appendSection("Creator", c.Creator)
	appendSection("Character Version", c.CharacterVersion)

	return strings.Join(sections, "\n\n")
}

func formatAlternateGreetings(values []string) string {
	greetings := nonEmpty(values)
	if len(greetings) == 0 {
		return ""
	}

	formatted := make([]string, 0, len(greetings))
	for i, greeting := range greetings {
		formatted = append(formatted, fmt.Sprintf("### Greeting %d\n%s", i+1, greeting))
	}

	return strings.Join(formatted, "\n\n")
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

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		canonical := strings.ToLower(trimmed)
		if _, ok := seen[canonical]; ok {
			continue
		}
		seen[canonical] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
