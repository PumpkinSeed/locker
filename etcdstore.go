package locker

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/client"
	"github.com/coreos/go-etcd/etcd"
	"github.com/coreos/go-log/log"
)

// EtcdStore is a backing store for Locker which uses Etcd for storage.
type EtcdStore struct {
	// Etcd client used for storing locks.
	Etcd *etcd.Client

	EtcdClient client.Client

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

	s.Log.Infof("GET %s", name)
	kapi := client.NewKeysAPI(s.EtcdClient)
	res, err := kapi.Get(context.Background(), s.lockPath(name), nil)
	if err == nil {
		fmt.Println(res.Node.Value)
		return res.Node.Value, nil
	}

	if etcderr, ok := err.(*etcd.EtcdError); ok && etcderr.ErrorCode == 100 {
		s.Log.Errorf("GET %s failed: Lock not found", name)
		return "", LockNotFound{name}
	}

	s.Log.Errorf("GET %s failed: %s", name, err)
	return "", err
}

// AcquireOrFreshenLock will aquires a named lock if it isn't already
// held, or updates its TTL if it is.
func (s EtcdStore) AcquireOrFreshenLock(name, value string) error {
	fmt.Printf("ACQUIRE %s\n", name)
	s.Log.Infof("ACQUIRE %s", name)
	if err := s.ensureLockDirectoryCreated(); err != nil {
		s.Log.Errorf("ACQUIRE %s failed ensuring lock directory exists: %s", name, err)
		return err
	}

	key := s.lockPath(name)

	// client

	// end client

	s.Log.Debugf("ACQUIRE %s CompareAndSwap on %s", name, key)
	_, err := s.Etcd.CompareAndSwap(key, value, s.lockTTL(), value, 0)
	if err == nil {
		// success!
		return nil
	}

	s.Log.Debugf("ACQUIRE %s CompareAndSwap on %s failed (%s) trying to recover", name, key, err)
	if etcderr, ok := err.(*etcd.EtcdError); ok {
		switch etcderr.ErrorCode {
		case 100:
			s.Log.Debugf("ACQUIRE %s CompareAndSwap on %s key didn't exist, trying to force set it", name, key)
			// key doesn't exist, set it. This seems to be odd behaviour for
			// CompareAndSwap. Surely, if it doesn't exist we should just set
			// it as part of CompareAndSwap. Potential for a race condition here,
			// where another client is able to do a CompareAndSwap and then we
			// stomp on it with our dumb Set.
			kapi := client.NewKeysAPI(s.EtcdClient)
			if _, err := kapi.Set(context.Background(), key, value, &client.SetOptions{TTL: 1}); err != nil {
				// wasn't able to force-set the key, no idea what happened
				s.Log.Errorf("ACQUIRE %s Set on %s key failed", name, key, err)
				return err
			}

			// Retry after stomping
			s.Log.Debugf("ACQUIRE %s retrying", name)
			return s.AcquireOrFreshenLock(name, value)
		case 101:
			// couldn't set the key, the prevValue we gave it differs from the
			// one in the server. Someone else has this key.
			s.Log.Errorf("ACQUIRE %s CompareAndSwap on %s key failed, lock denied", name, key)
			return LockDenied{name}
		}
	}

	return err
}

func (s EtcdStore) Delete(name string) error {
	fmt.Printf("DELETE %s\n", name)
	kapi := client.NewKeysAPI(s.EtcdClient)
	_, err := kapi.Delete(context.Background(), s.lockPath(name), &client.DeleteOptions{Recursive: true})
	fmt.Printf("DELETED %s\n", name)
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
