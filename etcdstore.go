package locker

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/go-log/log"
)

// EtcdStore is a backing store for Locker which uses Etcd for storage.
type EtcdStore struct {
	EtcdClientv3 *clientv3.Client

	// Version of the etcd
	Version string

	// Directory in Etcd to store locks. Default: locker.
	Directory string

	// TTL is the time-to-live for the lock. Default: 5s.
	TTL int64

	Log log.Logger
}

// Get returns the value of a lock. LockNotFound will be returned if a
// lock with the name isn't held.
func (s EtcdStore) Get(name string) (string, error) {
	resp, err := s.EtcdClientv3.Get(context.Background(), name)
	fmt.Println(resp)
	if resp.Count > 0 {
		return string(resp.Kvs[0].Value), err
	}
	return "", nil
}

// AcquireOrFreshenLock will aquires a named lock if it isn't already
// held, or updates its TTL if it is.
func (s EtcdStore) AcquireOrFreshenLock(name, value string) error {
	lresp, err := s.EtcdClientv3.Grant(context.Background(), s.lockTTL())
	if err != nil {
		return err
	}
	_, err = s.EtcdClientv3.Put(context.Background(), name, value, clientv3.WithLease(lresp.ID))
	//fmt.Println(resp)
	fmt.Println(err)
	return err
}

func (s EtcdStore) Delete(name string) error {
	resp, err := s.EtcdClientv3.Delete(context.Background(), name)
	fmt.Println(resp)
	return err
}

// directory will return the provided Directory or locker if nil.
func (s EtcdStore) directory() string {
	if s.Directory == "" {
		return "locker"
	}

	return s.Directory
}

// lockPath gets the path to a lock in Etcd. Defaults to /locker/name
func (s EtcdStore) lockPath(name string) string {
	return s.directory() + "/" + name
}

// lockTTL gets the TTL of the locks being stored in Etcd. Defaults to
// 5 seconds.
func (s EtcdStore) lockTTL() int64 {
	if s.TTL <= 0 {
		return 5
	}

	return s.TTL
}
