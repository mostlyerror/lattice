package youtube

// Transcript represents a YouTube video transcript
type Transcript struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

// Metadata represents YouTube video metadata
type Metadata struct {
	Title    string `json:"title"`
	Duration int    `json:"duration"` // in seconds
	Channel  string `json:"channel"`
}

// VideoInfo contains both transcript and metadata
type VideoInfo struct {
	Transcript *Transcript `json:"transcript"`
	Metadata   *Metadata   `json:"metadata"`
}
