package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"spotify/utils"
	"time"

	"github.com/batzz-00/goutils/logger"
	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type BulkIndexerWrapper struct {
	bulkIndexer esutil.BulkIndexer
}

func (b *BulkIndexerWrapper) Add(item esutil.BulkIndexerItem) error {
	return b.bulkIndexer.Add(context.Background(), item)
}

type MetricHandler struct {
	bulkIndexer *BulkIndexerWrapper
}

func NewMetricHandler(logstashHost string, logstashPort int) (MetricHandler, error) {
	retryBackoff := backoff.NewExponentialBackOff()

	logstashUrl := fmt.Sprintf("http://%s:%d", logstashHost, logstashPort)
	fmt.Println(logstashUrl)
	esClientRetryHandler := func(i int) time.Duration {
		fmt.Println("retrybackoff", i)
		if i == 1 {
			retryBackoff.Reset()
		}
		return retryBackoff.NextBackOff()
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			logstashUrl,
		},
		RetryOnStatus:     []int{502, 503, 504, 429},
		EnableDebugLogger: true,
		RetryBackoff:      esClientRetryHandler,
		MaxRetries:        5,
		Username:          utils.MustGetEnv("elastic_username"),
		Password:          utils.MustGetEnv("elastic_password"),
	})
	if err != nil {
		return MetricHandler{}, err
	}

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:  "spotfy_api", // The default index name
		Client: es,           // The Elasticsearch client
	})
	if err != nil {
		return MetricHandler{}, err
	}

	return MetricHandler{
		bulkIndexer: &BulkIndexerWrapper{bulkIndexer: bi},
	}, nil

}

func (metric *MetricHandler) OnSuccess(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem) {
	fmt.Println("SUCCESS--")
	logger.Log(fmt.Sprintf("Added item with id %s succesfully", item.DocumentID), logger.Trace)
	fmt.Println("SUCCESS--")
}

func (metric *MetricHandler) OnFailure(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem, err error) {
	fmt.Println("FAIL--")
	logger.Log(fmt.Sprintf("Failed to add item with ID %s, err %s", item.DocumentID, err.Error()), logger.Error)
	fmt.Println("FAIL--")
}

func (metric *MetricHandler) BiCtx() IndexItemContext {
	return IndexItemContext{
		OnSuccess: metric.OnSuccess,
		OnFailure: metric.OnFailure,
	}
}

func (m *MetricHandler) Close() error {
	fmt.Println("closing bulk indexer")
	fmt.Println(m.bulkIndexer.bulkIndexer.Stats())
	err := m.bulkIndexer.bulkIndexer.Close(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("closed bulk indexer")

	return errors.New("olaa")
}

func (m *MetricHandler) AddApiRequestIndex(url string, reqBody string) error {
	return m.bulkIndexer.Add(newApiRequestIndex(m.BiCtx(), url, reqBody))
}

type IndexItemContext struct {
	OnSuccess func(context.Context, esutil.BulkIndexerItem, esutil.BulkIndexerResponseItem)
	OnFailure func(context.Context, esutil.BulkIndexerItem, esutil.BulkIndexerResponseItem, error)
}

func newSongIndex(ctx IndexItemContext, id string, spotifyId string, reqBody string) esutil.BulkIndexerItem {
	data := make(map[string]interface{})
	data["spotifyId"] = spotifyId
	data["reqBody"] = reqBody
	body, _ := json.Marshal(data)

	fmt.Println(string(body))

	return esutil.BulkIndexerItem{
		Action:     "index",
		DocumentID: id,
		Body:       bytes.NewReader(body),
		OnSuccess:  ctx.OnSuccess,
		OnFailure:  ctx.OnFailure,
	}
}

func newApiRequestIndex(ctx IndexItemContext, url string, reqBody string) esutil.BulkIndexerItem {
	data := make(map[string]interface{})
	data["url"] = url
	data["reqBody"] = reqBody
	body, _ := json.Marshal(data)

	return esutil.BulkIndexerItem{
		Index:     "spotify",
		Action:    "index",
		Body:      bytes.NewReader(body),
		OnSuccess: ctx.OnSuccess,
		OnFailure: ctx.OnFailure,
	}
}
