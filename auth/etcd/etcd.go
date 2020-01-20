package etcd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/isi-nc/autentigo/api"
	"github.com/isi-nc/autentigo/auth"
)

// New Authenticator with etcd backend
func New(prefix string, endpoints []string) api.Authenticator {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})

	if err != nil {
		log.Fatal("failed to connect to etcd: ", err)
	}

	timeout := 5 * time.Second
	if timeoutEnv := os.Getenv("ETCD_TIMEOUT"); timeoutEnv != "" {
		timeout, err = time.ParseDuration(timeoutEnv)
		if err != nil {
			log.Fatalf("invalid ETCD_TIMEOUT %q: %v", timeoutEnv, timeout)
		}
	}

	return &etcdAuth{
		prefix:  prefix,
		client:  client,
		timeout: timeout,
	}
}

type etcdAuth struct {
	prefix  string
	client  *clientv3.Client
	timeout time.Duration
}

var _ api.Authenticator = &etcdAuth{}

// User describe an user stored in etcd
type User struct {
	PasswordHash string `json:"password_hash"`
	auth.ExtraClaims
}

func (a *etcdAuth) Authenticate(user, password string, expiresAt time.Time) (claims jwt.Claims, err error) {

	ba := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(ba[:])

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	resp, err := a.client.Get(ctx, path.Join(a.prefix, user))
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		err = api.ErrInvalidAuthentication
		return
	}

	u := User{}
	if err = json.Unmarshal(resp.Kvs[0].Value, &u); err != nil {
		return
	}

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
