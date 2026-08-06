# ADR 0001: React frontend and Go modular monolith

- Status: Accepted
- Date: 2026-08-05

## Context

DraftMeld must run as a Docker application or native Windows/Linux executable while keeping frontend, backend, contracts, and provider integrations understandable to occasional contributors.

## Decision

Use React with TypeScript for the frontend and Go for the backend. Expose a versioned REST API described by OpenAPI and use server-sent events for live draft notifications. Start with SQLite. Compile frontend assets into the production Go executable and publish the same executable inside the Docker image.

Organize the backend as a modular monolith with domain, application, API, persistence, and connector boundaries. Do not introduce independently deployed services until operational evidence requires them.

## Consequences

- Native releases are compact and require no Node.js runtime.
- Docker has one application service and one persistent data volume.
- Frontend and backend can be developed and tested independently.
- Cross-language API types require generation from the OpenAPI contract.
- Contributors work in two languages, but each language remains confined to a clearly defined project.
