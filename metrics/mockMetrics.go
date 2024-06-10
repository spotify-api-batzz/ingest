package metrics

import (
	"spotify/ingest"
)

type MockMetricHandler struct {
	bulkIndexer *BulkIndexerWrapper
}

func NewMockMetricHandler() MockMetricHandler {
	return MockMetricHandler{}
}

func (m *MockMetricHandler) Close() error {
	return nil
}

func (m *MockMetricHandler) AddApiRequestIndex(method string, url string, reqBody string) error {
	return nil
}

func (m *MockMetricHandler) AddNewSongIndex(spotifyId string, songName string) error {
	return nil
}

func (m *MockMetricHandler) AddNewAlbumIndex(spotifyId string, albumName string) error {
	return nil
}

func (m *MockMetricHandler) AddNewFailure(failureType string, err error) error {
	return nil
}

func (m *MockMetricHandler) AddNewThumbnailIndex(entity string, name string, url string) error {
	return nil
}

func (m *MockMetricHandler) AddNewEntity(entity string, name string, url string) error {
	return nil
}

func (m *MockMetricHandler) AddIngestFinishedIndex(stats ingest.SpotifyIngestStats) error {
	return nil
}
