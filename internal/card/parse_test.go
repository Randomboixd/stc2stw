package card

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestParseJSONV1(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"name":"Alice",
		"description":"Desc",
		"personality":"Kind",
		"scenario":"Scene",
		"first_mes":"Hello",
		"mes_example":"Alice: hi"
	}`)

	card, err := ParseJSON(input)
	if err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}

	if card.Name != "Alice" {
		t.Fatalf("expected name Alice, got %q", card.Name)
	}
	if card.CreatorNotes != "" {
		t.Fatalf("expected empty creator notes for v1, got %q", card.CreatorNotes)
	}
}

func TestParseJSONV2(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"spec":"chara_card_v2",
		"data":{
			"name":"Alice",
			"description":"Desc",
			"personality":"Kind",
			"scenario":"Scene",
			"first_mes":"Hello",
			"mes_example":"Alice: hi",
			"creator_notes":"Notes",
			"system_prompt":"Prompt",
			"post_history_instructions":"After",
			"creator":"Deck",
			"character_version":"1.0",
			"tags":["tag1","tag2"],
			"alternate_greetings":["Hi","Yo"]
		}
	}`)

	card, err := ParseJSON(input)
	if err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}

	if card.SystemPrompt != "Prompt" {
		t.Fatalf("expected Prompt, got %q", card.SystemPrompt)
	}
	if len(card.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(card.Tags))
	}
}

func TestParseJSONV2CharacterBook(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"spec":"chara_card_v2",
		"data":{
			"name":"Alice",
			"character_book":{
				"entries":[
					{
						"keys":["Town"],
						"content":"Big city",
						"name":"Capital",
						"enabled":true,
						"insertion_order":42,
						"secondary_keys":["Urban"],
						"selective":true,
						"position":"after_char"
					}
				]
			}
		}
	}`)

	card, err := ParseJSON(input)
	if err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}

	if len(card.EmbeddedLorebookEntries) != 1 {
		t.Fatalf("expected 1 embedded lore entry, got %d", len(card.EmbeddedLorebookEntries))
	}
	entry := card.EmbeddedLorebookEntries[0]
	if entry.Comment != "Capital" || entry.Order != 42 || entry.Position != 1 {
		t.Fatalf("unexpected embedded lore entry: %+v", entry)
	}
}

func TestParseEmbeddedStandaloneLorebook(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"entries":{
			"2":{
				"key":["Town"],
				"keysecondary":["Urban"],
				"comment":"Capital",
				"content":"Big city",
				"selective":true,
				"selectiveLogic":3,
				"position":4,
				"role":1
			}
		}
	}`)

	entries := ParseEmbeddedLorebook(input)
	if len(entries) != 1 {
		t.Fatalf("expected 1 embedded lore entry, got %d", len(entries))
	}
	if entries[0].Comment != "Capital" || entries[0].Role != 1 || entries[0].SelectiveLogic != 3 {
		t.Fatalf("unexpected parsed embedded lore entry: %+v", entries[0])
	}
}

func TestParsePNGCharaTextChunk(t *testing.T) {
	t.Parallel()

	payload := mustJSON(t, map[string]any{
		"name":        "Alice",
		"description": "Desc",
		"personality": "Kind",
		"scenario":    "Scene",
		"first_mes":   "Hello",
		"mes_example": "Alice: hi",
	})
	encoded := base64.StdEncoding.EncodeToString(payload)
	pngBytes := buildPNG(t, "tEXt", "chara", []byte(encoded), false)

	card, err := ParsePNG(pngBytes)
	if err != nil {
		t.Fatalf("ParsePNG returned error: %v", err)
	}

	if card.Name != "Alice" {
		t.Fatalf("expected name Alice, got %q", card.Name)
	}
}

func TestParsePNGITextRawJSON(t *testing.T) {
	t.Parallel()

	payload := mustJSON(t, map[string]any{
		"spec": "chara_card_v3",
		"data": map[string]any{
			"name":        "Alice",
			"description": "Desc",
			"personality": "Kind",
			"scenario":    "Scene",
			"first_mes":   "Hello",
			"mes_example": "Alice: hi",
		},
	})
	pngBytes := buildPNG(t, "iTXt", "ccv3", payload, false)

	card, err := ParsePNG(pngBytes)
	if err != nil {
		t.Fatalf("ParsePNG returned error: %v", err)
	}

	if card.Name != "Alice" {
		t.Fatalf("expected name Alice, got %q", card.Name)
	}
}

func TestParsePNGWithoutCardMetadataFails(t *testing.T) {
	t.Parallel()

	pngBytes := buildPlainPNG(t)
	if _, err := ParsePNG(pngBytes); err == nil {
		t.Fatal("expected ParsePNG to fail for missing metadata")
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	return data
}

func buildPlainPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode failed: %v", err)
	}

	return buf.Bytes()
}

func buildPNG(t *testing.T, chunkType, keyword string, payload []byte, compress bool) []byte {
	t.Helper()

	base := buildPlainPNG(t)
	if len(base) < 12 {
		t.Fatal("base png too short")
	}

	chunkData := buildTextChunkData(t, chunkType, keyword, payload, compress)
	chunk := buildChunk(chunkType, chunkData)

	insertAt := len(base) - 12
	result := make([]byte, 0, len(base)+len(chunk))
	result = append(result, base[:insertAt]...)
	result = append(result, chunk...)
	result = append(result, base[insertAt:]...)
	return result
}

func buildTextChunkData(t *testing.T, chunkType, keyword string, payload []byte, compress bool) []byte {
	t.Helper()

	switch chunkType {
	case "tEXt":
		return append(append([]byte(keyword), 0), payload...)
	case "iTXt":
		data := make([]byte, 0, len(keyword)+len(payload)+5)
		data = append(data, []byte(keyword)...)
		data = append(data, 0)
		if compress {
			data = append(data, 1, 0, 0, 0)
			return append(data, mustZlib(t, payload)...)
		}
		data = append(data, 0, 0, 0, 0)
		return append(data, payload...)
	default:
		t.Fatalf("unsupported chunk type in test: %s", chunkType)
		return nil
	}
}

func mustZlib(t *testing.T, payload []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("zlib write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("zlib close failed: %v", err)
	}
	return buf.Bytes()
}

func buildChunk(chunkType string, data []byte) []byte {
	chunk := make([]byte, 0, 12+len(data))
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(data)))
	chunk = append(chunk, lenBuf[:]...)
	chunk = append(chunk, chunkType...)
	chunk = append(chunk, data...)

	var crcBuf [4]byte
	crc := crc32.ChecksumIEEE(append([]byte(chunkType), data...))
	binary.BigEndian.PutUint32(crcBuf[:], crc)
	chunk = append(chunk, crcBuf[:]...)
	return chunk
}
