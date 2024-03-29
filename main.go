package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

	// TODO: maybe make it possible to have multiple archivers (multiple users)
	archiver := model.NewArchiver()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(middleware.Logger())
		e.Router.Use(middleware.Recover())

		// image/js/css static assets
		e.Router.GET("/static/*", apis.StaticDirectoryHandler(os.DirFS("./static"), false))

		// archive assets
		e.Router.GET("/archive/:filepath", func(c echo.Context) error {
			fp := c.PathParam("filepath")
			return c.Attachment(fmt.Sprintf("archive/%s", fp), fp)
		})

		e.Router.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, "/contacts")
		})

		// List Contacts
		e.Router.GET("/contacts", func(c echo.Context) error {
			search := c.QueryParam("q")
			pageStr := c.QueryParam("page")
			page, _ := strconv.ParseInt(pageStr, 10, 64) // skipping err check because we'll accept page 0 as a default
			if page < 0 {
				return fmt.Errorf("page cannot be negative (%d)", page)
			}
			contacts, err := db.ListContacts(search, page)
			if err != nil {
				return err
			}
			if c.Request().Header.Get("HX-TRIGGER") == "search" {
				return views.Render(c, views.Rows(page, contacts))
			}
			return views.Render(c, views.Index(search, page, contacts, archiver))
		})

		// Delete multiple Contacts
		e.Router.DELETE("/contacts", func(c echo.Context) error {
			// NOTE: a little weird, I expected the form data to be in c.FormValues or c.Request().Form rather
			//       than in the Body? I thought DELETE can't have a Body? Is this an issue with Echo or PocketBase?
			bs, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return err
			}
			values, err := url.ParseQuery(string(bs))
			if err != nil {
				return err
			}
			contactIDs := values["selected_contact_ids"]
			err = db.DeleteContacts(contactIDs)
			if err != nil {
				return err
			}
			return c.Redirect(http.StatusSeeOther, "/contacts")
		})

		// Download Archive
		e.Router.POST("/contacts/archive", func(c echo.Context) error {
			archiver.Run(db)
			return views.Render(c, views.Archive(archiver))
		})

		// Archive progress
		e.Router.GET("/contacts/archive", func(c echo.Context) error {
			return views.Render(c, views.Archive(archiver))
		})

		// Clear archiver state
		e.Router.DELETE("/contacts/archive", func(c echo.Context) error {
			archiver.Reset()
			return views.Render(c, views.Archive(archiver))
		})

		// Get Count of Contacts
		e.Router.GET("/contacts/count", func(c echo.Context) error {
			count, err := db.CountContacts()
			if err != nil {
				return err
			}
			return c.String(http.StatusOK, fmt.Sprintf("( %s total Contacts)", count))
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

		// Get Contact by ID, validate and check if exists
		e.Router.GET("/contacts/:id/email", func(c echo.Context) error {
			id := c.PathParam("id")
			email := c.QueryParam("email")
			contact, err := db.GetContactByID(id)
			if err != nil {
				return err
			}
			contact.Email = email
			if vErrs, ok := contact.Validate(); !ok {
				return c.String(http.StatusOK, vErrs.Email)
			}
			if exists := db.DoesEmailExist(id, email); exists {
				return c.String(http.StatusOK, "Contact with this email already exists")
			}
			return c.String(http.StatusOK, "")
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

		// ========================================
		// end of app.OnBeforeServe()
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
