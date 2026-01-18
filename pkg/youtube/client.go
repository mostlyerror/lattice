package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Client handles YouTube video operations
type Client struct {
	ytdlpPath string
	timeout   time.Duration
	parser    *SubtitleParser
}

// NewClient creates a new YouTube client
func NewClient() (*Client, error) {
	// Try to find yt-dlp in PATH or use env variable
	ytdlpPath := os.Getenv("YTDLP_PATH")
	if ytdlpPath == "" {
		// Try common locations
		for _, path := range []string{
			"/opt/homebrew/bin/yt-dlp",
			"/usr/local/bin/yt-dlp",
			"yt-dlp", // Will search in PATH
		} {
			if _, err := exec.LookPath(path); err == nil {
				ytdlpPath = path
				break
			}
		}
	}

	if ytdlpPath == "" {
		return nil, ErrYTDLPNotFound
	}

	return &Client{
		ytdlpPath: ytdlpPath,
		timeout:   120 * time.Second, // 2 minute timeout
		parser:    NewSubtitleParser(),
	}, nil
}

// ValidateURL checks if a URL is a valid YouTube URL
func ValidateURL(url string) error {
	// Support various YouTube URL formats
	patterns := []string{
		`^https?://(www\.)?youtube\.com/watch\?v=[\w-]+`,
		`^https?://(www\.)?youtu\.be/[\w-]+`,
		`^https?://(www\.)?youtube\.com/embed/[\w-]+`,
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, url)
		if err != nil {
			return fmt.Errorf("regex error: %w", err)
		}
		if matched {
			return nil
		}
	}

	return ErrInvalidURL
}

// GetTranscript fetches and parses the transcript for a YouTube video
func (c *Client) GetTranscript(ctx context.Context, videoURL string) (*Transcript, error) {
	// Validate URL first
	if err := ValidateURL(videoURL); err != nil {
		return nil, err
	}

	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Use yt-dlp to get full video JSON with subtitle information
	cmd := exec.CommandContext(cmdCtx, c.ytdlpPath,
		"--skip-download",
		"--write-auto-subs",
		"--sub-lang", "en",
		"--print-json",
		videoURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()

		if strings.Contains(stderrStr, "Private video") ||
		   strings.Contains(stderrStr, "Video unavailable") ||
		   strings.Contains(stderrStr, "This video is not available") {
			return nil, ErrVideoPrivate
		}

		return nil, fmt.Errorf("%w: %s", ErrCommandFailed, stderrStr)
	}

	// Parse JSON output
	var videoData map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &videoData); err != nil {
		return nil, fmt.Errorf("failed to parse video data: %w", err)
	}

	// Try to find subtitle URL (preferring JSON3 format)
	subtitleURL, subtitleFormat := c.findBestSubtitleURL(videoData)
	if subtitleURL == "" {
		return nil, ErrNoTranscript
	}

	// Download subtitle content
	subtitleData, err := c.downloadSubtitle(cmdCtx, subtitleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download subtitle: %w", err)
	}

	// Parse subtitle based on format
	var text string
	switch subtitleFormat {
	case "json3":
		text, err = c.parser.ParseJSON3(subtitleData)
	case "vtt":
		text, err = c.parser.ParseVTT(subtitleData)
	case "srv1", "srv2", "srv3":
		// YouTube's XML formats - try parsing as JSON3 first, fall back to VTT
		text, err = c.parser.ParseJSON3(subtitleData)
		if err != nil {
			text, err = c.parser.ParseVTT(subtitleData)
		}
	default:
		// Default to VTT parsing
		text, err = c.parser.ParseVTT(subtitleData)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse subtitle: %w", err)
	}

	// Clean up the transcript
	text = c.parser.CleanTranscript(text)

	if text == "" {
		return nil, ErrNoTranscript
	}

	return &Transcript{
		Text:     text,
		Language: "en",
	}, nil
}

// findBestSubtitleURL finds the best subtitle URL from video data
func (c *Client) findBestSubtitleURL(videoData map[string]interface{}) (string, string) {
	// Preference order: json3 > vtt > srv3 > srv2 > srv1
	formatPreference := []string{"json3", "vtt", "srv3", "srv2", "srv1"}

	// Check automatic_captions first (more reliable for most videos)
	if autoCaps, ok := videoData["automatic_captions"].(map[string]interface{}); ok {
		if url, format := c.extractSubtitleURL(autoCaps, formatPreference); url != "" {
			return url, format
		}
	}

	// Fall back to manual subtitles
	if subs, ok := videoData["subtitles"].(map[string]interface{}); ok {
		if url, format := c.extractSubtitleURL(subs, formatPreference); url != "" {
			return url, format
		}
	}

	return "", ""
}

// extractSubtitleURL extracts subtitle URL from subtitle data
func (c *Client) extractSubtitleURL(subsData map[string]interface{}, formatPreference []string) (string, string) {
	// Try to get English subtitles
	if enSubs, ok := subsData["en"].([]interface{}); ok && len(enSubs) > 0 {
		// Try each format in order of preference
		for _, preferredFormat := range formatPreference {
			for _, sub := range enSubs {
				if subInfo, ok := sub.(map[string]interface{}); ok {
					if ext, ok := subInfo["ext"].(string); ok && ext == preferredFormat {
						if url, ok := subInfo["url"].(string); ok {
							return url, preferredFormat
						}
					}
				}
			}
		}

		// If no preferred format found, use first available
		if subInfo, ok := enSubs[0].(map[string]interface{}); ok {
			if url, ok := subInfo["url"].(string); ok {
				format := "unknown"
				if ext, ok := subInfo["ext"].(string); ok {
					format = ext
				}
				return url, format
			}
		}
	}

	return "", ""
}

// downloadSubtitle downloads subtitle content from URL
func (c *Client) downloadSubtitle(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download subtitle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subtitle download failed with status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read subtitle data: %w", err)
	}

	return data, nil
}

// GetVideoMetadata fetches metadata for a YouTube video
func (c *Client) GetVideoMetadata(ctx context.Context, videoURL string) (*Metadata, error) {
	// Validate URL first
	if err := ValidateURL(videoURL); err != nil {
		return nil, err
	}

	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Use yt-dlp to get video info as JSON
	cmd := exec.CommandContext(cmdCtx, c.ytdlpPath,
		"--skip-download",
		"--print-json",
		videoURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()

		if strings.Contains(stderrStr, "Private video") ||
		   strings.Contains(stderrStr, "Video unavailable") ||
		   strings.Contains(stderrStr, "This video is not available") {
			return nil, ErrVideoPrivate
		}

		return nil, fmt.Errorf("%w: %s", ErrCommandFailed, stderrStr)
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse video metadata: %w", err)
	}

	metadata := &Metadata{
		Title:    "",
		Duration: 0,
		Channel:  "",
	}

	if title, ok := result["title"].(string); ok {
		metadata.Title = title
	}

	if duration, ok := result["duration"].(float64); ok {
		metadata.Duration = int(duration)
	}

	if channel, ok := result["channel"].(string); ok {
		metadata.Channel = channel
	} else if uploader, ok := result["uploader"].(string); ok {
		metadata.Channel = uploader
	}

	return metadata, nil
}

// GetVideoInfo fetches both transcript and metadata
func (c *Client) GetVideoInfo(ctx context.Context, videoURL string) (*VideoInfo, error) {
	// Get metadata first (it's more reliable)
	metadata, err := c.GetVideoMetadata(ctx, videoURL)
	if err != nil {
		return nil, err
	}

	// Try to get transcript
	transcript, err := c.GetTranscript(ctx, videoURL)
	if err != nil {
		// If transcript fails, return metadata only
		return &VideoInfo{
			Transcript: nil,
			Metadata:   metadata,
		}, fmt.Errorf("got metadata but transcript failed: %w", err)
	}

	return &VideoInfo{
		Transcript: transcript,
		Metadata:   metadata,
	}, nil
}
