package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AccessData struct {
	Token   string
	Refresh string
}

type ClientCreds struct {
	ID     string
	Secret string
}

type API struct {
	BaseURL string
	Creds   ClientCreds
	Client  http.Client
	Tokens  AccessData
}

type AuthResponse struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

type RefreshResponse struct {
	Access string `json:"access_token"`
}

type RecentlyPlayedResponse struct {
	Items []struct {
		Track struct {
			Album struct {
				AlbumType string `json:"album_type"`
				Artists   []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"artists"`
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					Height int    `json:"height"`
					URL    string `json:"url"`
					Width  int    `json:"width"`
				} `json:"images"`
				Name                 string `json:"name"`
				ReleaseDate          string `json:"release_date"`
				ReleaseDatePrecision string `json:"release_date_precision"`
				TotalTracks          int    `json:"total_tracks"`
				Type                 string `json:"type"`
				URI                  string `json:"uri"`
			} `json:"album"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       int      `json:"disc_number"`
			DurationMs       int      `json:"duration_ms"`
			Explicit         bool     `json:"explicit"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string `json:"href"`
			ID          string `json:"id"`
			IsLocal     bool   `json:"is_local"`
			Name        string `json:"name"`
			Popularity  int    `json:"popularity"`
			PreviewURL  string `json:"preview_url"`
			TrackNumber int    `json:"track_number"`
			Type        string `json:"type"`
			URI         string `json:"uri"`
		} `json:"track"`
		PlayedAt time.Time `json:"played_at"`
		Context  struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"context"`
	} `json:"items"`
	Next    string `json:"next"`
	Cursors struct {
		After  string `json:"after"`
		Before string `json:"before"`
	} `json:"cursors"`
	Limit int    `json:"limit"`
	Href  string `json:"href"`
}

type TopArtistsResponse struct {
	Items []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Followers struct {
			Href  interface{} `json:"href"`
			Total int         `json:"total"`
		} `json:"followers"`
		Genres []string `json:"genres"`
		Href   string   `json:"href"`
		ID     string   `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name       string `json:"name"`
		Popularity int    `json:"popularity"`
		Type       string `json:"type"`
		URI        string `json:"uri"`
	} `json:"items"`
	Total    int         `json:"total"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Href     string      `json:"href"`
	Previous interface{} `json:"previous"`
	Next     string      `json:"next"`
}

type TopTracksResponse struct {
	Items []struct {
		Album struct {
			AlbumType string `json:"album_type"`
			Artists   []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			ExternalUrls     struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				Height int    `json:"height"`
				URL    string `json:"url"`
				Width  int    `json:"width"`
			} `json:"images"`
			Name                 string `json:"name"`
			ReleaseDate          string `json:"release_date"`
			ReleaseDatePrecision string `json:"release_date_precision"`
			TotalTracks          int    `json:"total_tracks"`
			Type                 string `json:"type"`
			URI                  string `json:"uri"`
		} `json:"album"`
		Artists []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		DiscNumber       int      `json:"disc_number"`
		DurationMs       int      `json:"duration_ms"`
		Explicit         bool     `json:"explicit"`
		ExternalIds      struct {
			Isrc string `json:"isrc"`
		} `json:"external_ids"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href        string `json:"href"`
		ID          string `json:"id"`
		IsLocal     bool   `json:"is_local"`
		Name        string `json:"name"`
		Popularity  int    `json:"popularity"`
		PreviewURL  string `json:"preview_url"`
		TrackNumber int    `json:"track_number"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
	} `json:"items"`
	Total    int         `json:"total"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Href     string      `json:"href"`
	Previous interface{} `json:"previous"`
	Next     string      `json:"next"`
}

func NewAPI(baseURL string, secret string, clientID string, refresh string) API {
	return API{
		BaseURL: baseURL,
		Client:  http.Client{},
		Creds: ClientCreds{
			Secret: secret,
			ID:     clientID,
		},
		Tokens: AccessData{
			Refresh: refresh,
		},
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// scopes := []string{"user-read-playback-state", "user-read-currently-playing", "user-read-recently-played", "user-top-read"}
	api := NewAPI("https://accounts.spotify.com/", os.Getenv("secret"), os.Getenv("clientID"), os.Getenv("refresh"))
	// url := fmt.Sprintf("https://accounts.spotify.com/authorize?response_type=code&redirect_uri=http://localhost&client_id=%s", api.Creds.ID)

	// authUrl := fmt.Sprintf("%s&scope=%s", url, strings.Join(scopes, "%20"))
	// fmt.Println(authUrl)

	// err = api.Authorize("AQADBR9Yq7P6Mhb5V9cnKYKH3xWETG-Dpi8tkC371CsUK7C_3ZR--gejX85u2jhy5g9totBT6mw9SpBxjrDQQWVKRZ_x9npyExsWtprKb4Nv55BEY6jmkT6SRcob6ryJEBM2g_9iXaALUzGtFMAKwo8xRUJ4CmjYxXGnux93N5HKqL_CcKRaTAqg-KkTTpPWhRRnEPqKrZwKIHtm4RuBUlhe5qJKy60dGYNd6WFJ8CT5-kEPSmIGbVZASX-Fe6yw4qdi7D9QC08fpabpTifd97O0EYZ0")
	// if err != nil {
	// 	fmt.Println(err)
	// 	panic(err)
	// }
	fmt.Println("rfesrer")
	fmt.Println(os.Getenv("refresh"))
	err = api.Refresh()
	if err != nil {
		panic(err)
	}

	fmt.Println("rfesrer")
	fmt.Println(api.Tokens.Refresh)
	fmt.Println("rfesrer")
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	times := []string{"short", "medium", "long"}

	for _, period := range times {
		fmt.Println(fmt.Sprintf("Processing %s_term time range for artists endpoint", period))
		artists, err := api.GetTopArtists(period + "_term")
		if err != nil {
			panic(err)
		}
		fileName := fmt.Sprintf("%s.json", genNiceTime())
		filePath := path.Join(wd, "json", "artists", period, fileName)

		file, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}

		marshaledBody, err := json.Marshal(artists)
		if err != nil {
			panic(err)
		}

		file.Write(marshaledBody)
		file.Close()
	}

	for _, period := range times {
		fmt.Println(fmt.Sprintf("Processing %s_term time range for tracks endpoint", period))
		artists, err := api.GetTopTracks(period + "_term")
		if err != nil {
			panic(err)
		}
		fileName := fmt.Sprintf("%s.json", genNiceTime())
		filePath := path.Join(wd, "json", "tracks", period, fileName)

		file, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}

		marshaledBody, err := json.Marshal(artists)
		if err != nil {
			panic(err)
		}

		file.Write(marshaledBody)
		file.Close()
	}

	fmt.Println("Processing recently played endpoint")
	recentlyPlayed, err := api.GetRecentlyPlayed()
	if err != nil {
		panic(err)
	}

	fileName := fmt.Sprintf("%s.json", genNiceTime())
	filePath := path.Join(wd, "json", "recent", fileName)

	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	marshaledBody, err := json.Marshal(recentlyPlayed)
	if err != nil {
		panic(err)
	}

	file.Write(marshaledBody)
	file.Close()
}

func genNiceTime() string {
	timeFormat := "Mon 2 Jan 2006 15-04-05"
	time := time.Now()
	return time.Format(timeFormat)
}

func BasicAuth(clientID string, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}

func (api *API) GetRecentlyPlayed() (RecentlyPlayedResponse, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/recently-played", nil)
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
		fmt.Println(bytes)
		return RecentlyPlayedResponse{}, errors.New("Status code not 200")
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
		fmt.Println(bytes)
		return TopArtistsResponse{}, errors.New("Status code not 200")
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
		fmt.Println(bytes)
		return TopTracksResponse{}, errors.New("Status code not 200")
	}

	topPlayedResp := TopTracksResponse{}
	err = json.Unmarshal(bytes, &topPlayedResp)
	if err != nil {
		return TopTracksResponse{}, err
	}

	return topPlayedResp, nil
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
		fmt.Println(string(bytes))
		return errors.New("Status code not 200")
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
		fmt.Println(bytes)
		return errors.New("Status code not 200")
	}

	refreshResp := RefreshResponse{}
	err = json.Unmarshal(bytes, &refreshResp)
	if err != nil {
		return err
	}
	api.Tokens.Token = refreshResp.Access

	return nil
}

// https://accounts.spotify.com/authorize?client_id=5fe01282e44241328a84e7c5cc169165&response_type=code&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&scope=user-read-private%20user-read-email&state=34fFs29kd09

// {
//   "access_token": "BQBkz49PV7FjapOSXrWgrjglA9ARNial4Sim9dePiTqFqtPeKOWjkULi8UJdJxx2VXBsiRY5joFO29UB3DmU09u9-hApivJ73icQS0n2wVGMmqjTWFg0el_rF3KtLj0dO3nTcbZPHiLAqq8hSmP0",
//   "token_type": "Bearer",
//   "expires_in": 3600,
//   "refresh_token": "AQAcB3qPL3cld0iLTiiYOFBvoGXVFGX6tijiaWH9pCfvqzVYKezvCsCbjn4kWydsbFp-Rq-JVw2yGRS8Pd_ApxTwn0SKi7pwzdhPrudsE_cKQnu3KrF_gdfIYJPjxWlSTnM",
//   "scope": ""
// }
