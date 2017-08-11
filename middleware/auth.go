package middleware

import (
	"github.com/mpdroog/invoiced/config"
	"github.com/itshosted/webutils/encrypt"
	"github.com/mpdroog/invoiced/db"
	//"net/http"
	"crypto/sha256"
	"fmt"
	"time"
	"log"
)

type Sess struct {
	Email   string
	Created int64
	Version int
}

type Entities struct {
	IV string
	Version int

	Company map[string]Entity
	User []User
}

type Entity struct {
	Name string
	COC string
	VAT string
	IBAN string
	BIC string
	Salt string
}
type User struct {
	Email string
	Hash string
	Company []string
	Name string
	Address1 string
	Address2 string
}

var entities Entities

func Init() error {
	if e := db.Open("entities.toml", &entities); e != nil {
		return e
	}
	return nil	
}

// Authenticate user and return sess-cookie if valid
func Login(email, pass string) (string, error) {
	for _, user := range entities.User {
		if user.Email == email {
			// Found the user, validate pass!
			// TODO: protect against crash?
			salt := entities.Company[ user.Company[0] ].Salt

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

// Check if user is allowed to open this path
func companyAllowed(company, email string) (bool, error) {
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