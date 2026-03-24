module github.com/jtumidanski/home-hub/services/account-service

go 1.26.1

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/jtumidanski/api2go v1.0.4
	github.com/jtumidanski/home-hub/shared/go/auth v0.0.0-00010101000000-000000000000
	github.com/jtumidanski/home-hub/shared/go/database v0.0.0
	github.com/jtumidanski/home-hub/shared/go/logging v0.0.0
	github.com/jtumidanski/home-hub/shared/go/model v0.0.0
	github.com/jtumidanski/home-hub/shared/go/server v0.0.0
	github.com/jtumidanski/home-hub/shared/go/tenant v0.0.0
	github.com/sirupsen/logrus v1.9.3
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gedex/inflector v0.0.0-20170307190818-16278e9db813 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.42.0 // indirect
	go.opentelemetry.io/otel/metric v1.42.0 // indirect
	go.opentelemetry.io/otel/sdk v1.42.0 // indirect
	go.opentelemetry.io/otel/trace v1.42.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/grpc v1.79.2 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gorm.io/driver/postgres v1.5.11 // indirect
)

replace (
	github.com/jtumidanski/home-hub/shared/go/auth => ../../shared/go/auth
	github.com/jtumidanski/home-hub/shared/go/database => ../../shared/go/database
	github.com/jtumidanski/home-hub/shared/go/logging => ../../shared/go/logging
	github.com/jtumidanski/home-hub/shared/go/model => ../../shared/go/model
	github.com/jtumidanski/home-hub/shared/go/server => ../../shared/go/server
	github.com/jtumidanski/home-hub/shared/go/tenant => ../../shared/go/tenant
)
