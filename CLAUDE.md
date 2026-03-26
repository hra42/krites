# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Description

Krites — a full-stack platform for benchmarking LLMs: send identical prompts to multiple models, measure latency/throughput/cost/quality, and visually compare results.

## Tech Stack

- **Backend:** Go, Fiber web framework, DuckDB
- **Frontend:** SvelteKit, TypeScript, Chart.js
- **LLM Access:** OpenRouter API (300+ models)
- **Communication:** REST + SSE (Server-Sent Events) between frontend and backend
- **Language:** English for UI, comments, and code identifiers

## Architecture

### Layer Model

```
SvelteKit Frontend (Port 5173)
       │ REST + SSE (via Vite Proxy on /benchmarks and /v1)
Go Backend (Port 8080)
       │                        │
  OpenRouter API           DuckDB (per Suite)
```

### Core Concepts

- **Suite** — reusable test collection (prompts + models + configuration)
- **Run** — a single execution of a suite; status: pending → running → complete/failed/canceled
- **Result** — a model response to a prompt with metrics (TTFB, latency, tokens, cost)
- **Summary** — aggregated statistics per model after a run

### Backend Routes

All benchmark endpoints under `/benchmarks`:
- CRUD: `/benchmarks/suites`, `/benchmarks/suites/:id`
- Start run: `POST /benchmarks/suites/:id/run` (starts goroutine, returns 202)
- Run data: `/benchmarks/runs`, `/benchmarks/runs/:id`
- SSE stream: `GET /benchmarks/runs/:id/stream`
- Analytics: `/benchmarks/analytics/overview`, `/benchmarks/analytics/models`, `/benchmarks/analytics/trends`
- Export: `GET /benchmarks/runs/:id/export?format=csv|json`

### Benchmark Runner

The runner uses goroutines with a semaphore channel for parallelism. For each `(model, prompt, iteration)` tuple, an API call is made to OpenRouter. SSE events: `run_started`, `result_completed`, `judge_scored`, `run_completed`, `run_error`.

### LLM-as-Judge

Optionally enabled per suite. Evaluates responses on a 1-10 scale per criterion. Temperature 0.1, max 200 tokens, JSON response. On parse error, skip score — don't abort the run.

## Development Commands

### Backend (Go)

```bash
go build ./...
go test ./...
go test ./path/to/package -run TestName  # Run single test
go run main.go                            # Start server on :8080
```

### Frontend (SvelteKit)

```bash
cd frontend
npm install
npm run dev          # Dev server on :5173
npm run build        # Production build
npm run check        # Svelte-check + TypeScript
npm run lint         # Linting
```

### Docker

```bash
docker compose up -d --build    # Build and start all services
docker compose down              # Stop all services
```

## Project Structure

### Backend

```
├── main.go                  # Entry point, Fiber app setup, route registration
├── benchmark/
│   ├── models.go            # Core data structures (Suite, Run, Result, etc.)
│   ├── handler.go           # HTTP handlers + route registration
│   ├── runner.go            # Benchmark execution engine + judge
│   ├── store.go             # Storage interface
│   ├── duckdb_store.go      # DuckDB implementation
│   ├── summary.go           # Aggregation logic (percentiles, averages)
│   └── sse.go               # Server-Sent Events broadcaster
├── pricing/cache.go         # OpenRouter model pricing cache
├── database/                # DuckDB schema definitions
├── handler/                 # Non-benchmark handlers (chat, models, health)
├── middleware/               # CORS, auth, error handling
├── openrouter/              # OpenRouter API client
└── config/                  # YAML config loading
```

### Frontend

```
frontend/src/
├── lib/
│   ├── api/client.ts         # Fetch wrapper for all API calls
│   ├── types/index.ts        # TypeScript interfaces (mirror Go structs)
│   ├── stores/
│   │   ├── benchmark.ts      # Suite CRUD state
│   │   ├── run.ts            # Live run state + SSE event handling
│   │   └── toast.ts          # Toast notification system
│   ├── components/
│   │   ├── ModelPicker.svelte # Searchable model selector with OpenRouter data
│   │   ├── PromptEditor.svelte
│   │   ├── StatusBadge.svelte
│   │   ├── Toast.svelte
│   │   └── charts/           # Chart.js wrappers (6 chart types)
│   └── utils/chart-theme.ts  # Dark theme colors for Chart.js
├── routes/
│   ├── +layout.svelte        # App shell with sidebar
│   ├── +page.svelte          # Dashboard
│   ├── suites/               # Suite CRUD + inline edit
│   ├── runs/                 # Run history + detail with charts
│   └── analytics/            # Cross-run analytics
└── app.css                   # Global styles + design tokens
```

## Design Guidelines

- Dark-mode-first, industrial-precise aesthetic
- Base color: #0c0b0e, Accent: Violet/Purple (#a78bfa)
- Monospace for numbers/code (JetBrains Mono), display font (Outfit) for headings
- Chart.js with dark theme
- Toast notifications for user feedback on all actions
