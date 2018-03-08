package locker

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
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

	ctx context.Context
}

// New creates a default locker client using Etcd as a store. It requires
// you provide an etcd.Client, this is so locker doesn't make any dumb
// assumptions about how your Etcd cluster is configured.
//
//     client := locker.New(etcdclient)
//
func New(machines []string, timeout int64, ttl int64, ctx context.Context) (Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   machines,
		DialTimeout: time.Duration(timeout) * time.Second,
	})
	if err != nil {
		return Client{}, err
	}
	return Client{
		Store: EtcdStore{
			EtcdClientv3: cli,
			TTL:          ttl,
		},
		ctx: ctx,
	}, nil
}

// Get returns the value of a lock. LockNotFound will be returned if a
// lock with the name isn't held.
func (c Client) Get(name string) (string, error) {
	return c.Store.Get(c.ctx, name)
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
	Get(ctx context.Context, name string) (string, error)

	// AcquireOrFreshenLock will aquires a named lock if it isn't already
	// held, or updates its TTL if it is.
	AcquireOrFreshenLock(ctx context.Context, name, value string) error

	Delete(ctx context.Context, name string) error
}
