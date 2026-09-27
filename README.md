# Shalom

A hybrid adaptive communication platform that combines features from Google Meet, Telegram, Zoom, Discord, and WhatsApp — optimized for 500+ participants on low-connectivity networks.

## Features

- **Video Conferencing** — WebRTC SFU via LiveKit with adaptive bitrate
- **500+ Participants** — Speaker stage, role-based permissions (host, cohost, speaker, listener, moderator, guest)
- **Bandwidth Optimization** — Auto-switches between Ultra Saving (8-15 MB/hr) → HD (500-800 MB/hr)
- **Telegram-style Chat** — Text, files, voice messages, polls, reactions, reply threads, pinned messages
- **Connection Recovery** — Auto-reconnect within 1-5s, offline message sync
- **Meeting Links & Codes** — Shareable join links, 8-char join codes
- **Network Adaptive** — Detects 2G/3G/4G/5G and adjusts quality automatically

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React + Vite + Tailwind CSS |
| Backend | Go (Gin) |
| WebRTC | LiveKit (SFU) |
| Database | PostgreSQL + MongoDB + Redis |
| Deployment | Docker Compose / Kubernetes |

## Quick Start

### With Docker Compose (recommended)

```bash
# Start all services
docker compose up -d

# Run database migrations
docker compose exec postgres psql -U shalom -d shalom -f /docker-entrypoint-initdb.d/001_init.sql

# Access the app
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# LiveKit: ws://localhost:7880
# Grafana: http://localhost:3001
```

### Manual Setup

```bash
# Start PostgreSQL, MongoDB, Redis, LiveKit (see docker-compose.yml)

# Backend
cd backend
go mod tidy
go run ./cmd/api-gateway

# Frontend
cd frontend
npm install
npm run dev
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Backend port | `8080` |
| `POSTGRES_URL` | PostgreSQL connection string | `postgres://shalom:secret@localhost:5432/shalom` |
| `MONGO_URL` | MongoDB connection string | `mongodb://localhost:27017` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `JWT_SECRET` | JWT signing secret | Change in production |
| `LIVEKIT_URL` | LiveKit WebSocket URL | `ws://localhost:7880` |
| `LIVEKIT_API_KEY` | LiveKit API key | `devkey` |
| `LIVEKIT_API_SECRET` | LiveKit API secret | `devsecret` |

## Bandwidth Modes

| Mode | Quality | Data/Hour |
|------|---------|-----------|
| Ultra Saving | Audio only | 8-15 MB |
| Economy | 144p | 50-100 MB |
| Standard | 360p | 150-250 MB |
| HD | 720p | 500-800 MB |

## API Endpoints

### Auth
- `POST /api/auth/register` — Create account
- `POST /api/auth/login` — Sign in
- `POST /api/auth/logout` — Sign out (auth required)
- `GET /api/auth/profile` — Get profile (auth required)

### Meetings
- `POST /api/meetings` — Create meeting
- `GET /api/meetings/:id` — Get meeting details
- `POST /api/meetings/:id/start` — Start meeting (host only)
- `POST /api/meetings/join` — Join by code
- `POST /api/meetings/:id/end` — End meeting (host only)
- `GET /api/meetings/:id/participants` — List participants

### Chat
- `POST /api/chat/messages/text` — Send text message
- `POST /api/chat/messages/file` — Send file
- `POST /api/chat/messages/voice` — Send voice message
- `POST /api/chat/messages/poll` — Create poll
- `POST /api/chat/messages/reaction` — Add reaction
- `GET /api/chat/rooms/:roomId/messages` — Get messages
- `POST /api/chat/rooms` — Create chat room
- `GET /api/chat/rooms` — List rooms

### Bandwidth
- `GET /api/bandwidth/profile` — Get bandwidth profile
- `POST /api/bandwidth/switch` — Switch mode
- `POST /api/bandwidth/report` — Report network speed

## Architecture

```
Global Load Balancer
        │
   ┌────┼────┐
   │    │    │
API    Media  Chat
Gateway Gateway Gateway
   │    │    │
Auth  WebRTC  Message
Svc   SFU    Broker
   │    │    │
Postgres Redis MongoDB
```

## Project Structure

```
shalom/
├── backend/
│   ├── cmd/api-gateway/    # Main entry point
│   ├── internal/
│   │   ├── config/         # Configuration
│   │   ├── models/         # Data models
│   │   ├── service/        # Business logic
│   │   ├── handler/        # HTTP handlers
│   │   ├── middleware/      # Auth, CORS, rate limit
│   │   └── repository/     # Data access
│   ├── migrations/         # SQL migrations
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── pages/          # Login, Register, Dashboard, Meeting, Chat
│   │   ├── services/       # API client
│   │   ├── context/        # State management (Zustand)
│   │   └── styles/         # Tailwind CSS
│   ├── Dockerfile
│   └── nginx.conf
├── infra/
│   ├── k8s/                # Kubernetes manifests
│   ├── livekit.yaml        # LiveKit config
│   ├── prometheus.yml      # Monitoring
│   └── docker/
└── docker-compose.yml
```
