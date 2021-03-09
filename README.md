# cache elasticsearch adapter
[![Go Reference](https://pkg.go.dev/badge/github.com/morkid/gocache-elasticsearch/v5.svg)](https://pkg.go.dev/github.com/morkid/gocache-elasticsearch/v5)
[![Go](https://github.com/morkid/gocache-elasticsearch/actions/workflows/go.yml/badge.svg)](https://github.com/morkid/gocache-elasticsearch/actions/workflows/go.yml)
[![Build Status](https://travis-ci.com/morkid/gocache-elasticsearch.svg?branch=master)](https://travis-ci.com/morkid/gocache-elasticsearch)
[![Go Report Card](https://goreportcard.com/badge/github.com/morkid/gocache-elasticsearch/v5)](https://goreportcard.com/report/github.com/morkid/gocache-elasticsearch/v5)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/morkid/gocache-elasticsearch)](https://github.com/morkid/gocache-elasticsearch/releases)

This library is created by implementing [gocache](https://github.com/morkid/gocache) 
and require [elasticsearch](https://github.com/elastic/go-elasticsearch) v5.

## Installation

```bash
go get -d github.com/morkid/gocache-elasticsearch/v5
```

Available versions:
- [github.com/morkid/gocache-elasticsearch/v7](https://github.com/morkid/gocache-elasticsearch/tree/v7) for [elasticsearch client v7](https://github.com/elastic/go-elasticsearch/tree/v7)
- [github.com/morkid/gocache-elasticsearch/v6](https://github.com/morkid/gocache-elasticsearch/tree/v6) for [elasticsearch client v6](https://github.com/elastic/go-elasticsearch/tree/v6)
- [github.com/morkid/gocache-elasticsearch/v5](https://github.com/morkid/gocache-elasticsearch/tree/v5) for [elasticsearch client v5](https://github.com/elastic/go-elasticsearch/tree/v5)


## Example usage
```go
package main

import (
    "time"
    "fmt"
    cache "github.com/morkid/gocache-elasticsearch/v5"
    "github.com/elastic/go-elasticsearch/v5"
)

func latency() {
    // network latency simulation
    // just for testing
    time.Sleep(1 * time.Second)
}

func main() {
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

    adapter := *cache.NewElasticCache(config)
    adapter.Set("foo", "bar")

    if adapter.IsValid("foo") {
        value, err := adapter.Get("foo")
        if nil != err {
            fmt.Println(err)
        } else if value != "bar" {
            fmt.Println("value not equals to bar")
        } else {
            fmt.Println(value)
        }
        adapter.Clear("foo")

        latency()
        if adapter.IsValid("foo") {
            fmt.Println("Failed to remove key foo")
        }
    }
}

```

## License

Published under the [MIT License](https://github.com/morkid/gocache-elasticsearch/blob/master/LICENSE).