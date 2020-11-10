package main

import (
	"flag"

	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"

	companionapi "github.com/isi-nc/autentigo/pkg/companion-api/api"
	"github.com/isi-nc/autentigo/pkg/companion-api/backend"
	"github.com/isi-nc/autentigo/pkg/companion-api/backend/etcd"
	"github.com/isi-nc/autentigo/pkg/companion-api/backend/sql"
	"github.com/isi-nc/autentigo/pkg/companion-api/backend/users-file"
	"github.com/isi-nc/autentigo/pkg/rbac"
)

var (
	bind            = flag.String("bind", ":8181", "HTTP bind specification")
	rbacFile        = flag.String("rbac-file", "/etc/autentigo/rbac.yaml", "HTTP bind specification")
	disableCORS     = flag.Bool("no-cors", false, "Disable CORS support")
	adminToken      = flag.String("admin-token", "", "Admin Token")
	disableSecurity = flag.Bool("no-security", false, "Disable security, no auth required to call companion-api")

	validationCrt []byte
)

func main() {
	flag.Parse()

	var err error

	crtData := requireEnv("TLS_CRT", "certificate used to validate tokens")

	if os.Getenv("DISABLE_SECURITY") == "true" {
		*disableSecurity = true
		log.Println("Security disabled...")
	} else {
		*adminToken = requireEnv("ADMIN_TOKEN", "Admin token")
	}

	if rbac.Default, err = rbac.FromFile(*rbacFile); err != nil {
		log.Fatal("failed to load RBAC rules: ", err)
	}

	if len(crtData) == 0 {
		log.Fatal("Certificate empty, failed to load")
	}
	rbac.DefaultValidationCertificate = []byte(crtData)

	cAPI := &companionapi.CompanionAPI{
		Client:     getBackEndClient(),
		AdminToken: *adminToken,
		DisableSecurity: *disableSecurity,
	}

	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	restful.DefaultContainer.Router(restful.CurlyRouter{})

	for _, ws := range cAPI.WebServices() {
		restful.Add(ws)
	}

	config := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
		APIPath:     "/apidocs.json",
	}
	restful.Add(restfulspec.NewOpenAPIService(config))

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

	log.Fatal(http.Serve(l, restful.DefaultContainer))
}

func getBackEndClient() backend.Client {
	switch v := os.Getenv("AUTH_BACKEND"); v {
	case "stupid":
		log.Fatal("Stupid backend does not need the companion-api")
		return nil
	case "ldap-bind":
		log.Fatal("Please feel free to use a ldap client instead of the companion-api")
		return nil
	case "file":
		return usersfile.New(requireEnv("AUTH_FILE", "File containings users when using file auth"))
	case "etcd":
		return etcd.New(
			requireEnv("ETCD_PREFIX", "etcd prefix"),
			strings.Split(requireEnv("ETCD_ENDPOINTS", "etcd endpoints"), ","))
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

func requireEnv(name, description string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatal("Env ", name, " is required: ", description)
	}
	return v
}
