# Dashboard Designer — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce a new `dashboard-service` plus frontend designer/renderer so households can compose JSON-backed dashboards from a fixed widget registry, replacing the hand-coded `DashboardPage`.

**Architecture:** New Go service owning `dashboard.dashboards` (JSON:API CRUD + promote/copy/seed); new `household_preferences` typed table in `account-service`; new `shared/go/{dashboard,kafka,events}` libraries; `UserDeletedEvent` cascade over an externally-managed Kafka broker; React renderer (CSS Grid, library-free) and designer (lazy-loaded, `react-grid-layout`) behind `/app/dashboards/:id[/edit]` with a `DashboardShell` parent route; widget-type allowlist mirrored Go ↔ TS with a parity fixture.

**Tech Stack:** Go 1.26 / GORM / gorilla-mux / JSON:API via `api2go`; Postgres 16 (schema `dashboard`); `segmentio/kafka-go`; React 19 / Vite 8 / TypeScript; TanStack Query 5; `react-grid-layout` (designer only); `@dnd-kit/sortable` (sidebar reorder); zod 4 + react-hook-form 7; Tailwind 4 + shadcn-style primitives.

> Before starting any task, read `context.md` in the same folder. It documents reference files, non-obvious gotchas, and the rollout sequence.

---

## Phase index

- **Phase A** — Shared Go libraries (`shared/go/{dashboard,events,kafka}`) — Tasks 1–5
- **Phase B** — `dashboard-service` scaffold — Tasks 6–8
- **Phase C** — Layout validator — Tasks 9–13
- **Phase D** — Entity, model, provider, administrator — Tasks 14–17
- **Phase E** — Processor (CRUD, scope, reorder, promote, copy-to-mine, seed) — Tasks 18–24
- **Phase F** — REST layer — Tasks 25–30
- **Phase G** — Retention + Kafka consumer — Tasks 31–33
- **Phase H** — `account-service` changes (`household_preferences`, internal delete, producer) — Tasks 34–37
- **Phase I** — Infrastructure (compose, nginx, scripts) — Tasks 38–39
- **Phase J** — Service docs — Task 40
- **Phase K** — Frontend foundation (types, schema, widget registry, parity test) — Tasks 41–46
- **Phase L** — Frontend renderer + routing + seeding — Tasks 47–53
- **Phase M** — Sidebar integration — Tasks 54–56
- **Phase N** — Frontend designer — Tasks 57–64
- **Phase O** — Cutover (delete `DashboardPage`, flip `/app` index) — Task 65

Each task cites exact files and shows full code for novel pieces. Pattern-repeating scaffolding (builders, REST transforms, etc.) explicitly cites the reference file in `services/package-service/internal/tracking/` — copy that pattern, substituting the dashboard domain fields. See `context.md` for the reference table.

---

## Phase A — Shared Go libraries

### Task 1: `shared/go/dashboard` — widget-type allowlist + layout constants

**Files:**
- Create: `shared/go/dashboard/go.mod`
- Create: `shared/go/dashboard/types.go`
- Create: `shared/go/dashboard/types_test.go`
- Create: `shared/go/dashboard/fixtures/widget-types.json` (parity fixture)

- [ ] **Step 1: Write the failing test** — `shared/go/dashboard/types_test.go`

```go
package dashboard

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

func TestIsKnownWidgetType(t *testing.T) {
	for _, typ := range []string{
		"weather", "tasks-summary", "reminders-summary", "overdue-summary",
		"meal-plan-today", "calendar-today", "packages-summary",
		"habits-today", "workout-today",
	} {
		if !IsKnownWidgetType(typ) {
			t.Fatalf("expected %q to be known", typ)
		}
	}
	if IsKnownWidgetType("foo") {
		t.Fatalf("expected unknown type to be rejected")
	}
}

func TestWidgetTypesParityFixture(t *testing.T) {
	data, err := os.ReadFile("fixtures/widget-types.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture []string
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	got := make([]string, 0, len(WidgetTypes))
	for k := range WidgetTypes {
		got = append(got, k)
	}
	sort.Strings(got)
	want := append([]string(nil), fixture...)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("got %d types, fixture has %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("mismatch at %d: go=%q fixture=%q", i, got[i], want[i])
		}
	}
}

func TestLayoutConstants(t *testing.T) {
	if LayoutSchemaVersion != 1 {
		t.Fatalf("schema version must be 1")
	}
	if MaxWidgets != 40 || GridColumns != 12 {
		t.Fatalf("cap constants drifted")
	}
	if MaxLayoutBytes != 64*1024 {
		t.Fatalf("layout byte cap drifted")
	}
	if MaxWidgetConfigBytes != 4*1024 {
		t.Fatalf("widget config byte cap drifted")
	}
	if MaxWidgetConfigDepth != 5 {
		t.Fatalf("config depth cap drifted")
	}
}
```

- [ ] **Step 2: Create go.mod** — `shared/go/dashboard/go.mod`

```
module github.com/jtumidanski/home-hub/shared/go/dashboard

go 1.26.1
```

- [ ] **Step 3: Create fixture** — `shared/go/dashboard/fixtures/widget-types.json`

```json
[
  "calendar-today",
  "habits-today",
  "meal-plan-today",
  "overdue-summary",
  "packages-summary",
  "reminders-summary",
  "tasks-summary",
  "weather",
  "workout-today"
]
```

- [ ] **Step 4: Implement** — `shared/go/dashboard/types.go`

```go
// Package dashboard holds cross-service constants for the dashboard feature:
// the widget-type allowlist, the layout schema version, and the hard caps on
// payload size / widget count / per-widget config. The matching TypeScript
// allowlist lives at frontend/src/lib/dashboard/widget-types.ts; both sides
// are asserted against fixtures/widget-types.json.
package dashboard

const (
	LayoutSchemaVersion  = 1
	MaxWidgets           = 40
	MaxLayoutBytes       = 64 * 1024
	MaxWidgetConfigBytes = 4 * 1024
	MaxWidgetConfigDepth = 5
	GridColumns          = 12
)

var WidgetTypes = map[string]struct{}{
	"weather":           {},
	"tasks-summary":     {},
	"reminders-summary": {},
	"overdue-summary":   {},
	"meal-plan-today":   {},
	"calendar-today":    {},
	"packages-summary":  {},
	"habits-today":      {},
	"workout-today":     {},
}

func IsKnownWidgetType(t string) bool { _, ok := WidgetTypes[t]; return ok }
```

- [ ] **Step 5: Run and commit**

```bash
cd shared/go/dashboard && go test ./...
git add shared/go/dashboard
git commit -m "feat(shared): add dashboard widget-type allowlist and layout caps"
```

---

### Task 2: `shared/go/events` — cross-service event envelope

**Files:**
- Create: `shared/go/events/go.mod`
- Create: `shared/go/events/events.go`
- Create: `shared/go/events/events_test.go`

- [ ] **Step 1: Write the failing test** — `shared/go/events/events_test.go`

```go
package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserDeletedEventRoundtrip(t *testing.T) {
	in := UserDeletedEvent{
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		DeletedAt: time.Now().UTC().Truncate(time.Second),
	}
	env, err := NewEnvelope(TypeUserDeleted, in)
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != TypeUserDeleted {
		t.Fatalf("type: got %s", env.Type)
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var got Envelope
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	var out UserDeletedEvent
	if err := json.Unmarshal(got.Payload, &out); err != nil {
		t.Fatal(err)
	}
	if out.UserID != in.UserID || out.TenantID != in.TenantID {
		t.Fatalf("ids drifted")
	}
}
```

- [ ] **Step 2: Create module** — `shared/go/events/go.mod`

```
module github.com/jtumidanski/home-hub/shared/go/events

go 1.26.1

require github.com/google/uuid v1.6.0
```

- [ ] **Step 3: Implement** — `shared/go/events/events.go`

```go
// Package events defines versioned envelopes for domain events crossing
// service boundaries on the shared Kafka bus. Each event type has a stable
// string tag carried on the envelope so consumers can ignore types they do
// not handle.
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	TypeUserDeleted EventType = "USER_DELETED"
)

// Envelope is the wire format for every cross-service event.
type Envelope struct {
	Type    EventType       `json:"type"`
	Version int             `json:"version"`
	Payload json.RawMessage `json:"payload"`
}

func NewEnvelope(t EventType, payload any) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{Type: t, Version: 1, Payload: raw}, nil
}

// UserDeletedEvent is emitted by account-service when a user is hard-deleted
// from a tenant. Consumers remove any user-scoped rows they own.
type UserDeletedEvent struct {
	TenantID  uuid.UUID `json:"tenantId"`
	UserID    uuid.UUID `json:"userId"`
	DeletedAt time.Time `json:"deletedAt"`
}
```

- [ ] **Step 4: Run and commit**

```bash
cd shared/go/events && go mod tidy && go test ./...
git add shared/go/events
git commit -m "feat(shared): add events envelope and UserDeletedEvent"
```

---

### Task 3: `shared/go/kafka` — producer

**Files:**
- Create: `shared/go/kafka/go.mod`
- Create: `shared/go/kafka/producer/producer.go`
- Create: `shared/go/kafka/producer/producer_test.go`

Reference style: `/home/tumidanski/source/atlas-ms/atlas/libs/atlas-kafka/producer/producer.go`.

- [ ] **Step 1: Write the failing test** — `shared/go/kafka/producer/producer_test.go`

```go
package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type stubWriter struct {
	msgs []kafka.Message
	err  error
}

func (s *stubWriter) WriteMessages(_ context.Context, m ...kafka.Message) error {
	s.msgs = append(s.msgs, m...)
	return s.err
}

func (s *stubWriter) Close() error { return nil }

func TestProduceWritesMessage(t *testing.T) {
	sw := &stubWriter{}
	l := logrus.New()
	p := &Producer{writer: sw, logger: l}
	err := p.Produce(context.Background(), "topic", []byte("k"), []byte(`{"x":1}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.msgs) != 1 || sw.msgs[0].Topic != "topic" {
		t.Fatalf("unexpected: %+v", sw.msgs)
	}
}

func TestProduceRetriesThenFails(t *testing.T) {
	sw := &stubWriter{err: errors.New("boom")}
	l := logrus.New()
	p := &Producer{writer: sw, logger: l, maxAttempts: 3}
	err := p.Produce(context.Background(), "topic", []byte("k"), []byte("v"), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if len(sw.msgs) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(sw.msgs))
	}
}
```

- [ ] **Step 2: Create module** — `shared/go/kafka/go.mod`

```
module github.com/jtumidanski/home-hub/shared/go/kafka

go 1.26.1

require (
	github.com/segmentio/kafka-go v0.4.47
	github.com/sirupsen/logrus v1.9.4
)
```

- [ ] **Step 3: Implement** — `shared/go/kafka/producer/producer.go`

```go
// Package producer is a thin retry wrapper around segmentio/kafka-go's Writer.
// It is deliberately small: one Produce call, configurable max attempts,
// structured logging on each retry. No outbox, no partition sentinel — a
// failed final attempt is a logged warning so the caller can decide.
package producer

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Producer struct {
	writer      Writer
	logger      logrus.FieldLogger
	maxAttempts int
	backoff     time.Duration
}

type Config struct {
	Brokers     []string
	MaxAttempts int
	Backoff     time.Duration
}

func New(cfg Config, logger logrus.FieldLogger) *Producer {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.Backoff <= 0 {
		cfg.Backoff = 250 * time.Millisecond
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &Producer{writer: w, logger: logger, maxAttempts: cfg.MaxAttempts, backoff: cfg.Backoff}
}

func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	msg := kafka.Message{Topic: topic, Key: key, Value: value}
	for k, v := range headers {
		msg.Headers = append(msg.Headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	var err error
	attempts := p.maxAttempts
	if attempts <= 0 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		err = p.writer.WriteMessages(ctx, msg)
		if err == nil {
			return nil
		}
		p.logger.WithError(err).WithField("topic", topic).WithField("attempt", i+1).Warn("kafka produce failed")
		if i < attempts-1 && p.backoff > 0 {
			time.Sleep(p.backoff)
		}
	}
	return err
}

func (p *Producer) Close() error { return p.writer.Close() }
```

- [ ] **Step 4: Run and commit**

```bash
cd shared/go/kafka && go mod tidy && go test ./producer/...
git add shared/go/kafka
git commit -m "feat(shared): add kafka producer with retry and structured logging"
```

---

### Task 4: `shared/go/kafka` — consumer

**Files:**
- Create: `shared/go/kafka/consumer/consumer.go`
- Create: `shared/go/kafka/consumer/consumer_test.go`

Reference: `/home/tumidanski/source/atlas-ms/atlas/libs/atlas-kafka/consumer/manager.go` for the `Manager` + `Register` style (copy the shape, not the module path).

- [ ] **Step 1: Write the failing test** — `shared/go/kafka/consumer/consumer_test.go`

```go
package consumer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type fakeReader struct {
	mu   sync.Mutex
	msgs []kafka.Message
	idx  int
	done chan struct{}
}

func (f *fakeReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.idx >= len(f.msgs) {
		close(f.done)
		return kafka.Message{}, errors.New("EOF")
	}
	m := f.msgs[f.idx]
	f.idx++
	return m, nil
}
func (f *fakeReader) CommitMessages(_ context.Context, _ ...kafka.Message) error { return nil }
func (f *fakeReader) Close() error                                                { return nil }

func TestDispatchInvokesHandler(t *testing.T) {
	got := make(chan kafka.Message, 1)
	r := &fakeReader{msgs: []kafka.Message{{Topic: "t", Value: []byte(`x`)}}, done: make(chan struct{})}
	l := logrus.New()
	m := &Manager{reader: r, logger: l, handler: func(ctx context.Context, msg kafka.Message) error { got <- msg; return nil }}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Run(ctx)
	select {
	case msg := <-got:
		if msg.Topic != "t" {
			t.Fatalf("unexpected topic %q", msg.Topic)
		}
	case <-time.After(time.Second):
		t.Fatal("handler never invoked")
	}
}
```

- [ ] **Step 2: Implement** — `shared/go/kafka/consumer/consumer.go`

```go
// Package consumer is a small wrapper around kafka-go's Reader that routes
// each message to a single handler and commits only on success. Consumers
// that want to multiplex by event type can do so inside their handler by
// decoding the envelope's Type field.
package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Reader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Handler func(ctx context.Context, msg kafka.Message) error

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Manager struct {
	reader  Reader
	logger  logrus.FieldLogger
	handler Handler
}

func New(cfg Config, handler Handler, logger logrus.FieldLogger) *Manager {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})
	return &Manager{reader: r, logger: logger, handler: handler}
}

func (m *Manager) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		msg, err := m.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			m.logger.WithError(err).Warn("kafka fetch failed")
			continue
		}
		if err := m.handler(ctx, msg); err != nil {
			m.logger.WithError(err).WithField("topic", msg.Topic).Error("kafka handler failed; skipping commit")
			continue
		}
		if err := m.reader.CommitMessages(ctx, msg); err != nil {
			m.logger.WithError(err).Warn("kafka commit failed")
		}
	}
}

func (m *Manager) Close() error { return m.reader.Close() }
```

- [ ] **Step 3: Run and commit**

```bash
cd shared/go/kafka && go mod tidy && go test ./consumer/...
git add shared/go/kafka/consumer
git commit -m "feat(shared): add kafka consumer Manager with single-handler dispatch"
```

---

### Task 5: Add `dashboards` retention category to `shared/go/retention`

**Files:**
- Modify: `shared/go/retention/category.go`

- [ ] **Step 1: Write the failing test** — add to `shared/go/retention/category_test.go` (or create if missing). Add a new `TestDashboardCategory` that asserts:

```go
func TestDashboardCategory(t *testing.T) {
	if !CatDashboardDashboards.IsKnown() {
		t.Fatal("CatDashboardDashboards should be known")
	}
	if !CatDashboardDashboards.IsHouseholdScoped() {
		t.Fatal("dashboards should be household-scoped")
	}
	if Defaults[CatDashboardDashboards] != 0 {
		t.Fatalf("default for dashboards should be 0 (never auto-purge), got %d", Defaults[CatDashboardDashboards])
	}
}
```

- [ ] **Step 2: Implement** — in `shared/go/retention/category.go`, add the new const, `Defaults` entry, `scopeKindOf` entry, and add to the `knownCategories` check slice (or however the file tracks the known set).

```go
const CatDashboardDashboards Category = "dashboard.dashboards"

// In the Defaults map:
//   CatDashboardDashboards: 0, // v1: never auto-purge; plumbing only

// In scopeKindOf:
//   CatDashboardDashboards: ScopeHousehold,
```

- [ ] **Step 3: Run retention tests and commit**

```bash
cd shared/go/retention && go test ./...
git add shared/go/retention
git commit -m "feat(retention): register dashboard.dashboards category"
```

---

## Phase B — `dashboard-service` scaffold

### Task 6: Create service skeleton (module, Dockerfile, config, cmd)

**Files:**
- Create: `services/dashboard-service/go.mod`
- Create: `services/dashboard-service/Dockerfile`
- Create: `services/dashboard-service/internal/config/config.go`
- Create: `services/dashboard-service/cmd/main.go`
- Create: `services/dashboard-service/README.md`

Reference: `services/package-service/go.mod`, `Dockerfile`, `internal/config/config.go`, `cmd/main.go`.

- [ ] **Step 1: go.mod** — Mirror `services/package-service/go.mod`, drop carrier/poller deps, add:
```
github.com/jtumidanski/home-hub/shared/go/dashboard v0.0.0
github.com/jtumidanski/home-hub/shared/go/events v0.0.0
github.com/jtumidanski/home-hub/shared/go/kafka v0.0.0
```
and add the matching `replace (...)` entries pointing at `../../shared/go/{dashboard,events,kafka}`. Add `gorm.io/datatypes v1.1.1` for `datatypes.JSON`. Add `github.com/segmentio/kafka-go v0.4.47`.

- [ ] **Step 2: Dockerfile** — Copy `services/package-service/Dockerfile`, change the COPY lines so the shared libs copied are: `model, tenant, logging, database, server, auth, retention, dashboard, events, kafka`. Adjust the final service path to `services/dashboard-service/`.

- [ ] **Step 3: `internal/config/config.go`** — Copy the pattern from `services/package-service/internal/config/config.go`. Drop carrier/poller fields. Add:

```go
type Config struct {
	DB                  database.Config
	Port                string
	JWKSURL             string
	AccountServiceURL   string
	InternalToken       string
	RetentionInterval   time.Duration
	KafkaBrokers        []string
	TopicUserLifecycle  string
	KafkaConsumerGroup  string
}

func Load() Config {
	brokers := strings.Split(envOrDefault("BOOTSTRAP_SERVERS", "kafka-broker.kafka.svc.cluster.local:9092"), ",")
	return Config{
		DB: database.Config{
			Host: envOrDefault("DB_HOST", "postgres.home"),
			Port: envOrDefault("DB_PORT", "5432"),
			User: envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName: envOrDefault("DB_NAME", "home_hub"),
			Schema: "dashboard",
		},
		Port:                envOrDefault("PORT", "8080"),
		JWKSURL:             envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		AccountServiceURL:   envOrDefault("ACCOUNT_SERVICE_URL", "http://account-service:8080"),
		InternalToken:       os.Getenv("INTERNAL_SERVICE_TOKEN"),
		RetentionInterval:   parseDuration(os.Getenv("RETENTION_INTERVAL"), 6*time.Hour),
		KafkaBrokers:        brokers,
		TopicUserLifecycle:  envOrDefault("EVENT_TOPIC_USER_LIFECYCLE", "home-hub.user.lifecycle"),
		KafkaConsumerGroup:  envOrDefault("KAFKA_CONSUMER_GROUP", "dashboard-service"),
	}
}
```

(Copy `envOrDefault` + `parseDuration` helpers from the reference.)

- [ ] **Step 4: `cmd/main.go`** — Skeleton that boots the service. Start with a minimal version that will grow as the remaining phases complete. The minimum at this step:

```go
package main

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/config"
	sharedauth "github.com/jtumidanski/home-hub/shared/go/auth"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

func main() {
	l := logging.NewLogger()
	cfg := config.Load()

	shutdownTracing := logging.InitTracing(l, "dashboard-service")
	defer shutdownTracing(context.Background())

	db := database.Connect(l, cfg.DB) // migrations added in later phases

	authValidator := sharedauth.NewValidator(l, cfg.JWKSURL)
	si := server.GetServerInformation()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.New(l).
		WithAddr(":" + cfg.Port).
		AddRouteInitializer(func(router *mux.Router) {
			api := router.PathPrefix("/api/v1").Subrouter()
			api.Use(sharedauth.Middleware(l, authValidator))

			// Route registrations added in later tasks.
			_ = api
			_ = si
			_ = ctx
			_ = db
		}).
		Run()
}
```

- [ ] **Step 5: README** — Minimal stub noting `scripts/local-up.sh` starts the service and pointing at `docs/domain.md` (to be written later).

- [ ] **Step 6: Verify build and commit**

```bash
cd services/dashboard-service && go mod tidy && go build ./...
git add services/dashboard-service
git commit -m "feat(dashboard-service): scaffold module, config, Dockerfile, main entrypoint"
```

---

### Task 7: Add `appcontext` helper (copied pattern)

**Files:**
- Create: `services/dashboard-service/internal/appcontext/context.go`

Reference: `services/account-service/internal/appcontext/` (or any service's copy). The file is a thin re-export for consistency — or omit it if the service doesn't need one beyond `tenantctx`. Check the reference first.

- [ ] **Step 1:** Inspect `services/package-service/internal/appcontext/` (if exists) or `services/account-service/internal/appcontext/`. Copy the pattern byte-for-byte. If the reference packages simply re-export `shared/go/tenant`, this task is purely mechanical. If no such package exists in the reference services, skip this task.

- [ ] **Step 2: Commit** (if anything was added)

```bash
git add services/dashboard-service/internal/appcontext
git commit -m "feat(dashboard-service): add appcontext helpers"
```

---

### Task 8: CI / build smoke check

- [ ] **Step 1:** From repo root: `go work sync` if a `go.work` exists; then `go build ./services/dashboard-service/...` to make sure the skeleton compiles as part of the root workspace.

- [ ] **Step 2:** Verify `Dockerfile` builds locally:

```bash
docker build -f services/dashboard-service/Dockerfile -t dashboard-service:dev .
```

- [ ] **Step 3:** No commit needed unless fixups were required.

---

## Phase C — Layout validator (pure function, TDD)

All tasks in this phase live in `services/dashboard-service/internal/layout/`.

### Task 9: Validator happy path + `version==1` rule

**Files:**
- Create: `services/dashboard-service/internal/layout/validator.go`
- Create: `services/dashboard-service/internal/layout/validator_test.go`

- [ ] **Step 1: Write the failing test** — `validator_test.go`

```go
package layout

import (
	"encoding/json"
	"testing"
)

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestValidateEmptyWidgets(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{}})
	l, err := Validate(raw)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if l.Version != 1 || len(l.Widgets) != 0 {
		t.Fatalf("bad layout: %+v", l)
	}
}

func TestValidateRejectsWrongVersion(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 2, "widgets": []any{}})
	_, err := Validate(raw)
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeUnsupportedSchemaVersion {
		t.Fatalf("expected %s, got %v", CodeUnsupportedSchemaVersion, err)
	}
}
```

- [ ] **Step 2: Implement minimal validator** — `validator.go`

```go
// Package layout validates the dashboard layout JSON document. It is a pure
// function: no DB, no HTTP. It enforces PRD §4.9 rules and returns stable
// error codes that the REST layer maps to JSON:API error objects.
package layout

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	shared "github.com/jtumidanski/home-hub/shared/go/dashboard"
)

type Widget struct {
	ID     uuid.UUID       `json:"id"`
	Type   string          `json:"type"`
	X      int             `json:"x"`
	Y      int             `json:"y"`
	W      int             `json:"w"`
	H      int             `json:"h"`
	Config json.RawMessage `json:"config"`
}

type Layout struct {
	Version int      `json:"version"`
	Widgets []Widget `json:"widgets"`
}

type Code string

const (
	CodeUnsupportedSchemaVersion Code = "layout.unsupported_schema_version"
	CodeWidgetCountExceeded      Code = "layout.widget_count_exceeded"
	CodeWidgetUnknownType        Code = "layout.widget_unknown_type"
	CodeWidgetBadGeometry        Code = "layout.widget_bad_geometry"
	CodeWidgetBadID              Code = "layout.widget_bad_id"
	CodeWidgetDuplicateID        Code = "layout.widget_duplicate_id"
	CodeConfigTooLarge           Code = "layout.config_too_large"
	CodeConfigTooDeep            Code = "layout.config_too_deep"
	CodeConfigNotObject          Code = "layout.config_not_object"
	CodePayloadTooLarge          Code = "layout.payload_too_large"
	CodeMalformed                Code = "layout.malformed"
)

type ValidationError struct {
	Code    Code
	Pointer string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s at %s: %s", e.Code, e.Pointer, e.Message)
}

func Validate(raw json.RawMessage) (Layout, error) {
	if len(raw) > shared.MaxLayoutBytes {
		return Layout{}, ValidationError{Code: CodePayloadTooLarge, Pointer: "/data/attributes/layout",
			Message: fmt.Sprintf("layout exceeds %d bytes", shared.MaxLayoutBytes)}
	}
	var out Layout
	if err := json.Unmarshal(raw, &out); err != nil {
		return Layout{}, ValidationError{Code: CodeMalformed, Pointer: "/data/attributes/layout", Message: err.Error()}
	}
	if out.Version != shared.LayoutSchemaVersion {
		return Layout{}, ValidationError{Code: CodeUnsupportedSchemaVersion, Pointer: "/data/attributes/layout/version",
			Message: fmt.Sprintf("expected version %d, got %d", shared.LayoutSchemaVersion, out.Version)}
	}
	// Widget validations are added in later tasks.
	return out, nil
}
```

- [ ] **Step 3: Run and commit**

```bash
cd services/dashboard-service && go test ./internal/layout/...
git add services/dashboard-service/internal/layout
git commit -m "feat(dashboard-service): add layout Validate with version rule"
```

---

### Task 10: Widget count cap + unknown-type rule

**Files:**
- Modify: `services/dashboard-service/internal/layout/validator.go`
- Modify: `services/dashboard-service/internal/layout/validator_test.go`

- [ ] **Step 1: Add failing tests**

```go
func TestValidateRejectsTooManyWidgets(t *testing.T) {
	widgets := make([]map[string]any, 41)
	for i := range widgets {
		widgets[i] = map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": i, "w": 12, "h": 1, "config": map[string]any{}}
	}
	raw := mustJSON(map[string]any{"version": 1, "widgets": widgets})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeWidgetCountExceeded {
		t.Fatalf("expected %s, got %v", CodeWidgetCountExceeded, err)
	}
}

func TestValidateRejectsUnknownWidgetType(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "foo", "x": 0, "y": 0, "w": 1, "h": 1, "config": map[string]any{}},
	}})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeWidgetUnknownType || ve.Pointer != "/data/attributes/layout/widgets/0/type" {
		t.Fatalf("got %v", err)
	}
}
```

- [ ] **Step 2: Implement** — Inside `Validate`, after unmarshaling:

```go
if len(out.Widgets) > shared.MaxWidgets {
	return Layout{}, ValidationError{Code: CodeWidgetCountExceeded, Pointer: "/data/attributes/layout/widgets",
		Message: fmt.Sprintf("at most %d widgets allowed", shared.MaxWidgets)}
}
for i, w := range out.Widgets {
	ptr := func(f string) string { return fmt.Sprintf("/data/attributes/layout/widgets/%d/%s", i, f) }
	if !shared.IsKnownWidgetType(w.Type) {
		return Layout{}, ValidationError{Code: CodeWidgetUnknownType, Pointer: ptr("type"),
			Message: fmt.Sprintf("widget type %q is not in the registry", w.Type)}
	}
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/layout/... && git add -u && git commit -m "feat(dashboard-service): enforce widget count cap and known-type allowlist"
```

---

### Task 11: Geometry + ID rules

**Files:**
- Modify: `services/dashboard-service/internal/layout/{validator.go,validator_test.go}`

- [ ] **Step 1: Add failing tests** for: `x<0`, `w<1`, `x+w>12`, non-UUID id, duplicate id.

```go
func TestValidateRejectsBadGeometry(t *testing.T) {
	cases := []struct {
		name string
		x, y, w, h int
	}{
		{"negative x", -1, 0, 1, 1},
		{"negative y", 0, -1, 1, 1},
		{"zero w", 0, 0, 0, 1},
		{"zero h", 0, 0, 1, 0},
		{"overflows grid", 10, 0, 4, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
				map[string]any{"id": uuid.New().String(), "type": "weather", "x": c.x, "y": c.y, "w": c.w, "h": c.h, "config": map[string]any{}},
			}})
			_, err := Validate(raw)
			ve, ok := err.(ValidationError)
			if !ok || ve.Code != CodeWidgetBadGeometry {
				t.Fatalf("%s: got %v", c.name, err)
			}
		})
	}
}

func TestValidateRejectsBadID(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": "not-a-uuid", "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": map[string]any{}},
	}})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || (ve.Code != CodeWidgetBadID && ve.Code != CodeMalformed) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRejectsDuplicateID(t *testing.T) {
	id := uuid.New().String()
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": id, "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": map[string]any{}},
		map[string]any{"id": id, "type": "weather", "x": 0, "y": 1, "w": 1, "h": 1, "config": map[string]any{}},
	}})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeWidgetDuplicateID {
		t.Fatalf("got %v", err)
	}
}
```

- [ ] **Step 2: Implement** — extend the widget loop inside `Validate`:

```go
seen := make(map[uuid.UUID]struct{}, len(out.Widgets))
for i, w := range out.Widgets {
	ptr := func(f string) string { return fmt.Sprintf("/data/attributes/layout/widgets/%d/%s", i, f) }
	if !shared.IsKnownWidgetType(w.Type) { /* as in task 10 */ }
	if w.X < 0 || w.Y < 0 || w.W < 1 || w.H < 1 || w.X+w.W > shared.GridColumns {
		return Layout{}, ValidationError{Code: CodeWidgetBadGeometry, Pointer: ptr(""), Message: "widget geometry out of grid"}
	}
	if w.ID == uuid.Nil {
		return Layout{}, ValidationError{Code: CodeWidgetBadID, Pointer: ptr("id"), Message: "widget id is required and must be a uuid"}
	}
	if _, dup := seen[w.ID]; dup {
		return Layout{}, ValidationError{Code: CodeWidgetDuplicateID, Pointer: ptr("id"), Message: "widget id is duplicated"}
	}
	seen[w.ID] = struct{}{}
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/layout/... && git add -u && git commit -m "feat(dashboard-service): enforce widget geometry + unique UUID id rules"
```

---

### Task 12: Config size + depth rules

**Files:**
- Modify: `services/dashboard-service/internal/layout/{validator.go,validator_test.go}`

- [ ] **Step 1: Add failing tests** for: non-object config, oversized config, too-deep config.

```go
func TestValidateRejectsNonObjectConfig(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": []any{1, 2, 3}},
	}})
	_, err := Validate(raw)
	ve, _ := err.(ValidationError)
	if ve.Code != CodeConfigNotObject {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRejectsOversizedConfig(t *testing.T) {
	big := make(map[string]string)
	for i := 0; i < 500; i++ {
		big[fmt.Sprintf("k%d", i)] = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	}
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": big},
	}})
	_, err := Validate(raw)
	ve, _ := err.(ValidationError)
	if ve.Code != CodeConfigTooLarge {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRejectsDeepConfig(t *testing.T) {
	var nest any = "leaf"
	for i := 0; i < 10; i++ {
		nest = map[string]any{"x": nest}
	}
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": nest},
	}})
	_, err := Validate(raw)
	ve, _ := err.(ValidationError)
	if ve.Code != CodeConfigTooDeep {
		t.Fatalf("got %v", err)
	}
}
```

- [ ] **Step 2: Implement** — add a helper `validateConfig(raw json.RawMessage) (Code, string)` returning the code (empty on success) and detail:

```go
func validateConfig(raw json.RawMessage) (Code, string) {
	if len(raw) == 0 {
		return "", ""
	}
	if len(raw) > shared.MaxWidgetConfigBytes {
		return CodeConfigTooLarge, fmt.Sprintf("config exceeds %d bytes", shared.MaxWidgetConfigBytes)
	}
	trimmed := bytesTrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return CodeConfigNotObject, "config must be a JSON object"
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return CodeMalformed, err.Error()
	}
	if depth(generic, 0) > shared.MaxWidgetConfigDepth {
		return CodeConfigTooDeep, fmt.Sprintf("config depth exceeds %d", shared.MaxWidgetConfigDepth)
	}
	return "", ""
}

func bytesTrimSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\n' || b[0] == '\r') {
		b = b[1:]
	}
	return b
}

func depth(v any, current int) int {
	switch t := v.(type) {
	case map[string]any:
		max := current
		for _, c := range t {
			if d := depth(c, current+1); d > max {
				max = d
			}
		}
		return max
	case []any:
		max := current
		for _, c := range t {
			if d := depth(c, current+1); d > max {
				max = d
			}
		}
		return max
	default:
		return current
	}
}
```

Wire the helper into the widget loop:

```go
if code, msg := validateConfig(w.Config); code != "" {
	return Layout{}, ValidationError{Code: code, Pointer: ptr("config"), Message: msg}
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/layout/... && git add -u && git commit -m "feat(dashboard-service): cap widget config size, depth, and require object shape"
```

---

### Task 13: Payload-size cap edge case

**Files:**
- Modify: `services/dashboard-service/internal/layout/validator_test.go`

- [ ] **Step 1: Write test** asserting the 64KB cap triggers `CodePayloadTooLarge` (build a string blob inside a valid widget config that pushes total bytes over `shared.MaxLayoutBytes`).

```go
func TestValidateRejectsOversizedPayload(t *testing.T) {
	big := strings.Repeat("x", shared.MaxLayoutBytes+10)
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1,
			"config": map[string]any{"location": map[string]any{"label": big}}},
	}})
	_, err := Validate(raw)
	ve, _ := err.(ValidationError)
	if ve.Code != CodePayloadTooLarge {
		t.Fatalf("got %v", err)
	}
}
```

- [ ] **Step 2: Run tests and commit**

```bash
go test ./internal/layout/... && git add -u && git commit -m "test(dashboard-service): cover payload-size cap in validator"
```

---

## Phase D — Entity, model, provider, administrator

### Task 14: Entity + migration

**Files:**
- Create: `services/dashboard-service/internal/dashboard/entity.go`

Reference pattern: `services/package-service/internal/tracking/entity.go`.

- [ ] **Step 1: Write the entity** — per `data-model.md §1`.

```go
package dashboard

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Entity struct {
	Id            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantId      uuid.UUID      `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
	HouseholdId   uuid.UUID      `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
	UserId        *uuid.UUID     `gorm:"type:uuid;index:idx_dashboards_scope"`
	Name          string         `gorm:"type:varchar(80);not null"`
	SortOrder     int            `gorm:"not null;default:0"`
	Layout        datatypes.JSON `gorm:"type:jsonb;not null"`
	SchemaVersion int            `gorm:"not null;default:1"`
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
}

func (Entity) TableName() string { return "dashboards" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return db.Exec(`CREATE INDEX IF NOT EXISTS idx_dashboards_household_partial
		ON dashboards (tenant_id, household_id) WHERE user_id IS NULL`).Error
}
```

- [ ] **Step 2: Wire migration** into `cmd/main.go` — add `database.SetMigrations(dashboard.Migration)` to the `database.Connect` call.

- [ ] **Step 3: Build and commit**

```bash
cd services/dashboard-service && go build ./...
git add -A services/dashboard-service
git commit -m "feat(dashboard-service): add dashboards entity + migration"
```

---

### Task 15: Immutable model + builder

**Files:**
- Create: `services/dashboard-service/internal/dashboard/model.go`
- Create: `services/dashboard-service/internal/dashboard/builder.go`
- Create: `services/dashboard-service/internal/dashboard/builder_test.go`

Reference: `services/package-service/internal/tracking/model.go` + `builder.go`.

- [ ] **Step 1: Builder test** — cover required fields, then all setters round-trip via `Build()`.

```go
package dashboard

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func TestBuilderRequiresIDs(t *testing.T) {
	_, err := NewBuilder().Build()
	if err == nil {
		t.Fatal("expected error for missing ids")
	}
}

func TestBuilderRoundTrip(t *testing.T) {
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetUserID(&uid).
		SetName("Weekend").
		SetSortOrder(2).
		SetLayout(datatypes.JSON(raw)).
		SetSchemaVersion(1).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Name() != "Weekend" || m.SortOrder() != 2 {
		t.Fatalf("round trip mismatch: %+v", m)
	}
	if m.UserID() == nil || *m.UserID() != uid {
		t.Fatalf("user id mismatch")
	}
}

func TestBuilderTrimsName(t *testing.T) {
	tid, hid := uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetName("   Home   ").
		SetLayout(datatypes.JSON(raw)).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Name() != "Home" {
		t.Fatalf("expected trim, got %q", m.Name())
	}
}

func TestBuilderRejectsEmptyName(t *testing.T) {
	tid, hid := uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	_, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetName("  ").
		SetLayout(datatypes.JSON(raw)).
		Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuilderRejectsLongName(t *testing.T) {
	tid, hid := uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	_, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetName(string(make([]byte, 81))).
		SetLayout(datatypes.JSON(raw)).
		Build()
	if err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 2: Implement model + builder** — full getters for every field, setters that return `*Builder`, `Build()` returning `(Model, error)`. Validate trim + length in `Build()`; return `ErrNameRequired` / `ErrNameTooLong`. Define errors in `builder.go`:

```go
var (
	ErrNameRequired = errors.New("name is required")
	ErrNameTooLong  = errors.New("name exceeds 80 chars")
	ErrTenantRequired = errors.New("tenant id is required")
	ErrHouseholdRequired = errors.New("household id is required")
)
```

Add `ToEntity()` on `Model` and `Make(Entity) (Model, error)` mirroring the package-service pattern.

- [ ] **Step 3: Run and commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add immutable dashboard model + builder"
```

---

### Task 16: Provider (read-side)

**Files:**
- Create: `services/dashboard-service/internal/dashboard/provider.go`

Reference: `services/package-service/internal/tracking/provider.go`.

- [ ] **Step 1: Implement read queries** — no test in this task (the test comes in processor tests, because the provider is pure SQL composition).

```go
package dashboard

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

// visibleToCaller returns all dashboards for (tenant, household) either
// household-scoped or owned by the caller user.
func visibleToCaller(tenantID, householdID, callerUserID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ? AND household_id = ? AND (user_id IS NULL OR user_id = ?)",
			tenantID, householdID, callerUserID).
			Order("sort_order ASC, created_at ASC")
	})
}

func householdScoped(tenantID, householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ? AND household_id = ? AND user_id IS NULL", tenantID, householdID).
			Order("sort_order ASC, created_at ASC")
	})
}

func maxSortOrderInScope(db *gorm.DB, tenantID, householdID uuid.UUID, userID *uuid.UUID) (int, error) {
	var result struct{ Max *int }
	q := db.Model(&Entity{}).Select("MAX(sort_order) AS max").
		Where("tenant_id = ? AND household_id = ?", tenantID, householdID)
	if userID == nil {
		q = q.Where("user_id IS NULL")
	} else {
		q = q.Where("user_id = ?", *userID)
	}
	if err := q.Scan(&result).Error; err != nil {
		return 0, err
	}
	if result.Max == nil {
		return -1, nil
	}
	return *result.Max, nil
}

func countHouseholdScoped(db *gorm.DB, tenantID, householdID uuid.UUID) (int64, error) {
	var n int64
	err := db.Model(&Entity{}).
		Where("tenant_id = ? AND household_id = ? AND user_id IS NULL", tenantID, householdID).
		Count(&n).Error
	return n, err
}
```

- [ ] **Step 2: Build and commit**

```bash
go build ./... && git add -A && git commit -m "feat(dashboard-service): add provider read queries"
```

---

### Task 17: Administrator (write-side)

**Files:**
- Create: `services/dashboard-service/internal/dashboard/administrator.go`

Reference: `services/package-service/internal/tracking/administrator.go`.

- [ ] **Step 1: Implement writes**:

```go
package dashboard

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func insert(db *gorm.DB, e Entity) (Entity, error) {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateFields(db *gorm.DB, id uuid.UUID, fields map[string]any) (Entity, error) {
	fields["updated_at"] = time.Now().UTC()
	if err := db.Model(&Entity{}).Where("id = ?", id).Updates(fields).Error; err != nil {
		return Entity{}, err
	}
	var out Entity
	if err := db.First(&out, "id = ?", id).Error; err != nil {
		return Entity{}, err
	}
	return out, nil
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Delete(&Entity{}, "id = ?", id).Error
}

func updateSortOrders(db *gorm.DB, updates map[uuid.UUID]int) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for id, order := range updates {
			if err := tx.Model(&Entity{}).Where("id = ?", id).
				Updates(map[string]any{"sort_order": order, "updated_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// layoutAsJSON casts a validated Layout into datatypes.JSON safely.
func layoutAsJSON(raw []byte) datatypes.JSON { return datatypes.JSON(raw) }
```

- [ ] **Step 2: Build and commit**

```bash
go build ./... && git add -A && git commit -m "feat(dashboard-service): add administrator write functions"
```

---

## Phase E — Processor

All tasks in this phase touch `services/dashboard-service/internal/dashboard/processor.go` and `processor_test.go`. Use the shared service-test harness (`shared/go/testing`) — check `services/package-service/internal/tracking/rest_test.go` for the setup shape when building the integration-style tests in later tasks. For this phase, prefer unit-style tests with a real in-memory Postgres via the harness, because processors compose provider + administrator tightly.

### Task 18: Processor struct + Create

- [ ] **Step 1: Failing test** — `processor_test.go`

```go
func TestProcessorCreateHousehold(t *testing.T) {
	db := testutil.OpenTestDB(t)
	must(t, Migration(db))
	p := NewProcessor(logrus.New(), context.Background(), db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	m, err := p.Create(tid, hid, uid, CreateAttrs{
		Name:  "Home",
		Scope: "household",
		Layout: json.RawMessage(`{"version":1,"widgets":[]}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.Name() != "Home" || m.UserID() != nil {
		t.Fatalf("bad model: %+v", m)
	}
}
```

(Use whatever harness the repo exposes. If `shared/go/testing` exports an `OpenTestDB`, import it. Otherwise look at an existing processor_test for the harness pattern and copy it.)

- [ ] **Step 2: Implement** — minimal `Processor` with `Create`:

```go
package dashboard

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/layout"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

type CreateAttrs struct {
	Name      string
	Scope     string // "household" | "user"
	Layout    json.RawMessage
	SortOrder *int
}

func (p *Processor) Create(tenantID, householdID, callerUserID uuid.UUID, attrs CreateAttrs) (Model, error) {
	var userID *uuid.UUID
	switch attrs.Scope {
	case "household":
		userID = nil
	case "user":
		u := callerUserID
		userID = &u
	default:
		return Model{}, ErrInvalidScope
	}

	layoutBytes := attrs.Layout
	if len(layoutBytes) == 0 {
		layoutBytes = json.RawMessage(`{"version":1,"widgets":[]}`)
	}
	if _, err := layout.Validate(layoutBytes); err != nil {
		return Model{}, err
	}

	sortOrder := 0
	if attrs.SortOrder != nil {
		sortOrder = *attrs.SortOrder
	} else {
		max, err := maxSortOrderInScope(p.db.WithContext(p.ctx), tenantID, householdID, userID)
		if err != nil {
			return Model{}, err
		}
		sortOrder = max + 1
	}

	e := Entity{
		TenantId:      tenantID,
		HouseholdId:   householdID,
		UserId:        userID,
		Name:          attrs.Name,
		SortOrder:     sortOrder,
		Layout:        datatypes.JSON(layoutBytes),
		SchemaVersion: 1,
	}
	// Trim/length validated via Make on read-back.
	e.Name = trimName(e.Name)
	if err := validateNameLen(e.Name); err != nil {
		return Model{}, err
	}

	saved, err := insert(p.db.WithContext(p.ctx), e)
	if err != nil {
		return Model{}, err
	}
	return Make(saved)
}
```

Add helpers `trimName`, `validateNameLen`, `ErrInvalidScope` (in `processor.go` or a new `errors.go`).

- [ ] **Step 3: Commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add processor Create with scope + layout validation"
```

---

### Task 19: List + Get

- [ ] **Step 1: Failing tests** — cover (a) List returns household + caller's user dashboards but not another user's, (b) Get returns 404 equivalent for rows invisible to caller.

```go
func TestProcessorListScopesVisibility(t *testing.T) {
	// seed 3 rows: household, caller-user, other-user.
	// expect List to return 2: household + caller-user.
}

func TestProcessorGetBlocksInvisibleRow(t *testing.T) {
	// seed: one other-user dashboard.
	// GetByID should return ErrNotFound.
}
```

- [ ] **Step 2: Implement** `List(tenantID, householdID, callerUserID) ([]Model, error)` via `visibleToCaller`; `GetByID(id, tenantID, householdID, callerUserID)` that calls `getByID`, checks `TenantID==tenantID && HouseholdID==householdID && (UserId==nil || *UserId==callerUserID)`, else returns `ErrNotFound`.

```go
var ErrNotFound = errors.New("not found")
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add processor List/GetByID with visibility filter"
```

---

### Task 20: Update + Delete

- [ ] **Step 1: Failing tests** — (a) household-scoped update by any member, (b) user-scoped update by non-owner returns `ErrForbidden`, (c) same rules for Delete.

- [ ] **Step 2: Implement**

```go
var ErrForbidden = errors.New("forbidden")

type UpdateAttrs struct {
	Name      *string
	Layout    *json.RawMessage
	SortOrder *int
}

func (p *Processor) Update(id, tenantID, householdID, callerUserID uuid.UUID, attrs UpdateAttrs) (Model, error) {
	row, err := p.requireEditable(id, tenantID, householdID, callerUserID)
	if err != nil {
		return Model{}, err
	}
	fields := map[string]any{}
	if attrs.Name != nil {
		name := trimName(*attrs.Name)
		if err := validateNameLen(name); err != nil {
			return Model{}, err
		}
		fields["name"] = name
	}
	if attrs.Layout != nil {
		if _, err := layout.Validate(*attrs.Layout); err != nil {
			return Model{}, err
		}
		fields["layout"] = datatypes.JSON(*attrs.Layout)
	}
	if attrs.SortOrder != nil {
		fields["sort_order"] = *attrs.SortOrder
	}
	if len(fields) == 0 {
		return Make(row) // no-op update returns current
	}
	updated, err := updateFields(p.db.WithContext(p.ctx), id, fields)
	if err != nil {
		return Model{}, err
	}
	return Make(updated)
}

func (p *Processor) Delete(id, tenantID, householdID, callerUserID uuid.UUID) error {
	if _, err := p.requireEditable(id, tenantID, householdID, callerUserID); err != nil {
		return err
	}
	return deleteByID(p.db.WithContext(p.ctx), id)
}

// requireEditable returns the row and ErrNotFound/ErrForbidden per PRD §4.10.
// Household rows: any member can edit. User rows: owner only.
func (p *Processor) requireEditable(id, tenantID, householdID, callerUserID uuid.UUID) (Entity, error) {
	row, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Entity{}, ErrNotFound
	}
	if row.TenantId != tenantID || row.HouseholdId != householdID {
		return Entity{}, ErrNotFound
	}
	if row.UserId != nil && *row.UserId != callerUserID {
		return Entity{}, ErrForbidden
	}
	return row, nil
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add processor Update/Delete with scope enforcement"
```

---

### Task 21: Bulk reorder (single-scope rule)

- [ ] **Step 1: Failing tests**: (a) accepts single-scope ids, (b) rejects mixed scope with `ErrMixedScope`, (c) rejects ids not visible to caller.

- [ ] **Step 2: Implement** `Reorder(tenantID, householdID, callerUserID, pairs []ReorderPair) ([]Model, error)`:

```go
type ReorderPair struct {
	ID        uuid.UUID
	SortOrder int
}

var ErrMixedScope = errors.New("reorder requires single scope")

func (p *Processor) Reorder(tenantID, householdID, callerUserID uuid.UUID, pairs []ReorderPair) ([]Model, error) {
	if len(pairs) == 0 {
		return []Model{}, nil
	}
	ids := make([]uuid.UUID, 0, len(pairs))
	for _, pr := range pairs {
		ids = append(ids, pr.ID)
	}
	var rows []Entity
	if err := p.db.WithContext(p.ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != len(pairs) {
		return nil, ErrNotFound
	}
	var scope string
	for _, r := range rows {
		if r.TenantId != tenantID || r.HouseholdId != householdID {
			return nil, ErrNotFound
		}
		var rowScope string
		if r.UserId == nil {
			rowScope = "household"
		} else {
			if *r.UserId != callerUserID {
				return nil, ErrNotFound
			}
			rowScope = "user"
		}
		if scope == "" {
			scope = rowScope
		} else if scope != rowScope {
			return nil, ErrMixedScope
		}
	}
	upd := map[uuid.UUID]int{}
	for _, pr := range pairs {
		upd[pr.ID] = pr.SortOrder
	}
	if err := updateSortOrders(p.db.WithContext(p.ctx), upd); err != nil {
		return nil, err
	}
	list, err := visibleToCaller(tenantID, householdID, callerUserID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	out := make([]Model, 0, len(list))
	for _, e := range list {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add processor Reorder with single-scope guard"
```

---

### Task 22: Promote (user → household)

- [ ] **Step 1: Failing tests**: (a) owner promotes OK, (b) non-owner gets `ErrForbidden`, (c) already-household returns `ErrAlreadyHousehold`.

- [ ] **Step 2: Implement**

```go
var ErrAlreadyHousehold = errors.New("dashboard already household-scoped")

func (p *Processor) Promote(id, tenantID, householdID, callerUserID uuid.UUID) (Model, error) {
	row, err := p.requireEditable(id, tenantID, householdID, callerUserID)
	if err != nil {
		return Model{}, err
	}
	if row.UserId == nil {
		return Model{}, ErrAlreadyHousehold
	}
	updated, err := updateFields(p.db.WithContext(p.ctx), id, map[string]any{"user_id": nil})
	if err != nil {
		return Model{}, err
	}
	return Make(updated)
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add processor Promote (user→household)"
```

---

### Task 23: Copy-to-mine (deep copy, regenerate widget IDs)

- [ ] **Step 1: Failing tests**: (a) copies a household dashboard into caller's user scope with a new id, (b) all widget instance ids are freshly generated (asserted distinct from source), (c) `sort_order = max+1` within caller's user scope, (d) can't copy a user-scoped dashboard.

- [ ] **Step 2: Implement** — add a helper that regenerates widget ids inside the JSON:

```go
var ErrNotCopyable = errors.New("only household dashboards can be copied to mine")

func (p *Processor) CopyToMine(id, tenantID, householdID, callerUserID uuid.UUID) (Model, error) {
	row, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil || row.TenantId != tenantID || row.HouseholdId != householdID {
		return Model{}, ErrNotFound
	}
	if row.UserId != nil {
		return Model{}, ErrNotCopyable
	}

	regenerated, err := regenerateWidgetIDs([]byte(row.Layout))
	if err != nil {
		return Model{}, err
	}

	ownerID := callerUserID
	max, err := maxSortOrderInScope(p.db.WithContext(p.ctx), tenantID, householdID, &ownerID)
	if err != nil {
		return Model{}, err
	}

	copyName := row.Name + " (mine)"
	if len(copyName) > 80 {
		copyName = copyName[:80]
	}

	newRow := Entity{
		TenantId:      tenantID,
		HouseholdId:   householdID,
		UserId:        &ownerID,
		Name:          copyName,
		SortOrder:     max + 1,
		Layout:        datatypes.JSON(regenerated),
		SchemaVersion: row.SchemaVersion,
	}
	saved, err := insert(p.db.WithContext(p.ctx), newRow)
	if err != nil {
		return Model{}, err
	}
	return Make(saved)
}

func regenerateWidgetIDs(raw []byte) ([]byte, error) {
	var doc struct {
		Version int                      `json:"version"`
		Widgets []map[string]any         `json:"widgets"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	for i := range doc.Widgets {
		doc.Widgets[i]["id"] = uuid.New().String()
	}
	return json.Marshal(doc)
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add processor CopyToMine with regenerated widget ids"
```

---

### Task 24: Seed with Postgres advisory lock

- [ ] **Step 1: Failing tests**: (a) first call on empty household returns created=true + the new row, (b) second call returns created=false + the existing list, (c) two concurrent goroutines both call Seed — exactly one creates a row, the other sees the existing list.

```go
func TestProcessorSeedRace(t *testing.T) {
	db := testutil.OpenTestDB(t)
	must(t, Migration(db))
	p := NewProcessor(logrus.New(), context.Background(), db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	var wg sync.WaitGroup
	var createdCount int32
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			res, err := p.Seed(tid, hid, uid, "Home", layoutJSON)
			if err != nil {
				t.Error(err)
				return
			}
			if res.Created {
				atomic.AddInt32(&createdCount, 1)
			}
		}()
	}
	wg.Wait()
	if createdCount != 1 {
		t.Fatalf("expected exactly one create, got %d", createdCount)
	}
}
```

- [ ] **Step 2: Implement** — Use `pg_advisory_xact_lock` keyed on a hash of tenant+household:

```go
type SeedResult struct {
	Created   bool
	Dashboard Model
	Existing  []Model
}

func (p *Processor) Seed(tenantID, householdID, callerUserID uuid.UUID, name string, layoutJSON json.RawMessage) (SeedResult, error) {
	name = trimName(name)
	if err := validateNameLen(name); err != nil {
		return SeedResult{}, err
	}
	if _, err := layout.Validate(layoutJSON); err != nil {
		return SeedResult{}, err
	}

	var out SeedResult
	err := p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		key := seedLockKey(tenantID, householdID)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", key).Error; err != nil {
			return err
		}
		count, err := countHouseholdScoped(tx, tenantID, householdID)
		if err != nil {
			return err
		}
		if count > 0 {
			list, err := visibleToCaller(tenantID, householdID, callerUserID)(tx)()
			if err != nil {
				return err
			}
			for _, e := range list {
				m, err := Make(e)
				if err != nil {
					return err
				}
				out.Existing = append(out.Existing, m)
			}
			out.Created = false
			return nil
		}
		e := Entity{
			TenantId:      tenantID,
			HouseholdId:   householdID,
			UserId:        nil,
			Name:          name,
			SortOrder:     0,
			Layout:        datatypes.JSON(layoutJSON),
			SchemaVersion: 1,
		}
		saved, err := insert(tx, e)
		if err != nil {
			return err
		}
		m, err := Make(saved)
		if err != nil {
			return err
		}
		out.Dashboard = m
		out.Created = true
		return nil
	})
	return out, err
}

func seedLockKey(tenantID, householdID uuid.UUID) int64 {
	var combined [32]byte
	copy(combined[:16], tenantID[:])
	copy(combined[16:], householdID[:])
	sum := sha256.Sum256(combined[:])
	return int64(binary.BigEndian.Uint64(sum[:8]))
}
```

Add imports for `crypto/sha256`, `encoding/binary`, `sync/atomic` (in test only).

- [ ] **Step 3: Commit**

```bash
go test ./internal/dashboard/... && git add -A && git commit -m "feat(dashboard-service): add Seed with advisory lock for idempotent creation"
```

---

## Phase F — REST layer

All files in `services/dashboard-service/internal/dashboard/`. Reference: `services/package-service/internal/tracking/{rest,resource}.go`.

### Task 25: REST models

**Files:**
- Create: `services/dashboard-service/internal/dashboard/rest.go`

- [ ] **Step 1: Define types** — per `api-contracts.md`.

```go
package dashboard

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id            uuid.UUID       `json:"-"`
	Name          string          `json:"name"`
	Scope         string          `json:"scope"`
	SortOrder     int             `json:"sortOrder"`
	Layout        json.RawMessage `json:"layout"`
	SchemaVersion int             `json:"schemaVersion"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (r RestModel) GetName() string       { return "dashboards" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	v, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	r.Id = v
	return nil
}

type CreateRequest struct {
	Name      string          `json:"name"`
	Scope     string          `json:"scope"`
	Layout    json.RawMessage `json:"layout"`
	SortOrder *int            `json:"sortOrder"`
}

func (CreateRequest) GetName() string        { return "dashboards" }
func (CreateRequest) GetID() string           { return "" }
func (*CreateRequest) SetID(_ string) error   { return nil }

type UpdateRequest struct {
	Name      *string          `json:"name"`
	Layout    *json.RawMessage `json:"layout"`
	SortOrder *int             `json:"sortOrder"`
}

func (UpdateRequest) GetName() string        { return "dashboards" }
func (UpdateRequest) GetID() string           { return "" }
func (*UpdateRequest) SetID(_ string) error   { return nil }

type ReorderRequest struct {
	// Because api2go's input decoder expects a single resource, bulk reorder
	// is posted as a plain JSON body (non-JSON:API). See resource.go for the
	// custom handler wiring.
	Entries []ReorderEntry `json:"data"`
}

type ReorderEntry struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sortOrder"`
}

type SeedRequest struct {
	Name   string          `json:"name"`
	Layout json.RawMessage `json:"layout"`
}

func (SeedRequest) GetName() string       { return "dashboards" }
func (SeedRequest) GetID() string          { return "" }
func (*SeedRequest) SetID(_ string) error  { return nil }

func Transform(m Model) (RestModel, error) {
	scope := "household"
	if m.UserID() != nil {
		scope = "user"
	}
	return RestModel{
		Id:            m.Id(),
		Name:          m.Name(),
		Scope:         scope,
		SortOrder:     m.SortOrder(),
		Layout:        json.RawMessage(m.LayoutJSON()),
		SchemaVersion: m.SchemaVersion(),
		CreatedAt:     m.CreatedAt(),
		UpdatedAt:     m.UpdatedAt(),
	}, nil
}
```

(Add `Model.LayoutJSON()` accessor in `model.go` returning `[]byte(m.layout)`.)

- [ ] **Step 2: Commit**

```bash
go build ./... && git add -A && git commit -m "feat(dashboard-service): add REST models + Transform"
```

---

### Task 26: List + Get handlers

**Files:**
- Create: `services/dashboard-service/internal/dashboard/resource.go`

- [ ] **Step 1: Write a rest_test.go integration test** for List returning only caller-visible dashboards (use the service-test harness pattern from `services/package-service/internal/tracking/rest_test.go`).

- [ ] **Step 2: Implement resource.go** — register routes, implement List + Get with the JSON:API response shape. Pattern copied from `tracking/resource.go`:

```go
package dashboard

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)
		rihSeed := server.RegisterInputHandler[SeedRequest](l)(si)

		api.HandleFunc("/dashboards", rh("ListDashboards", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/dashboards/order", rh("ReorderDashboards", reorderHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/dashboards/seed", rihSeed("SeedDashboard", seedHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/dashboards", rihCreate("CreateDashboard", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/dashboards/{id}", rh("GetDashboard", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/dashboards/{id}", rihUpdate("UpdateDashboard", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/dashboards/{id}", rh("DeleteDashboard", deleteHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/dashboards/{id}/promote", rh("PromoteDashboard", promoteHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/dashboards/{id}/copy-to-mine", rh("CopyDashboardToMine", copyToMineHandler(db))).Methods(http.MethodPost)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			list, err := proc.List(t.Id(), t.HouseholdId(), t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("list dashboards")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			out := make([]RestModel, 0, len(list))
			for _, m := range list {
				rest, err := Transform(m)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				out = append(out, rest)
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(out)
		}
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.GetByID(id, t.Id(), t.HouseholdId(), t.UserId())
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "")
						return
					}
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}
```

(Stubs for the rest of the handlers come in subsequent tasks.)

- [ ] **Step 3: Wire into cmd/main.go** — add `dashboard.InitializeRoutes(db)(l, si, api)` to `AddRouteInitializer`.

- [ ] **Step 4: Commit**

```bash
go test ./... && git add -A && git commit -m "feat(dashboard-service): add resource + List/Get handlers"
```

---

### Task 27: Create / Update / Delete handlers

- [ ] **Step 1: Failing rest_test.go integration tests** covering:
  - Create household scope → 201 + body.
  - Create user scope → 201 + `scope: "user"`.
  - Create with `layout` missing an id → 422 with `layout.widget_bad_id` code.
  - Update another user's dashboard → 403.
  - Delete household dashboard as any member → 204.

- [ ] **Step 2: Implement handlers** — map `layout.ValidationError` to a JSON:API 422 with `code`, `pointer`, `detail`. Map `ErrNotFound`→404, `ErrForbidden`→403, `ErrInvalidScope`→400.

Helper (put it in `resource.go`):

```go
func writeLayoutError(w http.ResponseWriter, ve layout.ValidationError) {
	server.WriteJSONAPIError(w, http.StatusUnprocessableEntity, string(ve.Code), "Layout validation failed", ve.Message, ve.Pointer)
}
```

If `server.WriteJSONAPIError` does not yet exist, add it to `shared/go/server/` — otherwise inline the JSON:API error envelope write per `api-contracts.md §"Error envelope"`.

- [ ] **Step 3: Commit**

```bash
go test ./... && git add -A && git commit -m "feat(dashboard-service): add Create/Update/Delete handlers with stable error codes"
```

---

### Task 28: Reorder handler

- [ ] **Step 1: Failing rest_test.go** — asserts single-scope success + mixed-scope 400.

- [ ] **Step 2: Implement `reorderHandler`** — parse the body as `ReorderRequest` (plain JSON, not JSON:API-wrapped), call `proc.Reorder`, return a slice response of the resulting `RestModel`. Map `ErrMixedScope` → 400.

- [ ] **Step 3: Commit**

```bash
go test ./... && git add -A && git commit -m "feat(dashboard-service): add bulk reorder handler"
```

---

### Task 29: Promote + CopyToMine handlers

- [ ] **Step 1: Failing rest_test.go** — covers promote 200 / 409 `ErrAlreadyHousehold`, copy-to-mine 201 with sort_order = max+1.

- [ ] **Step 2: Implement** — `promoteHandler` → 200 on success, 409 on `ErrAlreadyHousehold`, 403 on non-owner. `copyToMineHandler` → 201, 400 on `ErrNotCopyable`, 404 on not-visible.

- [ ] **Step 3: Commit**

```bash
go test ./... && git add -A && git commit -m "feat(dashboard-service): add promote + copy-to-mine handlers"
```

---

### Task 30: Seed handler

- [ ] **Step 1: Failing rest_test.go** — (a) first call returns 201 with single resource body, (b) second call returns 200 with slice body; parallel goroutines only one gets 201.

- [ ] **Step 2: Implement `seedHandler`** — call `proc.Seed`. If `res.Created`, return `MarshalCreatedResponse` with single resource. Else return `MarshalSliceResponse` with the existing list.

- [ ] **Step 3: Commit**

```bash
go test ./... && git add -A && git commit -m "feat(dashboard-service): add seed handler (idempotent, advisory-lock-guarded)"
```

---

## Phase G — Retention + Kafka consumer

### Task 31: Retention wire + noop reaper

**Files:**
- Create: `services/dashboard-service/internal/retention/wire.go`
- Create: `services/dashboard-service/internal/retention/handlers.go`

Reference: `services/package-service/internal/retention/wire.go` + `handlers.go`.

- [ ] **Step 1: Implement `DashboardsRetention`** — category `CatDashboardDashboards`. `DiscoverScopes` returns every distinct (tenant_id, household_id) pair. `Reap` is a no-op (`return sr.ReapResult{}, nil`) — the plumbing exists, but v1 never prunes.

- [ ] **Step 2: Implement `Setup`** identical shape to package-service's — also register `AuditTrim` for `CatSystemRetentionAudit`.

- [ ] **Step 3: Wire** into `cmd/main.go`:

```go
if _, err := retention.Setup(ctx, l, db, router, cfg.AccountServiceURL, cfg.InternalToken, cfg.RetentionInterval); err != nil {
	l.WithError(err).Fatal("retention setup failed")
}
```

- [ ] **Step 4: Commit**

```bash
go build ./... && git add -A && git commit -m "feat(dashboard-service): register retention category with no-op reaper"
```

---

### Task 32: `UserDeletedEvent` consumer handler

**Files:**
- Create: `services/dashboard-service/internal/events/handler.go`
- Create: `services/dashboard-service/internal/events/handler_test.go`

- [ ] **Step 1: Failing test** — the handler decodes an envelope and deletes every user-scoped dashboard for the (tenant, user) pair; idempotent on re-receive.

```go
func TestHandleUserDeletedCascades(t *testing.T) {
	db := testutil.OpenTestDB(t)
	must(t, dashboard.Migration(db))
	tid, uid, other := uuid.New(), uuid.New(), uuid.New()
	hid := uuid.New()
	seedUserDashboard(db, tid, hid, uid)
	seedUserDashboard(db, tid, hid, other)
	h := NewHandler(logrus.New(), db)

	evt := events.UserDeletedEvent{TenantID: tid, UserID: uid, DeletedAt: time.Now()}
	env, _ := events.NewEnvelope(events.TypeUserDeleted, evt)
	raw, _ := json.Marshal(env)
	if err := h.Dispatch(context.Background(), kafka.Message{Value: raw}); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&dashboard.Entity{}).Where("user_id = ?", uid).Count(&n)
	if n != 0 {
		t.Fatalf("expected cascade delete, still %d", n)
	}
	// re-dispatch is a no-op
	if err := h.Dispatch(context.Background(), kafka.Message{Value: raw}); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Implement**

```go
package events

import (
	"context"
	"encoding/json"

	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/dashboard"
	sharedevents "github.com/jtumidanski/home-hub/shared/go/events"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handler struct {
	l  logrus.FieldLogger
	db *gorm.DB
}

func NewHandler(l logrus.FieldLogger, db *gorm.DB) *Handler {
	return &Handler{l: l, db: db}
}

func (h *Handler) Dispatch(ctx context.Context, msg kafka.Message) error {
	var env sharedevents.Envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		h.l.WithError(err).Warn("skipping malformed envelope")
		return nil
	}
	switch env.Type {
	case sharedevents.TypeUserDeleted:
		var evt sharedevents.UserDeletedEvent
		if err := json.Unmarshal(env.Payload, &evt); err != nil {
			h.l.WithError(err).Warn("skipping malformed UserDeletedEvent")
			return nil
		}
		res := h.db.WithContext(ctx).
			Where("tenant_id = ? AND user_id = ?", evt.TenantID, evt.UserID).
			Delete(&dashboard.Entity{})
		if res.Error != nil {
			return res.Error
		}
		h.l.WithField("tenant_id", evt.TenantID).WithField("user_id", evt.UserID).
			WithField("rows", res.RowsAffected).Info("user cascade")
	default:
		// ignore unknown types
	}
	return nil
}
```

- [ ] **Step 3: Commit**

```bash
go test ./internal/events/... && git add -A && git commit -m "feat(dashboard-service): add UserDeletedEvent cascade handler"
```

---

### Task 33: Wire consumer into `cmd/main.go`

- [ ] **Step 1:** In `cmd/main.go`, after router setup, start the consumer:

```go
import (
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/events"
	kcons "github.com/jtumidanski/home-hub/shared/go/kafka/consumer"
)

// ... after server.New(...)
h := events.NewHandler(l, db)
mgr := kcons.New(kcons.Config{
	Brokers: cfg.KafkaBrokers,
	Topic:   cfg.TopicUserLifecycle,
	GroupID: cfg.KafkaConsumerGroup,
}, h.Dispatch, l)
go mgr.Run(ctx)
defer mgr.Close()
```

- [ ] **Step 2:** `go build ./...` succeeds.

- [ ] **Step 3: Commit**

```bash
go build ./... && git add -A && git commit -m "feat(dashboard-service): wire kafka consumer for user-lifecycle topic"
```

---

## Phase H — `account-service` changes

### Task 34: `household_preferences` entity, builder, model

**Files:**
- Create: `services/account-service/internal/householdpreference/entity.go`
- Create: `services/account-service/internal/householdpreference/model.go`
- Create: `services/account-service/internal/householdpreference/builder.go`
- Create: `services/account-service/internal/householdpreference/builder_test.go`

Reference: `services/account-service/internal/preference/{entity,model,builder}.go`.

- [ ] **Step 1: Entity**

```go
package householdpreference

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                 uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId           uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	UserId             uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	HouseholdId        uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	DefaultDashboardId *uuid.UUID `gorm:"type:uuid"`
	CreatedAt          time.Time  `gorm:"not null"`
	UpdatedAt          time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "household_preferences" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }
```

- [ ] **Step 2: Model + builder + `Make` + `ToEntity`** — same shape as `preference` package. Write builder tests (required fields, round-trip).

- [ ] **Step 3: Wire migration** in `services/account-service/cmd/main.go` (add to `database.SetMigrations(...)` call).

- [ ] **Step 4: Commit**

```bash
cd services/account-service && go test ./internal/householdpreference/...
git add -A && git commit -m "feat(account-service): add household_preferences entity/model/builder"
```

---

### Task 35: `household_preferences` processor + REST

**Files:**
- Create: `services/account-service/internal/householdpreference/provider.go`
- Create: `services/account-service/internal/householdpreference/administrator.go`
- Create: `services/account-service/internal/householdpreference/processor.go`
- Create: `services/account-service/internal/householdpreference/rest.go`
- Create: `services/account-service/internal/householdpreference/resource.go`
- Create: `services/account-service/internal/householdpreference/rest_test.go`

- [ ] **Step 1: Failing rest_test.go**:
  - `GET /api/v1/household-preferences` on new user/household → 200 with auto-created row, `defaultDashboardId: null`.
  - `PATCH /api/v1/household-preferences/{id}` with `{defaultDashboardId: "<uuid>"}` → 200 with updated row.
  - PATCH with `null` explicitly → clears the field.

- [ ] **Step 2: Implement**:
  - `Processor.FindOrCreate(tenantID, userID, householdID)` — if no row for (t,u,h), create with `DefaultDashboardId: nil`.
  - `Processor.SetDefaultDashboard(id, *uuid.UUID)` — accepts nil to clear.
  - REST resource type: `householdPreferences`. Attributes: `defaultDashboardId`, `createdAt`, `updatedAt`.
  - `resource.go` routes:
    - `GET /household-preferences` → FindOrCreate and return slice-of-one.
    - `PATCH /household-preferences/{id}` → parse `UpdateRequest { DefaultDashboardId *string }` where `nil` means "not present" and the literal JSON `null` means "clear". Map `*string` pointer semantics accordingly (use a struct with an `IsSet bool` if the JSON:API input handler cannot distinguish).

Tip for PATCH disambiguation: register the body as `map[string]any` via the custom unmarshal path, or use a pointer-to-pointer (`**string`). Pick whichever matches the existing style; see `preference/resource.go`'s `UpdateRequest` as a baseline and adapt.

- [ ] **Step 3: Wire routes** in `services/account-service/cmd/main.go`: add `householdpreference.InitializeRoutes(db)(l, si, api)` alongside the existing route initializers.

- [ ] **Step 4: Commit**

```bash
go test ./internal/householdpreference/...
git add -A && git commit -m "feat(account-service): add household_preferences REST with default_dashboard_id"
```

---

### Task 36: `UserDeletedEvent` producer + internal endpoint

**Files:**
- Modify: `services/account-service/go.mod` (add shared/go/{kafka,events})
- Create: `services/account-service/internal/userlifecycle/producer.go`
- Create: `services/account-service/internal/userlifecycle/resource.go`
- Create: `services/account-service/internal/userlifecycle/resource_test.go`

- [ ] **Step 1: Failing test** — POST to internal endpoint `POST /api/v1/internal/users/{id}/deleted` with an `X-Internal-Token` header:
  - Deletes the user's `household_preferences` rows for that tenant.
  - Invokes the producer (assert via a stub producer injected into the handler).
  - Returns 204.
  - Rejects requests missing the internal token with 401.

- [ ] **Step 2: Implement**

```go
package userlifecycle

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/account-service/internal/householdpreference"
	sharedevents "github.com/jtumidanski/home-hub/shared/go/events"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Producer interface {
	Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error
}

type Config struct {
	Topic         string
	InternalToken string
}

func InitializeRoutes(db *gorm.DB, prod Producer, cfg Config) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		api.HandleFunc("/internal/users/{id}/deleted", deletedHandler(l, db, prod, cfg)).Methods(http.MethodPost)
	}
}

func deletedHandler(l logrus.FieldLogger, db *gorm.DB, prod Producer, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.InternalToken == "" || r.Header.Get("X-Internal-Token") != cfg.InternalToken {
			server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "")
			return
		}
		userIDStr := mux.Vars(r)["id"]
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "Bad Request", err.Error())
			return
		}
		t := tenantctx.MustFromContext(r.Context())
		if err := db.WithContext(r.Context()).
			Where("tenant_id = ? AND user_id = ?", t.Id(), userID).
			Delete(&householdpreference.Entity{}).Error; err != nil {
			l.WithError(err).Error("failed to delete household_preferences")
			server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
			return
		}
		evt := sharedevents.UserDeletedEvent{TenantID: t.Id(), UserID: userID, DeletedAt: time.Now().UTC()}
		env, err := sharedevents.NewEnvelope(sharedevents.TypeUserDeleted, evt)
		if err != nil {
			l.WithError(err).Warn("envelope build")
		} else {
			payload, _ := json.Marshal(env)
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(userID.ID()))
			if err := prod.Produce(r.Context(), cfg.Topic, key, payload, nil); err != nil {
				l.WithError(err).Warn("produce UserDeletedEvent failed; event dropped")
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

var _ = errors.New // keep imports tidy if added later
```

Note: this endpoint sits behind the **internal token** guard instead of the user JWT middleware, because the caller is a service-to-service action. Place it in a router subtree that does NOT apply the JWT middleware. Check `account-service/cmd/main.go` for how to split internal routes from authed routes — if there is no existing split, register this handler on the main router before the `/api/v1` JWT subrouter.

- [ ] **Step 3: Wire into main.go** — construct a `*producer.Producer` from `shared/go/kafka/producer`, instantiate routes with it, register.

- [ ] **Step 4: Commit**

```bash
go test ./internal/userlifecycle/...
git add -A && git commit -m "feat(account-service): add internal user-deleted endpoint and UserDeletedEvent producer"
```

---

### Task 37: Add config env + docs for account-service Kafka

- [ ] **Step 1:** Extend `services/account-service/internal/config/config.go` with `KafkaBrokers []string`, `TopicUserLifecycle string`. Defaults: `kafka-broker.kafka.svc.cluster.local:9092` and `home-hub.user.lifecycle`.

- [ ] **Step 2:** Build the service.

- [ ] **Step 3: Commit**

```bash
go build ./... && git add -A && git commit -m "feat(account-service): add kafka config"
```

---

## Phase I — Infrastructure

### Task 38: docker-compose + nginx

**Files:**
- Modify: `deploy/compose/docker-compose.yml`
- Modify: `deploy/compose/nginx.conf`

- [ ] **Step 1:** Add a `dashboard-service` compose entry mirroring `package-service`:

```yaml
  dashboard-service:
    container_name: hh-dashboard
    build:
      context: ../..
      dockerfile: services/dashboard-service/Dockerfile
    expose:
      - "8080"
    environment:
      - DB_HOST=${DB_HOST:-postgres.home}
      - DB_PORT=${DB_PORT:-5432}
      - DB_USER=${DB_USER:-home_hub}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME:-home_hub}
      - PORT=8080
      - JWKS_URL=http://auth-service:8080/api/v1/auth/.well-known/jwks.json
      - ACCOUNT_SERVICE_URL=http://account-service:8080
      - INTERNAL_SERVICE_TOKEN=${INTERNAL_SERVICE_TOKEN}
      - BOOTSTRAP_SERVERS=${BOOTSTRAP_SERVERS:-kafka-broker.kafka.svc.cluster.local:9092}
      - EVENT_TOPIC_USER_LIFECYCLE=${EVENT_TOPIC_USER_LIFECYCLE:-home-hub.user.lifecycle}
      - KAFKA_CONSUMER_GROUP=dashboard-service
      - RETENTION_INTERVAL=${RETENTION_INTERVAL:-6h}
```

Add `dashboard-service` to the nginx container's `depends_on` list (the block at line ~17 of docker-compose.yml).

Add kafka env to `account-service`'s environment block: `BOOTSTRAP_SERVERS`, `EVENT_TOPIC_USER_LIFECYCLE`.

- [ ] **Step 2:** In `deploy/compose/nginx.conf`, add:

```nginx
    # Dashboard service
    location /api/v1/dashboards {
        set $dashboard_upstream http://dashboard-service:8080;
        proxy_pass $dashboard_upstream;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Household preferences (account-service)
    location /api/v1/household-preferences {
        set $account_upstream http://account-service:8080;
        proxy_pass $account_upstream;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
```

- [ ] **Step 3: Commit**

```bash
git add deploy/compose && git commit -m "feat(deploy): add dashboard-service to compose + nginx"
```

---

### Task 39: `scripts/local-up.sh` + `.env` template

- [ ] **Step 1:** If `scripts/local-up.sh` references services explicitly, add `dashboard-service`. If the script only invokes `docker compose up`, no change is needed beyond env defaults.

- [ ] **Step 2:** If a `.env.example` exists at `deploy/compose/.env.example` or repo root, add placeholders:

```
BOOTSTRAP_SERVERS=kafka-broker.kafka.svc.cluster.local:9092
EVENT_TOPIC_USER_LIFECYCLE=home-hub.user.lifecycle
```

- [ ] **Step 3:** Smoke: `scripts/local-up.sh` brings up the new service.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(deploy): document dashboard kafka env vars"
```

---

## Phase J — Service docs

### Task 40: `services/dashboard-service/docs/{domain,rest,storage}.md`

Reference: `services/package-service/docs/*.md`.

- [ ] **Step 1: Write `domain.md`** — describe the aggregate (`Dashboard`), the `scope` derivation, the widget allowlist, layout validation rules, promote/copy semantics, retention category, cascade event. Keep it code-anchored — cite file paths/functions.

- [ ] **Step 2: Write `rest.md`** — mirror `api-contracts.md` but add error code tables and the `UserDeletedEvent` consumer note.

- [ ] **Step 3: Write `storage.md`** — schema, indexes, advisory-lock key derivation, migration touchpoints.

- [ ] **Step 4: Update `services/account-service/docs/domain.md`** — add a `household_preferences` section.

- [ ] **Step 5: Update `docs/architecture.md`** — short section on the new Kafka bus, `shared/go/{kafka,events}`, `UserDeletedEvent` flow.

- [ ] **Step 6: Commit**

```bash
git add services/dashboard-service/docs services/account-service/docs docs/architecture.md
git commit -m "docs: document dashboard-service + household_preferences + kafka bus"
```

---

## Phase K — Frontend foundation

All files under `frontend/src/`. Reference patterns: `frontend/src/lib/hooks/api/use-packages.ts`, `frontend/src/services/api/package.ts`, `frontend/src/components/features/navigation/nav-config.ts`.

### Task 41: Layout schema + types

**Files:**
- Create: `frontend/src/lib/dashboard/widget-types.ts`
- Create: `frontend/src/lib/dashboard/__tests__/widget-types.test.ts`
- Create: `frontend/src/lib/dashboard/schema.ts`
- Create: `frontend/src/lib/dashboard/__tests__/schema.test.ts`
- Copy (new): `frontend/src/lib/dashboard/fixtures/widget-types.json` (same content as `shared/go/dashboard/fixtures/widget-types.json`)

- [ ] **Step 1: Failing parity test** — `widget-types.test.ts`

```ts
import { describe, it, expect } from "vitest";
import fixture from "@/lib/dashboard/fixtures/widget-types.json";
import { WIDGET_TYPES } from "@/lib/dashboard/widget-types";

describe("widget-types parity with Go allowlist", () => {
  it("matches the shared fixture exactly", () => {
    const ts = [...WIDGET_TYPES].sort();
    const go = [...(fixture as string[])].sort();
    expect(ts).toEqual(go);
  });
});
```

- [ ] **Step 2: Implement** — `widget-types.ts`

```ts
export const WIDGET_TYPES = [
  "weather",
  "tasks-summary",
  "reminders-summary",
  "overdue-summary",
  "meal-plan-today",
  "calendar-today",
  "packages-summary",
  "habits-today",
  "workout-today",
] as const;

export type WidgetType = (typeof WIDGET_TYPES)[number];

export function isKnownWidgetType(t: string): t is WidgetType {
  return (WIDGET_TYPES as readonly string[]).includes(t);
}

export const LAYOUT_SCHEMA_VERSION = 1;
export const GRID_COLUMNS = 12;
export const MAX_WIDGETS = 40;
```

- [ ] **Step 3: Implement `schema.ts`** — the layout-level Zod schemas:

```ts
import { z } from "zod";
import { GRID_COLUMNS, LAYOUT_SCHEMA_VERSION, MAX_WIDGETS } from "@/lib/dashboard/widget-types";

export const widgetInstanceSchema = z.object({
  id: z.string().uuid(),
  type: z.string(),
  x: z.number().int().nonnegative(),
  y: z.number().int().nonnegative(),
  w: z.number().int().min(1),
  h: z.number().int().min(1),
  config: z.record(z.string(), z.unknown()).default({}),
}).refine((w) => w.x + w.w <= GRID_COLUMNS, { message: "x + w must be <= 12", path: ["w"] });

export type WidgetInstance = z.infer<typeof widgetInstanceSchema>;

export const layoutSchema = z.object({
  version: z.literal(LAYOUT_SCHEMA_VERSION),
  widgets: z.array(widgetInstanceSchema).max(MAX_WIDGETS),
});

export type Layout = z.infer<typeof layoutSchema>;
```

- [ ] **Step 4: Failing schema test** — covers: valid layout parses; invalid version rejected; 41 widgets rejected; bad geometry rejected.

- [ ] **Step 5: Run and commit**

```bash
cd frontend && npm run test -- lib/dashboard
git add -A && git commit -m "feat(frontend): add dashboard widget-type allowlist + layout schema"
```

---

### Task 42: `parseConfig` tolerant-read helper

**Files:**
- Create: `frontend/src/lib/dashboard/parse-config.ts`
- Create: `frontend/src/lib/dashboard/__tests__/parse-config.test.ts`

- [ ] **Step 1: Failing test**

```ts
import { describe, it, expect } from "vitest";
import { z } from "zod";
import { parseConfig } from "@/lib/dashboard/parse-config";

const def = {
  type: "tasks-summary" as const,
  defaultConfig: { status: "pending" as const, title: "" },
  configSchema: z.object({
    status: z.enum(["pending", "overdue", "completed"]),
    title: z.string().max(80).optional(),
  }),
} as const;

describe("parseConfig", () => {
  it("returns parsed config when valid", () => {
    const r = parseConfig(def, { status: "overdue" });
    expect(r.config.status).toBe("overdue");
    expect(r.lossy).toBe(false);
  });

  it("fills missing fields from defaultConfig", () => {
    const r = parseConfig(def, { title: "hi" });
    expect(r.config.status).toBe("pending");
    expect(r.lossy).toBe(false);
  });

  it("falls back to defaults on invalid input", () => {
    const r = parseConfig(def, { status: "not-real" });
    expect(r.config).toEqual(def.defaultConfig);
    expect(r.lossy).toBe(true);
  });

  it("treats non-object raw as empty", () => {
    const r = parseConfig(def, 42);
    expect(r.config).toEqual(def.defaultConfig);
  });
});
```

- [ ] **Step 2: Implement**

```ts
import { z } from "zod";

export type ParsedConfig<T> = { config: T; lossy: boolean };

export function parseConfig<T>(
  def: { defaultConfig: T; configSchema: z.ZodType<T> },
  raw: unknown,
): ParsedConfig<T> {
  const merged = { ...(def.defaultConfig as object), ...(isRecord(raw) ? raw : {}) };
  const result = def.configSchema.safeParse(merged);
  if (result.success) return { config: result.data, lossy: false };
  return { config: def.defaultConfig, lossy: true };
}

function isRecord(x: unknown): x is Record<string, unknown> {
  return typeof x === "object" && x !== null && !Array.isArray(x);
}
```

- [ ] **Step 3: Commit**

```bash
npm run test -- parse-config && git add -A && git commit -m "feat(frontend): add tolerant-read parseConfig helper"
```

---

### Task 43: Widget registry shape + first three widget modules

**Files:**
- Create: `frontend/src/lib/dashboard/widget-registry.ts`
- Create: `frontend/src/lib/dashboard/widgets/weather.ts`
- Create: `frontend/src/lib/dashboard/widgets/tasks-summary.ts`
- Create: `frontend/src/lib/dashboard/widgets/reminders-summary.ts`

- [ ] **Step 1: Define `WidgetDefinition`**

```ts
// widget-registry.ts
import type { ComponentType } from "react";
import type { z } from "zod";
import type { WidgetType } from "@/lib/dashboard/widget-types";

export type WidgetDefinition<TConfig> = {
  type: WidgetType;
  displayName: string;
  description: string;
  component: ComponentType<{ config: TConfig }>;
  configSchema: z.ZodType<TConfig>;
  defaultConfig: TConfig;
  defaultSize: { w: number; h: number };
  minSize: { w: number; h: number };
  maxSize: { w: number; h: number };
  dataScope: "household" | "user";
};

export type AnyWidgetDefinition = WidgetDefinition<unknown>;
```

- [ ] **Step 2: weather.ts** — wraps the existing `WeatherWidget`:

```ts
import { z } from "zod";
import { WeatherWidget } from "@/components/features/weather/weather-widget";
import type { WidgetDefinition } from "@/lib/dashboard/widget-registry";

const schema = z.object({
  units: z.enum(["imperial", "metric"]).default("imperial"),
  location: z.object({
    lat: z.number(),
    lon: z.number(),
    label: z.string().max(200),
  }).nullable().default(null),
});

type Cfg = z.infer<typeof schema>;

export const weatherWidget: WidgetDefinition<Cfg> = {
  type: "weather",
  displayName: "Weather",
  description: "Current conditions and forecast",
  component: ({ config }: { config: Cfg }) => <WeatherWidget units={config.units} locationOverride={config.location ?? undefined} />,
  configSchema: schema,
  defaultConfig: { units: "imperial", location: null },
  defaultSize: { w: 12, h: 3 },
  minSize: { w: 6, h: 2 },
  maxSize: { w: 12, h: 4 },
  dataScope: "household",
};
```

Note: the existing `WeatherWidget` may not currently accept `units`/`locationOverride` props. If not, extend its prop type with **optional** props — existing callers are unaffected. Same pattern in Task 44.

- [ ] **Step 3: tasks-summary.ts** — new component that reads `useTaskSummary`, renders a `Card` matching the existing Pending Tasks card, and renders different counts based on `status`. Copy the JSX from the Pending/Overdue cards inside `DashboardPage.tsx:104-148` and parameterize by status.

Component source location: `frontend/src/components/features/dashboard-widgets/tasks-summary.tsx` (new).

Registry entry:

```ts
const schema = z.object({
  status: z.enum(["pending", "overdue", "completed"]),
  title: z.string().max(80).optional(),
});
type Cfg = z.infer<typeof schema>;

export const tasksSummaryWidget: WidgetDefinition<Cfg> = {
  type: "tasks-summary",
  displayName: "Tasks Summary",
  description: "Count of pending, overdue, or completed tasks",
  component: TasksSummaryWidget,
  configSchema: schema,
  defaultConfig: { status: "pending" },
  defaultSize: { w: 4, h: 2 },
  minSize: { w: 3, h: 2 },
  maxSize: { w: 6, h: 2 },
  dataScope: "household",
};
```

- [ ] **Step 4: reminders-summary.ts** — symmetric, using `useReminderSummary`. Component at `frontend/src/components/features/dashboard-widgets/reminders-summary.tsx`. Config: `{ filter: "active" | "snoozed" | "upcoming", title?: string }`.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat(frontend): add widget registry shape + weather/tasks/reminders entries"
```

---

### Task 44: Remaining six widget modules

**Files (create each):**
- `frontend/src/lib/dashboard/widgets/overdue-summary.ts`
- `frontend/src/lib/dashboard/widgets/meal-plan-today.ts`
- `frontend/src/lib/dashboard/widgets/calendar-today.ts`
- `frontend/src/lib/dashboard/widgets/packages-summary.ts`
- `frontend/src/lib/dashboard/widgets/habits-today.ts`
- `frontend/src/lib/dashboard/widgets/workout-today.ts`
- `frontend/src/components/features/dashboard-widgets/overdue-summary.tsx` (new component)

- [ ] **Step 1: overdue-summary.ts** — `OverdueSummaryWidget` is a new thin component equivalent to `tasks-summary` with status=overdue. Config: `{ title?: string }`.

- [ ] **Step 2: meal-plan-today.ts** — wraps the existing `MealPlanWidget`. Config: `{ horizonDays: 1 | 3 | 7 }`. Extend `MealPlanWidget` to accept an optional `horizonDays` prop that influences its `useMealPlans` query (default preserves existing behavior → current value for back-compat).

- [ ] **Step 3: calendar-today.ts** — wraps `CalendarWidget`. Config: `{ horizonDays: 1 | 3 | 7, includeAllDay: boolean }`. Extend the widget similarly.

- [ ] **Step 4: packages-summary.ts** — wraps `PackageSummaryWidget`. Config: `{ title?: string }`.

- [ ] **Step 5: habits-today.ts** — wraps `HabitsWidget`. Config: `{ title?: string }`.

- [ ] **Step 6: workout-today.ts** — wraps `WorkoutWidget`. Config: `{ title?: string }`.

- [ ] **Step 7: Assemble registry** — extend `widget-registry.ts`:

```ts
export const widgetRegistry: readonly AnyWidgetDefinition[] = [
  weatherWidget,
  tasksSummaryWidget,
  remindersSummaryWidget,
  overdueSummaryWidget,
  mealPlanTodayWidget,
  calendarTodayWidget,
  packagesSummaryWidget,
  habitsTodayWidget,
  workoutTodayWidget,
] as unknown as readonly AnyWidgetDefinition[];

export function findWidget(type: string): AnyWidgetDefinition | undefined {
  return widgetRegistry.find((w) => w.type === type);
}
```

Add a test `widget-registry.test.ts`:

```ts
import { WIDGET_TYPES } from "@/lib/dashboard/widget-types";
import { widgetRegistry } from "@/lib/dashboard/widget-registry";

it("has one registry entry per widget type", () => {
  const registryTypes = widgetRegistry.map((w) => w.type).sort();
  expect(registryTypes).toEqual([...WIDGET_TYPES].sort());
});
```

- [ ] **Step 8: Commit**

```bash
npm run test -- dashboard
git add -A && git commit -m "feat(frontend): register all nine dashboard widgets"
```

---

### Task 45: Seed layout module

**Files:**
- Create: `frontend/src/lib/dashboard/seed-layout.ts`
- Create: `frontend/src/lib/dashboard/__tests__/seed-layout.test.ts`

- [ ] **Step 1: Failing test** — asserts the produced object: version 1, nine widgets with exactly the cells + configs from `data-model.md §6`, each `id` is a fresh UUID on every call.

- [ ] **Step 2: Implement**

```ts
import type { Layout } from "@/lib/dashboard/schema";

function uuid(): string {
  return (crypto as Crypto).randomUUID();
}

export function seedLayout(): Layout {
  return {
    version: 1,
    widgets: [
      { id: uuid(), type: "weather",           x: 0, y: 0, w: 12, h: 3, config: { units: "imperial", location: null } },
      { id: uuid(), type: "tasks-summary",     x: 0, y: 3, w: 4,  h: 2, config: { status: "pending", title: "Pending Tasks" } },
      { id: uuid(), type: "reminders-summary", x: 4, y: 3, w: 4,  h: 2, config: { filter: "active", title: "Active Reminders" } },
      { id: uuid(), type: "overdue-summary",   x: 8, y: 3, w: 4,  h: 2, config: { title: "Overdue" } },
      { id: uuid(), type: "meal-plan-today",   x: 0, y: 5, w: 4,  h: 3, config: { horizonDays: 1 } },
      { id: uuid(), type: "habits-today",      x: 4, y: 5, w: 4,  h: 3, config: {} },
      { id: uuid(), type: "packages-summary",  x: 8, y: 5, w: 4,  h: 3, config: {} },
      { id: uuid(), type: "calendar-today",    x: 0, y: 8, w: 6,  h: 3, config: { horizonDays: 1, includeAllDay: true } },
      { id: uuid(), type: "workout-today",     x: 6, y: 8, w: 6,  h: 3, config: {} },
    ],
  };
}
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add seed layout replicating legacy DashboardPage"
```

---

### Task 46: API client + query hooks

**Files:**
- Create: `frontend/src/services/api/dashboard.ts`
- Create: `frontend/src/services/api/household-preferences.ts`
- Create: `frontend/src/lib/hooks/api/use-dashboards.ts`
- Create: `frontend/src/lib/hooks/api/use-household-preferences.ts`

Reference: `frontend/src/services/api/package.ts`, `frontend/src/lib/hooks/api/use-packages.ts`.

- [ ] **Step 1: dashboard.ts service client** — methods: `list`, `get`, `create`, `update`, `remove`, `reorder`, `promote`, `copyToMine`, `seed`. Request shapes per `api-contracts.md`.

- [ ] **Step 2: household-preferences.ts** — `get` + `update(id, { defaultDashboardId })`.

- [ ] **Step 3: use-dashboards.ts** — React Query hooks. Query keys:

```ts
export const dashboardKeys = {
  all: (t, h) => ["dashboards", t?.id ?? "no-t", h?.id ?? "no-h"] as const,
  list: (t, h) => [...dashboardKeys.all(t, h), "list"] as const,
  detail: (t, h, id: string) => [...dashboardKeys.all(t, h), "detail", id] as const,
};
```

Hooks: `useDashboards`, `useDashboard(id)`, `useCreateDashboard`, `useUpdateDashboard`, `useDeleteDashboard`, `useReorderDashboards`, `usePromoteDashboard`, `useCopyDashboardToMine`, `useSeedDashboard`. Every mutation invalidates `dashboardKeys.all`.

- [ ] **Step 4: use-household-preferences.ts** — `useHouseholdPreferences()`, `useUpdateHouseholdPreferences()`.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat(frontend): add dashboard + household-preferences services and hooks"
```

---

## Phase L — Frontend renderer + routing

### Task 47: `DashboardShell` (parent route)

**Files:**
- Create: `frontend/src/pages/DashboardShell.tsx`

- [ ] **Step 1:** Implement a React Router `<Outlet>`-based shell that:
  - Reads `useParams<{ dashboardId: string }>()`.
  - Runs `useDashboard(dashboardId)`.
  - While loading, renders `<DashboardSkeleton />`.
  - On error (incl. 404), renders `<Error404Page />` inline (or redirect to default).
  - Renders a header `<DashboardHeader>` (name + Edit button + kebab) and `<Outlet context={{ dashboard }} />`.
  - On schema version check: if `dashboard.schemaVersion > LAYOUT_SCHEMA_VERSION`, show a full-bleed message ("newer schema — please refresh browser").

- [ ] **Step 2:** Ship `<DashboardHeader>` inline in `DashboardShell.tsx` — render `<h1>{dashboard.attributes.name}</h1>` + an Edit `<Button>` that navigates to `./edit`. The kebab menu is added in Task 55 (`DashboardKebabMenu`); in this task, render no kebab element (the header has just the title + Edit).

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add DashboardShell parent route with query + schema gate"
```

---

### Task 48: `DashboardRenderer` (CSS Grid, library-free)

**Files:**
- Create: `frontend/src/pages/DashboardRenderer.tsx`
- Create: `frontend/src/pages/__tests__/DashboardRenderer.test.tsx`

- [ ] **Step 1: Failing test** — renders a dashboard with two widgets from the registry, asserts both components mount with the right configs. Uses `@testing-library/react` + a MemoryRouter.

- [ ] **Step 2: Implement**

```tsx
import { useOutletContext } from "react-router-dom";
import { findWidget } from "@/lib/dashboard/widget-registry";
import { parseConfig } from "@/lib/dashboard/parse-config";
import { UnknownWidgetPlaceholder } from "@/components/features/dashboard-widgets/unknown-widget-placeholder";
import { LossyConfigBadge } from "@/components/features/dashboard-widgets/lossy-config-badge";
import type { Dashboard } from "@/types/models/dashboard";
import { GRID_COLUMNS } from "@/lib/dashboard/widget-types";

export function DashboardRenderer() {
  const { dashboard } = useOutletContext<{ dashboard: Dashboard }>();
  const layout = dashboard.attributes.layout;
  const sorted = [...layout.widgets].sort((a, b) => a.y - b.y || a.x - b.x);

  return (
    <div className="p-4 md:p-6">
      <div
        className="hidden md:grid gap-4"
        style={{ gridTemplateColumns: `repeat(${GRID_COLUMNS}, minmax(0, 1fr))` }}
      >
        {sorted.map((w) => {
          const def = findWidget(w.type);
          if (!def) {
            return (
              <div key={w.id} style={{ gridColumn: `span ${w.w}`, gridRow: `span ${w.h}` }}>
                <UnknownWidgetPlaceholder type={w.type} />
              </div>
            );
          }
          const { config, lossy } = parseConfig(def, w.config);
          const Comp = def.component as React.ComponentType<{ config: unknown }>;
          return (
            <div key={w.id} style={{ gridColumn: `span ${w.w}`, gridRow: `span ${w.h}` }} className="relative">
              <Comp config={config} />
              {lossy && <LossyConfigBadge />}
            </div>
          );
        })}
      </div>
      <div className="grid md:hidden grid-cols-1 gap-4">
        {sorted.map((w) => {
          const def = findWidget(w.type);
          if (!def) return <UnknownWidgetPlaceholder key={w.id} type={w.type} />;
          const { config } = parseConfig(def, w.config);
          const Comp = def.component as React.ComponentType<{ config: unknown }>;
          return <Comp key={w.id} config={config} />;
        })}
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add DashboardRenderer (CSS Grid, responsive, lossy badges)"
```

---

### Task 49: `UnknownWidgetPlaceholder` + `LossyConfigBadge`

**Files:**
- Create: `frontend/src/components/features/dashboard-widgets/unknown-widget-placeholder.tsx`
- Create: `frontend/src/components/features/dashboard-widgets/lossy-config-badge.tsx`

- [ ] **Step 1:** Implement both as small presentational components (`<Card>` with an alert icon for unknown; `<Badge>` with tooltip for lossy).

- [ ] **Step 2:** Write a brief render test for each.

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add unknown widget placeholder + lossy config badge"
```

---

### Task 50: `DashboardRedirect` (legacy route + default resolution)

**Files:**
- Create: `frontend/src/pages/DashboardRedirect.tsx`

Resolution order (per PRD §4.6 / data-model §5):

1. Read `household_preferences.defaultDashboardId`.
2. If set and present in `list`, navigate there.
3. Else, if any household-scoped in list → first one.
4. Else, if any user-scoped in list → first one.
5. Else, call `seedDashboard({ name: "Home", layout: seedLayout() })`.
   - If it returns `201`, navigate to the new dashboard.
   - If it returns `200` (idempotent no-op), re-read the list and pick step 3.

- [ ] **Step 1: Failing test** — stub hooks to simulate (a) preference hit, (b) preference stale → fallback to first household, (c) empty → seed flow.

- [ ] **Step 2: Implement**

```tsx
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useDashboards, useSeedDashboard } from "@/lib/hooks/api/use-dashboards";
import { useHouseholdPreferences } from "@/lib/hooks/api/use-household-preferences";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";

export function DashboardRedirect() {
  const navigate = useNavigate();
  const { data: prefs } = useHouseholdPreferences();
  const { data: dashboards, refetch } = useDashboards();
  const seed = useSeedDashboard();

  useEffect(() => {
    if (!prefs || !dashboards) return;
    const list = dashboards.data;
    const pref = prefs.data.attributes.defaultDashboardId;
    if (pref && list.some((d) => d.id === pref)) {
      navigate(`/app/dashboards/${pref}`, { replace: true });
      return;
    }
    const household = list.find((d) => d.attributes.scope === "household");
    if (household) {
      navigate(`/app/dashboards/${household.id}`, { replace: true });
      return;
    }
    const user = list.find((d) => d.attributes.scope === "user");
    if (user) {
      navigate(`/app/dashboards/${user.id}`, { replace: true });
      return;
    }
    // Seed
    seed.mutate(
      { name: "Home", layout: seedLayout() },
      {
        onSuccess: async (res) => {
          if ("id" in res.data) {
            navigate(`/app/dashboards/${(res.data as any).id}`, { replace: true });
          } else {
            const refreshed = await refetch();
            const first = refreshed.data?.data?.[0];
            if (first) navigate(`/app/dashboards/${first.id}`, { replace: true });
          }
        },
      },
    );
  }, [prefs, dashboards, navigate, seed, refetch]);

  return <DashboardSkeleton />;
}
```

Note: export the inline `DashboardSkeleton` from `DashboardPage.tsx` into `frontend/src/components/common/dashboard-skeleton.tsx` before Task 65's deletion.

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add DashboardRedirect with default-resolution + seed flow"
```

---

### Task 51: Extract `DashboardSkeleton` to a shared component

**Files:**
- Create: `frontend/src/components/common/dashboard-skeleton.tsx`
- Modify: `frontend/src/pages/DashboardPage.tsx` (re-import the extracted component)

- [ ] **Step 1:** Move the `DashboardSkeleton` function body from `DashboardPage.tsx:26-51` into the new file as an exported function. Update `DashboardPage.tsx` to import it. This is a prerequisite for deleting `DashboardPage.tsx` in Task 65.

- [ ] **Step 2: Commit**

```bash
npm run build && git add -A && git commit -m "refactor(frontend): extract DashboardSkeleton to components/common"
```

---

### Task 52: Pull-to-refresh on renderer

- [ ] **Step 1:** Wrap `DashboardRenderer` output in `<PullToRefresh>` and invalidate widget query keys on refresh. Re-use the existing `handleRefresh` logic from `DashboardPage.tsx:74-84` — mirror its invalidation set.

- [ ] **Step 2:** Test: asserts that calling `onRefresh` invalidates the expected keys via a mock queryClient.

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add pull-to-refresh to dashboard renderer"
```

---

### Task 53: Wire routes in App.tsx (keep legacy `/app` intact)

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1:** Add routes **without** removing the existing `<Route index element={<DashboardPage />} />`:

```tsx
<Route path="dashboard" element={<DashboardRedirect />} />
<Route path="dashboards" element={<DashboardsIndexRedirect />} />
<Route path="dashboards/:dashboardId" element={<DashboardShell />}>
  <Route index element={<DashboardRenderer />} />
  <Route path="edit" element={<DashboardDesigner />} />
</Route>
```

Import `DashboardDesigner` via `const DashboardDesigner = React.lazy(() => import("@/pages/DashboardDesigner"));` and wrap it in `<Suspense fallback={<DashboardSkeleton />}>`. Create a stub `DashboardDesigner.tsx` exporting a placeholder component — the real designer lands in Phase N.

- [ ] **Step 2:** Add `DashboardsIndexRedirect` — thin wrapper that runs the same resolution as `DashboardRedirect`. Put both in `frontend/src/pages/` alongside `DashboardRedirect.tsx`.

- [ ] **Step 3: Verify** — `npm run dev`, visit `/app/dashboards/<id>` manually against a test dashboard, confirm render. Cutover to `/app` happens in Task 65.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(frontend): add /app/dashboards routes with lazy designer"
```

---

## Phase M — Sidebar integration

### Task 54: `DashboardsNavGroup`

**Files:**
- Create: `frontend/src/components/features/navigation/dashboards-nav-group.tsx`
- Modify: `frontend/src/components/features/navigation/nav-config.ts` (remove the hardcoded Dashboard entry from the `home` group — happens in Task 65 cutover)
- Modify: `frontend/src/components/features/navigation/app-shell.tsx` (render the new component above/inside the existing nav)

- [ ] **Step 1:** Implement the component rendering household-scoped then user-scoped groups, ordered by `sortOrder ASC, createdAt ASC`. Each entry is a `<NavLink to={`/app/dashboards/${id}`} />` with active highlight (NavLink's `isActive`). Kebab menu wired in Task 55.

- [ ] **Step 2:** Render it inside the existing sidebar (check `app-shell.tsx` for how `navGroups` are mapped; inject the new group either by extending `nav-config.ts` to declare a dynamic section or by adding the component next to the group list in `app-shell.tsx`).

- [ ] **Step 3:** Render test — given a mocked `useDashboards` response, asserts the right number of links + grouping + active state.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(frontend): add DashboardsNavGroup sidebar component"
```

---

### Task 55: New dashboard modal + kebab actions

**Files:**
- Create: `frontend/src/components/features/dashboards/new-dashboard-modal.tsx`
- Create: `frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx`

- [ ] **Step 1: New-dashboard modal** — shadcn `Dialog` + `react-hook-form` + zod. Fields: name (required 1-80), scope (radio), optional "Copy of" dropdown (list of visible dashboards). On submit:
  - Scope + empty layout → `useCreateDashboard`, then navigate to `/app/dashboards/<newId>/edit`.
  - Scope + "copy of" → if scope=user and source is household → `useCopyDashboardToMine` then navigate; otherwise create + patch with deep-copied layout.

- [ ] **Step 2: Kebab menu** — per entry: Rename, Set as my default, Delete, and conditionally Promote (user-owned) or Copy to mine (household). Actions use the hooks from Task 46 + preference updates.

- [ ] **Step 3:** Wire both into `DashboardsNavGroup` — kebab per entry + "+ New dashboard" row at bottom.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(frontend): add new-dashboard modal + sidebar kebab actions"
```

---

### Task 56: Drag-reorder in sidebar

**Files:**
- Install: `@dnd-kit/core` `@dnd-kit/sortable` (add to `frontend/package.json`)
- Modify: `frontend/src/components/features/navigation/dashboards-nav-group.tsx`

- [ ] **Step 1:** `npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities --workspace=frontend` (or at the frontend root). Commit the lockfile update separately if generators produce noise.

- [ ] **Step 2:** Wrap each scope's list in `<SortableContext>`. On drop, compute the new `sortOrder` for each item and call `useReorderDashboards` with the single-scope payload.

- [ ] **Step 3:** Test via user-event: simulate drag, assert reorder mutation called with expected payload.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(frontend): add dnd-kit-based drag reorder to dashboards sidebar"
```

---

## Phase N — Frontend designer

### Task 57: Install `react-grid-layout`

- [ ] **Step 1:** `npm install react-grid-layout --workspace=frontend` and `npm install -D @types/react-grid-layout --workspace=frontend`.

- [ ] **Step 2:** Commit

```bash
git add -A && git commit -m "chore(frontend): add react-grid-layout dependency"
```

---

### Task 58: `DashboardDesigner` reducer + tests

**Files:**
- Create: `frontend/src/pages/dashboard-designer/state.ts`
- Create: `frontend/src/pages/dashboard-designer/__tests__/state.test.ts`

- [ ] **Step 1: Failing test** — covers each action from design §3.5. Start with `add`, `remove`, `move-or-resize`, `update-config`, `rename`, `reset`, `saved`. `dirty` true after any mutation, cleared by `reset` or `saved`.

- [ ] **Step 2: Implement**

```ts
import type { Dashboard } from "@/types/models/dashboard";
import type { Layout, WidgetInstance } from "@/lib/dashboard/schema";

export type DraftState = {
  name: string;
  layout: Layout;
  dirty: boolean;
  selectedWidgetId: string | null;
  paletteOpen: boolean;
};

export type DraftAction =
  | { type: "rename"; name: string }
  | { type: "move-or-resize"; widgets: WidgetInstance[] }
  | { type: "add"; widget: WidgetInstance }
  | { type: "remove"; id: string }
  | { type: "update-config"; id: string; config: Record<string, unknown> }
  | { type: "select"; id: string | null }
  | { type: "toggle-palette"; open: boolean }
  | { type: "reset"; server: Dashboard }
  | { type: "saved"; server: Dashboard };

export function fromServer(server: Dashboard): DraftState {
  return {
    name: server.attributes.name,
    layout: server.attributes.layout,
    dirty: false,
    selectedWidgetId: null,
    paletteOpen: false,
  };
}

export function draftReducer(state: DraftState, action: DraftAction): DraftState {
  switch (action.type) {
    case "rename":
      return { ...state, name: action.name, dirty: true };
    case "move-or-resize":
      return { ...state, layout: { ...state.layout, widgets: action.widgets }, dirty: true };
    case "add":
      return { ...state, layout: { ...state.layout, widgets: [...state.layout.widgets, action.widget] }, dirty: true, selectedWidgetId: action.widget.id };
    case "remove":
      return {
        ...state,
        layout: { ...state.layout, widgets: state.layout.widgets.filter((w) => w.id !== action.id) },
        dirty: true,
        selectedWidgetId: state.selectedWidgetId === action.id ? null : state.selectedWidgetId,
      };
    case "update-config":
      return {
        ...state,
        layout: { ...state.layout, widgets: state.layout.widgets.map((w) => (w.id === action.id ? { ...w, config: action.config } : w)) },
        dirty: true,
      };
    case "select":
      return { ...state, selectedWidgetId: action.id };
    case "toggle-palette":
      return { ...state, paletteOpen: action.open };
    case "reset":
      return fromServer(action.server);
    case "saved":
      return fromServer(action.server);
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add designer reducer with dirty tracking"
```

---

### Task 59: Grid integration (react-grid-layout)

**Files:**
- Create: `frontend/src/pages/dashboard-designer/designer-grid.tsx`
- Create: `frontend/src/pages/DashboardDesigner.tsx` (replace stub)

- [ ] **Step 1: Implement `DesignerGrid`** — render `<GridLayout cols={12} isBounded compactType="vertical" ...>` with per-widget `data-grid={{ i, x, y, w, h, minW, minH, maxW, maxH }}`. `onLayoutChange` dispatches `move-or-resize`.

- [ ] **Step 2: Implement `DashboardDesigner`** — reads `useOutletContext` for the server dashboard, instantiates the reducer via `useReducer(draftReducer, server, fromServer)`, renders header (Save / Discard / Rename), `<DesignerGrid>`, `<PaletteDrawer>`, `<ConfigPanel>`.

- [ ] **Step 3:** Render test — mount designer, assert 9 grid items for a seeded dashboard, simulate `onLayoutChange`, assert reducer state updated.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(frontend): add react-grid-layout-backed designer grid"
```

---

### Task 60: Widget edit chrome (drag/gear/trash)

**Files:**
- Create: `frontend/src/pages/dashboard-designer/widget-chrome.tsx`

- [ ] **Step 1:** Each grid item wraps its widget in `<WidgetChrome>`: drag handle (top-left, class used by `react-grid-layout` as `draggableHandle`), gear icon → `dispatch({ type: "select", id })`, trash icon → `dispatch({ type: "remove", id })`. If `findWidget(type)` is undefined, the gear is disabled with a tooltip.

- [ ] **Step 2:** Commit

```bash
git add -A && git commit -m "feat(frontend): add widget edit-mode chrome (drag/gear/trash)"
```

---

### Task 61: Palette drawer

**Files:**
- Create: `frontend/src/pages/dashboard-designer/palette-drawer.tsx`

- [ ] **Step 1:** shadcn `<Sheet>` listing every `widgetRegistry` entry. Each palette item is draggable (HTML5 drag or `react-grid-layout`'s `onDrop` flow — follow rgl's docs). On drop: generate a new instance with `defaultSize` + `defaultConfig` + freshly-minted UUID, dispatch `add`.

- [ ] **Step 2:** Commit

```bash
git add -A && git commit -m "feat(frontend): add designer palette drawer"
```

---

### Task 62: Config panel + `ZodForm`

**Files:**
- Create: `frontend/src/pages/dashboard-designer/config-panel.tsx`
- Create: `frontend/src/pages/dashboard-designer/zod-form.tsx`

- [ ] **Step 1: `ZodForm`** — recursive walk of a Zod object schema producing labeled inputs:
  - `ZodEnum` → `<RadioGroup>`.
  - `ZodBoolean` → shadcn `<Switch>`.
  - `ZodNumber` → `<Input type="number">` with min/max from `_def.checks` if present.
  - `ZodString` → `<Input>` with char counter if `.max()` exists.
  - `ZodObject` → nested `<fieldset>`.
  - `ZodOptional`/`ZodDefault`/`ZodNullable` → unwrap and recurse.

Use `react-hook-form` + `zodResolver`. `Apply` calls `configSchema.parse` and dispatches `update-config`. `Cancel` closes. `Reset to defaults` seeds the form with `defaultConfig`.

- [ ] **Step 2: Render test** for tasks-summary config form.

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat(frontend): add Zod-rendered widget config panel"
```

---

### Task 63: Save / Discard + dirty-state guard

**Files:**
- Modify: `frontend/src/pages/DashboardDesigner.tsx`
- Create: `frontend/src/pages/dashboard-designer/use-unsaved-guard.ts`

- [ ] **Step 1: Save** — `useUpdateDashboard()` mutation, body `{ name, layout }`. On success: `dispatch({ type: "saved", server: res.data })` and `navigate("../")`.

- [ ] **Step 2: Discard** — if dirty, confirm dialog; otherwise just `navigate("../")`.

- [ ] **Step 3: `useUnsavedGuard(dirty)`** — installs `useBlocker` for in-app navigation (React Router v7 API; check if it's available — if not, fall back to a route-level guard via an event listener) + a `beforeunload` listener installed only while `dirty === true`.

- [ ] **Step 4:** Tests: (a) save mutation posts expected payload, (b) `beforeunload` listener removed once cleared.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat(frontend): add designer save/discard and dirty-state guard"
```

---

### Task 64: Below-tablet blocker + header tooltip

- [ ] **Step 1:** In `DashboardDesigner`, check `window.matchMedia('(min-width: 768px)')` (or existing `useIsMobile` hook). If below, render a blocker pane: "Editing is only available on tablet-or-wider screens" + a "View only" button linking back to view mode.

- [ ] **Step 2:** In `DashboardShell`'s header, disable the Edit button with a tooltip when below 768px.

- [ ] **Step 3:** Test both paths.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(frontend): block designer below tablet breakpoint"
```

---

## Phase O — Cutover

### Task 65: Flip `/app` index + delete `DashboardPage.tsx`

**Files:**
- Modify: `frontend/src/App.tsx` (replace `<Route index element={<DashboardPage />} />` with `<Route index element={<DashboardRedirect />} />`)
- Modify: `frontend/src/components/features/navigation/nav-config.ts` (remove the hardcoded `{ to: "/app", icon: Home, label: "Dashboard", end: true }` entry from the `home` group)
- Delete: `frontend/src/pages/DashboardPage.tsx`

- [ ] **Step 1: Parity snapshot test** — add `frontend/src/pages/__tests__/dashboard-renderer-parity.test.tsx` that mounts `<DashboardRenderer>` with `{ data: { attributes: { layout: seedLayout(), name: "Home", scope: "household", ... } } }` in an Outlet context, and asserts (at minimum) one card per widget type is rendered. (Data-level, not pixel-level.)

- [ ] **Step 2: Run the full frontend test suite.** Manually verify in `npm run dev`:
   - `/app` redirects to a dashboard (household or seeded).
   - `/app/dashboard` redirects too (legacy).
   - Sidebar shows Dashboards group with at least one entry.
   - Edit button opens designer at `/app/dashboards/:id/edit`.

- [ ] **Step 3: Delete legacy** — `rm frontend/src/pages/DashboardPage.tsx`. Remove its import from `App.tsx`. Ensure no broken imports remain (`grep -r "DashboardPage"` should only hit test files for the new renderer — or nothing).

- [ ] **Step 4: Commit**

```bash
cd frontend && npm run build && npm run test
git add -A
git commit -m "feat(frontend): cut over /app to DashboardRedirect and remove legacy DashboardPage"
```

---

## Verification (run before claiming done)

- [ ] From repo root: `go build ./...` (if workspace) OR for each of `services/{dashboard-service,account-service}`, `shared/go/{dashboard,events,kafka,retention}`: `go test ./...`.
- [ ] `docker build -f services/dashboard-service/Dockerfile -t dashboard-service:dev .` succeeds.
- [ ] `scripts/local-up.sh` brings the full stack up with no crash loop in `hh-dashboard`.
- [ ] From `frontend/`: `npm run build && npm run test`.
- [ ] Manual (per acceptance criteria in PRD §10): seed, render, edit, save, reorder, promote, copy-to-mine, delete, schema-version gate, unknown-widget placeholder, below-tablet blocker, default preference round-trip, cascade on internal delete endpoint.

Per CLAUDE.md: always run tests, not just builds, before marking verification complete.

---

## Notes on scope seams

- No existing `account-service` user-delete handler exists today (as of 2026-04-23). Task 36 adds an internal-only endpoint (`POST /api/v1/internal/users/{id}/deleted`). Integrating it into a user-facing "delete my account" flow is out of scope for this task; when that flow is built, it calls this endpoint.
- Prometheus scrape config changes (design §4.7) are mechanical and left to the ops-facing follow-up. The service already exposes `/metrics` via the retention framework.
- `@dnd-kit/sortable` choice for sidebar reorder is deliberate (design §4.7 flagged this as interchangeable). Swap if the project has already standardized on a different DnD lib — check `frontend/package.json` before Task 56 installs.
