package cache_test

import (
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	cache "github.com/morkid/gocache-elasticsearch/v7"
)

func TestCache(t *testing.T) {
	config := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	es, err := elasticsearch.NewClient(config)
	if nil != err {
		panic(err)
	}

	adapterConfig := cache.ElasticCacheConfig{
		Client:    es,
		Index:     "example",
		ExpiresIn: 10 * time.Second,
	}

	latency := func() {
		// network latency simulation
		time.Sleep(1 * time.Second)
	}

	adapter := *cache.NewElasticCache(adapterConfig)
	adapter.Set("foo", "bar")

	latency()
	if adapter.IsValid("foo") {
		value, err := adapter.Get("foo")
		if nil != err {
			t.Error(err)
		} else if value != "bar" {
			t.Error("value not equals to bar")
		} else {
			t.Log("Value ok", value)
		}
		adapter.Clear("foo")

		latency()
		if adapter.IsValid("foo") {
			t.Error("Failed to remove cache with key foo")
		} else {
			t.Log("cache with key foo was removed")
		}
	} else {
		t.Error("invalid foo value")
	}

	if err := adapter.Set("hello", "world"); nil != err {
		t.Error(err)
	}
	if err := adapter.Set("heli", "copter"); nil != err {
		t.Error(err)
	}
	if err := adapter.Set("kitty", "hello"); nil != err {
		t.Error(err)
	}

	latency()
	if err := adapter.ClearPrefix("hel"); nil != err {
		t.Error(err)
	}

	latency()
	if value, _ := adapter.Get("heli"); value != "" {
		t.Log("Value not removed", value)
		t.Error("Failed to remove key with prefix hel")
	} else {
		t.Log("All keys with prefix hel was removed")
	}

	latency()
	if err := adapter.ClearAll(); nil != err {
		t.Error(err)
	}

	latency()
	if value, _ := adapter.Get("kitty"); value != "" {
		t.Error("Failed to remove all keys")
	} else {
		t.Log("All keys was removed")
	}
}
