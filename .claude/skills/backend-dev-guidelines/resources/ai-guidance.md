
---
title: AI Code Generation Guidance
description: Rules for AI agents generating or editing Golang services.
---

# AI Code Generation Guidance

## Core Rules
1. Respect file responsibilities.
2. Maintain immutability and functional composition.
3. Use curried functions for dependency injection.
4. Only emit messages via `AndEmit` + buffer.
5. Never include tenant fields in public APIs.

## Generation Workflow
1. Create `model.go`
2. Map `entity.go` to DB
3. Implement `builder.go`
4. Define processors and providers
5. Add Kafka producers and REST handlers
6. Write table-driven tests

## Useful Composition
```go
model.Map(Make)(entityProvider)(model.ParallelMap())
```
