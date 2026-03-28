package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/deck/stc2stw/internal/card"
	"github.com/deck/stc2stw/internal/lorebook"
	"github.com/deck/stc2stw/internal/persona"
)

// Run executes the CLI and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if err := run(args, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "stc2stw: %v\n", err)
		return 1
	}

	return 0
}

func run(args []string, stdout, stderr io.Writer) error {
	args, err := normalizeArgs(args)
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("stc2stw", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outPath string
	var personaName string
	var mass bool
	var verbose bool

	fs.StringVar(&outPath, "out", "", "write output JSON to a file")
	fs.StringVar(&outPath, "o", "", "write output JSON to a file")
	fs.StringVar(&personaName, "persona", "", "select a persona from a persona export json")
	fs.StringVar(&personaName, "p", "", "select a persona from a persona export json")
	fs.BoolVar(&mass, "mass", false, "combine multiple inputs into one lorebook")
	fs.BoolVar(&mass, "m", false, "combine multiple inputs into one lorebook")
	fs.BoolVar(&verbose, "v", false, "print progress logs to stderr")
	fs.BoolVar(&verbose, "verbose", false, "print progress logs to stderr")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: stc2stw <Character Card.{png,json}> [--out=output.json|-o output.json] [-v|--verbose]")
		fmt.Fprintln(stderr, "   or: stc2stw <Persona Export.json> --persona \"Name\" [--out=output.json|-o output.json] [-v|--verbose]")
		fmt.Fprintln(stderr, "   or: stc2stw <Input1> <Input2> [...] --mass [--out=output.json|-o output.json] [-v|--verbose]")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if fs.NArg() == 0 {
		fs.Usage()
		return fmt.Errorf("expected at least one input")
	}
	if fs.NArg() > 1 && !mass {
		fs.Usage()
		return fmt.Errorf("multiple inputs require --mass")
	}
	if mass && fs.NArg() < 2 {
		fs.Usage()
		return fmt.Errorf("--mass requires at least two inputs")
	}
	if mass && strings.TrimSpace(personaName) != "" {
		return fmt.Errorf("--mass cannot be combined with --persona; use <persona_export.json>:<Persona Name> inputs instead")
	}

	logf := func(format string, values ...any) {}
	if verbose {
		logf = func(format string, values ...any) {
			fmt.Fprintf(stderr, format+"\n", values...)
		}
	}

	parsedCards := make([]card.Card, 0, fs.NArg())
	for _, input := range fs.Args() {
		parsedCard, err := resolveInput(input, mass, personaName, logf)
		if err != nil {
			return err
		}
		parsedCards = append(parsedCards, parsedCard)
	}

	logf("building lorebook with %d entries", len(parsedCards))
	book := lorebook.BuildMany(parsedCards)

	data, err := lorebook.Marshal(book)
	if err != nil {
		return err
	}

	if strings.TrimSpace(outPath) != "" {
		logf("writing lorebook json to: %s", outPath)
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		return nil
	}

	logf("writing lorebook json to stdout")
	_, err = stdout.Write(data)
	return err
}

func normalizeArgs(args []string) ([]string, error) {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if needsValue(arg) {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag needs an argument: %s", arg)
				}
				i++
				flags = append(flags, args[i])
			}
			continue
		}

		positionals = append(positionals, arg)
	}

	return append(flags, positionals...), nil
}

func needsValue(arg string) bool {
	return arg == "--out" || arg == "-o" || arg == "--persona" || arg == "-p"
}

func resolveInput(input string, mass bool, personaName string, logf func(string, ...any)) (card.Card, error) {
	if mass {
		if looksLikeMalformedPersonaReference(input) {
			return card.Card{}, fmt.Errorf("invalid persona reference %q; use <persona_export.json>:<Persona Name>", input)
		}
		if path, personaRefName, ok := parsePersonaReference(input); ok {
			logf("reading persona export: %s", path)
			logf("selecting persona: %s", personaRefName)
			parsedCard, err := persona.ParseFile(path, personaRefName)
			if err != nil {
				return card.Card{}, err
			}
			logf("normalizing card: %s", parsedCard.Name)
			return parsedCard, nil
		}
	}

	if strings.TrimSpace(personaName) != "" {
		if !strings.EqualFold(filepathExt(input), ".json") {
			return card.Card{}, errors.New("--persona requires a json persona export input")
		}
		logf("reading persona export: %s", input)
		logf("selecting persona: %s", personaName)
		parsedCard, err := persona.ParseFile(input, personaName)
		if err != nil {
			return card.Card{}, err
		}
		logf("normalizing card: %s", parsedCard.Name)
		return parsedCard, nil
	}

	logf("reading card: %s", input)
	if strings.EqualFold(filepathExt(input), ".json") {
		if suggested := maybePersonaHint(input); suggested != "" {
			return card.Card{}, fmt.Errorf("%s", suggested)
		}
	}
	parsedCard, err := card.ParseFile(input)
	if err != nil {
		return card.Card{}, err
	}
	logf("normalizing card: %s", parsedCard.Name)
	return parsedCard, nil
}

func maybePersonaHint(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if persona.LooksLikeExportJSON(data) {
		return "input looks like a persona export; use --persona \"<Name>\" to select one persona"
	}
	return ""
}

func filepathExt(path string) string {
	lastDot := strings.LastIndexByte(path, '.')
	if lastDot < 0 {
		return ""
	}
	return path[lastDot:]
}

func parsePersonaReference(input string) (string, string, bool) {
	lower := strings.ToLower(input)
	idx := strings.Index(lower, ".json:")
	if idx < 0 {
		return "", "", false
	}

	path := input[:idx+len(".json")]
	personaName := strings.TrimSpace(input[idx+len(".json:"):])
	if personaName == "" {
		return "", "", false
	}

	return path, personaName, true
}

func looksLikeMalformedPersonaReference(input string) bool {
	lower := strings.ToLower(input)
	idx := strings.Index(lower, ".json:")
	if idx < 0 {
		return false
	}

	return strings.TrimSpace(input[idx+len(".json:"):]) == ""
}
