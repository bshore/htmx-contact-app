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

const (
	// Close enough... 123-456-7890
	PhoneRegexPattern string = "[0-9]{3}-[0-9]{3}-[0-9]{4}"

	Limit  int64 = 10
	Offset int64 = 10
)

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

func (d *DBClient) ListContacts(search string, page int64) ([]Contact, error) {
	if search == "" {
		return d.allContacts(page)
	}
	return d.searchContacts(search, page)
}

func (d *DBClient) allContacts(page int64) ([]Contact, error) {
	contacts := []Contact{}
	err := d.dao.DB().Select("*").From("contacts").Limit(10).Offset(page * Offset).All(&contacts)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (d *DBClient) searchContacts(search string, page int64) ([]Contact, error) {
	contacts := []Contact{}
	err := d.dao.DB().Select("*").From("contacts").
		Where(dbx.Like("first", search)).
		OrWhere(dbx.Like("last", search)).
		OrWhere(dbx.Like("phone", search)).
		OrWhere(dbx.Like("email", search)).
		Limit(10).
		Offset(page * Offset).
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

func (d *DBClient) DoesEmailExist(id, email string) bool {
	c := Contact{}
	err := d.dao.DB().Select("*").From("contacts").
		Where(
			dbx.NewExp("email = {:email}", dbx.Params{"email": email}),
		).Limit(1).One(&c)
	if c.ID != id {
		return true // this is someone else's email
	}
	if c.Email == "" {
		return false // it wasn't found
	}
	// if the query failed, there's no record
	// if the query succeded and it passed the id/email checks it should be okay
	return err != nil
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
