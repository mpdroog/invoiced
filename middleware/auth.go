package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/itshosted/webutils/encrypt"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
)

const (
	// SessionMaxAge is the maximum age of a session in seconds (8 hours)
	SessionMaxAge = 8 * 60 * 60
	// IVSize is the size of random bytes to generate (24 bytes = 32 base64 chars)
	IVSize = 24
)

type Sess struct {
	Email   string
	Created int64
	Version int
}

type Entities struct {
	IV      string `json:"-"`
	Version int

	GitUser  string `json:"-"`
	GitToken string `json:"-"`

	Company map[string]Entity
	User    []User
}

type Entity struct {
	Years       []string
	YearRevenue map[string]string // Revenue per year (EUR)

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

// generateIV creates a cryptographically secure random IV for AES encryption
func generateIV() (string, error) {
	iv := make([]byte, IVSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}
	return base64.StdEncoding.EncodeToString(iv), nil
}

// encryptSession encrypts session data with a unique IV and returns IV:ciphertext
func encryptSession(sess *Sess) (string, error) {
	iv, err := generateIV()
	if err != nil {
		return "", err
	}
	enc, err := encrypt.EncryptBase64("aes", iv, sess)
	if err != nil {
		return "", err
	}
	// Prepend IV to ciphertext (format: IV:ciphertext)
	return iv + ":" + enc, nil
}

// decryptSession extracts IV from session cookie and decrypts
func decryptSession(sessCipher string, sess *Sess) error {
	// Check for new format (IV:ciphertext)
	if idx := strings.Index(sessCipher, ":"); idx > 0 {
		iv := sessCipher[:idx]
		ciphertext := sessCipher[idx+1:]
		return encrypt.DecryptBase64("aes", iv, ciphertext, sess)
	}
	// Fallback to legacy format (static IV) for existing sessions
	return encrypt.DecryptBase64("aes", entities.IV, sessCipher, sess)
}

// isSessionExpired checks if the session has exceeded the maximum age
func isSessionExpired(sess *Sess) bool {
	return time.Now().Unix()-sess.Created > SessionMaxAge
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
			if subtle.ConstantTimeCompare([]byte(user.Hash), []byte(hash)) != 1 {
				// Invalid pass
				if config.Verbose {
					log.Printf("Login(%s) invalid hash, expect=%s got=%s\n", email, user.Hash, hash)
				}
				return "", nil
			}

			// Create sess with unique IV per session
			u := Sess{Email: email, Created: time.Now().Unix(), Version: entities.Version}
			return encryptSession(&u)
		}
	}

	return "", nil
}

// Resolve user companies by sess
func Companies(sessCipher string) (map[string]Entity, error) {
	var sess Sess
	if e := decryptSession(sessCipher, &sess); e != nil {
		return nil, e
	}

	// Check session expiration
	if isSessionExpired(&sess) {
		return nil, fmt.Errorf("session expired")
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

// GitCredentials returns configured git user and token
func GitCredentials() (string, string) {
	return entities.GitUser, entities.GitToken
}

// Check if user is allowed to open this path
func CompanyAllowed(company, email string) (bool, error) {
	for _, user := range entities.User {
		// Only check the user that matches the email
		if user.Email != email {
			continue
		}
		// Found the user, check if they have access to this company
		for _, name := range user.Company {
			if name == company {
				return true, nil
			}
		}
		// User found but doesn't have access to this company
		return false, nil
	}
	// User not found
	return false, nil
}
