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
}
