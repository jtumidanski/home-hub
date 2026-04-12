module github.com/jtumidanski/home-hub/shared/go/retention

go 1.26.1

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/jtumidanski/home-hub/shared/go/database v0.0.0
	github.com/prometheus/client_golang v1.20.5
	github.com/sirupsen/logrus v1.9.3
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jtumidanski/home-hub/shared/go/model v0.0.0 // indirect
	github.com/jtumidanski/home-hub/shared/go/tenant v0.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gorm.io/driver/postgres v1.5.11 // indirect
)

replace (
	github.com/jtumidanski/home-hub/shared/go/database => ../database
	github.com/jtumidanski/home-hub/shared/go/model => ../model
	github.com/jtumidanski/home-hub/shared/go/tenant => ../tenant
)
