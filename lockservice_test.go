package locker

import (
	"context"
	"fmt"
	_ "fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

var machines = []string{"http://127.0.0.1:2379", "http://127.0.0.1:2380", "http://127.0.0.1:2381"}

var etcdclient *etcd.Client

func TestTestLockAndUnlock(t *testing.T) {
	client, err := New(machines, 5, 5, context.Background())
	if err != nil {
		t.Error(err)
	}

	key := randomKey()

	quit := make(chan bool)
	report := client.Lock(key, DefaultValue, quit)
	if report.Err != nil {
		t.Error(report.Err)
	}
	if report.Msg != Success {
		t.Errorf("Report message should be '%s', instead of %s", Success, report.Msg)
	}

	report = client.Lock(key, DefaultValue, quit)
	if report.Err != nil {
		t.Error(report.Err)
	}
	if report.Msg != Fail {
		t.Errorf("Report message should be '%s', instead of %s", Fail, report.Msg)
	}

	client.Unlock(key, quit)

	report = client.Lock(key, DefaultValue, quit)
	if report.Err != nil {
		t.Error(report.Err)
	}
	if report.Msg != Success {
		t.Errorf("Report message should be '%s', instead of %s", Success, report.Msg)
	}

	client.Unlock(key, quit)
}

func TestTTL(t *testing.T) {
	ttl := int64(3)
	client, err := New(machines, 3, ttl, context.Background())
	if err != nil {
		t.Error(err)
	}

	key := randomKey()

	quit := make(chan bool)
	report := client.Lock(key, DefaultValue, quit)
	if report.Err != nil {
		t.Error(report.Err)
	}
	if report.Msg != Success {
		t.Errorf("Report message should be '%s', instead of %s", Success, report.Msg)
	}

	time.Sleep(time.Duration(ttl*1000+500) * time.Millisecond)

	report = client.Lock(key, DefaultValue, quit)
	if report.Err != nil {
		t.Error(report.Err)
	}
	if report.Msg != Success {
		t.Errorf("Report message should be '%s', instead of %s", Success, report.Msg)
	}

	client.Unlock(key, quit)
}

func BenchmarkLockAndUnlock(b *testing.B) {
	client, err := New(machines, 3, 3, context.Background())
	if err != nil {
		b.Error(err)
	}

	key := randomKey()
	quit := make(chan bool)
	for i := 0; i < b.N; i++ {
		client.Lock(key, DefaultValue, quit)

		client.Unlock(key, quit)
	}
}
func BenchmarkLock(b *testing.B) {
	client, err := New(machines, 3, 3, context.Background())
	if err != nil {
		b.Error(err)
	}

	key := randomKey()
	quit := make(chan bool)
	for i := 0; i < b.N; i++ {
		client.Lock(key+strconv.Itoa(i), DefaultValue, quit)
	}
}

func randomKey() string {
	return fmt.Sprintf("key%s", strconv.Itoa(random(10000, 99999)))
}

func random(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
