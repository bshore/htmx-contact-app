package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		settings, _ := dao.FindSettings()
		// Rename App
		settings.Meta.AppName = "htmx-contact-app"
		// Disable Create/Edit Collection Controls
		settings.Meta.HideControls = true

		return dao.SaveSettings(settings)
	}, nil)
}
