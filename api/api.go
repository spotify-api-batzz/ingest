package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"spotify/utils"
	"strings"
)

type AccessCreds struct {
	Token   string
	Refresh string
}

type ClientInfo struct {
	ID     string
	Secret string
}

type AuthResponse struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

type RefreshResponse struct {
	Access string `json:"access_token"`
}

func NewAPIOptions(refreshRetries int) APIOptions {
	return APIOptions{
		RefreshRetries: refreshRetries,
	}
}

type APIOptions struct {
	RefreshRetries int
}

func (a *APIOptions) Retries() int {
	return a.RefreshRetries
}

type spotifyAPI struct {
	BaseURL string
	Auth    SpotifyAPIAuth
	Client  http.Client
	Metrics MetricHandler
	opts    APIOptions
}

type SpotifyAPIAuth struct {
	Secret       string
	ClientID     string
	RefreshToken string
	AccessToken  string
}

func newBadRespError(code int, body string) *BadRespError {
	return &BadRespError{
		Code: code,
		Body: body,
	}
}

type BadRespError struct {
	Code int
	Body string
}

func (b *BadRespError) Error() string {
	return fmt.Sprintf("got a bad response from the api status code %d, body %s", b.Code, b.Body)
}

type MetricHandler interface {
	AddApiRequestIndex(method string, url string, reqBody string, timeTakenMs int64, bodySize int) error
}

func NewSpotifyAPI(baseURL string, Metrics MetricHandler, auth SpotifyAPIAuth, options APIOptions) spotifyAPI {
	return spotifyAPI{
		BaseURL: baseURL,
		Client:  http.Client{},
		Metrics: Metrics,
		Auth:    auth,
		opts:    options,
	}
}

func (api *spotifyAPI) Options() APIOptions {
	return api.opts
}

func safeCloneReader(body io.Reader) (string, io.Reader, error) {
	if body == nil {
		return "", nil, nil
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", nil, err
	}

	return string(bodyBytes), bytes.NewReader(bodyBytes), nil
}

func (api *spotifyAPI) Request(method string, url string, body io.Reader) ([]byte, error) {
	bodyString, bodyReader, err := safeCloneReader(body)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Auth.AccessToken))
	if method == "POST" {
		req.Header.Set("Authorization", BasicAuth(api.Auth.ClientID, api.Auth.Secret))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	reqStart := time.Now()
	resp, err := api.Client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	reqEnd := time.Now()
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	sanitizedBody := utils.ScrubSensitiveData(bodyString, []string{"refresh_token"})
	sanitizedUrl := utils.ScrubSensitiveData(url, []string{"refresh_token"})
	err = api.Metrics.AddApiRequestIndex(method, sanitizedUrl, sanitizedBody, reqStart.Sub(reqEnd).Milliseconds(), len(bytes))
	if resp.StatusCode != 200 {
		return []byte{}, newBadRespError(resp.StatusCode, string(bytes))
	}

	return bytes, nil
}

func (api *spotifyAPI) Me() (MeResponse, error) {
	bytes, err := api.Request("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return MeResponse{}, err
	}

	meResp := MeResponse{}
	err = json.Unmarshal(bytes, &meResp)
	if err != nil {
		return MeResponse{}, err
	}

	return meResp, nil
}

func (api *spotifyAPI) RecentlyPlayedByUser() (RecentlyPlayedResponse, error) {
	data := url.Values{}
	data.Set("limit", "50")
	url := fmt.Sprintf("https://api.spotify.com/v1/me/player/recently-played?%s", data.Encode())
	bytes, err := api.Request("GET", url, nil)
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}

	recentlyPlayedResp := RecentlyPlayedResponse{}
	err = json.Unmarshal(bytes, &recentlyPlayedResp)
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}

	return recentlyPlayedResp, nil
}

func (api *spotifyAPI) TopArtistsForUser(period string) (TopArtistsResponse, error) {
	data := url.Values{}
	data.Set("time_range", period)
	data.Set("limit", "50")

	url := fmt.Sprintf("https://api.spotify.com/v1/me/top/artists?%s", data.Encode())
	bytes, err := api.Request("GET", url, nil)
	if err != nil {
		return TopArtistsResponse{}, err
	}

	topPlayedResp := TopArtistsResponse{}
	err = json.Unmarshal(bytes, &topPlayedResp)
	if err != nil {
		return TopArtistsResponse{}, err
	}

	return topPlayedResp, nil
}

func (api *spotifyAPI) TopTracksForUser(period string) (TopTracksResponse, error) {
	data := url.Values{}
	data.Set("time_range", period)
	data.Set("limit", "50")

	url := fmt.Sprintf("https://api.spotify.com/v1/me/top/tracks?%s", data.Encode())
	bytes, err := api.Request("GET", url, nil)
	if err != nil {
		return TopTracksResponse{}, err
	}

	topPlayedResp := TopTracksResponse{}
	err = json.Unmarshal(bytes, &topPlayedResp)
	if err != nil {
		return TopTracksResponse{}, err
	}

	return topPlayedResp, nil
}

func (api *spotifyAPI) ArtistsBySpotifyID(ids []string) ([]Artist, error) {
	chunkedIDs := utils.ChunkSlice(ids, 50)
	artistList := []Artist{}

	for _, chunk := range chunkedIDs {
		data := url.Values{}
		data.Set("ids", strings.Join(chunk, ","))
		url := fmt.Sprintf("https://api.spotify.com/v1/artists?%s", data.Encode())
		bytes, err := api.Request("GET", url, nil)
		if err != nil {
			return nil, err
		}

		artistsResp := ArtistsResponse{}
		err = json.Unmarshal(bytes, &artistsResp)
		if err != nil {
			return nil, err
		}

		artistList = append(artistList, artistsResp.Artists...)
	}

	return artistList, nil
}

func (api *spotifyAPI) TracksBySpotifyID(ids []string) ([]Song, error) {
	chunkedIDs := utils.ChunkSlice(ids, 50)
	trackList := []Song{}

	for _, chunk := range chunkedIDs {
		data := url.Values{}
		data.Set("ids", strings.Join(chunk, ","))

		url := fmt.Sprintf("https://api.spotify.com/v1/tracks?%s", data.Encode())
		bytes, err := api.Request("GET", url, nil)
		if err != nil {
			return nil, err
		}

		tracksResp := TracksResponse{}
		err = json.Unmarshal(bytes, &tracksResp)
		if err != nil {
			return nil, err
		}

		trackList = append(trackList, tracksResp.Tracks...)
	}

	return trackList, nil
}

func (api *spotifyAPI) AlbumsBySpotifyID(ids []string) ([]Album, error) {
	chunkedIDs := utils.ChunkSlice(ids, 20)
	albumList := []Album{}

	// #TODO: Fire these in parallel
	for _, chunk := range chunkedIDs {
		albums, err := api.albumsBySpotifyID(chunk)
		if err != nil {
			return nil, err
		}
		albumList = append(albumList, albums.Albums...)
	}
	return albumList, nil
}

func (api *spotifyAPI) albumsBySpotifyID(ids []string) (AlbumResponse, error) {
	data := url.Values{}
	data.Set("ids", strings.Join(ids, ","))
	bytes, err := api.Request("GET", fmt.Sprintf("https://api.spotify.com/v1/albums?%s", data.Encode()), nil)
	if err != nil {
		return AlbumResponse{}, err
	}

	albumResp := AlbumResponse{}
	err = json.Unmarshal(bytes, &albumResp)
	if err != nil {
		return AlbumResponse{}, err
	}

	return albumResp, nil
}

func (api *spotifyAPI) Authorize(code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost")

	bytes, err := api.Request("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	authResp := AuthResponse{}
	err = json.Unmarshal(bytes, &authResp)
	if err != nil {
		return err
	}

	api.Auth.AccessToken = authResp.Access
	api.Auth.RefreshToken = authResp.Refresh

	return nil
}

func (api *spotifyAPI) Refresh() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", api.Auth.RefreshToken)

	bytes, err := api.Request("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	refreshResp := RefreshResponse{}
	err = json.Unmarshal(bytes, &refreshResp)
	if err != nil {
		return err
	}

	api.Auth.AccessToken = refreshResp.Access

	return nil
}

func BasicAuth(clientID string, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}
