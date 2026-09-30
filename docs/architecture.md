# Architecture

The code is arranged around stable domain and service boundaries so external SDKs remain replaceable.

```text
HTTP handlers
    │
    ▼
application services
    │
    ├── ATS dispatcher ── ATS adapters ── public JSON APIs
    ├── drafting port ─── Azure AI Foundry
    ├── mail port ─────── Gmail API (draft creation only)
    ├── repository ────── Cosmos DB
    └── cache ─────────── Redis Cloud
```

Only the HTTP-to-manual-input path exists in the initial skeleton. Future integrations should expose narrow interfaces in `internal/service` and keep provider-specific code in a package dedicated to that provider.

Redis may contain normalized postings, generated draft cache entries, and rate-limit counters. Application records, job-description snapshots, profile versions, and draft versions belong in Cosmos DB.

The browser extension passes an active-tab URL into the Angular frontend. The frontend owns review and user intent; it calls the Go API only after an explicit user action.

