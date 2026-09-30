# Job Application Assistant

A single-user application that captures a public ATS job URL, normalizes the posting, drafts a tailored cover letter, and stages it as a Gmail draft for manual review.

This repository currently contains the initial project skeleton. External services are represented by boundaries and configuration only; no ATS, Azure AI Foundry, Cosmos DB, Redis Cloud, Gmail, or Entra ID integration is active yet.

## Repository layout

```text
apps/
  api/                 Go HTTP API and domain packages
  web/                 Angular frontend
extensions/
  shared/              Shared browser-extension source
  chrome/              Chrome manifest
  firefox/             Firefox manifest
infra/
  bicep/               Azure deployment entry point and modules
docs/                  Architecture notes
```

## Local development

Requirements: Go 1.24+, Node.js 22+, and npm.

```bash
cp .env.example .env
make api
```

In another terminal:

```bash
cd apps/web
npm install
npm start
```

The API listens on `http://localhost:8080`; the Angular dev server listens on `http://localhost:4200` and proxies `/api` to the API.

## Current API

- `GET /healthz` returns service health.
- `POST /api/v1/jobs/resolve` accepts a URL or manually pasted job text. URL-based ATS resolution is scaffolded and currently returns a clear not-implemented error.

See [docs/architecture.md](docs/architecture.md) for package responsibilities and dependency direction.

