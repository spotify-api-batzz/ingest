package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"spotify/ingest"
	"spotify/models"
	"spotify/utils"
	"time"

	"github.com/batzz-00/goutils/logger"
	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type BulkIndexerWrapper struct {
	bulkIndexer esutil.BulkIndexer
	context     interface{}
}

func (b *BulkIndexerWrapper) Add(eventBody interface{}) error {
	body := make(map[string]interface{})
	body["ctx"] = b.context
	body["body"] = eventBody
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

func NewMetricHandler(logstashHost string, logstashPort int, context interface{}) (MetricHandler, error) {
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
		bulkIndexer: &BulkIndexerWrapper{bulkIndexer: bi, context: context},
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
	return m.bulkIndexer.Add(newCommonEntityBody(spotifyId, "songName", songName))
}

func (m *MetricHandler) AddNewAlbumIndex(spotifyId string, albumName string) error {
	return m.bulkIndexer.Add(newCommonEntityBody(spotifyId, "albumName", albumName))
}

func (m *MetricHandler) AddNewArtistIndex(spotifyId string, artistName string) error {
	return m.bulkIndexer.Add(newCommonEntityBody(spotifyId, "artistName", artistName))
}

func (m *MetricHandler) AddNewFailure(failureType string, err error) error {
	return m.bulkIndexer.Add(newFailureIndexBody(failureType, err))
}

func (m *MetricHandler) AddNewThumbnailIndex(entity string, name string, url string) error {
	return m.bulkIndexer.Add(newThumbnailIndexBody(entity, name, url))
}

func (m *MetricHandler) AddNewModel(model models.Model) {
	m.bulkIndexer.Add(newModel(model.TableName(), model))
}

func (m *MetricHandler) AddIngestFinishedIndex(stats ingest.SpotifyIngestStats) {
	m.bulkIndexer.Add(stats)
}

func newModel(tableName string, value interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	data["tableName"] = tableName
	data["data"] = value

	return data
}

func newThumbnailIndexBody(entity string, name string, url string) map[string]interface{} {
	data := make(map[string]interface{})
	data["entity"] = entity
	data["name"] = name
	data["url"] = url

	return data
}

func newFailureIndexBody(failureType string, err error) map[string]interface{} {
	data := make(map[string]interface{})
	data["failureType"] = failureType
	data["err"] = err.Error()

	return data
}

func newCommonEntityBody(spotifyId string, entityKey string, entityValue string) map[string]interface{} {
	data := make(map[string]interface{})
	data["spotifyId"] = spotifyId
	data[entityKey] = entityValue

	return data
}

func newApiRequestIndexBody(method string, url string, reqBody string) map[string]interface{} {
	data := make(map[string]interface{})
	data["url"] = url
	data["method"] = method
	data["reqBody"] = reqBody

	return data
}
