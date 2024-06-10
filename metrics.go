package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"spotify/ingest"
	"spotify/utils"
	"time"

	"github.com/batzz-00/goutils/logger"
	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type BulkIndexerWrapper struct {
	bulkIndexer   esutil.BulkIndexer
	ingestContext ingest.SpotifyIngestContext
}

func (b *BulkIndexerWrapper) Add(body map[string]interface{}) error {
	body["ctx"] = b.ingestContext
	byteBody, _ := json.Marshal(body)

	item := esutil.BulkIndexerItem{
		Action:    "index",
		Body:      bytes.NewReader(byteBody),
		OnSuccess: b.OnSuccess,
		OnFailure: b.OnFailure,
	}

	return b.bulkIndexer.Add(context.Background(), item)
}

func (wrapper *BulkIndexerWrapper) OnSuccess(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem) {
	logger.Log(fmt.Sprintf("Added item with id %s (index %s) to event log", item.DocumentID, item.Index), logger.Trace)
}

func (wrapper *BulkIndexerWrapper) OnFailure(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem, err error) {
	logger.Log(fmt.Sprintf("Failed to add item with ID %s (index %s), err %s to event log", item.DocumentID, item.Index, err.Error()), logger.Error)
}

type MetricHandler struct {
	bulkIndexer *BulkIndexerWrapper
}

func NewMetricHandler(logstashHost string, logstashPort int, context ingest.SpotifyIngestContext) (MetricHandler, error) {
	retryBackoff := backoff.NewExponentialBackOff()

	logstashUrl := fmt.Sprintf("http://%s:%d", logstashHost, logstashPort)
	esClientRetryHandler := func(i int) time.Duration {
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
		Index:  "spotfy_api",
		Client: es,
	})

	if err != nil {
		return MetricHandler{}, err
	}

	return MetricHandler{
		bulkIndexer: &BulkIndexerWrapper{bulkIndexer: bi, ingestContext: context},
	}, nil

}

func (m *MetricHandler) Close() error {
	err := m.bulkIndexer.bulkIndexer.Close(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (m *MetricHandler) AddApiRequestIndex(method string, url string, reqBody string) error {
	return m.bulkIndexer.Add(newApiRequestIndexBody(method, url, reqBody))
}

func (m *MetricHandler) AddNewSongIndex(spotifyId string, songName string) error {
	return m.bulkIndexer.Add(newSongIndexBody(spotifyId, songName))
}

func newSongIndexBody(spotifyId string, songName string) map[string]interface{} {
	data := make(map[string]interface{})
	data["spotifyId"] = spotifyId
	data["songName"] = songName

	return data
}

func newApiRequestIndexBody(method string, url string, reqBody string) map[string]interface{} {
	data := make(map[string]interface{})
	data["url"] = url
	data["method"] = method
	data["reqBody"] = reqBody

	return data
}
