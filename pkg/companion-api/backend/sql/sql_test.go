package sql

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/isi-nc/autentigo/auth"
	"github.com/isi-nc/autentigo/pkg/companion-api/api"
	"github.com/isi-nc/autentigo/pkg/companion-api/backend"
	"github.com/isi-nc/autentigo/pkg/test"
	"testing"
)

const (
	authUsersTableName = "auth_users"
)

func setup(t *testing.T) {
	test.StartPool(t)
	test.StartPostgres(t)
	// automatic docker expire after 120s
	test.TestPostgresRsource.Expire(120)

	t.Cleanup(func() {
		test.TestPool.Purge(test.TestPostgresRsource)
	})
}

func TestDbConnect(t *testing.T) {
	setup(t)
	DbConnect("postgres",fmt.Sprintf("postgres://postgres:postgres@%s:%s/%s?sslmode=disable", test.PostgresHost(), test.TestPostgresRsource.GetPort("5432/tcp"), "postgres"))
}

func TestCreateUsersTableIfNotExists(t *testing.T) {
	setup(t)
	db := DbConnect("postgres",fmt.Sprintf("postgres://postgres:postgres@%s:%s/%s?sslmode=disable", test.PostgresHost(), test.TestPostgresRsource.GetPort("5432/tcp"), "postgres"))
	err := CreateUsersTableIfNotExists(db,authUsersTableName)
	if err != nil {
		t.Fatalf("Error while creating table: %v", err)
	}

	tables, err := test.GetTables(db)
	if err != nil {
		t.Fatalf("Error while getting tables name: %v", err)
	}

	if len(tables) == 0 {
		t.Fatalf("Table list is empty")
	}

	if tables[0] != authUsersTableName {
		t.Fatalf("Tables[0] should be %s but is %s", authUsersTableName, tables[0])
	}
}

func TestSqlClient_CreateUser(t *testing.T) {
	setup(t)
	db := DbConnect("postgres",fmt.Sprintf("postgres://postgres:postgres@%s:%s/%s?sslmode=disable", test.PostgresHost(), test.TestPostgresRsource.GetPort("5432/tcp"), "postgres"))
	err := CreateUsersTableIfNotExists(db,authUsersTableName)
	if err != nil {
		t.Fatalf("Error while creating table: %v", err)
	}

	client := &sqlClient{
		db: db,
		table: authUsersTableName,
	}

	user1Id := "toto"
	user1Claims := auth.ExtraClaims{
		DisplayName:   "toto",
		Email:         "toto@test.net",
		EmailVerified: false,
		Groups:        []string{"group1"},
	}
	user1 := &backend.UserData{
		PasswordHash: "hash",
		ExtraClaims:  user1Claims,
	}

	err = client.CreateUser(user1Id, user1)
	if err != nil {
		t.Fatalf("Error while creating user: %v", err)
	}

	userFromDb, err := client.GetUser(user1Id)
	if err != nil {
		t.Fatalf("Error getting creating user: %v", err)
	}

	if !cmp.Equal(user1, userFromDb) {
		t.Fatalf("User is different: %s", cmp.Diff(user1, userFromDb))
	}
}

func TestSqlClient_UpdateUser(t *testing.T) {
	setup(t)
	db := DbConnect("postgres",fmt.Sprintf("postgres://postgres:postgres@%s:%s/%s?sslmode=disable", test.PostgresHost(), test.TestPostgresRsource.GetPort("5432/tcp"), "postgres"))
	err := CreateUsersTableIfNotExists(db,authUsersTableName)
	if err != nil {
		t.Fatalf("Error while creating table: %v", err)
	}

	client := &sqlClient{
		db: db,
		table: authUsersTableName,
	}

	user1Id := "toto"
	user1Claims := auth.ExtraClaims{
		DisplayName:   "toto",
		Email:         "toto@test.net",
		EmailVerified: false,
		Groups:        []string{"group1"},
	}
	user1 := &backend.UserData{
		PasswordHash: "hash",
		ExtraClaims:  user1Claims,
	}

	err = client.CreateUser(user1Id, user1)
	if err != nil {
		t.Fatalf("Error while creating user: %v", err)
	}

	userFromDb, err := client.GetUser(user1Id)
	if err != nil {
		t.Fatalf("Error while getting user: %v", err)
	}

	if !cmp.Equal(user1, userFromDb) {
		t.Fatalf("User is different: %s", cmp.Diff(user1, userFromDb))
	}

	// now we update
	user1bClaims := auth.ExtraClaims{
		DisplayName:   "newtoto",
		Email:         "newtoto@test.net",
		EmailVerified: true,
		Groups:        []string{"group1","newgroup"},
	}
	user1b := &backend.UserData{
		PasswordHash: "newhash",
		ExtraClaims:  user1bClaims,
	}

	err = client.UpdateUser(user1Id, func(user *backend.UserData) error {
		user.ExtraClaims = user1bClaims
		user.PasswordHash = user1b.PasswordHash
		return nil
	})
	if err != nil {
		t.Fatalf("Error while updating user: %v", err)
	}

	// we get the updated user
	updateUserFromDb, err := client.GetUser(user1Id)
	if err != nil {
		t.Fatalf("Error while getting user: %v", err)
	}

	if !cmp.Equal(user1b, updateUserFromDb) {
		t.Fatalf("Updated user is different: %s", cmp.Diff(user1b, updateUserFromDb))
	}

}

func TestSqlClient_DeleteUser(t *testing.T) {
	setup(t)
	db := DbConnect("postgres",fmt.Sprintf("postgres://postgres:postgres@%s:%s/%s?sslmode=disable", test.PostgresHost(), test.TestPostgresRsource.GetPort("5432/tcp"), "postgres"))
	err := CreateUsersTableIfNotExists(db,authUsersTableName)
	if err != nil {
		t.Fatalf("Error while creating table: %v", err)
	}

	client := &sqlClient{
		db: db,
		table: authUsersTableName,
	}

	user1Id := "toto"
	user1 := &backend.UserData{
		PasswordHash: "hash",
	}

	err = client.CreateUser(user1Id, user1)
	if err != nil {
		t.Fatalf("Error while creating user: %v", err)
	}

	count, err := test.CountRows(db, authUsersTableName)
	if err != nil {
		t.Fatalf("Error while counting table: %v", err)
	}
	if count != 1 {
		t.Fatalf("Table %s should have one user but has %d rows", authUsersTableName, count)
	}

	err = client.DeleteUser(user1Id)
	if err != nil {
		t.Fatalf("Error while deleting user: %v", err)
	}

	count, err = test.CountRows(db, authUsersTableName)
	if err != nil {
		t.Fatalf("Error while counting table: %v", err)
	}
	if count != 0 {
		t.Fatalf("Table %s should be empty but has %d rows", authUsersTableName, count)
	}

	_, err = client.GetUser(user1Id)
	if !cmp.Equal(err, api.ErrMissingUser) {
		t.Fatal("User should be missing")
	}

}