# Lattice MVP - Learning-to-Credibility Pipeline

> Transform every piece of content you consume into personal knowledge, quiz questions, and marketing assets.

## What is Lattice?

Lattice solves the problem of **zero ROI on learning time**. Right now when you watch valuable content:
- ❌ Zero content output (time spent = no tangible assets)
- ❌ No proof of expertise (you understand it, but prospects can't see that)
- ❌ No ROI on learning time (consumed but not leveraged)

**Lattice transforms learning into leverage:**

1. **YouTube Video** → Submit URL to Lattice
2. **Concept Extraction** → AI extracts 3-7 core learnable concepts
3. **Quiz Generation** → Tests your understanding (spaced repetition)
4. **Content Generation** → LinkedIn case studies + X threads + blog tutorials
5. **Publish & Profit** → Prospects see expertise → Trust transferred → Deals closed

## Prerequisites

### Required
- **Go 1.25+** - [Install Go](https://golang.org/dl/)
- **PostgreSQL** - [Install Postgres](https://www.postgresql.org/download/)
- **yt-dlp** - For YouTube transcript extraction
  ```bash
  brew install yt-dlp
  ```
- **Claude API Key** - Get from [Anthropic Console](https://console.anthropic.com/)

### Recommended
- **pgAdmin** or **psql** - For database management
- **Postman** or **curl** - For API testing

## Quick Start

### 1. Clone and Install

```bash
cd brain-api
go mod download
```

### 2. Database Setup

```bash
# Create database
createdb brain

# Or using psql
psql postgres
CREATE DATABASE brain;
\q
```

### 3. Environment Configuration

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your values
nano .env
```

**Required environment variables:**

```bash
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/brain?sslmode=disable

# Claude API (get from https://console.anthropic.com/)
CLAUDE_API_KEY=sk-ant-your-api-key-here
CLAUDE_MODEL=claude-sonnet-4-5-20250929

# Server
PORT=8080
ENV=development
CORS_ORIGIN=http://localhost:3000

# YouTube (optional - will auto-detect yt-dlp)
YTDLP_PATH=/opt/homebrew/bin/yt-dlp

# Concept Extraction (optional)
CONCEPTS_MIN=3
CONCEPTS_MAX=7
```

### 4. Run the Server

```bash
cd brain-api
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

**Expected output:**
```
Database connection established
Applied migration: 001_initial_schema.sql
All migrations applied successfully
Starting Brain API server on port 8080...
```

## API Endpoints

### Source Content (Main Pipeline)

#### **POST /api/source-content** - Process YouTube Video
Runs the full pipeline: transcript → concepts → quizzes → content generation

**Request:**
```bash
curl -X POST http://localhost:8080/api/source-content \
  -H "Content-Type: application/json" \
  -d '{
    "type": "youtube",
    "url": "https://www.youtube.com/watch?v=Yr9O6KFwbW4"
  }'
```

**Response (201 Created):**
```json
{
  "source_content": {
    "id": 1,
    "type": "youtube",
    "url": "https://www.youtube.com/watch?v=Yr9O6KFwbW4",
    "title": "RALF loops in Claude - prompt engineering",
    "transcript": "Full transcript text...",
    "processed_at": "2026-01-17T10:30:00Z",
    "created_at": "2026-01-17T10:30:00Z"
  },
  "concepts": [
    {
      "id": 1,
      "title": "RALF Loop Pattern",
      "description": "A prompt engineering technique involving Research, Apply, Learn, and Feedback in iterative cycles...",
      "source_content_id": 1
    }
  ],
  "quizzes": [
    {
      "id": 1,
      "concept_id": 1,
      "question": "What is the primary benefit of using RALF loops?",
      "option_a": "Faster execution",
      "option_b": "Iterative refinement of AI outputs",
      "option_c": "Reduced token usage",
      "option_d": "Simplified prompts",
      "correct_answer": "B",
      "explanation": "RALF loops enable iterative refinement..."
    }
  ],
  "generated_content": [
    {
      "id": 1,
      "platform": "linkedin",
      "title": "How I Used RALF Loops to Improve Client AI Systems",
      "body": "Last week, a client came to me...",
      "concept_ids": [1, 2, 3],
      "status": "draft"
    },
    {
      "id": 2,
      "platform": "twitter",
      "title": "Why RALF loops matter for AI consulting",
      "body": "1/ Most teams struggle with AI prompts...",
      "concept_ids": [1, 2],
      "status": "draft"
    },
    {
      "id": 3,
      "platform": "blog",
      "title": "Complete Guide to RALF Loops in AI Development",
      "body": "If you're building AI systems...",
      "concept_ids": [1, 2, 3],
      "status": "draft"
    }
  ]
}
```

#### **GET /api/source-content** - List All Content
```bash
curl http://localhost:8080/api/source-content
```

#### **GET /api/source-content/:id** - Get Specific Content
```bash
curl http://localhost:8080/api/source-content/1
```

#### **GET /api/source-content/:id/concepts** - Get Concepts
```bash
curl http://localhost:8080/api/source-content/1/concepts
```

#### **GET /api/source-content/:id/quizzes** - Get Quizzes
```bash
curl http://localhost:8080/api/source-content/1/quizzes
```

#### **GET /api/source-content/:id/content** - Get Generated Content
```bash
curl http://localhost:8080/api/source-content/1/content
```

#### **DELETE /api/source-content/:id** - Delete Content
```bash
curl -X DELETE http://localhost:8080/api/source-content/1
```

### Concepts (Direct Management)

#### **GET /api/concepts** - List All Concepts
```bash
curl http://localhost:8080/api/concepts
```

#### **POST /api/concepts** - Create Concept
```bash
curl -X POST http://localhost:8080/api/concepts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "RALF Loop Pattern",
    "description": "A prompt engineering technique...",
    "source_content_id": 1
  }'
```

#### **PATCH /api/concepts/:id** - Update Concept
```bash
curl -X PATCH http://localhost:8080/api/concepts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Advanced RALF Loop Pattern",
    "description": "Updated description..."
  }'
```

#### **DELETE /api/concepts/:id** - Delete Concept
```bash
curl -X DELETE http://localhost:8080/api/concepts/1
```

### Health Check

```bash
curl http://localhost:8080/api/health
```

## How It Works

### 1. YouTube Transcript Extraction
- Uses `yt-dlp` to fetch video metadata and subtitles
- Supports auto-generated captions and manual subtitles
- Parses VTT, SRT, and JSON3 subtitle formats
- Cleans up timestamps and formatting artifacts

### 2. Concept Extraction (Claude AI)
- Sends transcript to Claude with expert educator prompt
- Extracts 3-7 core learnable concepts
- Each concept has:
  - **Title**: Clear, concise name (max 100 chars)
  - **Description**: Detailed explanation (2-4 sentences)
- Focuses on fundamental ideas, actionable techniques, key mental models

### 3. Quiz Generation (Claude AI)
- Generates 2-3 quiz questions per concept
- Questions test understanding and application (not just recall)
- Each question includes:
  - 4 plausible options (A, B, C, D)
  - Correct answer
  - Detailed explanation
- Designed for spaced repetition learning

### 4. Content Generation (Claude AI)
- Creates 3 platform-specific content pieces:

**LinkedIn (Case Study Format):**
- Hook: Client problem
- Body: How you used concepts to solve it
- Result: Measurable outcome
- CTA: Invite discussion
- Tone: Professional, credible, approachable
- Length: 1200-1500 characters

**X/Twitter (Thread Format):**
- 5 tweets total
- Tweet 1: Hook - why this matters
- Tweets 2-4: Key insights from concepts
- Tweet 5: Actionable takeaway + CTA
- Tone: Casual but authoritative
- Length: <280 chars per tweet

**Blog (Tutorial Format):**
- Introduction: Why this matters
- Sections: One per concept (explanation + how to apply)
- Conclusion: Summary + next steps
- Tone: Teaching, detailed, actionable
- Length: 800-1200 words
- Uses Markdown formatting

## Database Schema

### Tables

- **source_contents** - Original YouTube videos, PDFs, articles
- **concepts** - Learnable units extracted from content
- **quiz_questions** - Generated quiz questions for concepts
- **quiz_attempts** - User answers tracking (future)
- **learning_progress** - Spaced repetition tracking (future)
- **generated_contents** - Marketing content (LinkedIn, X, blog)
- **concept_relationships** - Relationships between concepts (future)
- **publishing_events** - Publishing history (future)

### Relationships

```
source_contents (1) ──< (many) concepts
concepts (1) ──< (many) quiz_questions
concepts (1) ──< (many) learning_progress
concepts (many) ──< (many) generated_contents (via JSONB array)
```

## Project Structure

```
brain-api/
├── cmd/
│   └── server/
│       └── main.go              # Server entry point
├── internal/
│   ├── db/
│   │   ├── postgres.go          # Database connection
│   │   ├── concept_repo.go      # Concept database operations
│   │   ├── source_content_repo.go
│   │   ├── quiz_repo.go
│   │   ├── generated_content_repo.go
│   │   └── migrations/
│   │       └── 001_initial_schema.sql
│   ├── handlers/
│   │   ├── concept_handler.go   # HTTP handlers for concepts
│   │   └── source_content_handler.go
│   ├── middleware/
│   │   └── cors.go              # CORS middleware
│   ├── models/
│   │   ├── concept.go           # Data models
│   │   ├── source_content.go
│   │   ├── quiz.go
│   │   └── generated_content.go
│   └── services/
│       ├── claude_service.go    # Claude AI integration
│       └── source_content_service.go # Orchestration
├── pkg/
│   ├── claude/
│   │   ├── client.go            # Claude API client
│   │   └── errors.go
│   └── youtube/
│       ├── client.go            # YouTube transcript fetching
│       ├── subtitle_parser.go   # Parse VTT/SRT/JSON3
│       ├── models.go
│       └── errors.go
├── .env.example                 # Environment template
├── go.mod
├── go.sum
└── README.md
```

## Architecture

### Service Layer Architecture

```
Handler Layer (HTTP)
    ↓
Service Layer (Business Logic)
    ├── SourceContentService (orchestration)
    ├── ClaudeService (AI interactions)
    └── YouTubeClient (transcript fetching)
    ↓
Repository Layer (Database)
    ├── SourceContentRepo
    ├── ConceptRepo
    ├── QuizRepo
    └── GeneratedContentRepo
    ↓
External Integrations
    ├── pkg/youtube (yt-dlp wrapper)
    └── pkg/claude (Anthropic API client)
```

### Processing Flow

```
YouTube URL
    ↓
1. Duplicate Check (GetSourceContentByURL)
   ├─ If exists → Return existing data
   └─ If new → Continue
    ↓
2. Fetch Transcript (yt-dlp)
   └─ GetVideoInfo() → metadata + transcript
    ↓
3. Save Source Content
   └─ CreateSourceContent()
    ↓
4. Extract Concepts (Claude AI)
   └─ ExtractConcepts() → 3-7 concepts
    ↓
5. Save Concepts
   └─ CreateConceptsBatch()
    ↓
6. Generate Quizzes (Claude AI)
   └─ For each concept: GenerateQuiz() → 2-3 questions
    ↓
7. Save Quizzes
   └─ CreateQuizBatch()
    ↓
8. Generate Content (Claude AI)
   └─ For each platform: GenerateContent()
       ├─ LinkedIn (case study)
       ├─ Twitter (thread)
       └─ Blog (tutorial)
    ↓
9. Save Generated Content
   └─ CreateGeneratedContentBatch()
    ↓
10. Return Complete Result
```

## Error Handling

### Partial Failure Strategy

The system is designed to handle partial failures gracefully:

- **If YouTube fails** → Return error (can't proceed without transcript)
- **If concept extraction fails** → Save source content, return warning
- **If quiz generation fails** → Save concepts, skip quizzes for that concept
- **If content generation fails** → Save everything else, skip that platform

This ensures you always get **some** value even if parts fail.

### Common Errors

**"yt-dlp not found"**
```bash
brew install yt-dlp
```

**"No transcript available"**
- Video has no auto-captions or manual subtitles
- Try a different video or enable captions on YouTube

**"Claude API error"**
- Check CLAUDE_API_KEY in .env
- Verify API key at https://console.anthropic.com/
- Check for rate limits

**"Database connection failed"**
- Verify PostgreSQL is running: `pg_isready`
- Check DATABASE_URL in .env
- Create database: `createdb brain`

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o brain-server cmd/server/main.go
./brain-server
```

### Code Organization

- **pkg/** - Reusable packages (YouTube, Claude clients)
- **internal/** - Application-specific code
- **cmd/** - Application entry points
- **Handlers** - HTTP request handling
- **Services** - Business logic
- **Repositories** - Database operations
- **Models** - Data structures

## Future Enhancements

### Phase 2: Learning Features
- [ ] Quiz tracking with spaced repetition
- [ ] Mastery level calculation
- [ ] Next review scheduling
- [ ] Learning progress analytics

### Phase 3: Publishing Integration
- [ ] One-click publish to LinkedIn
- [ ] One-click publish to X/Twitter
- [ ] Blog platform integration (Medium, Dev.to)
- [ ] Content scheduling

### Phase 4: Knowledge Graph
- [ ] Visualize concept relationships
- [ ] Detect prerequisite concepts
- [ ] Suggest related content
- [ ] Learning path generation

### Phase 5: Multi-Format Support
- [ ] PDF document processing
- [ ] Article URL scraping
- [ ] Podcast transcript support
- [ ] Browser extension for one-click processing

### Phase 6: Analytics & ROI
- [ ] Track engagement metrics (likes, shares, comments)
- [ ] Content performance dashboard
- [ ] ROI calculation (learning time → content output → deals closed)
- [ ] A/B testing for content variations

## Success Metrics

**MVP Success** = You can:
1. ✅ Submit a YouTube URL
2. ✅ Get back 3-7 solid concepts you didn't know before
3. ✅ Quiz yourself and actually learn them
4. ✅ Get 3 content drafts (LinkedIn, X, blog) that sound credible
5. ⏳ Edit and publish at least one piece
6. ⏳ Prospect sees it and thinks "this person knows AI"

**Long-term Success** =
- Every video watched → 3+ marketing assets created
- Consistent content output (3-5 posts/week from learning)
- Prospects mention your content in sales calls
- "Transfer of trust" working → more consulting deals

## License

MIT

## Support

For issues or questions:
- Create an issue on GitHub
- Email: support@lattice.ai

---

**Built with:**
- Go 1.25.6
- PostgreSQL
- Claude AI (Anthropic)
- yt-dlp
- Gin Web Framework
