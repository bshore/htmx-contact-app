package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/bshore/htmx-contact-app/internal/migrations"
	"github.com/bshore/htmx-contact-app/internal/model"
	"github.com/bshore/htmx-contact-app/internal/views"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Dir:         "internal/migrations",
		Automigrate: false,
	})

	db := model.NewDBClient()
	// Register the Dao DB Instance after server startup/bootstrap
	// https://github.com/pocketbase/pocketbase/discussions/1813#discussioncomment-4910730
	app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
		db.RegisterDao(app.Dao())
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(middleware.Logger())
		e.Router.Use(middleware.Recover())

		// image/js/css static assets
		e.Router.GET("/static/*", apis.StaticDirectoryHandler(os.DirFS("./static"), false))

		e.Router.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, "/contacts")
		})

		// List Contacts
		e.Router.GET("/contacts", func(c echo.Context) error {
			search := c.QueryParam("q")
			contacts, err := db.ListContacts(search)
			if err != nil {
				return err
			}
			return views.Render(c, views.Index(search, contacts))
		})

		// Create Contact Form
		e.Router.GET("/contacts/new", func(c echo.Context) error {
			return views.Render(c, views.New(model.Contact{}, model.ContactError{}))
		})

		// Create Contact Form Submission
		e.Router.POST("/contacts/new", func(c echo.Context) error {
			contact := model.Contact{
				First: c.FormValue("first_name"),
				Last:  c.FormValue("last_name"),
				Phone: c.FormValue("phone"),
				Email: c.FormValue("email"),
			}
			vErrs, ok := contact.Validate()
			if !ok {
				return views.Render(c, views.New(contact, vErrs))
			}
			err := db.CreateContact(contact)
			if err != nil {
				if strings.Contains(err.Error(), "constraint failed") {
					vErrs.Email = "Contact with this email already exists"
					return views.Render(c, views.New(contact, vErrs))
				}
				return err
			}
			return c.Redirect(http.StatusSeeOther, "/contacts")
		})

		// Get Contact by ID
		e.Router.GET("/contacts/:id", func(c echo.Context) error {
			id := c.PathParam("id")
			contact, err := db.GetContactByID(id)
			if err != nil {
				return err
			}
			return views.Render(c, views.Show(contact))
		})

		// Get Edit Contact Form
		e.Router.GET("/contacts/:id/edit", func(c echo.Context) error {
			id := c.PathParam("id")
			contact, err := db.GetContactByID(id)
			if err != nil {
				return err
			}
			return views.Render(c, views.Edit(contact, model.ContactError{}))
		})

		// Edit Contact Form Submission
		e.Router.POST("/contacts/:id/edit", func(c echo.Context) error {
			contact := model.Contact{
				ID:    c.PathParam("id"),
				First: c.FormValue("first_name"),
				Last:  c.FormValue("last_name"),
				Phone: c.FormValue("phone"),
				Email: c.FormValue("email"),
			}

			vErrs, ok := contact.Validate()
			if !ok {
				return views.Render(c, views.Edit(contact, vErrs))
			}

			err := db.SaveContact(contact)
			if err != nil {
				if strings.Contains(err.Error(), "constraint failed") {
					vErrs.Email = "Contact with this email already exists"
					return views.Render(c, views.Edit(contact, vErrs))
				}
				return err
			}
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/contacts/%s", contact.ID))
		})

		// Delete Contact by ID
		e.Router.DELETE("/contacts/:id", func(c echo.Context) error {
			id := c.PathParam("id")
			err := db.DeleteContactByID(id)
			if err != nil {
				return err
			}
			return c.Redirect(http.StatusSeeOther, "/contacts")

		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
