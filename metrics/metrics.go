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

type LogstashAuth struct {
	Hostname string
	Port     int
}

type ElasticAuth struct {
	Username string
	Password string
}

func (b *BulkIndexerWrapper) Add(eventBody interface{}) error {
	body := make(map[string]interface{})
	body["ctx"] = b.context
	body["body"] = eventBody
	body["timestamp"] = utils.NewTime().String()
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

func NewMetricHandler(logstashAuth LogstashAuth, elasticAuth ElasticAuth, context interface{}) (MetricHandler, error) {
	retryBackoff := backoff.NewExponentialBackOff()

	logstashUrl := fmt.Sprintf("http://%s:%d", logstashAuth.Hostname, logstashAuth.Port)
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
		Username:          elasticAuth.Username,
		Password:          elasticAuth.Password,
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

const (
	FAILURE    = "FAILURE"
	APIREQUEST = "API_REQUEST"
	ENTITY     = "ENTITY"
)

func (m *MetricHandler) AddApiRequestIndex(method string, url string, reqBody string, timeTakenMS int64, bodySize int) error {
	return m.bulkIndexer.Add(newApiRequestIndexBody(method, url, reqBody, timeTakenMS, bodySize))
}

func (m *MetricHandler) AddNewFailure(failureType string, err error) error {
	return m.bulkIndexer.Add(newFailureIndexBody(failureType, err))
}

type MetricModel interface {
	MetricIdentifier() string
}

func (m *MetricHandler) AddNewModel(model models.Model) {
	m.bulkIndexer.Add(newModel(model.TableName(), model))
}

func (m *MetricHandler) AddIngestFinishedIndex(stats ingest.SpotifyIngestStats) {
	m.bulkIndexer.Add(stats)
}

func newModel(tableName string, value interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	data["type"] = ENTITY
	data["tableName"] = tableName
	data["data"] = value

	return data
}

func newFailureIndexBody(failureType string, err error) map[string]interface{} {
	data := make(map[string]interface{})
	data["type"] = FAILURE
	data["failureType"] = failureType
	data["err"] = err.Error()

	return data
}

func newApiRequestIndexBody(method string, url string, reqBody string, timeTakenMS int64, bodySize int) map[string]interface{} {
	data := make(map[string]interface{})
	data["type"] = APIREQUEST
	data["url"] = url
	data["method"] = method
	data["reqBody"] = reqBody
	data["timeTaken"] = timeTakenMS
	data["bodySize"] = bodySize

	return data
}
