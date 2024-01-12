package model

import (
	"net/mail"
	"regexp"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/models"
)

type Contact struct {
	ID    string `db:"id"`
	First string `db:"first"`
	Last  string `db:"last"`
	Phone string `db:"phone"`
	Email string `db:"email"`
}

type ContactError struct {
	First string
	Last  string
	Phone string
	Email string
}

// Close enough... 123-456-7890
const PhoneRegexPattern string = "[0-9]{3}-[0-9]{3}-[0-9]{4}"

var phoneRegexp = regexp.MustCompile(PhoneRegexPattern)

func (c *Contact) Validate() (vErrs ContactError, ok bool) {
	ok = true
	if !phoneRegexp.MatchString(c.Phone) {
		vErrs.Phone = "Phone number is invalid: XXX-XXX-XXXX"
		ok = false
	} else {
		vErrs.Phone = ""
	}

	if _, err := mail.ParseAddress(c.Email); err != nil {
		vErrs.Email = "Email address is invalid"
		ok = false
	} else {
		vErrs.Email = ""
	}

	return vErrs, ok
}

func (d *DBClient) ListContacts(search string) ([]Contact, error) {
	if search == "" {
		return d.allContacts()
	}
	return d.searchContacts(search)
}

func (d *DBClient) allContacts() ([]Contact, error) {
	contacts := []Contact{}
	err := d.dao.DB().Select("*").From("contacts").Limit(10).All(&contacts)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (d *DBClient) searchContacts(search string) ([]Contact, error) {
	contacts := []Contact{}
	err := d.dao.DB().Select("*").From("contacts").
		Where(dbx.Like("first", search)).
		OrWhere(dbx.Like("last", search)).
		OrWhere(dbx.Like("phone", search)).
		OrWhere(dbx.Like("email", search)).
		All(&contacts)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (d *DBClient) CreateContact(c Contact) error {
	col, err := d.dao.FindCollectionByNameOrId("contacts")
	if err != nil {
		return err
	}

	rec := models.NewRecord(col)
	rec.Set("first", c.First)
	rec.Set("last", c.Last)
	rec.Set("phone", c.Phone)
	rec.Set("email", c.Email)

	return d.dao.SaveRecord(rec)
}

func (d *DBClient) GetContactByID(id string) (Contact, error) {
	c := Contact{}
	err := d.dao.DB().Select("*").From("contacts").
		Where(
			dbx.NewExp("id = {:id}", dbx.Params{"id": id}),
		).Limit(1).One(&c)
	return c, err
}

func (d *DBClient) SaveContact(c Contact) error {
	rec, err := d.dao.FindRecordById("contacts", c.ID)
	if err != nil {
		return err
	}
	rec.Set("first", c.First)
	rec.Set("last", c.Last)
	rec.Set("phone", c.Phone)
	rec.Set("email", c.Email)

	return d.dao.SaveRecord(rec)
}

func (d *DBClient) DeleteContactByID(id string) error {
	rec, err := d.dao.FindRecordById("contacts", id)
	if err != nil {
		return err
	}

	return d.dao.DeleteRecord(rec)
}
