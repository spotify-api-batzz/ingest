package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"spotify/models"
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

type spotify struct {
	Database *Database
	API      *API

	ExistingArtists map[string]models.Artist
	ExistingSongs   map[string]models.Song

	TopArtists map[string]TopArtistsResponse
	TopTracks  map[string]TopTracksResponse

	Times []string
}

func newSpotify(database *Database, api *API) spotify {
	return spotify{
		Database: database,
		API:      api,

		ExistingArtists: make(map[string]models.Artist),
		ExistingSongs:   make(map[string]models.Song),

		TopArtists: make(map[string]TopArtistsResponse),
		TopTracks:  make(map[string]TopTracksResponse),

		Times: []string{"short", "medium", "long"},
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

	database := Database{}
	database.Connect()

	spotify := newSpotify(&database, &api)

	err = api.Refresh()
	if err != nil {
		panic(err)
	}

	spotify.Artists()
	spotify.Tracks()
	spotify.Recents()
	// songs := map[string]models.Song{}
	// albums := map[string]models.Album{}

}
func (spotify *spotify) Artists() {
	artistsToQuery := []interface{}{}
	artistList := []Artist{}
	// songsToQuery := []interface{}{}

	for _, period := range spotify.Times {
		fmt.Println(fmt.Sprintf("Processing %s_term time range for artists endpoint", period))
		artists, err := spotify.API.GetTopArtists(period + "_term")
		if err != nil {
			panic(err)
		}

		for _, artist := range artists.Items {
			if _, ok := spotify.ExistingArtists[artist.ID]; ok {
				continue
			}
			artistsToQuery = append(artistsToQuery, artist.ID)
			artistList = append(artistList, artist)
		}
		spotify.TopArtists[period] = artists
	}

	artists, err := spotify.Database.FetchArtistsBySpotifyID(artistsToQuery)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	artistsToInsert := []interface{}{}

	for _, artist := range artistList {
		validArtist := false
		for _, artist := range artists {
			for _, artistID := range artistsToQuery {
				if artist.SpotifyID == artistID {
					spotify.ExistingArtists[artist.SpotifyID] = artist
					validArtist = true
				}
			}
		}
		if !validArtist {
			newArtist := models.NewArtist(artist.Name, artist.ID)
			spotify.ExistingArtists[newArtist.SpotifyID] = newArtist
			artistsToInsert = append(artistsToInsert, newArtist.ToSlice()...)
		}
	}

	err = spotify.createArtists(artistsToInsert)
	if err != nil {
		panic(err)
	}

	topArtistValues := []interface{}{}

	for _, period := range spotify.Times {
		topArtistResp := spotify.TopArtists[period]
		for order, artist := range topArtistResp.Items {
			newTopArtist := models.NewTopArtist(artist.Name, spotify.ExistingArtists[artist.ID].ID.String(), order+1, period)
			topArtistValues = append(topArtistValues, newTopArtist.ToSlice()...)
		}
	}

	err = spotify.createTopArtists(topArtistValues)
	if err != nil {
		panic(err)
	}
}

func (spotify *spotify) Recents() {
	fmt.Println("Processing recently played endpoint")
	recentlyPlayed, err := spotify.API.GetRecentlyPlayed()
	if err != nil {
		panic(err)
	}
	songsToCreate := []interface{}{}
	for _, recent := range recentlyPlayed.Items {
		if _, ok := spotify.ExistingSongs[recent.Track.ID]; !ok {
			newSong := models.NewSong(recent.Track.Name, recent.Track.ID, recent.Track.Album.ID)
			songsToCreate = append(songsToCreate, newSong.ToSlice()...)
			spotify.ExistingSongs[newSong.SpotifyID] = newSong
		}
	}

	err = spotify.createSongs(songsToCreate)
	if err != nil {
		panic(err)
	}

	recentlyToCreate := []interface{}{}
	for _, recent := range recentlyPlayed.Items {
		existingSong := spotify.ExistingSongs[recent.Track.ID]
		newRecentlyListened := models.NewRecentListen(existingSong.ID.String(), "1232", recent.PlayedAt)
		recentlyToCreate = append(recentlyToCreate, newRecentlyListened.ToSlice()...)
	}

	err = spotify.createRecentlyListened(recentlyToCreate)
	if err != nil {
		panic(err)
	}
}

func (spotify *spotify) Tracks() {
	songsToQuery := []interface{}{}
	songList := []Song{}

	for _, period := range spotify.Times {
		fmt.Println(fmt.Sprintf("Processing %s_term time range for tracks endpoint", period))
		tracks, err := spotify.API.GetTopTracks(period + "_term")
		if err != nil {
			panic(err)
		}
		for _, track := range tracks.Items {
			if _, ok := spotify.ExistingSongs[track.ID]; ok {
				continue
			}
			songsToQuery = append(songsToQuery, track.ID)
			songList = append(songList, track)
		}
		spotify.TopTracks[period] = tracks
	}

	songs, err := spotify.Database.FetchSongsBySpotifyID(songsToQuery)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	songsToInsert := []interface{}{}

	for _, song := range songList {
		validSong := false
		for _, song := range songs {
			for _, artistID := range songsToQuery {
				if song.SpotifyID == artistID {
					spotify.ExistingSongs[song.SpotifyID] = song
					validSong = true
				}
			}
		}
		if !validSong {
			newSong := models.NewSong(song.Name, song.ID, "123")
			spotify.ExistingSongs[newSong.SpotifyID] = newSong
			songsToInsert = append(songsToInsert, newSong.ToSlice()...)
		}
	}

	err = spotify.createSongs(songsToInsert)
	if err != nil {
		panic(err)
	}

	topSongValues := []interface{}{}

	for _, period := range spotify.Times {
		topSongResp := spotify.TopTracks[period]
		for order, artist := range topSongResp.Items {
			newTopSong := models.NewTopSong(artist.Name, spotify.ExistingSongs[artist.ID].ID.String(), order+1, period)
			topSongValues = append(topSongValues, newTopSong.ToSlice()...)
		}
	}

	err = spotify.createTopSongs(topSongValues)
	if err != nil {
		panic(err)
	}

}

func (spotify *spotify) createRecentlyListened(recentlyListenedValues []interface{}) error {
	if len(recentlyListenedValues) == 0 {
		return nil
	}
	err := spotify.Database.CreateRecentlyListened(recentlyListenedValues)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *spotify) createArtists(artistValues []interface{}) error {
	if len(artistValues) == 0 {
		return nil
	}
	err := spotify.Database.CreateArtist(artistValues)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *spotify) createTopArtists(topArtistValues []interface{}) error {
	if len(topArtistValues) == 0 {
		return nil
	}
	err := spotify.Database.CreateTopArtist(topArtistValues)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *spotify) createSongs(songValues []interface{}) error {
	if len(songValues) == 0 {
		return nil
	}
	err := spotify.Database.CreateSong(songValues)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *spotify) createTopSongs(topSongValues []interface{}) error {
	if len(topSongValues) == 0 {
		return nil
	}
	err := spotify.Database.CreateTopSong(topSongValues)
	if err != nil {
		return err
	}
	return nil
}

func genNiceTime() string {
	timeFormat := "Mon 2 Jan 2006 15-04-05"
	time := time.Now()
	return time.Format(timeFormat)
}

func fetchOrCreateArtist(database Database, artist Artist) models.Artist {
	dbArtist, err := database.FetchArtistBySpotifyID(artist.ID)
	if err == sql.ErrNoRows {
		dbArtist = models.NewArtist(artist.Name, artist.ID)
		database.CreateArtist(dbArtist.ToSlice())
	}
	return dbArtist
}

func BasicAuth(clientID string, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
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
