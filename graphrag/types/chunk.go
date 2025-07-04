package types

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/kun/log"
)

// TextWChars returns the text split into individual UTF-8 wide characters/runes as strings
func (chunk *Chunk) TextWChars() []string {
	if chunk.Text == "" {
		return []string{}
	}

	var lines []string
	text := chunk.Text
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		if r == utf8.RuneError {
			// Skip invalid UTF-8 sequences
			text = text[1:]
			continue
		}
		lines = append(lines, string(r))
		text = text[size:]
	}

	return lines
}

// TextWCharsJSON returns the text split into individual UTF-8 wide characters/runes as JSON
func (chunk *Chunk) TextWCharsJSON() (string, error) {
	lines := chunk.TextWChars()
	json, err := jsoniter.Marshal(lines)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

// TextLines returns the text split into individual lines
func (chunk *Chunk) TextLines() []string {
	if chunk.Text == "" {
		return []string{}
	}

	// Handle different line ending styles:
	// \r\n (Windows), \n (Unix/Linux), \r (Mac Classic)
	text := chunk.Text

	// First, normalize \r\n to \n
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Then, replace remaining \r with \n
	text = strings.ReplaceAll(text, "\r", "\n")

	// Finally, split by \n
	return strings.Split(text, "\n")
}

// TextLinesJSON returns the text split into individual lines as JSON
func (chunk *Chunk) TextLinesJSON() (string, error) {
	lines := chunk.TextLines()
	json, err := jsoniter.Marshal(lines)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

// TextLinesToWChars returns the text split into lines, then each line split into individual UTF-8 wide characters
func (chunk *Chunk) TextLinesToWChars() [][]string {
	if chunk.Text == "" {
		return [][]string{}
	}

	lines := chunk.TextLines()
	result := make([][]string, len(lines))

	for i, line := range lines {
		if line == "" {
			result[i] = []string{}
			continue
		}

		var slices []string
		text := line
		for len(text) > 0 {
			r, size := utf8.DecodeRuneInString(text)
			if r == utf8.RuneError {
				// Skip invalid UTF-8 sequences
				text = text[1:]
				continue
			}
			slices = append(slices, string(r))
			text = text[size:]
		}
		result[i] = slices
	}

	return result
}

// TextLinesToWCharsJSON returns the text split into lines and wide characters as JSON
func (chunk *Chunk) TextLinesToWCharsJSON() (string, error) {
	linesSlices := chunk.TextLinesToWChars()
	json, err := jsoniter.Marshal(linesSlices)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

// ValidatePositions validates the positions against the characters
func ValidatePositions(chars []string, positions []Position) error {
	if len(chars) == 0 {
		return nil
	}

	if len(positions) == 0 {
		return nil
	}

	for idx, pos := range positions {
		if pos.StartPos < 0 {
			return fmt.Errorf("position %d has negative StartPos (%d)", idx, pos.StartPos)
		}
		if pos.EndPos < 0 {
			return fmt.Errorf("position %d has negative EndPos (%d)", idx, pos.EndPos)
		}
		if pos.StartPos >= pos.EndPos {
			return fmt.Errorf("position %d has StartPos (%d) >= EndPos (%d)", idx, pos.StartPos, pos.EndPos)
		}

		// Check if the position is out of bounds
		if pos.StartPos >= len(chars) {
			return fmt.Errorf("position %d has StartPos (%d) out of bounds (%d)", idx, pos.StartPos, len(chars))
		}
		if pos.EndPos > len(chars) {
			return fmt.Errorf("position %d has EndPos (%d) out of bounds (%d)", idx, pos.EndPos, len(chars))
		}
	}
	return nil
}

// Split splits the chunk into multiple sub-chunks based on the given positions
// It performs bounds checking and logs warnings for out-of-bounds positions
// Returns chunks with proper cascading relationships set up
func (chunk *Chunk) Split(chars []string, positions []Position) []*Chunk {
	if chunk == nil {
		log.Warn("Split called on nil chunk")
		return []*Chunk{}
	}

	if chunk.Text == "" {
		log.Warn("Split called on chunk with empty text, chunk ID: %s", chunk.ID)
		return []*Chunk{}
	}

	if len(positions) == 0 {
		log.Warn("Split called with empty positions array, chunk ID: %s", chunk.ID)
		return []*Chunk{}
	}

	if len(chars) == 0 {
		log.Warn("Split called with empty chars array, chunk ID: %s", chunk.ID)
		return []*Chunk{}
	}

	charsLen := len(chars)
	var validPositions []Position

	// Filter valid positions and log warnings for invalid ones
	for i, pos := range positions {
		if pos.StartPos < 0 {
			log.Warn("Position %d has negative StartPos (%d), ignoring. Chunk ID: %s", i, pos.StartPos, chunk.ID)
			continue
		}
		if pos.EndPos < 0 {
			log.Warn("Position %d has negative EndPos (%d), ignoring. Chunk ID: %s", i, pos.EndPos, chunk.ID)
			continue
		}
		if pos.StartPos >= pos.EndPos {
			log.Warn("Position %d has StartPos (%d) >= EndPos (%d), ignoring. Chunk ID: %s", i, pos.StartPos, pos.EndPos, chunk.ID)
			continue
		}
		if pos.StartPos >= charsLen {
			log.Warn("Position %d StartPos (%d) exceeds chars length (%d), ignoring. Chunk ID: %s", i, pos.StartPos, charsLen, chunk.ID)
			continue
		}
		if pos.EndPos > charsLen {
			log.Warn("Position %d EndPos (%d) exceeds chars length (%d), clamping to chars length. Chunk ID: %s", i, pos.EndPos, charsLen, chunk.ID)
			pos.EndPos = charsLen
		}

		validPositions = append(validPositions, pos)
	}

	if len(validPositions) == 0 {
		log.Warn("No valid positions found after filtering, chunk ID: %s", chunk.ID)
		return []*Chunk{}
	}

	// Create sub-chunks from valid positions
	var subChunks []*Chunk
	for i, pos := range validPositions {
		// Extract text from chars array based on position indices
		chunkChars := chars[pos.StartPos:pos.EndPos]
		chunkText := strings.Join(chunkChars, "")

		// Skip empty chunks
		if strings.TrimSpace(chunkText) == "" {
			log.Debug("Skipping empty chunk at position %d (%d-%d), chunk ID: %s", i, pos.StartPos, pos.EndPos, chunk.ID)
			continue
		}

		// Calculate byte positions in original text for TextPosition
		// We need to find where these characters start and end in the original text
		var textPos *TextPosition
		if chunk.TextPos != nil {
			// Calculate byte positions by counting bytes up to the character positions
			byteStartPos := 0
			byteEndPos := 0

			// Count bytes up to StartPos
			for charIdx := 0; charIdx < pos.StartPos && charIdx < len(chars); charIdx++ {
				byteStartPos += len(chars[charIdx])
			}

			// Count bytes up to EndPos
			for charIdx := 0; charIdx < pos.EndPos && charIdx < len(chars); charIdx++ {
				byteEndPos += len(chars[charIdx])
			}

			// Calculate line numbers for the extracted text
			textBeforeStart := ""
			if byteStartPos < len(chunk.Text) {
				textBeforeStart = chunk.Text[:byteStartPos]
			}
			textBeforeEnd := ""
			if byteEndPos <= len(chunk.Text) {
				textBeforeEnd = chunk.Text[:byteEndPos]
			} else {
				textBeforeEnd = chunk.Text
			}

			startLine := chunk.TextPos.StartLine + strings.Count(textBeforeStart, "\n")
			endLine := chunk.TextPos.StartLine + strings.Count(textBeforeEnd, "\n")

			textPos = &TextPosition{
				StartIndex: chunk.TextPos.StartIndex + byteStartPos,
				EndIndex:   chunk.TextPos.StartIndex + byteEndPos,
				StartLine:  startLine,
				EndLine:    endLine,
			}
		}

		// Create new sub-chunk with cascading relationship
		subChunk := &Chunk{
			ID:       uuid.NewString(),
			Text:     chunkText,
			Type:     chunk.Type,
			ParentID: chunk.ID,                // Set parent relationship
			Depth:    chunk.Depth + 1,         // Increase depth (restored for generic Split method)
			Leaf:     true,                    // Sub-chunks are leaf nodes by default
			Root:     false,                   // Sub-chunks are not root
			Index:    i,                       // Index within this split operation
			Status:   ChunkingStatusCompleted, // Sub-chunks are completed
			TextPos:  textPos,
			MediaPos: chunk.MediaPos, // Inherit media position if any
		}

		// Set up parent chain
		subChunk.Parents = make([]Chunk, len(chunk.Parents)+1)
		copy(subChunk.Parents, chunk.Parents)
		subChunk.Parents[len(chunk.Parents)] = *chunk // Add current chunk as immediate parent

		subChunks = append(subChunks, subChunk)
	}

	// Update original chunk to no longer be a leaf since it now has children
	if len(subChunks) > 0 {
		chunk.Leaf = false
	}

	log.Debug("Successfully split chunk %s into %d sub-chunks", chunk.ID, len(subChunks))
	return subChunks
}

// CalculateTextPos calculates and sets the TextPos field for the chunk
// based on its text content and optional parent position information
func (chunk *Chunk) CalculateTextPos(parentPos *TextPosition, offsetInParent int) {
	if chunk == nil || chunk.Text == "" {
		chunk.TextPos = nil
		return
	}

	textLen := len(chunk.Text)

	// Count lines in the chunk text
	lines := strings.Split(chunk.Text, "\n")
	lineCount := len(lines)

	var startIndex, endIndex int
	var startLine, endLine int

	if parentPos != nil {
		// Calculate positions relative to parent
		startIndex = parentPos.StartIndex + offsetInParent
		endIndex = startIndex + textLen

		// Calculate line numbers by counting newlines before the offset
		if offsetInParent > 0 && len(chunk.Parents) > 0 {
			// Get parent text to count newlines before this chunk
			parentChunk := &chunk.Parents[len(chunk.Parents)-1]
			if offsetInParent <= len(parentChunk.Text) {
				textBeforeChunk := parentChunk.Text[:offsetInParent]
				newlinesBeforeChunk := strings.Count(textBeforeChunk, "\n")
				startLine = parentPos.StartLine + newlinesBeforeChunk
			} else {
				startLine = parentPos.StartLine
			}
		} else {
			startLine = parentPos.StartLine
		}

		// Calculate end line
		if lineCount > 1 {
			endLine = startLine + lineCount - 1
		} else {
			endLine = startLine
		}
	} else {
		// Calculate positions as if this is the root chunk
		startIndex = 0
		endIndex = textLen
		startLine = 1

		if lineCount > 1 {
			endLine = lineCount
		} else {
			endLine = 1
		}
	}

	chunk.TextPos = &TextPosition{
		StartIndex: startIndex,
		EndIndex:   endIndex,
		StartLine:  startLine,
		EndLine:    endLine,
	}
}

// UpdateTextPosFromText updates the TextPos based on the current text content
// This is useful when the chunk text has been modified
func (chunk *Chunk) UpdateTextPosFromText() {
	if chunk == nil {
		return
	}

	// If there's an existing TextPos, preserve the StartIndex and StartLine
	// and update EndIndex and EndLine based on current text
	if chunk.TextPos != nil {
		existingStartIndex := chunk.TextPos.StartIndex
		existingStartLine := chunk.TextPos.StartLine

		textLen := len(chunk.Text)
		lines := strings.Split(chunk.Text, "\n")
		lineCount := len(lines)

		chunk.TextPos.EndIndex = existingStartIndex + textLen
		if lineCount > 1 {
			chunk.TextPos.EndLine = existingStartLine + lineCount - 1
		} else {
			chunk.TextPos.EndLine = existingStartLine
		}
	} else {
		// Calculate from scratch if no existing TextPos
		chunk.CalculateTextPos(nil, 0)
	}
}

// CalculateRelativeTextPos calculates TextPos for a substring within this chunk
// startOffset and endOffset are byte positions within chunk.Text
func (chunk *Chunk) CalculateRelativeTextPos(startOffset, endOffset int) *TextPosition {
	if chunk == nil || chunk.Text == "" || chunk.TextPos == nil {
		return nil
	}

	textLen := len(chunk.Text)
	if startOffset < 0 || endOffset < 0 || startOffset >= endOffset || startOffset >= textLen {
		log.Warn("Invalid offsets for CalculateRelativeTextPos: start=%d, end=%d, textLen=%d", startOffset, endOffset, textLen)
		return nil
	}

	// Clamp endOffset if it exceeds text length
	if endOffset > textLen {
		endOffset = textLen
	}

	// Calculate absolute positions
	absoluteStartIndex := chunk.TextPos.StartIndex + startOffset
	absoluteEndIndex := chunk.TextPos.StartIndex + endOffset

	// Calculate line numbers by counting newlines
	textBeforeStart := chunk.Text[:startOffset]
	textBeforeEnd := chunk.Text[:endOffset]

	newlinesBeforeStart := strings.Count(textBeforeStart, "\n")
	newlinesBeforeEnd := strings.Count(textBeforeEnd, "\n")

	startLine := chunk.TextPos.StartLine + newlinesBeforeStart
	endLine := chunk.TextPos.StartLine + newlinesBeforeEnd

	return &TextPosition{
		StartIndex: absoluteStartIndex,
		EndIndex:   absoluteEndIndex,
		StartLine:  startLine,
		EndLine:    endLine,
	}
}

// GetTextAtPosition extracts text content at the specified position
// Returns empty string if position is invalid or outside chunk bounds
func (chunk *Chunk) GetTextAtPosition(pos *TextPosition) string {
	if chunk == nil || chunk.Text == "" || pos == nil || chunk.TextPos == nil {
		return ""
	}

	// Check if the requested position is within this chunk's bounds
	if pos.StartIndex < chunk.TextPos.StartIndex || pos.EndIndex > chunk.TextPos.EndIndex {
		log.Warn("Position (%d-%d) is outside chunk bounds (%d-%d)", pos.StartIndex, pos.EndIndex, chunk.TextPos.StartIndex, chunk.TextPos.EndIndex)
		return ""
	}

	// Calculate relative offsets within this chunk
	relativeStart := pos.StartIndex - chunk.TextPos.StartIndex
	relativeEnd := pos.EndIndex - chunk.TextPos.StartIndex

	if relativeStart < 0 || relativeEnd > len(chunk.Text) || relativeStart >= relativeEnd {
		return ""
	}

	return chunk.Text[relativeStart:relativeEnd]
}
