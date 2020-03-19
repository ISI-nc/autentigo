package main

import (
	"flag"
	"github.com/isi-nc/autentigo/auth/mongo"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	restful "github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"

	"github.com/isi-nc/autentigo/api"
	"github.com/isi-nc/autentigo/auth/etcd"
	ldapbind "github.com/isi-nc/autentigo/auth/ldap-bind"
	"github.com/isi-nc/autentigo/auth/sql"
	stupidauth "github.com/isi-nc/autentigo/auth/stupid-auth"
	usersfile "github.com/isi-nc/autentigo/auth/users-file"
)

var (
	tokenDuration = flag.Duration("token-duration", 1*time.Hour, "Duration of emitted tokens")
	bind          = flag.String("bind", ":8080", "HTTP bind specification")
	tlsBind       = flag.String("tls-bind", ":8443", "HTTPS bind specification")
	tlsKeyFile    = flag.String("tls-bind-key", "", "File containing the TLS listener's key")
	tlsCertFile   = flag.String("tls-bind-cert", "", "File containing the TLS listener's certificate")
	disableCORS   = flag.Bool("no-cors", false, "Disable CORS support")
)

func main() {
	flag.Parse()

	key, pubKey, sm, crtData := initJWT()

	hAPI := &api.API{
		CRTData:       []byte(crtData),
		Authenticator: getAuthenticator(),
		PrivateKey:    key,
		PublicKey:     pubKey,
		SigningMethod: sm,
		TokenDuration: *tokenDuration,
	}

	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	restful.DefaultContainer.Router(restful.CurlyRouter{})

	restful.Add(hAPI.Register())

	config := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
		APIPath:     "/apidocs.json",
	}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	if !*disableCORS {
		restful.Filter(restful.CrossOriginResourceSharing{
			CookiesAllowed: true,
			Container:      restful.DefaultContainer,
		}.Filter)
	}

	l, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("listening on ", *bind)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Kill, os.Interrupt, syscall.SIGTERM)
		<-sig

		log.Print("closing listener")
		l.Close()
	}()

	if *tlsKeyFile != "" && *tlsCertFile != "" {
		go func() {
			log.Print("TLS listening on ", *bind)
			log.Fatal(http.ListenAndServeTLS(
				*tlsBind, *tlsKeyFile, *tlsCertFile,
				restful.DefaultContainer))
		}()

	} else if *tlsKeyFile != "" || *tlsCertFile != "" {
		log.Fatal("please specify both tls-key and tls-cert, or none.")
	}

	log.Fatal(http.Serve(l, restful.DefaultContainer))
}

func initJWT() (key interface{}, cert interface{}, method jwt.SigningMethod, crtData string) {
	crtData = requireEnv("TLS_CRT", "certificate used to sign/verify tokens")
	keyData := requireEnv("TLS_KEY", "key used to sign tokens")
	sm := requireEnv("SIGNING_METHOD", "signature method to use (must match the key)")

	method = jwt.GetSigningMethod(sm)

	if method == nil {
		log.Fatal("unknown signing method: ", sm)
	}

	switch sm[:2] {
	case "RS":
		if x, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyData)); err == nil {
			key = x
		} else {
			log.Fatal("Failed to load private key: ", err)
		}
		if x, err := jwt.ParseRSAPublicKeyFromPEM([]byte(crtData)); err == nil {
			cert = x
		} else {
			log.Fatal("Failed to load public key: ", err)
		}

	case "ES":
		if x, err := jwt.ParseECPrivateKeyFromPEM([]byte(keyData)); err == nil {
			key = x
		} else {
			log.Fatal("Failed to load private key: ", err)
		}
		if x, err := jwt.ParseECPublicKeyFromPEM([]byte(crtData)); err == nil {
			cert = x
		} else {
			log.Fatal("Failed to load public key: ", err)
		}

	default:
		log.Fatal("Invalid SIGNING_METHOD: ", sm)
	}

	return
}

func requireEnv(name, description string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatal("Env ", name, " is required: ", description)
	}
	return v
}

func getAuthenticator() api.Authenticator {

	switch v := os.Getenv("AUTH_BACKEND"); v {
	case "", "stupid":
		return stupidauth.New()

	case "file":
		return usersfile.New(requireEnv("AUTH_FILE", "File containings users when using file auth"))

	case "ldap-bind":
		return ldapbind.New(
			requireEnv("LDAP_SERVER", "LDAP server"),
			requireEnv("LDAP_USER", "LDAP user template (%s is substituted)"))

	case "etcd":
		return etcd.New(
			requireEnv("ETCD_PREFIX", "etcd prefix"),
			strings.Split(requireEnv("ETCD_ENDPOINTS", "etcd endpoints"), ","))

	case "mongo":
		return mongo.New(
			requireEnv("MONGO_DATABASE", "mongo database"),
			requireEnv("MONGO_COLLECTION", "mongo collection"),
			requireEnv("MONGO_FIELD", "field where to look the user (default: _id)"),
			requireEnv("MONGO_ENDPOINT", "mongo endpoint"))

	case "sql":
		return sql.New(
			requireEnv("SQL_DRIVER", "SQL driver (ex: postgres)"),
			requireEnv("SQL_DSN", "SQL destination"),
			requireEnv("SQL_USER_TABLE", "sql table with stored users"))
	default:
		log.Fatal("Unknown authenticator: ", v)
		return nil
	}
}
