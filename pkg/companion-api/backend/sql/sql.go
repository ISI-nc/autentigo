package sql

import (
	"database/sql"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"log"
	"strings"

	"github.com/isi-nc/autentigo/pkg/companion-api/api"
	"github.com/isi-nc/autentigo/pkg/companion-api/backend"

	_ "github.com/lib/pq"
)

type sqlClient struct {
	db    *sql.DB
	table string
}

func New(driver, dsn, table string) backend.Client {
	db := DbConnect(driver, dsn)

	sqlClient := &sqlClient{
		db: db,
		table: table,
	}

	if err := CreateUsersTableIfNotExists(db, table); err != nil {
		panic(err)
	}

	return sqlClient
}

var _ backend.Client = &sqlClient{}

func (s *sqlClient) GetUser(id string) (user *backend.UserData, err error) {
	return s.getUser(id)
}

func (s *sqlClient) CreateUser(id string, user *backend.UserData) (err error) {
	oldUser := &backend.UserData{}
	oldUser, err = s.getUser(id)

	if oldUser != nil {
		err = api.ErrUserAlreadyExist
	} else if cmp.Equal(err, api.ErrMissingUser) {
		err = s.createUser(id, user)
	}

	return
}

func (s *sqlClient) UpdateUser(id string, update func(user *backend.UserData) error) (err error) {
	user := &backend.UserData{}
	user, err = s.getUser(id)

	if err == nil && user != nil {
		err = update(user)
		if err == nil {
			err = s.updateUser(id, user)
		}
	}

	return
}

func (s *sqlClient) DeleteUser(id string) (err error) {
	user := &backend.UserData{}
	user, err = s.getUser(id)

	if err == nil && user != nil {
		//ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
		//defer cancel()
		err = s.deleteUser(id)
	}

	return
}

func DbConnect(driver, dsn string) *sql.DB {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic(err)
	}

	// try to connect
	if err := db.Ping(); err != nil {
		panic(err)
	}

	log.Println("Connected to the database...")
	return db
}

func CreateUsersTableIfNotExists(db *sql.DB, table string) (err error) {

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(" +
		"id VARCHAR PRIMARY KEY NOT NULL," +
		"password_hash VARCHAR NOT NULL," +
		"display_name VARCHAR NOT NULL," +
		"email VARCHAR NOT NULL," +
		"email_verified BOOLEAN," +
		"groups VARCHAR NOT NULL" +
		");", table)

	_, err = db.Exec(query)

	return
}

func (s *sqlClient) getUser(id string) (u *backend.UserData, err error) {
	//ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	//defer cancel()

	groups := ""
	query := fmt.Sprintf("select password_hash, display_name, email, email_verified, groups from %s where id=$1;", s.table)

	u = &backend.UserData{}
	//xtraClaims := auth.ExtraClaims{}

	err = s.db.
		QueryRow(query, id).
		Scan(&u.PasswordHash, &u.ExtraClaims.DisplayName, &u.ExtraClaims.Email, &u.ExtraClaims.EmailVerified, &groups)
	if err != nil {
		return nil, api.ErrMissingUser
	}

	u.ExtraClaims.Groups = strings.Split(groups, ",")

	return
}

func (s *sqlClient) createUser(id string, user *backend.UserData) (err error) {
	//ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	//defer cancel()

	preparedQuery := fmt.Sprintf("INSERT INTO %s(id, password_hash, display_name,email, email_verified, groups) VALUES($1,$2,$3,$4,$5,$6)", s.table)
	var stmt *sql.Stmt
	stmt, err = s.db.Prepare(preparedQuery)
	if err != nil {
		return
	}

	_, err = stmt.Exec(id, user.PasswordHash, user.ExtraClaims.DisplayName, user.ExtraClaims.Email, user.ExtraClaims.EmailVerified, strings.Join(user.ExtraClaims.Groups,","))

	return
}

func (s *sqlClient) updateUser(id string, user *backend.UserData) (err error) {
	//ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	//defer cancel()

	preparedQuery := fmt.Sprintf("UPDATE %s SET password_hash=$2, display_name=$3, email=$4, email_verified=$5, groups=$6 WHERE id=$1", s.table)
	var stmt *sql.Stmt
	stmt, err = s.db.Prepare(preparedQuery)
	if err != nil {
		return
	}

	_, err = stmt.Exec(id, user.PasswordHash, user.ExtraClaims.DisplayName, user.ExtraClaims.Email, user.ExtraClaims.EmailVerified, strings.Join(user.ExtraClaims.Groups,","))

	return
}

func (s *sqlClient) deleteUser(id string) (err error) {
	//ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	//defer cancel()

	preparedQuery := fmt.Sprintf("DELETE FROM %s WHERE id=$1", s.table)
	var stmt *sql.Stmt
	stmt, err = s.db.Prepare(preparedQuery)
	if err != nil {
		return
	}

	_, err = stmt.Exec(id)

	return
}
