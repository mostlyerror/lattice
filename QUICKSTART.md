# Lattice MVP - Quick Start Guide

## ✅ What's Already Done

All implementation is complete! Here's what's been built:

### Backend Infrastructure ✅
- [x] YouTube transcript fetching (yt-dlp integration)
- [x] Claude API client (Anthropic integration)
- [x] Concept extraction service
- [x] Quiz generation service
- [x] Content generation service (LinkedIn, Twitter, Blog)
- [x] Orchestration service (coordinates full pipeline)
- [x] Database repositories (Postgres)
- [x] HTTP API handlers
- [x] REST API endpoints
- [x] Database migrations
- [x] Environment configuration

### System Prerequisites ✅
- [x] PostgreSQL running (`/tmp:5432 - accepting connections`)
- [x] Database 'lattice' created
- [x] yt-dlp installed (`/opt/homebrew/bin/yt-dlp`)
- [x] .env file exists
- [x] Code compiles successfully

## 🔧 What You Need to Do

### 1. Add Your Claude API Key

**Get your API key:**
1. Go to https://console.anthropic.com/
2. Sign up or log in
3. Navigate to API Keys
4. Create a new API key (starts with `sk-ant-`)

**Add it to your .env file:**
```bash
# Edit the .env file
nano .env

# Find this line:
CLAUDE_API_KEY=your_claude_api_key_here

# Replace with your actual key:
CLAUDE_API_KEY=sk-ant-api03-xxxxxxxxxxxxxxxxxxxxx

# Save and exit (Ctrl+X, then Y, then Enter)
```

**Verify your DATABASE_URL is correct:**
```bash
# In .env, make sure this matches your PostgreSQL setup:
DATABASE_URL=postgresql://user:password@localhost:5432/lattice?sslmode=disable

# Replace 'user' and 'password' with your PostgreSQL credentials
# Common defaults:
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/lattice?sslmode=disable
```

### 2. Start the Server

```bash
# From the root directory
go run cmd/server/main.go
```

**Expected output:**
```
Database connection established
Applied migration: 001_initial_schema.sql
All migrations applied successfully
Starting Lattice API server on port 8080...
[GIN-debug] POST   /api/source-content       --> ...
[GIN-debug] GET    /api/source-content       --> ...
[GIN-debug] GET    /api/source-content/:id   --> ...
...
[GIN-debug] Listening and serving HTTP on :8080
```

### 3. Test the Full Pipeline

**In a new terminal, run this test:**

```bash
# Test with the RALF loops video
curl -X POST http://localhost:8080/api/source-content \
  -H "Content-Type: application/json" \
  -d '{
    "type": "youtube",
    "url": "https://www.youtube.com/watch?v=Yr9O6KFwbW4"
  }' | jq '.'
```

> **Note:** If you don't have `jq` installed, you can pipe to `json_pp` instead or just remove `| jq '.'`

**What should happen:**

1. **YouTube Phase** (10-15 seconds)
   - Fetches video metadata
   - Downloads and parses transcript
   - Saves to database

2. **Claude AI Phase** (30-60 seconds)
   - Extracts 3-7 concepts
   - Generates 2-3 quiz questions per concept
   - Creates LinkedIn case study
   - Creates Twitter thread
   - Creates blog tutorial

3. **Response** (JSON)
   - Full source content with transcript
   - All extracted concepts
   - All quiz questions
   - All generated content pieces

**Expected response structure:**
```json
{
  "source_content": {
    "id": 1,
    "type": "youtube",
    "url": "https://www.youtube.com/watch?v=Yr9O6KFwbW4",
    "title": "RALF loops in Claude - prompt engineering",
    "transcript": "...",
    "processed_at": "2026-01-17T...",
    "created_at": "2026-01-17T..."
  },
  "concepts": [
    {
      "id": 1,
      "title": "RALF Loop Pattern",
      "description": "A prompt engineering technique...",
      "source_content_id": 1
    }
    // ... more concepts
  ],
  "quizzes": [
    {
      "id": 1,
      "concept_id": 1,
      "question": "What is the primary benefit of using RALF loops?",
      "option_a": "Faster execution",
      "option_b": "Iterative refinement",
      "option_c": "Reduced tokens",
      "option_d": "Simplified prompts",
      "correct_answer": "B",
      "explanation": "..."
    }
    // ... more quizzes
  ],
  "generated_content": [
    {
      "id": 1,
      "platform": "linkedin",
      "title": "How I Used RALF Loops...",
      "body": "Last week, a client came to me...",
      "concept_ids": [1, 2, 3],
      "status": "draft"
    },
    {
      "id": 2,
      "platform": "twitter",
      "title": "Why RALF loops matter...",
      "body": "1/ Most teams struggle...",
      "concept_ids": [1, 2],
      "status": "draft"
    },
    {
      "id": 3,
      "platform": "blog",
      "title": "Complete Guide to RALF Loops...",
      "body": "If you're building AI systems...",
      "concept_ids": [1, 2, 3],
      "status": "draft"
    }
  ]
}
```

## 🎯 What to Try Next

### Test Different Videos

Try shorter videos first to verify the system works:

```bash
# Shorter video for faster testing
curl -X POST http://localhost:8080/api/source-content \
  -H "Content-Type: application/json" \
  -d '{
    "type": "youtube",
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  }'
```

### Query Your Data

```bash
# List all processed content
curl http://localhost:8080/api/source-content | jq '.source_contents'

# Get concepts for source content ID 1
curl http://localhost:8080/api/source-content/1/concepts | jq '.concepts'

# Get quizzes
curl http://localhost:8080/api/source-content/1/quizzes | jq '.quizzes'

# Get generated marketing content
curl http://localhost:8080/api/source-content/1/content | jq '.generated_content'
```

### Duplicate URL Handling

```bash
# Submit the same URL again - should return cached data instantly
curl -X POST http://localhost:8080/api/source-content \
  -H "Content-Type: application/json" \
  -d '{
    "type": "youtube",
    "url": "https://www.youtube.com/watch?v=Yr9O6KFwbW4"
  }'
```

## 🐛 Troubleshooting

### "Failed to create Claude client: CLAUDE_API_KEY environment variable is not set"
- Add your Claude API key to `.env`
- Restart the server

### "yt-dlp not found"
```bash
brew install yt-dlp
```

### "No transcript available for this video"
- Video doesn't have auto-captions or manual subtitles
- Try a different video
- Most popular YouTube videos have auto-captions

### "Database connection failed"
```bash
# Check PostgreSQL is running
pg_isready

# If not running:
brew services start postgresql@14
# or
brew services start postgresql

# Verify database exists
psql -l | grep lattice
```

### "Failed to extract concepts"
- Check Claude API key is valid
- Check you have API credits
- Check network connection to Anthropic API

### Server won't start
```bash
# Check if something is using port 8080
lsof -i :8080

# Kill the process if needed
kill -9 <PID>

# Or use a different port in .env
PORT=8081
```

## 📊 Monitoring Logs

The server logs show detailed progress:

```
Processing YouTube URL: https://www.youtube.com/watch?v=Yr9O6KFwbW4
Fetching YouTube video info...
Source content saved with ID: 1
Extracting concepts from transcript...
Saving 5 concepts to database...
Concepts saved successfully
Generating quizzes for concepts...
Saving 12 quizzes to database...
Quizzes saved successfully
Generating marketing content...
Saving 3 generated content pieces to database...
Generated content saved successfully
Processing complete for source content ID: 1
```

## 🎉 Success Criteria

You know it's working when:

1. ✅ Server starts without errors
2. ✅ POST request returns 201 Created
3. ✅ Response includes 3-7 concepts
4. ✅ Each concept has 2-3 quiz questions
5. ✅ You get LinkedIn, Twitter, and Blog drafts
6. ✅ Content sounds credible and demonstrates expertise
7. ✅ Duplicate URL returns cached data immediately

## 📝 Next Steps

Once the system is working:

1. **Test with your own learning content**
   - Submit videos you're actually watching
   - Review the extracted concepts
   - Take the quizzes to test your understanding
   - Edit the generated content to match your voice

2. **Publish your first piece**
   - Pick the best LinkedIn/Twitter/Blog draft
   - Edit it to add your personal touch
   - Publish it on the platform
   - Track engagement

3. **Iterate and improve**
   - Adjust the prompts in `claude_service.go`
   - Experiment with different content types
   - Build a content calendar

## 💡 Tips

- **Start with shorter videos** (5-15 minutes) for faster iteration
- **Watch for quota limits** on Claude API during testing
- **Save good content** - mark drafts you like as published
- **Build a backlog** - process videos faster than you publish
- **Mix content types** - use blog for deep dives, Twitter for quick insights, LinkedIn for case studies

---

**Ready to test?** Just add your Claude API key and run the server!

```bash
# 1. Add API key to .env
nano .env

# 2. Start server
go run cmd/server/main.go

# 3. Test (in new terminal)
curl -X POST http://localhost:8080/api/source-content \
  -H "Content-Type: application/json" \
  -d '{"type": "youtube", "url": "https://www.youtube.com/watch?v=Yr9O6KFwbW4"}'
```
