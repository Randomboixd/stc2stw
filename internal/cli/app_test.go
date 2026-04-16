package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesJSONToStdoutByDefault(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "card.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"spec":"chara_card_v2",
		"data":{
			"name":"Alice",
			"description":"Desc",
			"personality":"Kind",
			"scenario":"Scene",
			"first_mes":"Hello",
			"mes_example":"Alice: hi",
			"character_book":{
				"entries":[
					{"keys":["Town"],"content":"Big city","name":"Capital"}
				]
			}
		}
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"entries\"") {
		t.Fatalf("expected lorebook json on stdout, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "\"position\": 4") || !strings.Contains(stdout.String(), "\"role\": 1") {
		t.Fatalf("expected default @duser position/role, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "\"comment\": \"(src: Alice) -\\u003e Capital\"") {
		t.Fatalf("expected compacted embedded entry, got %q", stdout.String())
	}
}

func TestRunWritesFileWhenOutIsSet(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "card.json")
	outputPath := filepath.Join(tempDir, "out.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"name":"Alice",
		"description":"Desc",
		"personality":"Kind",
		"scenario":"Scene",
		"first_mes":"Hello",
		"mes_example":"Alice: hi"
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "--out=" + outputPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(data), "\"entries\"") {
		t.Fatalf("expected lorebook json file, got %q", string(data))
	}
}

func TestRunWritesFileWhenShortOutIsSet(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "card.json")
	outputPath := filepath.Join(tempDir, "out.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"name":"Alice",
		"description":"Desc"
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "-o", outputPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
}

func TestRunVerboseLogsToStderr(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "card.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"name":"Alice",
		"description":"Desc",
		"personality":"Kind",
		"scenario":"Scene",
		"first_mes":"Hello",
		"mes_example":"Alice: hi"
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--verbose", inputPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stderr.String(), "reading card:") {
		t.Fatalf("expected verbose stderr logs, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"entries\"") {
		t.Fatalf("expected lorebook json on stdout, got %q", stdout.String())
	}
}

func TestRunInvalidInputFails(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"missing.json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "read input:") {
		t.Fatalf("expected read error, got %q", stderr.String())
	}
}

func TestRunPersonaExportWritesJSONToStdout(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "personas.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"personas": {
			"1740236705309-Alice.png": "Alice",
			"1740236705310-Bob.png": "Bob"
		},
		"persona_descriptions": {
			"1740236705309-Alice.png": {"description":"Analyst"},
			"1740236705310-Bob.png": {
				"description":"Builder",
				"lorebook":{
					"entries":{
						"0":{"key":["Hammer"],"comment":"Tool","content":"Heavy"}
					}
				}
			}
		}
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "--persona", "bob"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "\"comment\": \"Bob\"") {
		t.Fatalf("expected selected persona in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "## Description\\nBuilder") && !strings.Contains(stdout.String(), "## Description\nBuilder") {
		t.Fatalf("expected persona description in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "\"comment\": \"(src: Bob) -\\u003e Tool\"") {
		t.Fatalf("expected compacted persona lorebook entry, got %q", stdout.String())
	}
}

func TestRunPersonaExportShortFlagsWriteFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "personas.json")
	outputPath := filepath.Join(tempDir, "out.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"personas": {
			"1740236705309-Alice.png": "Alice",
			"1740236705310-Bob.png": "Bob"
		},
		"persona_descriptions": {
			"1740236705309-Alice.png": {"description":"Analyst"},
			"1740236705310-Bob.png": {"description":"Builder"}
		}
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "-p", "Alice", "-o", outputPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(data), "\"comment\": \"Alice\"") {
		t.Fatalf("expected selected persona in output file, got %q", string(data))
	}
}

func TestRunSuggestsPersonaFlagForPersonaExport(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "personas.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"personas": {
			"1740236705309-Alice.png": "Alice",
			"1740236705310-Bob.png": "Bob"
		},
		"persona_descriptions": {
			"1740236705309-Alice.png": {"description":"Analyst"},
			"1740236705310-Bob.png": {"description":"Builder"}
		}
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "use --persona") {
		t.Fatalf("expected persona hint in stderr, got %q", stderr.String())
	}
}

func TestRunPersonaExportSupportsKeyedSillyTavernFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "personas.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"personas": {
			"1740236705309-User.png": "User"
		},
		"persona_descriptions": {
			"1740236705309-User.png": {
				"description": "",
				"position": 0,
				"depth": 2,
				"role": 0
			}
		},
		"default_persona": null
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "-p", "User"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"comment\": \"User\"") {
		t.Fatalf("expected selected persona in output, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "1740236705309-User.png") {
		t.Fatalf("expected storage key to stay out of output, got %q", stdout.String())
	}
}

func TestRunMassModeBuildsMultipleEntries(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	personasPath := filepath.Join(tempDir, "personas.json")
	cardPath := filepath.Join(tempDir, "bob.json")
	if err := os.WriteFile(personasPath, []byte(`{
		"personas": {
			"1740236705309-User.png": "User"
		},
		"persona_descriptions": {
			"1740236705309-User.png": {"description":"Analyst"}
		}
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(cardPath, []byte(`{
		"name":"Bob",
		"description":"Builder"
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{personasPath + ":User", cardPath, "--mass"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}
	text := stdout.String()
	if !strings.Contains(text, "\"0\": {") || !strings.Contains(text, "\"1\": {") {
		t.Fatalf("expected multiple entries, got %q", text)
	}
	if !strings.Contains(text, "\"comment\": \"User\"") || !strings.Contains(text, "\"comment\": \"Bob\"") {
		t.Fatalf("expected both inputs in lorebook, got %q", text)
	}
}

func TestRunPositionPresetAppliesToAllMassEntries(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cardAPath := filepath.Join(tempDir, "alice.json")
	cardBPath := filepath.Join(tempDir, "bob.json")
	if err := os.WriteFile(cardAPath, []byte(`{"name":"Alice","description":"A"}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(cardBPath, []byte(`{"name":"Bob","description":"B"}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{cardAPath, cardBPath, "--mass", "--position", "bchar"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}
	text := stdout.String()
	if strings.Count(text, "\"position\": 0") != 2 {
		t.Fatalf("expected both entries at before-char position, got %q", text)
	}
}

func TestRunNoCompactDisablesEmbeddedLoreCopy(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "card.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"spec":"chara_card_v2",
		"data":{
			"name":"Alice",
			"character_book":{
				"entries":[
					{"keys":["Town"],"content":"Big city","name":"Capital"}
				]
			}
		}
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "--no-compact"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "(src: Alice) -> Capital") {
		t.Fatalf("expected no compacted entries, got %q", stdout.String())
	}
}

func TestRunPositionDoesNotOverrideCompactedEntryPosition(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "card.json")
	if err := os.WriteFile(inputPath, []byte(`{
		"spec":"chara_card_v2",
		"data":{
			"name":"Alice",
			"character_book":{
				"entries":[
					{"keys":["Town"],"content":"Big city","name":"Capital","position":"before_char"}
				]
			}
		}
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "--position", "outlet"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}
	text := stdout.String()
	if !strings.Contains(text, "\"outletName\": \"stc2stw\"") {
		t.Fatalf("expected primary entry to use outlet preset, got %q", text)
	}
	if !strings.Contains(text, "\"comment\": \"(src: Alice) -\\u003e Capital\"") || !strings.Contains(text, "\"position\": 0") {
		t.Fatalf("expected compacted entry to keep embedded position, got %q", text)
	}
}

func TestRunOutletPositionWritesOutletName(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "card.json")
	if err := os.WriteFile(inputPath, []byte(`{"name":"Alice","description":"Desc"}`), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{inputPath, "-P", "outlet"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"position\": 7") || !strings.Contains(stdout.String(), "\"outletName\": \"stc2stw\"") {
		t.Fatalf("expected outlet position and outlet name, got %q", stdout.String())
	}
}

func TestRunInvalidPositionFails(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"card.json", "--position", "nope"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "invalid --position") {
		t.Fatalf("expected invalid position error, got %q", stderr.String())
	}
}

func TestRunMultipleInputsWithoutMassFails(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"one.json", "two.json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "require --mass") {
		t.Fatalf("expected --mass guidance, got %q", stderr.String())
	}
}

func TestRunMassCannotUsePersonaFlag(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"one.json", "two.json", "--mass", "--persona", "User"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "cannot be combined with --persona") {
		t.Fatalf("expected mass/persona conflict, got %q", stderr.String())
	}
}

func TestRunMassMalformedPersonaReferenceFallsBackToError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"personas.json:", "other.json", "--mass"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "invalid persona reference") {
		t.Fatalf("expected malformed persona reference error, got %q", stderr.String())
	}
}
