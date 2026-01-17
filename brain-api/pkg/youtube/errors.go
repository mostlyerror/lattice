package youtube

import "errors"

var (
	// ErrInvalidURL is returned when the YouTube URL is invalid
	ErrInvalidURL = errors.New("invalid YouTube URL")

	// ErrNoTranscript is returned when no transcript is available for the video
	ErrNoTranscript = errors.New("no transcript available for this video")

	// ErrVideoPrivate is returned when the video is private or deleted
	ErrVideoPrivate = errors.New("video is private, deleted, or unavailable")

	// ErrYTDLPNotFound is returned when yt-dlp is not installed
	ErrYTDLPNotFound = errors.New("yt-dlp not found - please install with 'brew install yt-dlp'")

	// ErrCommandFailed is returned when yt-dlp command execution fails
	ErrCommandFailed = errors.New("yt-dlp command failed")
)
