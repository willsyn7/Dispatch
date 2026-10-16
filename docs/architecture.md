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

The HTTP layer calls application services without importing provider SDKs. ATS adapters convert known public URL shapes into provider-owned JSON endpoints, so the backend never performs a request against an arbitrary URL supplied by a user. The drafting client calls Azure AI Foundry's OpenAI-compatible v1 chat-completions endpoint and authenticates with an API key locally or managed identity in Azure.

Redis may contain normalized postings, generated draft cache entries, and rate-limit counters. Application records, job-description snapshots, profile versions, and draft versions belong in Cosmos DB.

Cosmos containers follow the partitioning in the design: applications and profiles use `/userId`, and draft versions use `/applicationId`. A memory repository supports local development. Redis failures do not make source-of-truth operations fail; a miss or cache error falls through to the ATS or model call.

The browser extension passes an active-tab URL into the Angular frontend. The frontend owns review and user intent; it calls the Go API only after an explicit user action. The current UI can save the candidate profile, normalize a posting, request a draft, and display the result. Gmail draft creation is the remaining external service.
