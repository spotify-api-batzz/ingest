package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	bulkIndexer BulkIndexerWrapper
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
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff:  esClientRetryHandler,
		MaxRetries:    5,
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
		bulkIndexer: BulkIndexerWrapper{bulkIndexer: bi},
	}, nil

}

func (metric *MetricHandler) OnSuccess(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem) {
	logger.Log(fmt.Sprintf("Added item with id %s succesfully", item.DocumentID), logger.Trace)
}

func (metric *MetricHandler) OnFailure(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem, err error) {
	logger.Log(fmt.Sprintf("Failed to add item with ID %s, err %s", item.DocumentID, err.Error()), logger.Error)
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

type IndexItemContext struct {
	OnSuccess func(context.Context, esutil.BulkIndexerItem, esutil.BulkIndexerResponseItem)
	OnFailure func(context.Context, esutil.BulkIndexerItem, esutil.BulkIndexerResponseItem, error)
}

func newSongIndex(ctx IndexItemContext, id string, spotifyId string, name string) esutil.BulkIndexerItem {
	data := make(map[string]interface{})
	data["spotifyId"] = spotifyId
	body, _ := json.Marshal(data)

	fmt.Println(body)

	return esutil.BulkIndexerItem{
		Action:     "index",
		DocumentID: id,
		Body:       bytes.NewReader(body),
		OnSuccess:  ctx.OnSuccess,
		OnFailure:  ctx.OnFailure,
	}
}

func newApiRequestIndex(ctx IndexItemContext) esutil.BulkIndexerItem {
	data := make(map[string]interface{})
	fmt.Println("heh")
	data["spotifyId"] = "test"
	body, _ := json.Marshal(data)
	fmt.Println(body)

	return esutil.BulkIndexerItem{
		Index:      "tester",
		Action:     "index",
		DocumentID: "123",
		Body:       bytes.NewReader(body),
		OnSuccess:  ctx.OnSuccess,
		OnFailure:  ctx.OnFailure,
	}
}
