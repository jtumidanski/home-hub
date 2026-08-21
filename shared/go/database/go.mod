module github.com/jtumidanski/home-hub/shared/go/database

go 1.26.1

require (
	github.com/jtumidanski/home-hub/shared/go/model v0.0.0
	github.com/jtumidanski/home-hub/shared/go/tenant v0.0.0
	github.com/sirupsen/logrus v1.9.4
	gorm.io/driver/postgres v1.6.2
	gorm.io/gorm v1.31.2
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/stretchr/testify v1.12.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace (
	github.com/jtumidanski/home-hub/shared/go/model => ../model
	github.com/jtumidanski/home-hub/shared/go/tenant => ../tenant
)
