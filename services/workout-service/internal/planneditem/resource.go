package planneditem

import (
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes is intentionally a no-op. Planned-item HTTP handlers live
// in the `weekview` package alongside the week handlers — both need to call
// each other's processors, and weekview is the cycle-breaking layer.
func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	}
}
