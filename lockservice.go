package locker

import (
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/go-etcd/etcd"
)

const (
	Failure string = "failure"
	Success        = "success"
)

// Client is the main locker type. Use it to manage your locks. A locker
// Client has a Store which it uses to persist the locks.
type Client struct {
	// Store is what locker uses to persist locks.
	Store Store

	Quit chan bool
}

// New creates a default locker client using Etcd as a store. It requires
// you provide an etcd.Client, this is so locker doesn't make any dumb
// assumptions about how your Etcd cluster is configured.
//
//     client := locker.New(etcdclient)
//
func New(etcdclient *etcd.Client, etcdClient client.Client, etcdClientv3 *clientv3.Client) Client {
	return Client{
		Store: EtcdStore{
			Etcd:         etcdclient,
			EtcdClient:   etcdClient,
			EtcdClientv3: etcdClientv3,
		},
	}
}

// Get returns the value of a lock. LockNotFound will be returned if a
// lock with the name isn't held.
func (c Client) Get(name string) (string, error) {
	return c.Store.Get(name)
}

func (c Client) Inspect(name string) Report {
	v, err := c.Get(name)
	if err == nil && len(v) > 0 {
		return Report{
			Err: nil,
			Msg: Failure,
		}
	}
	return Report{
		Err: nil,
		Msg: Success,
	}
}

type Report struct {
	Err error
	Msg string
}

// Store is a persistance mechaism for locker to store locks. Needs to be
// able to support querying and an atomic compare-and-swap. Currently, the
// only implementation of a Store is EtcdStore.
type Store interface {
	// Get returns the value of a lock. LockNotFound will be returned if a
	// lock with the name isn't held.
	Get(name string) (string, error)

	// AcquireOrFreshenLock will aquires a named lock if it isn't already
	// held, or updates its TTL if it is.
	AcquireOrFreshenLock(name, value string) error

	Delete(name string) error
}
