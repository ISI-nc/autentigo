package auth

import (
	jwt "github.com/golang-jwt/jwt/v4"
)

// ExtraClaims are our standard extensions to JWT tokens.
type ExtraClaims struct {
	DisplayName   string   `json:"display_name,omitempty" bson:"display_name,omitempty"`
	Email         string   `json:"email,omitempty" bson:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty" bson:"email_verified,omitempty"`
	Groups        []string `json:"groups,omitempty" bson:"groups,omitempty"`
}

// Claims supporting our ExtraClaims.
type Claims struct {
	jwt.StandardClaims
	ExtraClaims
}
