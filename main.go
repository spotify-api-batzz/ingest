package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"spotify/models"

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
	ExistingAlbums  map[string]models.Album

	TopArtists map[string]TopArtistsResponse
	TopTracks  map[string]TopTracksResponse

	thumbnailsToInsert []interface{}

	Times []string
}

func newSpotify(database *Database, api *API) spotify {
	return spotify{
		Database: database,
		API:      api,

		ExistingArtists: make(map[string]models.Artist),
		ExistingSongs:   make(map[string]models.Song),
		ExistingAlbums:  make(map[string]models.Album),

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

	spotify.Tracks()
	// spotify.Artists()
	// spotify.Recents()

	// if len(spotify.thumbnailsToInsert) != 0 {
	// 	err = spotify.Database.CreateThumbnail(spotify.thumbnailsToInsert)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

}

func (spotify *spotify) Artists() {
	artistsToQuery := []interface{}{}
	artistList := []Artist{}

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
			newThumbnail := models.NewThumbnail("artist", newArtist.ID.String(), artist.Images[0].URL)
			spotify.thumbnailsToInsert = append(spotify.thumbnailsToInsert, newThumbnail.ToSlice()...)
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
			newTopArtist := models.NewTopArtist(artist.Name, spotify.ExistingArtists[artist.ID].ID.String(), order+1, period, os.Getenv("USER_ID"))
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

	artistsToFetch := []string{}

	for _, song := range recentlyPlayed.Items {
		if _, ok := spotify.ExistingArtists[song.Track.Artists[0].ID]; ok {
			continue
		}
		artistsToFetch = append(artistsToFetch, song.Track.Artists[0].ID)
	}

	artists, err := spotify.API.GetArtists(artistsToFetch)
	if err != nil {
		panic(err)
	}

	artistsToInsert := []interface{}{}
	for _, artist := range artists {
		newArtist := models.NewArtist(artist.Name, artist.ID)
		spotify.ExistingArtists[newArtist.SpotifyID] = newArtist
		artistsToInsert = append(artistsToInsert, newArtist.ToSlice()...)
	}

	err = spotify.createArtists(artistsToInsert)
	if err != nil {
		panic(err)
	}

	songsToCreate := []interface{}{}
	for _, recent := range recentlyPlayed.Items {
		if _, ok := spotify.ExistingSongs[recent.Track.ID]; !ok {
			newSong := models.NewSong(recent.Track.Name, recent.Track.ID, recent.Track.Album.ID, spotify.ExistingArtists[recent.Track.Artists[0].ID].ID.String())
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
		newRecentlyListened := models.NewRecentListen(existingSong.ID.String(), os.Getenv("USER_ID"), recent.PlayedAt)
		recentlyToCreate = append(recentlyToCreate, newRecentlyListened.ToSlice()...)
	}

	err = spotify.createRecentlyListened(recentlyToCreate)
	if err != nil {
		panic(err)
	}
}

// func (spotify *spotify) fetchSongs(songIDs []interface{}) ([]models.Song, error) {
// 	songs, err := spotify.Database.FetchSongsBySpotifyID(songIDs)
// 	if err != nil && err != sql.ErrNoRows {
// 		return nil, err
// 	}

// 	songsToFetch := []string{}
// Outer:
// 	for _, songID := range songIDs {
// 		for _, song := range songs {
// 			if song.SpotifyID == songID {
// 				continue Outer
// 			}
// 		}
// 		songsToFetch = append(songsToFetch, songID.(string))
// 	}

// 	apiSongs, err := spotify.API.GetTracks(songsToFetch)
// 	if err != nil {
// 		return nil, err
// 	}

// 	songsToInsert := []interface{}{}
// 	for _, song := range apiSongs {
//     newSong := models.NewSong(song.Name, song.ID, song.Album.ID, song.Artists[] )
// 	}

// }

func (spotify *spotify) aaa() (map[string]TopTracksResponse, error) {
	topTrackResp := make(map[string]TopTracksResponse)

	for _, period := range spotify.Times {
		fmt.Printf("Processing %s_term time range for tracks endpoint\n", period)
		tracks, err := spotify.API.GetTopTracks(period + "_term")
		if err != nil {
			return nil, err
		}

		topTrackResp[period] = tracks
	}
	return topTrackResp, nil
}

func (spotify *spotify) Tracks() {
	songsToQuery := []interface{}{}
	albumsToQuery := []interface{}{}
	songList := []Song{}
	artistsToFetch := []string{}

	for _, period := range spotify.Times {
		fmt.Printf("Processing %s_term time range for tracks endpoint\n", period)
		tracks, err := spotify.API.GetTopTracks(period + "_term")
		if err != nil {
			panic(err)
		}

		for _, track := range tracks.Items {
			if _, ok := spotify.ExistingArtists[track.Artists[0].ID]; !ok {
				artistsToFetch = append(artistsToFetch, track.Artists[0].ID)
			}
			if _, ok := spotify.ExistingSongs[track.ID]; !ok {
				songsToQuery = append(songsToQuery, track.ID)
			}
			albumsToQuery = append(albumsToQuery, track.Album.ID)
			songList = append(songList, track)
		}
		spotify.TopTracks[period] = tracks
	}

	albums, err := spotify.Database.FetchAlbumsBySpotifyID(albumsToQuery)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	songsToInsert := []interface{}{}
	albumsToInsert := []interface{}{}

	artists, err := spotify.API.GetArtists(artistsToFetch)
	if err != nil {
		panic(err)
	}

	artistsToInsert := []interface{}{}
	for _, artist := range artists {
		newArtist := models.NewArtist(artist.Name, artist.ID)
		spotify.ExistingArtists[newArtist.SpotifyID] = newArtist
		artistsToInsert = append(artistsToInsert, newArtist.ToSlice()...)
	}

	for _, song := range songList {
		validAlbum := false
		if _, ok := spotify.ExistingAlbums[song.Album.ID]; ok {
			continue
		}
		for _, album := range albums {
			if song.Album.ID == album.SpotifyID {
				spotify.ExistingAlbums[album.SpotifyID] = album
				validAlbum = true
			}
		}
		if !validAlbum {
			newAlbum := models.NewAlbum(song.Album.Name, spotify.ExistingArtists[song.Artists[0].ID].ID.String(), song.Album.ID)
			spotify.ExistingAlbums[newAlbum.SpotifyID] = newAlbum
			newThumbnail := models.NewThumbnail("album", newAlbum.ID.String(), song.Album.Images[0].URL)
			spotify.thumbnailsToInsert = append(spotify.thumbnailsToInsert, newThumbnail.ToSlice()...)
			albumsToInsert = append(albumsToInsert, newAlbum.ToSlice()...)
		}
	}

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
			newSong := models.NewSong(song.Name, song.ID, spotify.ExistingAlbums[song.Album.ID].ID.String(), spotify.ExistingArtists[song.Artists[0].ID].ID.String())

			spotify.ExistingSongs[newSong.SpotifyID] = newSong
			songsToInsert = append(songsToInsert, newSong.ToSlice()...)
		}
	}

	err = spotify.createSongs(songsToInsert)
	if err != nil {
		panic(err)
	}

	err = spotify.createAlbums(albumsToInsert)
	if err != nil {
		panic(err)
	}

	topSongValues := []interface{}{}

	for _, period := range spotify.Times {
		topSongResp := spotify.TopTracks[period]
		for order, song := range topSongResp.Items {
			songID := spotify.ExistingSongs[song.ID].ID.String()
			if songID == "00000000-0000-0000-0000-000000000000" {
				songID = song.ID
			}
			newTopSong := models.NewTopSong(os.Getenv("USER_ID"), songID, order+1, period)
			topSongValues = append(topSongValues, newTopSong.ToSlice()...)
		}
	}

	err = spotify.createTopSongs(topSongValues)
	if err != nil {
		panic(err)
	}

}

func getSongByName()

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

func (spotify *spotify) createAlbums(albumValues []interface{}) error {
	if len(albumValues) == 0 {
		return nil
	}
	err := spotify.Database.CreateAlbum(albumValues)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *spotify) createThumbnails(thumbnailValues []interface{}) error {
	if len(thumbnailValues) == 0 {
		return nil
	}
	err := spotify.Database.CreateThumbnail(thumbnailValues)
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

func BasicAuth(clientID string, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}
