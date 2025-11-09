
---
title: File Responsibilities
description: Responsibilities of each standard file in a domain package.
---

# File Responsibilities

## `model.go`
Defines immutable domain objects with private fields and accessor methods.

## `entity.go`
Database entity definitions and migration helpers using GORM. Provides `Make(Entity) (Model, error)` and `Model.ToEntity()`.

## `builder.go`
Fluent API for constructing validated domain models. `Build()` enforces invariants.

## `processor.go`
Pure curried business functions plus `AndEmit` variants for Kafka emission. Dependency order: `NewProcessor(log, ctx, db)`.

## `administrator.go`
High-level curried functions coordinating transactional DB operations; return `model.Provider[Entity]`.

## `provider.go`
Lazy data access layer returning `model.Provider[T]`. Compose with `model.Map`, `model.SliceMap`, `model.ParallelMap`.

## `producer.go`
Kafka message creation using context-aware header decorators via `producer.ProviderImpl(log)(ctx)`.

## `resource.go`
Route registration and handler definitions for REST endpoints. Perform input validation before processor invocation.

## `rest.go`
Serialization and transformation between domain models and JSON:API. No tenant data in payloads.

## `state.go`
Domain-specific enums and constants; state transition helpers.
