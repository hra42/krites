# Krites — Projektplan

## Überblick

Eine Full-Stack-Plattform zum Durchführen eigener AI-Benchmarks: gleiche Prompts an mehrere LLMs senden, Latenz/Throughput/Kosten/Qualität messen und die Ergebnisse visuell vergleichen.

**Tech-Stack:** Go-Backend (erweitert [krites](https://github.com/hra42/krites)) + SvelteKit-Frontend

---

## 1. Architektur

### Schichtenmodell

```
┌─────────────────────────────────────────────────────┐
│  SvelteKit Frontend (Port 5173)                     │
│  Suite Editor · Live Runner · Dashboard · History   │
└──────────────────────┬──────────────────────────────┘
                       │ REST + SSE
┌──────────────────────▼──────────────────────────────┐
│  Go Backend — krites + Benchmark-Erweiterung      │
│  Benchmark Runner · Metrics Collector · LLM-Judge   │
└───────┬─────────────────────────────┬───────────────┘
        │ API Calls                   │ Ergebnisse
   ┌────▼────┐                   ┌────▼────┐
   │OpenRouter│                   │ DuckDB  │
   │300+ LLMs │                   │per Suite│
   └──────────┘                   └─────────┘
```

### Kernkonzepte

- **Suite** — eine wiederverwendbare Testsammlung (Prompts + Modelle + Konfiguration)
- **Run** — eine einzelne Ausführung einer Suite mit allen Ergebnissen
- **Result** — ein Modell-Antwort auf einen Prompt mit Metriken
- **Summary** — aggregierte Statistiken pro Modell nach einem Run

---

## 2. Backend-Erweiterung (Go)

### 2.1 Neue Datenmodelle

Diese Structs definieren die gesamte Benchmark-Domäne. Lege sie z.B. unter `models/benchmark/` ab.

**Suite:**
- `ID`, `Name`, `Description`
- `Prompts []Prompt` — jeder mit `ID`, `Name`, `SystemMessage`, `UserMessage`, `ExpectedOutput`, `Category`
- `Models []string` — OpenRouter Model-IDs (z.B. `openai/gpt-4o`, `anthropic/claude-3.5-sonnet`)
- `Config RunConfig`
- `CreatedAt`, `UpdatedAt`

**RunConfig:**
- `Temperature`, `MaxTokens`, `TopP` — Standard-LLM-Parameter
- `Iterations int` — wie oft jeder Prompt pro Modell laufen soll (für statistische Signifikanz)
- `Concurrency int` — max parallele API-Aufrufe
- `JudgeEnabled bool`, `JudgeModel string`, `JudgeCriteria []string`
- `TimeoutSeconds int`

**Run:**
- `ID`, `SuiteID`, `SuiteName`, `Status` (pending/running/complete/failed/canceled)
- `Results []Result`, `Summary *Summary`
- `StartedAt`, `EndedAt`, `Config`

**Result:**
- `ID`, `RunID`, `PromptID`, `PromptName`, `Model`, `Iteration`
- `Response string`
- `Metrics` — TTFB (ms), TotalLatency (ms), PromptTokens, CompletionTokens, TokensPerSecond, EstimatedCost
- `JudgeScores []JudgeScore` — Criterion, Score (1-10), Explanation
- `Status` (success/error/timeout), `Error string`

**ModelSummary** (pro Modell im Summary):
- Avg/P50/P95 für TTFB und Latenz
- AvgTokensPerSecond, AvgCost, SuccessRate
- AvgJudgeScores als Map (Criterion → Durchschnitt)

### 2.2 Neue API-Endpunkte

Registriere eine neue Route-Gruppe `/benchmarks` in deiner Fiber-App.

| Methode | Pfad | Beschreibung |
|---------|------|--------------|
| `GET` | `/benchmarks/suites` | Alle Suites auflisten |
| `POST` | `/benchmarks/suites` | Neue Suite erstellen |
| `GET` | `/benchmarks/suites/:id` | Suite-Details abrufen |
| `PUT` | `/benchmarks/suites/:id` | Suite aktualisieren |
| `DELETE` | `/benchmarks/suites/:id` | Suite löschen |
| `POST` | `/benchmarks/suites/:id/run` | Benchmark-Run starten |
| `GET` | `/benchmarks/runs` | Alle Runs auflisten (ohne Results) |
| `GET` | `/benchmarks/runs/:id` | Run mit allen Results + Summary |
| `GET` | `/benchmarks/runs/:id/stream` | SSE-Stream für Live-Updates |

### 2.3 Benchmark Runner — Ablauf

Der Runner wird als Goroutine gestartet, wenn `POST /suites/:id/run` aufgerufen wird.

**Schritt-für-Schritt:**

1. **Initialisierung:** Run-Objekt mit Status `pending` erstellen, in DB/Map speichern, `202 Accepted` zurückgeben
2. **Status → `running`:** SSE-Event `run_started` an alle Listener senden
3. **Parallelausführung:** Für jedes `(model, prompt, iteration)`-Tupel eine Goroutine starten, begrenzt durch einen Semaphore-Channel mit `Concurrency`-Größe
4. **Einzelner API-Call:** 
   - HTTP POST an OpenRouter `/chat/completions` mit `Authorization: Bearer <key>`
   - Zeitmessung starten → TTFB bei erstem Response-Byte → TotalLatency bei Abschluss
   - Token-Counts aus der `usage`-Sektion der Antwort extrahieren
   - TokensPerSecond = CompletionTokens / (TotalLatency / 1000)
   - Kosten schätzen aus Model-Pricing (optional: über `/v1/models` Preise cachen)
5. **SSE-Event `result_completed`** für jedes fertige Result senden
6. **LLM-as-Judge** (wenn aktiviert): Für jedes erfolgreiche Result den Judge-Prompt zusammenbauen und an das Judge-Modell schicken. Format: "Bewerte folgende Antwort von 1-10 für {Kriterium}. Antworte nur mit JSON." → SSE-Event `judge_scored`
7. **Summary berechnen:** Nach allen Results die Aggregate pro Modell berechnen (Avg, Percentile, Kosten)
8. **Status → `complete`:** SSE-Event `run_completed` mit vollem Run-Objekt, Streams schließen

### 2.4 SSE-Streaming

Verwende Fiber's `StreamWriter` für Server-Sent Events:

- Pro Run eine Map von Channels `map[string][]chan StreamEvent`
- `StreamRun`-Handler: Channel registrieren, Events in `data: {json}\n\n`-Format schreiben
- Nach `run_completed`: Alle Channels schließen und aus der Map entfernen
- Events: `run_started`, `result_completed`, `judge_scored`, `run_completed`, `run_error`

### 2.5 LLM-as-Judge — Prompt-Design

Für jedes Kriterium (z.B. "accuracy", "coherence", "helpfulness") einen separaten Judge-Call machen:

```
Du bist ein Experte. Bewerte die folgende AI-Antwort auf einer Skala von 1-10 für "{criterion}".

Original-Prompt: {user_message}
AI-Antwort: {response}
[Optional: Erwartete Antwort: {expected_output}]

Antworte NUR mit einem JSON-Objekt: {"score": <zahl>, "explanation": "<kurze Begründung>"}
```

- Temperature 0.1 für konsistente Bewertungen
- Max 200 Tokens für die Antwort
- JSON parsen, bei Fehler Score überspringen (nicht den Run abbrechen)

### 2.6 Integration in krites

Drei Änderungen an `main.go`:

1. **Import:** `import "github.com/hra42/krites/benchmark"`
2. **CORS:** Middleware mit `AllowOrigins: "http://localhost:5173"` hinzufügen
3. **Routes:** `benchmark.NewHandler(cfg.OpenRouter.BaseURL, cfg.OpenRouter.APIKey).RegisterRoutes(app)`

### 2.7 Persistenz-Strategie

**Phase 1 (Prototyp):** In-Memory mit `sync.RWMutex`-geschützten Maps — reicht zum Testen.

**Phase 2 (Production):** DuckDB nutzen, das krites schon mitbringt. Neue Tabellen:

- `benchmark_suites` — Suites als JSON-Blob oder normalisiert
- `benchmark_runs` — Run-Metadaten
- `benchmark_results` — einzelne Ergebnisse mit allen Metriken
- DuckDB eignet sich hervorragend für analytische Queries auf den Metriken (Percentile, Aggregationen)

---

## 3. Frontend (SvelteKit)

### 3.1 Projektstruktur

```
frontend/
├── src/
│   ├── lib/
│   │   ├── api/
│   │   │   └── client.ts          # Fetch-Wrapper für alle API-Calls
│   │   ├── stores/
│   │   │   └── benchmark.ts       # Svelte Stores für reaktiven State
│   │   ├── types/
│   │   │   └── index.ts           # TypeScript-Interfaces (spiegeln Go-Structs)
│   │   └── components/
│   │       ├── ModelChip.svelte    # Modell-Badge
│   │       ├── StatusBadge.svelte  # Status-Anzeige
│   │       ├── MetricCard.svelte   # Einzelne Kennzahl
│   │       ├── LatencyChart.svelte # Chart.js Wrapper
│   │       ├── CompareTable.svelte # Modell-Vergleichstabelle
│   │       └── ResultCard.svelte   # Einzelnes Benchmark-Ergebnis
│   ├── routes/
│   │   ├── +layout.svelte         # App-Shell mit Sidebar
│   │   ├── +page.svelte           # Dashboard/Home
│   │   ├── suites/
│   │   │   ├── +page.svelte       # Suite-Liste + Erstellen
│   │   │   └── [id]/
│   │   │       ├── +page.svelte   # Suite-Detail
│   │   │       └── run/
│   │   │           └── +page.svelte # Live Run mit SSE
│   │   ├── runs/
│   │   │   ├── +page.svelte       # Run-History
│   │   │   └── [id]/
│   │   │       └── +page.svelte   # Run-Detail mit Charts
│   │   └── dashboard/
│   │       └── +page.svelte       # Analytics über alle Runs
│   └── app.css                    # Globale Styles
├── svelte.config.js
├── vite.config.ts                 # Proxy /benchmarks → :8080
└── package.json
```

### 3.2 API-Client (`lib/api/client.ts`)

Ein dünner Fetch-Wrapper mit:

- `request<T>(path, options)` — generische Fetch-Funktion mit JSON-Handling und Error-Mapping
- Exportierte Funktionen: `listSuites()`, `getSuite(id)`, `createSuite(data)`, `updateSuite(id, data)`, `deleteSuite(id)`, `listRuns()`, `getRun(id)`, `startRun(suiteId)`, `listModels()`
- `streamRun(runId, onEvent, onError)` — erstellt einen `EventSource`, parst `data:`-Zeilen als JSON, gibt eine Cleanup-Funktion zurück

### 3.3 Svelte Stores (`lib/stores/benchmark.ts`)

Nutze Svelte's `writable` und `derived` Stores:

- `suites`, `currentSuite` — Suite-Daten
- `runs`, `currentRun` — Run-Daten
- `liveResults` — Array von Results, das während eines Runs live wächst
- `isRunning` — Boolean für UI-State
- `handleStreamEvent(event)` — Switch über Event-Typen, aktualisiert die Stores
- `resultsByModel` — derived Store, gruppiert liveResults nach Modell
- `progress` — derived Store, berechnet completed/total/percent

### 3.4 Seiten im Detail

#### Dashboard (`/`)
- Statistik-Karten: Anzahl Suites, Anzahl Runs, letzter Status
- Tabelle der letzten 5 Runs mit Suite-Name, Status, Datum
- Quick-Links: "Neue Suite", "Analytics"

#### Suites (`/suites`)
- Liste aller Suites als Karten mit Name, Beschreibung, Modell-Chips, Prompt-Anzahl
- "Neue Suite"-Button öffnet ein Inline-Formular
- **Suite-Formular:**
  - Name + Beschreibung
  - Modell-Auswahl: Text-Input + "Hinzufügen"-Button, Modelle als entfernbare Chips
  - Prompts: Dynamische Liste, jeder mit Name, System Message, User Message
  - Konfiguration: Temperature, Max Tokens, Iterations, Concurrency als Zahlenfelder
  - Judge: Checkbox zum Aktivieren, dann Judge-Modell und Kriterien-Auswahl

#### Suite-Detail (`/suites/[id]`)
- Name, Beschreibung, Modell-Chips
- Alle Prompts mit System/User Message als formatierte Blöcke
- Konfigurationsübersicht als Grid
- "Benchmark starten"-Button → navigiert zu `/suites/[id]/run`

#### Live Run (`/suites/[id]/run`)
- **Vor dem Start:** Zusammenfassung (Modelle × Prompts × Iterationen = erwartete Calls), Start-Button
- **Während des Runs:** 
  - Progress-Bar mit Zähler (12/48 abgeschlossen)
  - Status-Badge "Läuft" mit Pulse-Animation
  - Ergebnisse gruppiert nach Modell, jedes mit:
    - Prompt-Name, Status-Badge, Latenz
    - Response-Preview (max 300 Zeichen)
    - Judge-Scores als farbige Badges
  - Live-Aggregat pro Modell: Avg Latenz, tok/s
- **Nach dem Run:** Summary-Tabelle mit allen Modellen nebeneinander

#### Run-History (`/runs`)
- Tabelle aller Runs: Suite-Name, Status, Datum, Ergebnis-Anzahl
- Filter nach Status
- Klick → Run-Detail

#### Run-Detail (`/runs/[id]`)
- Summary-Karten: Avg Latenz, Avg tok/s, Kosten pro Modell
- **Charts (Chart.js):**
  - Grouped Bar Chart: Latenz pro Modell
  - Radar Chart: Judge-Scores pro Modell/Kriterium
  - Line Chart: Latenz über Iterationen (Konsistenz)
  - Horizontal Bar: Kosten-Vergleich
- Vollständige Results-Tabelle mit Filter/Sortierung

#### Analytics (`/dashboard`)
- Modell-Übergreifende Trends über alle Runs
- Kosten-Tracker
- Best/Worst per Kategorie

### 3.5 Vite-Konfiguration

Proxy im `vite.config.ts` einrichten, damit der Dev-Server API-Calls an das Go-Backend weiterleitet:

```js
server: {
  proxy: {
    '/benchmarks': 'http://localhost:8080',
    '/v1': 'http://localhost:8080'
  }
}
```

### 3.6 Design-Richtung

Das Frontend sollte sich deutlich von generischen Admin-Panels abheben. Empfehlung:

- **Ästhetik:** Dark-Mode-first, industriell-präzise ("Krites")
- **Typografie:** Eine Display-Font wie Outfit oder Satoshi + JetBrains Mono für Daten/Code
- **Farben:** Warmes Schwarz (#0c0b0e) als Basis, Violet/Purple (#a78bfa) als Akzent, semantische Farben für Status
- **Daten-Darstellung:** Monospace für Zahlen, dezente Badges für Status, Modell-Chips als Pills
- **Animationen:** Dezente fadeIn für Karten, Pulse für Live-Status, Progress-Bar mit smooth Transition
- **Charts:** Chart.js mit dunklem Theme, angepasste Farben pro Modell

---

## 4. Umsetzungsreihenfolge

### Phase 1 — Fundament (Tag 1-2)

1. Go: Datenmodelle definieren (Structs für Suite, Run, Result, etc.)
2. Go: CRUD-Endpunkte für Suites (In-Memory-Maps, noch kein Runner)
3. Go: CORS-Middleware + Integration in `main.go`
4. SvelteKit: Projekt scaffolden (`npm create svelte@latest`)
5. SvelteKit: TypeScript-Types, API-Client, Stores anlegen
6. SvelteKit: Layout mit Sidebar, Suite-Liste + Erstellformular

### Phase 2 — Benchmark Runner (Tag 3-4)

7. Go: Benchmark Runner mit Goroutines + Semaphore
8. Go: Metrics-Erfassung (TTFB, Latenz, Tokens, Cost)
9. Go: SSE-Streaming-Endpunkt
10. SvelteKit: Live-Run-Seite mit EventSource + Progress
11. SvelteKit: Echtzeit-Ergebnisanzeige gruppiert nach Modell

### Phase 3 — Qualität & Analyse (Tag 5-6)

12. Go: LLM-as-Judge Implementation
13. Go: Summary-Berechnung (Percentile, Aggregate)
14. SvelteKit: Run-Detail-Seite mit Charts (Chart.js)
15. SvelteKit: Vergleichstabellen + Judge-Score-Darstellung

### Phase 4 — Polish (Tag 7+)

16. Go: DuckDB-Persistenz statt In-Memory
17. SvelteKit: Analytics-Dashboard über mehrere Runs
18. SvelteKit: Export (CSV/JSON)
19. Docker Compose für Backend + Frontend
20. Modell-Preise aus OpenRouter `/v1/models` cachen für Kosten-Berechnung

---

## 5. Wichtige Design-Entscheidungen

| Entscheidung | Empfehlung | Begründung |
|---|---|---|
| Persistenz | DuckDB | Schon in krites vorhanden, exzellent für analytische Queries |
| Streaming | SSE (nicht WebSocket) | Einfacher, unidirektional reicht, Fiber unterstützt es nativ |
| Charts | Chart.js | Leichtgewichtig, funktioniert mit SSR, gute Svelte-Integration |
| Judge-Modell | Separat konfigurierbar | Nicht jeder will GPT-4o als Judge, manche bevorzugen Claude |
| Parallelität | Semaphore-Channel | Go-idiomatisch, verhindert Rate-Limiting bei OpenRouter |
| Kosten-Tracking | Optional/Best-Effort | Pricing ändert sich oft, lieber aus API lesen statt hardcoden |

---

## 6. Erweiterungsmöglichkeiten

- **Custom Grading Functions** — neben LLM-as-Judge auch Regex-Checks, Code-Execution-Tests
- **A/B-Testing** — gleicher Prompt, zwei Temperature-Settings vergleichen
- **Scheduled Runs** — Cron-basiert regelmäßig Benchmarks laufen lassen
- **Team-Features** — Suites teilen, Kommentare an Ergebnissen
- **Webhook-Notifications** — Slack/Discord bei Run-Ende benachrichtigen
- **Model Routing** — automatisch das beste Modell pro Kategorie empfehlen basierend auf historischen Daten
