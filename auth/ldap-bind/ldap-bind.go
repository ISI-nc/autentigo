package ldapbind

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/isi-nc/autentigo/api"
	"gopkg.in/ldap.v2"
)

// New Authenticator with ldap backend
func New(server, userTemplate string) api.Authenticator {
	u, err := url.Parse(server)
	if err != nil {
		log.Fatal("Bad LDAP server URL: ", err)
	}
	return &auth{
		url:          u,
		userTemplate: userTemplate,
	}
}

type auth struct {
	url          *url.URL
	userTemplate string
}

var _ api.Authenticator = auth{}

func (a auth) Authenticate(user, password string, expiresAt time.Time) (jwt.Claims, error) {
	var (
		l   *ldap.Conn
		err error
	)

	switch a.url.Scheme {
	case "ldaps":
		l, err = ldap.DialTLS("tcp", a.url.Host, &tls.Config{
			InsecureSkipVerify: true,
		})
	case "ldap":
		l, err = ldap.Dial("tcp", a.url.Host)
	default:
		log.Fatal("ldap: bad protocol: ", a.url.Scheme)
	}

	if err != nil {
		log.Print("LDAP dial error: ", err)
		return nil, err
	}

	if err := l.Bind(fmt.Sprintf(a.userTemplate, user), password); err != nil {
		log.Print("LDAP bind error: ", err)
		return nil, api.ErrInvalidAuthentication
	}

	return jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: expiresAt.Unix(),
		Subject:   user,
	}, nil
}
