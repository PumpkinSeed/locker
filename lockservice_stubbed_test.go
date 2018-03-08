package locker

import (
	"context"
	"testing"
	"time"
)

const name = "myservice"

func TestWatch(t *testing.T) {
	store := &memoryStore{}
	client := Client{store, context.Background()}

	valueChanges := make(chan string)
	quit := make(chan bool)

	go client.Watch(name, valueChanges, quit)

	select {
	case <-timeout():
		t.Fatal("Timeout")
	case change := <-valueChanges:
		if change != "" {
			t.Error("Expected initial blank value inidicating missing lock")
		}
	}

	store.set(name, "value")

	select {
	case <-timeout():
		t.Fatal("Timeout")
	case change := <-valueChanges:
		if change != "value" {
			t.Error("Expected value to change to 'value'")
		}
	}

	quit <- true
}

// For testing purposes
type memoryStore struct {
	cache map[string]string
}

func (c *memoryStore) ensureCache() {
	if c.cache == nil {
		c.cache = make(map[string]string)
	}
}

func (c *memoryStore) set(key, value string) {
	c.ensureCache()
	c.cache[key] = value
}

func (c *memoryStore) Get(ctx context.Context, name string) (string, error) {
	c.ensureCache()

	if v, ok := c.cache[name]; ok {
		return v, nil
	}

	return "", LockNotFound{name}
}

func (c *memoryStore) AcquireOrFreshenLock(ctx context.Context, name, value string) error {
	c.ensureCache()

	if v, ok := c.cache[name]; ok {
		if v != value {
			return LockDenied{name}
		}
	}

	c.cache[name] = value
	return nil
}

func (c *memoryStore) Delete(ctx context.Context, name string) error {
	return nil
}

func timeout() <-chan time.Time {
	return time.After(10 * time.Second)
}
