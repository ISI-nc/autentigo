package backend

import (
	"github.com/isi-nc/autentigo/auth"
)

// UserData is a simple user struct with paswordhash and claims
type UserData struct {
	PasswordHash string           `json:"password"`
	ExtraClaims  auth.ExtraClaims `json:"claims"`
}

type User struct {
	PasswordHash string `json:"password_hash"`
	auth.ExtraClaims
}

func (u *UserData) ToUser() *User {
	return &User{
		PasswordHash: u.PasswordHash,
		ExtraClaims:  u.ExtraClaims,
	}
}

func (u *User) ToUserData() *UserData {
	return &UserData{
		PasswordHash: u.PasswordHash,
		ExtraClaims:  u.ExtraClaims,
	}
}

// Client is the interface for all backends clients
type Client interface {
	CreateUser(id string, user *UserData) error
	UpdateUser(id string, update func(user *UserData) error) error
	DeleteUser(id string) error
}
