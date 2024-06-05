package mocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"spotify/types"
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

func (mockAPI *MockSpotifyAPI) Me() (types.MeResponse, error) {
	data := mockAPI.loader("get-me")

	meResponse := types.MeResponse{}
	err := json.Unmarshal(data, &meResponse)
	if err != nil {
		return types.MeResponse{}, err
	}

	return meResponse, nil
}

func (mockAPI *MockSpotifyAPI) RecentlyPlayedByUser() (types.RecentlyPlayedResponse, error) {
	data := mockAPI.loader("get-recently-played")

	recentlyPlayedResponse := types.RecentlyPlayedResponse{}
	err := json.Unmarshal(data, &recentlyPlayedResponse)
	if err != nil {
		return types.RecentlyPlayedResponse{}, err
	}

	return recentlyPlayedResponse, nil
}

func (mockAPI *MockSpotifyAPI) TopArtistsForUser(period string) (types.TopArtistsResponse, error) {
	data := mockAPI.loader(fmt.Sprintf("get-top-artists-%s", period))

	topArtistsResponse := types.TopArtistsResponse{}
	err := json.Unmarshal(data, &topArtistsResponse)
	if err != nil {
		return types.TopArtistsResponse{}, err
	}

	return topArtistsResponse, nil
}

func (mockAPI *MockSpotifyAPI) TopTracksForUser(period string) (types.TopTracksResponse, error) {
	data := mockAPI.loader(fmt.Sprintf("get-top-tracks-%s", period))

	topTracksResponse := types.TopTracksResponse{}
	err := json.Unmarshal(data, &topTracksResponse)
	if err != nil {
		return types.TopTracksResponse{}, err
	}

	return topTracksResponse, nil
}

func (mockAPI *MockSpotifyAPI) TracksBySpotifyID(ids []string) ([]types.Song, error) {
	data := mockAPI.loader("get-tracks")

	tracksResponse := types.TracksResponse{}
	err := json.Unmarshal(data, &tracksResponse)
	if err != nil {
		return []types.Song{}, err
	}

	return tracksResponse.Tracks, nil
}

func (mockAPI *MockSpotifyAPI) ArtistsBySpotifyID(ids []string) ([]types.Artist, error) {
	data := mockAPI.loader("get-artists")

	artistResponse := types.ArtistsResponse{}
	err := json.Unmarshal(data, &artistResponse)
	if err != nil {
		return []types.Artist{}, err
	}

	return artistResponse.Artists, nil
}

func (mockAPI *MockSpotifyAPI) AlbumsBySpotifyID(ids []string) ([]types.Album, error) {
	data := mockAPI.loader("get-albums")

	albumResponse := types.AlbumResponse{}
	err := json.Unmarshal(data, &albumResponse)
	if err != nil {
		return []types.Album{}, err
	}

	return albumResponse.Albums, nil
}

func (mockAPI *MockSpotifyAPI) Authorize(code string) error {
	return nil
}

func (mockAPI *MockSpotifyAPI) Options() types.APIOptions {
	return types.APIOptions{
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
