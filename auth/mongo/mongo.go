package mongo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/isi-nc/autentigo/api"
	"github.com/isi-nc/autentigo/auth"
)

// New Authenticator with mongo backend
func New(database string, collection string, field string, endpoint string) api.Authenticator {
	timeout := 5 * time.Second
	if timeoutEnv := os.Getenv("MONGO_TIMEOUT"); timeoutEnv != "" {
		timeout, err := time.ParseDuration(timeoutEnv)
		if err != nil {
			log.Fatalf("invalid MONGO_TIMEOUT %q: %v", timeoutEnv, timeout)
		}
	}

	// create client
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	mongoc, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		log.Fatal("failed to create mongo client: ", err)
	}

	// check connection
	err = mongoc.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("failed to connect to mongo: ", err)
	}

	// default field
	if field == "" {
		field = "_id"
	}

	return &mongoAuth{
		database:  database,
		collection: collection,
		field: field,
		client:  mongoc,
		timeout: timeout,
	}
}

type mongoAuth struct {
	database string
	collection  string
	field string
	client  *mongo.Client
	timeout time.Duration
}

var _ api.Authenticator = &mongoAuth{}

// User describe an user stored in mongo
type User struct {
	PasswordHash string `json:"password_hash"`
	auth.ExtraClaims
}

func (a *mongoAuth) Authenticate(user string, password string, expiresAt time.Time) (claims jwt.Claims, err error) {

	ba := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(ba[:])

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	filter := bson.M{a.field:user}

	u := &User{}

	sr := a.client.Database(a.database).
		Collection(a.collection).
		FindOne(ctx, filter)

	if sr.Err() == mongo.ErrNoDocuments {
		err = api.ErrInvalidAuthentication
		return
	}

	err = sr.Decode(u)
	if err != nil {
		log.Printf("Error while decoding user %s from database", user)
		err = api.ErrInvalidAuthentication
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
