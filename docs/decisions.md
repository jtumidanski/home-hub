## 2026-03-24: Fork api2go to jtumidanski/api2go
Status: accepted

Context:
The project uses the `manyminds/api2go` library for JSON:API serialization and deserialization. The upstream repository is no longer maintained, and we needed updates to better support JSON:API compliance.

Decision:
Replace `github.com/manyminds/api2go` with `github.com/jtumidanski/api2go` (v1.0.4) across all modules (shared/go/server, account-service, productivity-service).

Why:
The upstream `manyminds/api2go` is unmaintained. Our fork includes updates to better support JSON:API patterns used across Home Hub services.

Consequences:
- All Go import paths changed from `github.com/manyminds/api2go/jsonapi` to `github.com/jtumidanski/api2go/jsonapi`.
- Three `go.mod` files updated: `shared/go/server`, `services/account-service`, `services/productivity-service`.
- We own the dependency lifecycle and can push fixes as needed.
- Transitive dependency versions (gin, validator, etc.) were upgraded as part of `go mod tidy`.

Alternatives considered:
- Continue using the unmaintained upstream — rejected due to inability to fix issues or add features.
- Vendor the library locally in `shared/` — rejected as a Go module fork is cleaner and allows versioned releases.
