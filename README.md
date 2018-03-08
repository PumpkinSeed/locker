# locker

A distributed lock service client for [etcd](https://github.com/coreos/etcd).

**Forked from github.com/jagregory/locker**

#### Features

- Lock/Unlock mechanism
- Migrate to `github.com/coreos/etcd/clientv3`
- TTL


[![Godoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/PumpkinSeed/locker)
[![Build Status](https://travis-ci.org/PumpkinSeed/locker.svg?branch=master)](https://travis-ci.org/PumpkinSeed/locker)

## What? Why?

A distributed lock service is somewhat self-explanatory. Locking (mutexes) as a service that's distributed across a cluster.

What a lock service gives you is a key-value store with locking semantics, and a mechanism for watching state changes. The distributed part of a distributed lock service is thanks to the etcd underpinnings, and it simply means your locks are persisted across a cluster.

## Usage

### Creating a lock and release it with the Unlock

```go
import "github.com/PumpkinSeed/locker"

// Add etcd cluster nodes
var machines = []string{"http://127.0.0.1:2379", "http://127.0.0.1:2380", "http://127.0.0.1:2381"}

// Add timeout for etcd DialTimeout option
var timeout = int64(5)

// Add ttl (time-to-live)
var ttl = int64(5)

// Create the locker client.
client, err := locker.New(machines, timeout, ttl, context.Background())

var key = "1231231-123123-123"
quit := make(chan bool)
report := client.Lock(key, locker.DefaultValue, quit)
// report has a Msg and an Err field, Msg will contains 'success' or 'fail' operations.
```

### Report

- Report returned by the `Lock`, it has a Msg and an Err field
- The Msg Success if the key not locked yet, and Fail if the key locked
- The Err will return an error occured among the lock mechanism

### Releasing the lock

The third argument to `Lock` is a `quit` channel. Push anything into this channel to kill the locking, which will delete the locked key from the etcd.

```go
client.Unlock(key, quit)
```

### Watching a lock (deprecated)

An interesting aspect of lock services is the ability to watch a lock that isn't owned by you. A service can alter its behaviour depending on the value of a lock. You can use the `Watch` function to watch the value of a lock.

```go
valueChanges := make(chan string)
go client.Watch("key", valueChanges, nil)

select {
case change := <-valueChanges
	fmt.Printf("Lock value changed: %s", change)
}
```

Quitting works the same way as `Lock`.

## Contribution

- Docker environment provided for testing etcd in the docker directory.
- utils.go has a debug option, currently in `false` if it's `true` than it will log debug messages to the os.Stdout
