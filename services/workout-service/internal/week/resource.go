package week

import (
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes is intentionally a no-op. The week-related HTTP handlers
// live in the `weekview` package because they need to import both the week
// and planneditem packages — keeping the routes in `week` would create an
// import cycle. The shim is preserved here so main.go's wiring remains
// symmetrical with every other domain package.
func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	}
}
