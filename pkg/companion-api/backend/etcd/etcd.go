package etcd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/isi-nc/autentigo/pkg/companion-api/api"
	"github.com/isi-nc/autentigo/pkg/companion-api/backend"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdClient struct {
	prefix  string
	client  *clientv3.Client
	timeout time.Duration
}

// New Client to manage users with an etcd backend
func New(prefix string, endpoints []string) backend.Client {
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

	return &etcdClient{
		prefix:  prefix,
		client:  client,
		timeout: timeout,
	}
}

var _ backend.Client = &etcdClient{}

func (e *etcdClient) CreateUser(id string, user *backend.UserData) (err error) {
	oldUser := &backend.UserData{}
	oldUser, err = e.getUser(id)

	if oldUser != nil {
		err = api.ErrUserAlreadyExist
	} else if cmp.Equal(err, api.ErrMissingUser) {
		err = e.putUser(id, user)
	}

	return
}

func (e *etcdClient) UpdateUser(id string, update func(user *backend.UserData) error) (err error) {
	user := &backend.UserData{}
	user, err = e.getUser(id)

	if err == nil && user != nil {
		err = update(user)
		if err == nil {
			err = e.putUser(id, user)
		}
	}

	return
}

func (e *etcdClient) DeleteUser(id string) (err error) {
	user := &backend.UserData{}
	user, err = e.getUser(id)

	if err == nil && user != nil {
		ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
		defer cancel()

		_, err = e.client.Delete(ctx, path.Join(e.prefix, id))
	}
	return
}

func (e *etcdClient) getUser(id string) (userData *backend.UserData, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.client.Get(ctx, path.Join(e.prefix, id))
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		err = api.ErrMissingUser
		return
	}

	user := &backend.User{}
	err = json.Unmarshal(resp.Kvs[0].Value, user)
	userData = user.ToUserData()
	return
}

func (e *etcdClient) putUser(id string, user *backend.UserData) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	u, err := json.Marshal(*user.ToUser())
	if err == nil {
		_, err = e.client.Put(ctx, path.Join(e.prefix, id), string(u))
	}

	return

}
