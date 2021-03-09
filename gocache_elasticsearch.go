package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/morkid/gocache"

	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
)

// ElasticCacheConfig struct config
type ElasticCacheConfig struct {
	Client    *elasticsearch.Client
	Index     string
	ExpiresIn time.Duration
}

// NewElasticCache func
func NewElasticCache(config ElasticCacheConfig) *gocache.AdapterInterface {
	if nil == config.Client {
		panic("Client config is required")
	}

	if config.Index == "" {
		config.Index = "gocache"
	}

	if config.ExpiresIn <= 0 {
		config.ExpiresIn = 3600 * time.Second
	}

	var adapter gocache.AdapterInterface = &elasticCache{
		Client:    config.Client,
		Index:     config.Index,
		ExpiresIn: config.ExpiresIn,
	}

	return &adapter
}

type hit struct {
	Source *documentObject `json:"_source"`
}

type hits struct {
	Hits *[]hit `json:"hits"`
}

type response struct {
	Hits hits `json:"hits"`
}

type documentObject struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type elasticCache struct {
	Client    *elasticsearch.Client
	Index     string
	ExpiresIn time.Duration
}

func (e elasticCache) Set(key string, value string) error {
	es := e.Client
	data := documentObject{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
	}
	bte, err := json.Marshal(data)
	if nil != err {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(es *elasticsearch.Client, index string, data string) {
		req := esapi.IndexRequest{
			Index:      index,
			DocumentID: key,
			Body:       strings.NewReader(data),
			Refresh:    "true",
		}
		res, err := req.Do(context.Background(), es)
		if nil != err {
			log.Println(err)
		}
		defer res.Body.Close()
		wg.Done()
	}(es, e.Index, string(bte))
	wg.Wait()
	return nil
}

func (e elasticCache) Get(key string) (string, error) {
	result, err := e.find(key)
	if nil != err {
		return "", err
	}

	if e.isExpired(result) {
		return "", errors.New("Cache expired")
	}

	return result.Value, nil
}

func (e elasticCache) IsValid(key string) bool {
	result, err := e.Get(key)
	if err == nil && result != "" {
		return true
	}
	return false
}

func (e elasticCache) Clear(key string) error {
	query := map[string]map[string]map[string]string{
		"query": {
			"match": {
				"key": key,
			},
		},
	}

	return e.deleteByQuery(query)
}

func (e elasticCache) ClearPrefix(keyPrefix string) error {
	query := map[string]map[string]map[string]string{
		"query": {
			"prefix": {
				"key": keyPrefix,
			},
		},
	}

	return e.deleteByQuery(query)
}

func (e elasticCache) ClearAll() error {
	query := map[string]map[string]map[string]string{
		"query": {
			"match_all": {},
		},
	}

	return e.deleteByQuery(query)
}
func (e elasticCache) isExpired(source *documentObject) bool {
	if nil == source {
		return true
	}
	expired := time.Now().Sub(source.CreatedAt) > e.ExpiresIn
	if expired {
		var wg sync.WaitGroup
		wg.Add(1)
		go func(key string) {
			e.Clear(key)
			wg.Done()
		}(source.Key)
		wg.Wait()
	}
	return expired
}

func (e elasticCache) find(key string) (*documentObject, error) {
	var search bytes.Buffer
	query := map[string]map[string]map[string]string{
		"query": {
			"match": {
				"key": key,
			},
		},
	}
	if err := json.NewEncoder(&search).Encode(query); err != nil {
		return nil, err
	}
	es := e.Client
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(e.Index),
		es.Search.WithBody(&search),
		es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var er map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&er); err != nil {
			return nil, err
		}
		if nil != er["error"] {
			return nil, fmt.Errorf("[%s] %s: %s",
				res.Status(),
				er["error"].(map[string]interface{})["type"],
				er["error"].(map[string]interface{})["reason"],
			)
		}
	}

	var r response
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	if nil != r.Hits.Hits && len(*r.Hits.Hits) > 0 {
		h := *r.Hits.Hits
		if result := h[0]; nil != result.Source {
			return result.Source, nil
		}
	}

	return nil, errors.New("Not found")
}

func (e elasticCache) deleteByQuery(query map[string]map[string]map[string]string) error {
	es := e.Client
	var wg sync.WaitGroup
	wg.Add(1)
	go func(query map[string]map[string]map[string]string) {
		var search bytes.Buffer
		if err := json.NewEncoder(&search).Encode(query); err != nil {
			log.Println(err)
			return
		}
		res, err := es.DeleteByQuery([]string{e.Index}, &search)
		if nil != err {
			log.Println(err)
			return
		}

		defer res.Body.Close()
		if res.IsError() {
			var er map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&er); err != nil {
				log.Println(err)
				return
			}
			if nil != er["error"] {
				log.Println(fmt.Errorf("[%s] %s: %s",
					res.Status(),
					er["error"].(map[string]interface{})["type"],
					er["error"].(map[string]interface{})["reason"],
				))
			}
		}
		wg.Done()
	}(query)
	wg.Wait()

	return nil
}
