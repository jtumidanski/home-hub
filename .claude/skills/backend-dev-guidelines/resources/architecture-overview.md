
---
title: Architecture Overview
description: Core architectural principles of Home Hub Golang microservices.
---

# Architecture Overview

Services follow a strict layered design with functional composition and immutability.

## Layers
1. **Domain Layer** — Core logic, models, and validation.
2. **Infrastructure Layer** — Database and external systems.
3. **Transport Layer** — REST endpoints.
4. **Application Layer** — Orchestration and configuration.

## Core Technologies
- Go 1.24+
- GORM ORM (PostgreSQL)
- JSON:API standard
- Logrus structured logging
- OpenTelemetry tracing
- Functional programming with curried functions

## Principles

- **Immutability:** Domain models never mutate.
- **Separation of Concerns:** Domain logic isolated from persistence and transport.
- **Functional Composition:** Use `Provider`, `Map`, and `ParallelMap` for chaining.
- **Stateless Services:** All state in database, services are stateless.

## Startup Example

```go

db := database.Connect(logger, database.SetMigrations(domain.Migration))
server.New(logger).
  AddRouteInitializer(domain.InitializeRoutes(db)(GetServer())).
  Run()
```
