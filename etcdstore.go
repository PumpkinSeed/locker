package locker

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/go-etcd/etcd"
	"github.com/coreos/go-log/log"
)

// EtcdStore is a backing store for Locker which uses Etcd for storage.
type EtcdStore struct {
	// Etcd client used for storing locks.
	Etcd *etcd.Client

	EtcdClient client.Client

	EtcdClientv3 *clientv3.Client

	// Version of the etcd
	Version string

	// Directory in Etcd to store locks. Default: locker.
	Directory string

	// TTL is the time-to-live for the lock. Default: 5s.
	TTL int

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
	_, err := s.EtcdClientv3.Put(context.Background(), name, value)
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

// ensureLockDirectoryCreated tries to create the root locker directory in
// etcd. This is to compensate for etcd sometimes getting upset when all
// the nodes expire.
func (s EtcdStore) ensureLockDirectoryCreated() error {
	_, err := s.Etcd.CreateDir(s.directory(), 0)

	if eerr, ok := err.(*etcd.EtcdError); ok {
		if eerr.ErrorCode == 105 {
			return nil // key already exists, cool
		}
	}

	// not an etcderr, or a etcderror we want to propagate, or there was no error
	return err
}

// lockPath gets the path to a lock in Etcd. Defaults to /locker/name
func (s EtcdStore) lockPath(name string) string {
	return s.directory() + "/" + name
}

// lockTTL gets the TTL of the locks being stored in Etcd. Defaults to
// 5 seconds.
func (s EtcdStore) lockTTL() uint64 {
	if s.TTL <= 0 {
		return 5
	}

	return uint64(s.TTL)
}
