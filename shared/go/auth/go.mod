module github.com/jtumidanski/home-hub/shared/go/auth

go 1.26.1

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jtumidanski/home-hub/shared/go/logging v0.0.0
	github.com/jtumidanski/home-hub/shared/go/tenant v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace (
	github.com/jtumidanski/home-hub/shared/go/logging => ../logging
	github.com/jtumidanski/home-hub/shared/go/tenant => ../tenant
)
