module github.com/jtumidanski/home-hub/shared/go/database

go 1.26.1

require (
	github.com/jtumidanski/home-hub/shared/go/model v0.0.0
	github.com/jtumidanski/home-hub/shared/go/tenant v0.0.0
	github.com/sirupsen/logrus v1.9.3
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.12
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	github.com/jtumidanski/home-hub/shared/go/model => ../model
	github.com/jtumidanski/home-hub/shared/go/tenant => ../tenant
)
