package youtube

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// SubtitleParser handles parsing various subtitle formats
type SubtitleParser struct{}

// NewSubtitleParser creates a new subtitle parser
func NewSubtitleParser() *SubtitleParser {
	return &SubtitleParser{}
}

// ParseJSON3 parses YouTube's JSON3 subtitle format
// JSON3 format looks like:
// {"events": [{"segs": [{"utf8": "text"}], ...}]}
func (p *SubtitleParser) ParseJSON3(data []byte) (string, error) {
	var result struct {
		Events []struct {
			Segs []struct {
				UTF8 string `json:"utf8"`
			} `json:"segs"`
		} `json:"events"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON3: %w", err)
	}

	var text strings.Builder
	for _, event := range result.Events {
		for _, seg := range event.Segs {
			if seg.UTF8 != "" && seg.UTF8 != "\n" {
				text.WriteString(seg.UTF8)
				text.WriteString(" ")
			}
		}
	}

	return strings.TrimSpace(text.String()), nil
}

// ParseSRT parses SRT (SubRip) subtitle format
// SRT format looks like:
// 1
// 00:00:00,000 --> 00:00:02,000
// First subtitle text
//
// 2
// 00:00:02,000 --> 00:00:04,000
// Second subtitle text
func (p *SubtitleParser) ParseSRT(data []byte) (string, error) {
	content := string(data)

	// Split by double newlines (subtitle blocks)
	blocks := strings.Split(content, "\n\n")

	var text strings.Builder
	for _, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")

		// Skip sequence number and timestamp lines
		// The actual text starts from line 3 (index 2)
		if len(lines) >= 3 {
			// Join all lines after the timestamp (in case subtitle spans multiple lines)
			subtitleText := strings.Join(lines[2:], " ")
			text.WriteString(subtitleText)
			text.WriteString(" ")
		}
	}

	return strings.TrimSpace(text.String()), nil
}

// ParseVTT parses WebVTT subtitle format
// VTT format looks like:
// WEBVTT
//
// 00:00:00.000 --> 00:00:02.000
// First subtitle text
//
// 00:00:02.000 --> 00:00:04.000
// Second subtitle text
func (p *SubtitleParser) ParseVTT(data []byte) (string, error) {
	content := string(data)

	// Remove WEBVTT header
	content = regexp.MustCompile(`(?i)^WEBVTT[^\n]*\n`).ReplaceAllString(content, "")

	// Remove styling/note blocks (between double colons or NOTE)
	content = regexp.MustCompile(`(?m)^NOTE.*$`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`(?s)STYLE\s*\n.*?\n\n`).ReplaceAllString(content, "")

	// Split by double newlines
	blocks := strings.Split(content, "\n\n")

	var text strings.Builder
	for _, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")

		// Skip timestamp lines (contain -->)
		for _, line := range lines {
			if !strings.Contains(line, "-->") && strings.TrimSpace(line) != "" {
				// Remove VTT tags like <c>, <v>, etc.
				line = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(line, "")
				text.WriteString(strings.TrimSpace(line))
				text.WriteString(" ")
			}
		}
	}

	return strings.TrimSpace(text.String()), nil
}

// CleanTranscript removes duplicate words and extra whitespace
func (p *SubtitleParser) CleanTranscript(text string) string {
	// Remove excessive whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Remove common subtitle artifacts
	text = strings.ReplaceAll(text, "[Music]", "")
	text = strings.ReplaceAll(text, "[Applause]", "")
	text = strings.ReplaceAll(text, "[Laughter]", "")

	return strings.TrimSpace(text)
}
