package middleware

import (
	"github.com/itshosted/webutils/encrypt"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	//"net/http"
	"crypto/sha256"
	"fmt"
	"log"
	"time"
)

type Sess struct {
	Email   string
	Created int64
	Version int
}

type Entities struct {
	IV      string `json:"-"`
	Version int

	Company map[string]Entity
	User    []User
}

type Entity struct {
	Years   []string

	Name string
	COC  string
	VAT  string
	IBAN string
	BIC  string
	Salt string `json:"-"`
}
type User struct {
	Email    string
	Hash     string `json:"-"`
	Company  []string
	Name     string
	Address1 string
	Address2 string
}

var entities Entities

func Init() error {
	return db.View(func(t *db.Txn) error {
		e := t.Open("entities.toml", &entities)
		if config.Verbose {
			fmt.Printf("entities=%+v\n", entities)
		}
		return e
	})
}

// Authenticate user and return sess-cookie if valid
func Login(email, pass string) (string, error) {
	for _, user := range entities.User {
		if user.Email == email {
			// Found the user, validate pass!
			// TODO: protect against crash?
			salt := entities.Company[user.Company[0]].Salt

			h := sha256.New()
			h.Write([]byte(salt))
			h.Write([]byte(pass))
			hash := fmt.Sprintf("%x", h.Sum(nil))
			if user.Hash != hash {
				// Invalid pass
				if config.Verbose {
					log.Printf("Login(%s) invalid hash, expect=%s got=%s\n", email, user.Hash, hash)
				}
				return "", nil
			}

			// Create sess
			u := Sess{Email: email, Created: time.Now().Unix(), Version: entities.Version}
			enc, e := encrypt.EncryptBase64("aes", entities.IV, &u)
			return enc, e
		}
	}

	return "", nil
}

// Resolve user companies by sess
func Companies(sessCipher string) (map[string]Entity, error) {
	var sess Sess
	if e := encrypt.DecryptBase64("aes", entities.IV, sessCipher, &sess); e != nil {
		return nil, e
	}

	out := make(map[string]Entity)
	for _, user := range entities.User {
		if user.Email == sess.Email {
			// Found user
			for _, entity := range user.Company {
				// TODO: not set exception?
				out[entity] = entities.Company[entity]
			}
			return out, nil
		}
	}
	return nil, nil
}

func UserByEmail(email string) *User {
	for _, user := range entities.User {
		if user.Email == email {
			return &user
		}
	}
	return nil
}

func CompanyByName(name string) *Entity {
	for cname, entity := range entities.Company {
		if cname == name {
			return &entity
		}
	}
	return nil
}

// Check if user is allowed to open this path
func CompanyAllowed(company, email string) (bool, error) {
	// TODO: security bug here? no email check???
	for _, user := range entities.User {
		// Found the user
		for _, name := range user.Company {
			if name == company {
				return true, nil
			}
		}
	}

	return false, nil
}
