package migrations

import (
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"

	"github.com/bshore/htmx-contact-app/internal/model"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		collection := &models.Collection{
			Name:       "contacts",
			CreateRule: nil,
			ListRule:   nil,
			ViewRule:   nil,
			UpdateRule: nil,
			DeleteRule: nil,
			Type:       models.CollectionTypeBase,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "first",
					Type:     schema.FieldTypeText,
					Required: false,
					Options:  &schema.TextOptions{},
				},
				&schema.SchemaField{
					Name:     "last",
					Type:     schema.FieldTypeText,
					Required: false,
					Options:  &schema.TextOptions{},
				},
				&schema.SchemaField{
					Name:     "phone",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Pattern: model.PhoneRegexPattern,
					},
				},
				&schema.SchemaField{
					Name:     "email",
					Type:     schema.FieldTypeEmail,
					Required: true,
					Options:  &schema.TextOptions{},
				},
			),
			Indexes: []string{
				`CREATE UNIQUE INDEX IF NOT EXISTS "idx__contacts__email_unique" ON "contacts" ("email")`,
			},
		}

		dao := daos.New(db)

		err := dao.SaveCollection(collection)
		if err != nil {
			return fmt.Errorf("failed to save collection for 'contacts': %v", err)
		}

		for _, c := range bootstrapContacts {
			record := models.NewRecord(collection)
			record.Set("first", c.First)
			record.Set("last", c.Last)
			record.Set("phone", c.Phone)
			record.Set("email", c.Email)

			err = dao.SaveRecord(record)
			if err != nil {
				return fmt.Errorf("failed to save record: %v", err)
			}
		}
		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("contacts")
		if err != nil {
			return err
		}

		return daos.New(db).DeleteCollection(collection)
	})
}

// https://jsonplaceholder.typicode.com/users
var bootstrapContacts = []model.Contact{
	{
		First: "Leanne",
		Last:  "Graham",
		Email: "Sincere@april.biz",
		Phone: "770-736-8031",
	},
	{
		First: "Ervin",
		Last:  "Howell",
		Email: "Shanna@melissa.tv",
		Phone: "010-692-6593",
	},
	{
		First: "Clementine",
		Last:  "Bauch",
		Email: "Nathan@yesenia.net",
		Phone: "463-123-4447",
	},
	{
		First: "Patricia",
		Last:  "Lebsack",
		Email: "Julianne.OConner@kory.org",
		Phone: "493-170-9623",
	},
	{
		First: "Chelsey",
		Last:  "Dietrich",
		Email: "Lucio_Hettinger@annie.ca",
		Phone: "254-954-1289",
	},
	{
		First: "Dennis",
		Last:  "Schulist",
		Email: "Karley_Dach@jasper.info",
		Phone: "477-935-8478",
	},
	{
		First: "Kurtis",
		Last:  "Weissnat",
		Email: "Telly.Hoeger@billy.biz",
		Phone: "210-067-6132",
	},
	{
		First: "Nicholas",
		Last:  "Runolfsdottir",
		Email: "Sherwood@rosamond.me",
		Phone: "586-493-6943",
	},
	{
		First: "Glenna",
		Last:  "Reichert",
		Email: "Chaim_McDermott@dana.io",
		Phone: "775-976-6794",
	},
	{
		First: "Clementina",
		Last:  "DuBuque",
		Email: "Rey.Padberg@karina.biz",
		Phone: "024-648-3804",
	},
}
