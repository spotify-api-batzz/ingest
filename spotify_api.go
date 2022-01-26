package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

type API interface {
	Authorize(code string) error
	Options() *APIOptions
	Refresh() error
}

type SpotifyAPI interface {
	API
	Me() (MeResponse, error)
	RecentlyPlayedByUser() (RecentlyPlayedResponse, error)
	TopArtistsForUser(period string) (TopArtistsResponse, error)
	TopTracksForUser(period string) (TopTracksResponse, error)
	TracksBySpotifyID(ids []string) ([]Song, error)
	ArtistsBySpotifyID(ids []string) ([]Artist, error)
	AlbumsBySpotifyID(ids []string) ([]Album, error)
}

type APIOptions struct {
	RefreshRetries int
}

func NewAPIOptions(refreshRetries int) *APIOptions {
	return &APIOptions{
		RefreshRetries: refreshRetries,
	}
}

type spotifyAPI struct {
	BaseURL string
	Creds   ClientInfo
	Tokens  AccessCreds
	Client  http.Client
	opts    *APIOptions
}

func NewSpotifyAPI(baseURL string, secret string, clientID string, refresh string, options *APIOptions) SpotifyAPI {
	return &spotifyAPI{
		BaseURL: baseURL,
		Client:  http.Client{},
		Creds: ClientInfo{
			Secret: secret,
			ID:     clientID,
		},
		Tokens: AccessCreds{
			Refresh: refresh,
		},
		opts: options,
	}
}

func (api *spotifyAPI) Options() *APIOptions {
	return api.opts
}

func (api *spotifyAPI) Me() (MeResponse, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return MeResponse{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Tokens.Token))

	body, err := api.Client.Do(req)
	if err != nil {
		return MeResponse{}, err
	}
	defer body.Body.Close()

	bytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return MeResponse{}, err
	}

	if body.StatusCode != 200 {
		fmt.Println(string(bytes))
		return MeResponse{}, errors.New("status code not 200")
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
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/me/player/recently-played?%s", data.Encode()), nil)
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Tokens.Token))

	body, err := api.Client.Do(req)
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}
	defer body.Body.Close()

	bytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}

	if body.StatusCode != 200 {
		return RecentlyPlayedResponse{}, errors.New("status code not 200")
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
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/me/top/artists?%s", data.Encode()), nil)

	if err != nil {
		return TopArtistsResponse{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Tokens.Token))

	body, err := api.Client.Do(req)
	if err != nil {
		return TopArtistsResponse{}, err
	}
	defer body.Body.Close()

	bytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return TopArtistsResponse{}, err
	}

	if body.StatusCode != 200 {
		return TopArtistsResponse{}, errors.New("status code not 200")
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
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/me/top/tracks?%s", data.Encode()), nil)
	if err != nil {
		return TopTracksResponse{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Tokens.Token))

	body, err := api.Client.Do(req)
	if err != nil {
		return TopTracksResponse{}, err
	}
	defer body.Body.Close()

	bytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return TopTracksResponse{}, err
	}

	if body.StatusCode != 200 {
		return TopTracksResponse{}, errors.New("status code not 200")
	}

	topPlayedResp := TopTracksResponse{}
	err = json.Unmarshal(bytes, &topPlayedResp)
	if err != nil {
		return TopTracksResponse{}, err
	}

	return topPlayedResp, nil
}

func (api *spotifyAPI) ArtistsBySpotifyID(ids []string) ([]Artist, error) {
	fmt.Println(ids)
	chunkedIDs := utils.ChunkSlice(ids, 50)
	artistList := []Artist{}

	for _, chunk := range chunkedIDs {
		data := url.Values{}
		data.Set("ids", strings.Join(chunk, ","))
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/artists?%s", data.Encode()), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Tokens.Token))

		body, err := api.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer body.Body.Close()

		bytes, err := ioutil.ReadAll(body.Body)
		if err != nil {
			return nil, err
		}

		if body.StatusCode != 200 {
			return nil, errors.New("status code not 200")
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
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/tracks?%s", data.Encode()), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Tokens.Token))

		body, err := api.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer body.Body.Close()

		bytes, err := ioutil.ReadAll(body.Body)
		if err != nil {
			return nil, err
		}

		if body.StatusCode != 200 {
			return nil, errors.New("status code not 200")
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
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/albums?%s", data.Encode()), nil)
	if err != nil {
		return AlbumResponse{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Tokens.Token))

	body, err := api.Client.Do(req)
	if err != nil {
		return AlbumResponse{}, err
	}
	defer body.Body.Close()

	bytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return AlbumResponse{}, err
	}

	if body.StatusCode != 200 {
		fmt.Println(string(bytes))
		return AlbumResponse{}, errors.New("status code not 200")
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

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", BasicAuth(api.Creds.ID, api.Creds.Secret))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return err
	}

	body, err := api.Client.Do(req)
	if err != nil {
		return err
	}

	defer body.Body.Close()
	bytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return err
	}

	if body.StatusCode != 200 {
		return errors.New("status code not 200")
	}

	authResp := AuthResponse{}
	err = json.Unmarshal(bytes, &authResp)
	if err != nil {
		return err
	}

	api.Tokens = AccessCreds{
		Token:   authResp.Access,
		Refresh: authResp.Refresh,
	}

	return nil
}

func (api *spotifyAPI) Refresh() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", api.Tokens.Refresh)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", BasicAuth(api.Creds.ID, api.Creds.Secret))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}

	body, err := api.Client.Do(req)
	if err != nil {
		return err
	}
	defer body.Body.Close()

	bytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return err
	}

	if body.StatusCode != 200 {
		return errors.New("status code not 200")
	}

	refreshResp := RefreshResponse{}
	err = json.Unmarshal(bytes, &refreshResp)
	if err != nil {
		return err
	}

	fmt.Println(refreshResp)
	api.Tokens.Token = refreshResp.Access

	return nil
}

func BasicAuth(clientID string, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}
