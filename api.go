package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type API struct {
	BaseURL string
	Creds   ClientCreds
	Client  http.Client
	Tokens  AccessData
}

func (api *API) GetRecentlyPlayed() (RecentlyPlayedResponse, error) {
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

func (api *API) GetTopArtists(period string) (TopArtistsResponse, error) {
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

func (api *API) GetTopTracks(period string) (TopTracksResponse, error) {
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

func (api *API) GetArtists(ids []string) ([]Artists, error) {
	chunkedIDs := chunkSlice(ids, 50)
	artistList := []Artists{}

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

func (api *API) GetTracks(ids []string) ([]Song, error) {
	chunkedIDs := chunkSlice(ids, 50)
	trackList := []Song{}

	for _, chunk := range chunkedIDs {
		data := url.Values{}
		data.Set("ids", strings.Join(chunk, ","))
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/tracks ?%s", data.Encode()), nil)
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

func (api *API) Authorize(code string) error {
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

	api.Tokens = AccessData{
		Token:   authResp.Access,
		Refresh: authResp.Refresh,
	}

	return nil
}

func (api *API) Refresh() error {
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
	api.Tokens.Token = refreshResp.Access

	return nil
}
