package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"spotify/utils"
)

type MockSpotifyAPI struct {
	albums         []string
	artists        []string
	tracks         []string
	refreshRetries int
	test           string
	loader         func(fileName string) []byte
}

func NewMockSpotifyApi(test string) MockSpotifyAPI {
	return MockSpotifyAPI{
		albums:  []string{"1", "2", "3"},
		artists: []string{"1", "2"},
		tracks:  []string{"1", "2", "3"},
		test:    test,
		loader:  utils.LoadJSON("fixtures", "recent-listens"),
	}
}

func (mockAPI *MockSpotifyAPI) Me() (MeResponse, error) {
	data := mockAPI.loader("get-me")

	meResponse := MeResponse{}
	err := json.Unmarshal(data, &meResponse)
	if err != nil {
		return MeResponse{}, err
	}

	return meResponse, nil
}

func (mockAPI *MockSpotifyAPI) RecentlyPlayedByUser() (RecentlyPlayedResponse, error) {
	data := mockAPI.loader("get-recently-played")

	recentlyPlayedResponse := RecentlyPlayedResponse{}
	err := json.Unmarshal(data, &recentlyPlayedResponse)
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}

	return recentlyPlayedResponse, nil
}

func (mockAPI *MockSpotifyAPI) TopArtistsForUser(period string) (TopArtistsResponse, error) {
	data := mockAPI.loader(fmt.Sprintf("get-top-artists-%s", period))

	topArtistsResponse := TopArtistsResponse{}
	err := json.Unmarshal(data, &topArtistsResponse)
	if err != nil {
		return TopArtistsResponse{}, err
	}

	return topArtistsResponse, nil
}

func (mockAPI *MockSpotifyAPI) TopTracksForUser(period string) (TopTracksResponse, error) {
	data := mockAPI.loader(fmt.Sprintf("get-top-tracks-%s", period))

	topTracksResponse := TopTracksResponse{}
	err := json.Unmarshal(data, &topTracksResponse)
	if err != nil {
		return TopTracksResponse{}, err
	}

	return topTracksResponse, nil
}

func (mockAPI *MockSpotifyAPI) TracksBySpotifyID(ids []string) ([]Song, error) {
	data := mockAPI.loader("get-tracks")

	tracksResponse := TracksResponse{}
	err := json.Unmarshal(data, &tracksResponse)
	if err != nil {
		return []Song{}, err
	}

	return tracksResponse.Tracks, nil
}

func (mockAPI *MockSpotifyAPI) ArtistsBySpotifyID(ids []string) ([]Artist, error) {
	data := mockAPI.loader("get-artists")

	artistResponse := ArtistsResponse{}
	err := json.Unmarshal(data, &artistResponse)
	if err != nil {
		return []Artist{}, err
	}

	return artistResponse.Artists, nil
}

func (mockAPI *MockSpotifyAPI) AlbumsBySpotifyID(ids []string) ([]Album, error) {
	data := mockAPI.loader("get-albums")

	albumResponse := AlbumResponse{}
	err := json.Unmarshal(data, &albumResponse)
	if err != nil {
		return []Album{}, err
	}

	return albumResponse.Albums, nil
}

func (mockAPI *MockSpotifyAPI) Authorize(code string) error {
	return nil
}

func (mockAPI *MockSpotifyAPI) Options() APIOptions {
	return APIOptions{
		RefreshRetries: 3,
	}
}

func (mockAPI *MockSpotifyAPI) Refresh() error {
	if mockAPI.refreshRetries == 3 {
		return nil
	}

	mockAPI.refreshRetries += 1
	return errors.New("force retry")
}
