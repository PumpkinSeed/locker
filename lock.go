package locker

import (
	"errors"
	"fmt"
	"time"
)

type lockState int

const (
	unknown  lockState = iota
	acquired lockState = iota
	released lockState = iota
)

const DefaultValue = "ok"

// Lock will create a lock for a key and set its value. If the owned
// channel is provided a bool will be pushed whenever our ownership of
// the lock changes. Pushing true into the quit channel will stop the
// locker from refreshing the lock and let it expire if we own it.
//
//     owned := make(chan bool)
//
//     go client.Lock("my-service", "http://10.0.0.1:9292", owned, nil)
//
//     for {
//       select {
//       case v := <-owned:
//         fmt.Printf("Lock ownership changed: %t\n", v)
//       }
//     }
//
// Lock is a blocking call, so it's recommended to run it in a goroutine.
func (c Client) Lock(name, value string, owned chan<- bool, quit <-chan bool) Report {
	report := c.Inspect(name)

	var doneCh = make(chan bool)
	go c.lock(name, value, owned, quit, doneCh)
	var done = <-doneCh
	if done {
		return report
	}

	report.Err = errors.New("error happened")
	return report
}

func (c Client) lock(name, value string, owned chan<- bool, quit <-chan bool, done chan<- bool) error {
	fmt.Println("UPDATE")
	state, err := c.updateNode(name, value)
	if err != nil {
		return err
	}
	lastState := state
	tick := time.Tick(time.Millisecond * 500)

	fmt.Println("DONE")
	done <- true

	for {
		select {
		case <-quit:
			return nil
		case <-tick:
			if lastState != state {
				lastState = state
			}

		}
	}

	panic("unreachable")
}

func (c Client) Unlock(name string, quit chan<- bool) error {
	quit <- true
	return c.Store.Delete(name)
}

// updateNode will update the lock node in the cluster, effectively just
// updating the TTL of the key and ensuring our value is still in it.
func (c Client) updateNode(name, value string) (lockState, error) {
	if err := c.Store.AcquireOrFreshenLock(name, value); err != nil {
		if _, ok := err.(LockDenied); ok {
			return released, nil
		}

		// no idea what just happened, just return the error
		return unknown, err
	}

	return acquired, nil
}
