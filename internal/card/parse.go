package card

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

type rawCardData struct {
	Name                    string         `json:"name"`
	Description             string         `json:"description"`
	Personality             string         `json:"personality"`
	Scenario                string         `json:"scenario"`
	FirstMessage            string         `json:"first_mes"`
	MessageExamples         string         `json:"mes_example"`
	CreatorNotes            string         `json:"creator_notes"`
	SystemPrompt            string         `json:"system_prompt"`
	PostHistoryInstructions string         `json:"post_history_instructions"`
	Creator                 string         `json:"creator"`
	CharacterVersion        string         `json:"character_version"`
	Tags                    []string       `json:"tags"`
	AlternateGreetings      []string       `json:"alternate_greetings"`
	Extensions              map[string]any `json:"extensions"`
}

type rawCard struct {
	Spec string       `json:"spec"`
	Data *rawCardData `json:"data"`

	Name            string `json:"name"`
	Description     string `json:"description"`
	Personality     string `json:"personality"`
	Scenario        string `json:"scenario"`
	FirstMessage    string `json:"first_mes"`
	MessageExamples string `json:"mes_example"`
}

type pngTextChunk struct {
	keyword string
	text    string
}

// ParseFile loads and normalizes a character card from a .png or .json file.
func ParseFile(path string) (Card, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Card{}, fmt.Errorf("read input: %w", err)
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return ParseJSON(data)
	case ".png":
		return ParsePNG(data)
	default:
		return Card{}, fmt.Errorf("unsupported input extension %q", filepath.Ext(path))
	}
}

// ParseJSON parses and normalizes card JSON.
func ParseJSON(data []byte) (Card, error) {
	var raw rawCard
	if err := json.Unmarshal(data, &raw); err != nil {
		return Card{}, fmt.Errorf("parse card json: %w", err)
	}

	card := normalizeRawCard(raw)
	if strings.TrimSpace(card.Name) == "" {
		return Card{}, errors.New("card is missing a character name")
	}

	return card, nil
}

// ParsePNG extracts embedded card JSON from PNG metadata and normalizes it.
func ParsePNG(data []byte) (Card, error) {
	chunks, err := extractPNGTextChunks(data)
	if err != nil {
		return Card{}, err
	}

	candidates := prioritizeChunkCandidates(chunks)
	for _, candidate := range candidates {
		card, err := ParseJSON([]byte(candidate))
		if err == nil {
			return card, nil
		}

		decoded, ok := decodeMaybeBase64(candidate)
		if !ok {
			continue
		}

		card, err = ParseJSON(decoded)
		if err == nil {
			return card, nil
		}
	}

	return Card{}, errors.New("png does not contain recognizable character card metadata")
}

func normalizeRawCard(raw rawCard) Card {
	if raw.Data != nil {
		return Card{
			Name:                    strings.TrimSpace(raw.Data.Name),
			Description:             raw.Data.Description,
			Personality:             raw.Data.Personality,
			Scenario:                raw.Data.Scenario,
			FirstMessage:            raw.Data.FirstMessage,
			MessageExamples:         raw.Data.MessageExamples,
			CreatorNotes:            raw.Data.CreatorNotes,
			SystemPrompt:            raw.Data.SystemPrompt,
			PostHistoryInstructions: raw.Data.PostHistoryInstructions,
			Creator:                 raw.Data.Creator,
			CharacterVersion:        raw.Data.CharacterVersion,
			Tags:                    slices.Clone(raw.Data.Tags),
			AlternateGreetings:      slices.Clone(raw.Data.AlternateGreetings),
		}
	}

	return Card{
		Name:            strings.TrimSpace(raw.Name),
		Description:     raw.Description,
		Personality:     raw.Personality,
		Scenario:        raw.Scenario,
		FirstMessage:    raw.FirstMessage,
		MessageExamples: raw.MessageExamples,
	}
}

func prioritizeChunkCandidates(chunks []pngTextChunk) []string {
	prioritized := make([]string, 0, len(chunks))
	seen := make(map[string]struct{}, len(chunks))

	appendChunk := func(chunk pngTextChunk) {
		if _, ok := seen[chunk.text]; ok {
			return
		}
		seen[chunk.text] = struct{}{}
		prioritized = append(prioritized, chunk.text)
	}

	for _, chunk := range chunks {
		if strings.EqualFold(chunk.keyword, "chara") || strings.EqualFold(chunk.keyword, "ccv3") {
			appendChunk(chunk)
		}
	}
	for _, chunk := range chunks {
		appendChunk(chunk)
	}

	return prioritized
}

func decodeMaybeBase64(text string) ([]byte, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, false
	}

	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(trimmed)
		if err == nil && json.Valid(decoded) {
			return decoded, true
		}
	}

	return nil, false
}

func extractPNGTextChunks(data []byte) ([]pngTextChunk, error) {
	if len(data) < len(pngSignature) || !bytes.Equal(data[:len(pngSignature)], pngSignature) {
		return nil, errors.New("input is not a valid png")
	}

	var chunks []pngTextChunk
	offset := len(pngSignature)

	for offset+8 <= len(data) {
		length := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		if offset+4 > len(data) {
			return nil, errors.New("png chunk header is truncated")
		}

		chunkType := string(data[offset : offset+4])
		offset += 4

		if offset+length+4 > len(data) {
			return nil, errors.New("png chunk data is truncated")
		}

		chunkData := data[offset : offset+length]
		offset += length
		offset += 4 // crc

		textChunk, ok, err := decodeTextChunk(chunkType, chunkData)
		if err != nil {
			return nil, err
		}
		if ok {
			chunks = append(chunks, textChunk)
		}

		if chunkType == "IEND" {
			break
		}
	}

	if len(chunks) == 0 {
		return nil, errors.New("png does not contain text metadata chunks")
	}

	return chunks, nil
}

func decodeTextChunk(chunkType string, data []byte) (pngTextChunk, bool, error) {
	switch chunkType {
	case "tEXt":
		idx := bytes.IndexByte(data, 0)
		if idx <= 0 {
			return pngTextChunk{}, false, nil
		}
		return pngTextChunk{
			keyword: string(data[:idx]),
			text:    string(data[idx+1:]),
		}, true, nil
	case "zTXt":
		idx := bytes.IndexByte(data, 0)
		if idx <= 0 || idx+2 > len(data) {
			return pngTextChunk{}, false, nil
		}
		if data[idx+1] != 0 {
			return pngTextChunk{}, false, nil
		}
		text, err := decompressZlib(data[idx+2:])
		if err != nil {
			return pngTextChunk{}, false, fmt.Errorf("decode zTXt chunk: %w", err)
		}
		return pngTextChunk{
			keyword: string(data[:idx]),
			text:    text,
		}, true, nil
	case "iTXt":
		keywordEnd := bytes.IndexByte(data, 0)
		if keywordEnd <= 0 || keywordEnd+5 > len(data) {
			return pngTextChunk{}, false, nil
		}

		keyword := string(data[:keywordEnd])
		offset := keywordEnd + 1
		compressionFlag := data[offset]
		offset++
		compressionMethod := data[offset]
		offset++

		languageEnd := bytes.IndexByte(data[offset:], 0)
		if languageEnd < 0 {
			return pngTextChunk{}, false, nil
		}
		offset += languageEnd + 1

		translatedEnd := bytes.IndexByte(data[offset:], 0)
		if translatedEnd < 0 {
			return pngTextChunk{}, false, nil
		}
		offset += translatedEnd + 1

		compressed := compressionFlag == 1
		textBytes := data[offset:]
		if compressed {
			if compressionMethod != 0 {
				return pngTextChunk{}, false, nil
			}
			text, err := decompressZlib(textBytes)
			if err != nil {
				return pngTextChunk{}, false, fmt.Errorf("decode iTXt chunk: %w", err)
			}
			return pngTextChunk{keyword: keyword, text: text}, true, nil
		}

		return pngTextChunk{keyword: keyword, text: string(textBytes)}, true, nil
	default:
		return pngTextChunk{}, false, nil
	}
}

func decompressZlib(data []byte) (string, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
