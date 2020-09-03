## Running

#### With file backend

```sh
export AUTH_BACKEND=file \
export AUTH_FILE="autentigo-users.csv" \
companion-api
```

#### With etcd backend

```sh
export AUTH_BACKEND=etcd \
export ETCD_ENDPOINTS=http://localhost:2379 \
export ETCD_PREFIX=/users \
companion-api
```

#### With sql backend

```sh
AUTH_BACKEND=sql SQL_DRIVER=postgres SQL_USER_TABLE=auth_users SQL_DSN="user=postgres password=postgres host=localhost dbname=postgres sslmode=disable" ./ag-companion-api --validation-cert=../../test/tls.crt --admin-token=toto --rbac-file=../../test/rbac.yaml
```

### Flags

```
companion-api --help
```

### Environment

| Variable         | Description                                                                            |
| ---------------- | -------------------------------------------------------------------------------------- |
| `AUTH_BACKEND`   | Choose an authentication backend (required)                                            |
| `AUTH_FILE`      | Backend file (required if `AUTH_BACKEND`=file)                                         |
| `ETCD_TIMEOUT`   | Simple etcd timeout (default: 5s)                                                      |
| `ETCD_PREFIX`    | Prefix before the etcd key (default: none)                                             |
| `ETCD_ENDPOINTS` | Etcd endpoints (format: `ETCD_ENDPOINTS`=http://localhost:2379,http://localhost:4001 ) |

### Auth backends

#### stupid

Stupid backend does not need the companion-api.

#### file

Reads or update a content file, defined by the `AUTH_FILE` env, in the format:

```
<user name>:<password SHA256 (hex)>:email:email_validated:groups
```

#### LDAP simple bind

Please feel free to use a ldap client instead of the companion-api.

#### etcd lookup

Update or looks up the user in etcd, with a key like `prefix/user-name`. Takes an optionnal `ETCD_TIMEOUT` to change the lookup timeout.

### Tests

```sh
curl -i -H'Content-Type: application/json' -H 'Authorization: Bearer toto' localhost:8181/users -d '{"id":"hahaguy","user":{"password":"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8","display_name":"Hahaguy","email":"hahaguy@toto.net","email_verified":false,"groups":["self-service"]}}'
curl -i --basic --user hahaguy:password localhost:8080/basic
```