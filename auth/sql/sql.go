package sql

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/isi-nc/autentigo/api"
	"github.com/isi-nc/autentigo/auth"

	_ "github.com/lib/pq"
)

// User describe an user stored in db
type User struct {
	Id           string
	PasswordHash string `json:"password_hash"`
	auth.ExtraClaims
}

type sqlAuth struct {
	db    *sql.DB
	table string
}

// New Authenticator with no backend
func New(driver, dsn, table string) api.Authenticator {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic(err)
	}

	// try to connect
	if err := db.Ping(); err != nil {
		panic(err)
	}

	log.Println("Connected to the database...")
	return &sqlAuth{
		db:    db,
		table: table,
	}
}

var _ api.Authenticator = sqlAuth{}

func (sa sqlAuth) Authenticate(user, password string, expiresAt time.Time) (claims jwt.Claims, err error) {
	ba := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(ba[:])

	u := User{}
	groups := ""
	query := fmt.Sprintf("select id, password_hash, display_name, email, email_verified, groups from %s where id=$1;", sa.table)

	err = sa.db.
		QueryRow(query, user).
		Scan(&u.Id, &u.PasswordHash, &u.DisplayName, &u.Email, &u.EmailVerified, &groups)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User %s not found", user)
			err = api.ErrInvalidAuthentication
			return
		}
		return
	}

	u.Groups = strings.Split(groups, ",")

	if u.PasswordHash != passwordHash {
		err = api.ErrInvalidAuthentication
		return
	}

	claims = auth.Claims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expiresAt.Unix(),
			Subject:   user,
		},
		ExtraClaims: u.ExtraClaims,
	}

	return
}
