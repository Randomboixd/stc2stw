package card

// Card is the normalized character-card model used by the converter.
type Card struct {
	Name                    string
	Description             string
	Personality             string
	Scenario                string
	FirstMessage            string
	MessageExamples         string
	CreatorNotes            string
	SystemPrompt            string
	PostHistoryInstructions string
	Creator                 string
	CharacterVersion        string
	Tags                    []string
	AlternateGreetings      []string
	EmbeddedLorebookEntries []EmbeddedLorebookEntry
}

// EmbeddedLorebookEntry is a normalized lorebook entry found inside a card or persona.
type EmbeddedLorebookEntry struct {
	Key                       []string
	KeySecondary              []string
	Comment                   string
	Name                      string
	Content                   string
	Constant                  bool
	Vectorized                bool
	Selective                 bool
	SelectiveLogic            int
	AddMemo                   bool
	Order                     int
	Position                  int
	Disable                   bool
	ExcludeRecursion          bool
	PreventRecursion          bool
	DelayUntilRecursion       bool
	Probability               int
	UseProbability            bool
	Depth                     int
	Group                     string
	GroupOverride             bool
	GroupWeight               int
	ScanDepth                 *int
	CaseSensitive             *bool
	MatchWholeWords           *bool
	UseGroupScoring           *bool
	AutomationID              string
	Role                      int
	OutletName                string
	Sticky                    int
	Cooldown                  int
	Delay                     int
	CharacterFilterExclude    *bool
	CharacterFilterNames      []string
	CharacterFilterTags       []string
	MatchCharacterDepthPrompt *bool
	MatchCharacterDescription *bool
	MatchCharacterPersonality *bool
	MatchCreatorNotes         *bool
	MatchPersonaDescription   *bool
	MatchScenario             *bool
}
