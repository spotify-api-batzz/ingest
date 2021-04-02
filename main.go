package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"spotify/logger"
	"spotify/models"
	"time"

	"github.com/google/uuid"
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

type spotify struct {
	Database *Database
	UserID   string
	API      *API
	Times    []string
}

func newSpotify(database *Database, api *API, userID string) spotify {
	return spotify{
		Database: database,
		API:      api,

		Times:  []string{"short", "medium", "long"},
		UserID: userID,
	}
}

func main() {
	if len(os.Args) == 1 {
		panic("You must specify a userID for your first command line arg!")
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	api := NewAPI("https://accounts.spotify.com/", os.Getenv("secret"), os.Getenv("clientID"), os.Getenv("refresh"))

	err = api.Refresh()
	if err != nil {
		panic(err)
	}

	database := Database{}
	err = database.Connect()
	if err != nil {
		logger.Log("Failed to connect to database", logger.Error)
		panic(err)
	}

	me, err := api.Me()
	if err != nil {
		logger.Log("Failed to fetch Me endpoint", logger.Error)
		panic(err)
	}

	user, err := HandleBaseUsers(database, os.Args[1], me)
	if err != nil {
		logger.Log("Failed when handling base user routine", logger.Error)
		panic(err)
	}

	spotify := newSpotify(&database, &api, user.ID)

	err = spotify.DataInserts()
	if err != nil {
		panic(err)
	}
}

func HandleBaseUsers(db Database, usernameToReturn string, user MeResponse) (models.User, error) {
	logger.Log("Handling base user data", logger.Notice)
	baseUsers := []interface{}{"bungusbuster", "anneteresa-gb"}
	users, err := db.FetchUsersByNames(baseUsers)
	if err != nil {
		panic(err)
	}

	userValues := []interface{}{}
	userToReturn := models.User{}
Outer:
	for _, username := range baseUsers {
		for _, user := range users {
			if user.SpotifyID == username {
				continue Outer
			}
			if user.Username == usernameToReturn {
				userToReturn = user
			}
		}
		newUser := models.NewUser(username.(string), "123", username.(string))
		userValues = append(userValues, newUser.ToSlice()...)
	}

	if len(userValues) == 0 {
		logger.Log("syke bitch the database already contains all necessary users", logger.Notice)
		return userToReturn, nil
	}

	err = db.CreateUser(userValues)
	if err != nil {
		return models.User{}, err
	}

	return userToReturn, nil
}

func (spotify *spotify) DataInserts() error {
	logger.Log("Attempting to fetch spotify top tracks", logger.Notice)
	songs, err := spotify.Tracks()
	if err != nil {
		logger.Log("Failed to fetch spotify top tracks!", logger.Error)
		return err
	}

	logger.Log("Attempting to fetch spotify top artists", logger.Notice)
	artists, err := spotify.Artists()
	if err != nil {
		logger.Log("Failed to fetch spotify top artists!", logger.Error)
		return err
	}

	logger.Log("Attempting to fetch spotify recently played tracks", logger.Notice)
	recents, err := spotify.Recents()
	if err != nil {
		logger.Log("Failed to fetch spotify recently played tracks!", logger.Error)
		return err
	}

	logger.Log("Fetching tracks from database, then API if missing", logger.Notice)
	dbSongs, err := spotify.PopulateTracks(songs, recents)
	if err != nil {
		logger.Log("Failed to fetch tracks from database", logger.Error)
		return err
	}

	logger.Log("Fetching artists from database, then API if missing", logger.Notice)
	dbArtists, err := spotify.PopulateArtists(songs, artists, recents)
	if err != nil {
		logger.Log("Failed to fetch spotify artists!", logger.Error)
		return err
	}

	logger.Log("Fetching albums from database, then API if missing", logger.Notice)
	dbAlbums, err := spotify.PopulateAlbums(songs, recents)
	if err != nil {
		logger.Log("Failed to fetch spotify recently played albums!", logger.Error)
		return err
	}

	logger.Log("Inserting all relevant thumbnails into DB", logger.Notice)
	err = spotify.InsertThumbnails(songs, recents, artists, dbArtists, dbAlbums)
	if err != nil {
		logger.Log("Failed to insert thumbnails!", logger.Error)
		return err
	}

	logger.Log("Attaching appropriate UUIDs to tracks to be inserted, then inserting", logger.Notice)
	dbSongs, err = spotify.AttachTrackUUIDs(dbSongs, dbArtists, dbAlbums)
	if err != nil {
		logger.Log("Failed to attach and insert songs into the database", logger.Error)
		return err
	}

	logger.Log("Attaching appropriate UUIDs to albums to be inserted, then inserting", logger.Notice)
	err = spotify.AttachAlbumUUIDs(dbAlbums, dbArtists)
	if err != nil {
		logger.Log("Failed to attach and insert albums into the database", logger.Error)
		return err
	}

	logger.Log("Artists dont need related data, simply inserting", logger.Notice)
	err = spotify.InsertArtists(dbArtists)
	if err != nil {
		logger.Log("Failed to insert artists into the database", logger.Error)
		return err
	}

	logger.Log("Prefetch and insert phase complete, moving on to user data insert", logger.Notice)

	logger.Log("Fetching all existing recently listened to songs", logger.Notice)
	existingRecents, err := spotify.FetchExistingRecentListens(recents)
	if err != nil {
		logger.Log("Failed to fetch recently listened to songs from the database", logger.Error)
		return err
	}

	logger.Log("Inserting all recently listened to songs", logger.Notice)
	err = spotify.InsertRecentListens(recents, dbSongs, existingRecents)
	if err != nil {
		logger.Log("Failed to insert artists into the database database", logger.Error)
		return err
	}

	logger.Log("Inserting all top songs", logger.Notice)
	err = spotify.InsertTopSongs(songs, dbSongs)
	if err != nil {
		logger.Log("Failed to insert artists into the database database", logger.Error)
		return err
	}

	logger.Log("Inserting all top artists", logger.Notice)
	err = spotify.InsertTopArtists(artists, dbArtists)
	if err != nil {
		logger.Log("Failed to insert artists into the database database", logger.Error)
		return err
	}

	return nil
}

func (spotify *spotify) InsertTopSongs(songs map[string]TopTracksResponse, dbSongs []models.Song) error {
	newTopSongValues := []interface{}{}

	for term, resp := range songs {
		for i, song := range resp.Items {
			dbSong, exists := getSongBySpotifyID(dbSongs, song.ID)
			newTopSong := models.NewTopSong(spotify.UserID, "", i+1, term)
			if exists {
				newTopSong.SongID = dbSong.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach song ID for song %s", song.Name), logger.Warning)
			}
			newTopSongValues = append(newTopSongValues, newTopSong.ToSlice()...)
		}
	}

	logger.Log("Inserting top songs", logger.Notice)
	err := spotify.createTopSongs(newTopSongValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *spotify) InsertTopArtists(songs map[string]TopArtistsResponse, dbArtists []models.Artist) error {
	newTopArtistValues := []interface{}{}

	for term, resp := range songs {
		for i, artist := range resp.Items {
			dbArtist, exists := getArtistBySpotifyID(dbArtists, artist.ID)
			newTopArtist := models.NewTopArtist(artist.Name, "", i+1, term, spotify.UserID)
			if exists {
				newTopArtist.ArtistID = dbArtist.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach artist ID for artist %s", artist.Name), logger.Warning)
			}
			newTopArtistValues = append(newTopArtistValues, newTopArtist.ToSlice()...)
		}
	}

	logger.Log("Inserting top artists", logger.Notice)
	err := spotify.createTopArtists(newTopArtistValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *spotify) InsertRecentListens(recents RecentlyPlayedResponse, songs []models.Song, existingRecentListens []models.RecentListen) error {
	recentListenValues := []interface{}{}
Outer:
	for _, recentListen := range recents.Items {
		for _, existingRecentListen := range existingRecentListens {
			if recentListen.PlayedAt.Format(time.RFC3339) == existingRecentListen.PlayedAt.Format(time.RFC3339) {
				continue Outer
			}
		}

		newRecentListen := models.NewRecentListen("", spotify.UserID, recentListen.PlayedAt)
		song, exists := getSongBySpotifyID(songs, recentListen.Track.ID)
		if exists {
			newRecentListen.SongID = song.ID
		} else {
			fmt.Printf("Failed to find %s\n", recentListen.Track.ID)
		}
		recentListenValues = append(recentListenValues, newRecentListen.ToSlice()...)
	}

	if len(recentListenValues) == 0 {
		logger.Log("syke bitch the database already contains recent listens for this user", logger.Notice)
		return nil
	}

	logger.Log("Inserting new recently listened to songs", logger.Notice)
	err := spotify.Database.CreateRecentlyListened(recentListenValues)
	if err != nil {
		return err

	}

	return nil
}

func (spotify *spotify) FetchExistingRecentListens(recents RecentlyPlayedResponse) ([]models.RecentListen, error) {
	recentPlayedAtList := []interface{}{}
	for _, recent := range recents.Items {
		recentPlayedAtList = append(recentPlayedAtList, recent.PlayedAt.Format(time.RFC3339))
	}

	recentListens, err := spotify.Database.FetchRecentListensByTime(recentPlayedAtList, spotify.UserID)
	if err != nil {
		return nil, err
	}

	return recentListens, nil
}

func (spotify *spotify) AttachTrackUUIDs(songs []models.Song, artists []models.Artist, albums []models.Album) ([]models.Song, error) {
	songValues := []interface{}{}
	for i, song := range songs {
		if !song.NeedsUpdate {
			continue
		}

		for _, artist := range artists {
			if songs[i].ArtistID == artist.SpotifyID {
				songs[i].ArtistID = artist.ID
				break
			}
		}
		for _, album := range albums {
			if songs[i].AlbumID == album.SpotifyID {
				songs[i].AlbumID = album.ID
				break
			}
		}
		songValues = append(songValues, songs[i].ToSlice()...)
	}

	if len(songValues) == 0 {
		logger.Log("syke bitch the database already contains all necessary songs", logger.Notice)
		return songs, nil
	}

	logger.Log("Inserting new songs", logger.Notice)
	err := spotify.createSongs(songValues)
	if err != nil {
		return nil, err
	}

	return songs, nil
}

func (spotify *spotify) InsertArtists(artists []models.Artist) error {
	artistValues := []interface{}{}
	for _, artist := range artists {
		if !artist.NeedsUpdate {
			continue
		}
		artistValues = append(artistValues, artist.ToSlice()...)
	}

	if len(artistValues) == 0 {
		logger.Log("syke bitch the database already contains all necessary artists", logger.Notice)
		return nil
	}

	logger.Log("Inserting new artists", logger.Notice)
	err := spotify.createArtists(artistValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *spotify) AttachAlbumUUIDs(albums []models.Album, artists []models.Artist) error {
	albumValues := []interface{}{}

	for i, album := range albums {
		if !album.NeedsUpdate {
			continue
		}
		for _, artist := range artists {
			if albums[i].ArtistID == artist.SpotifyID {
				albums[i].ArtistID = artist.ID
				break
			}
		}
		albumValues = append(albumValues, albums[i].ToSlice()...)
	}

	if len(albumValues) == 0 {
		logger.Log("syke bitch the database already contains all necessary albums", logger.Notice)
		return nil
	}

	logger.Log("Inserting new albums", logger.Notice)
	err := spotify.createAlbums(albumValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *spotify) PopulateTracks(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse) ([]models.Song, error) {
	// Songs to attempt to fetch from DB
	songSpotifyIDs := []interface{}{}
	for _, resp := range songs {
		for _, song := range resp.Items {
			songSpotifyIDs = append(songSpotifyIDs, song.ID)
		}
	}

	for _, song := range recents.Items {
		songSpotifyIDs = append(songSpotifyIDs, song.Track.ID)
	}

	dbSongs, err := spotify.Database.FetchSongsBySpotifyID(songSpotifyIDs)
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	songsToFetch := []string{}
Outer:
	for _, id := range songSpotifyIDs {
		for _, dbSong := range dbSongs {
			if dbSong.SpotifyID == id {
				continue Outer
			}
		}
		songsToFetch = append(songsToFetch, id.(string))
	}

	if len(songsToFetch) == 0 {
		logger.Log("Skipping querying API for tracks as database contains all necessary data", logger.Notice)
		return dbSongs, nil
	}

	apiSongs, err := spotify.API.GetTracks(songsToFetch)
	if err != nil {
		return nil, err
	}

	for _, song := range apiSongs {
		dbSongs = append(dbSongs, models.Song{Name: song.Name, SpotifyID: song.ID, AlbumID: song.Album.ID, ArtistID: song.Artists[0].ID, NeedsUpdate: true, ID: uuid.New().String()})
	}

	return dbSongs, nil
}

func (spotify *spotify) PopulateArtists(songs map[string]TopTracksResponse, artists map[string]TopArtistsResponse, recents RecentlyPlayedResponse) ([]models.Artist, error) {
	// Songs to attempt to fetch from DB
	artistSpotifyIDs := []interface{}{}
	for _, resp := range songs {
		for _, song := range resp.Items {
			artistSpotifyIDs = append(artistSpotifyIDs, song.Artists[0].ID)
		}
	}

	for _, resp := range artists {
		for _, artist := range resp.Items {
			artistSpotifyIDs = append(artistSpotifyIDs, artist.ID)
		}
	}

	for _, song := range recents.Items {
		artistSpotifyIDs = append(artistSpotifyIDs, song.Track.Album.Artists[0].ID)
	}

	dbArtists, err := spotify.Database.FetchArtistsBySpotifyID(artistSpotifyIDs)
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	artistsToFetch := []string{}
Outer:
	for _, id := range artistSpotifyIDs {
		for _, dbArtist := range dbArtists {
			if dbArtist.SpotifyID == id {
				continue Outer
			}
		}
		artistsToFetch = append(artistsToFetch, id.(string))
	}

	if len(artistsToFetch) == 0 {
		logger.Log("Skipping querying API for tracks as database contains all necessary data", logger.Notice)
		return dbArtists, nil
	}

	apiArtists, err := spotify.API.GetArtists(artistsToFetch)
	if err != nil {
		return nil, err
	}

	for _, artist := range apiArtists {
		dbArtists = append(dbArtists, models.Artist{Name: artist.Name, SpotifyID: artist.ID, NeedsUpdate: true, ID: uuid.New().String()})
	}

	return dbArtists, nil
}

// too many args should use struct tbh
func (spotify *spotify) InsertThumbnails(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse, artists map[string]TopArtistsResponse, dbArtists []models.Artist, dbAlbums []models.Album) error {
	thumbnails := []models.Thumbnail{}
	for _, resp := range songs {
		for _, song := range resp.Items {
			dbAlbum, exists := getAlbumBySpotifyID(dbAlbums, song.Album.ID)
			if !exists {
				logger.Log(fmt.Sprintf("Failed to attach album ID for album %s", song.Album.Name), logger.Warning)
			}
			for _, image := range song.Album.Images {
				thumbnail := models.NewThumbnail("album", "", image.URL, image.Height, image.Width)
				if exists {
					thumbnail.EntityID = dbAlbum.ID
				}
				thumbnails = append(thumbnails, thumbnail)
			}
		}
	}

	for _, resp := range artists {
		for _, artist := range resp.Items {
			dbArtist, exists := getArtistBySpotifyID(dbArtists, artist.ID)
			if !exists {
				logger.Log(fmt.Sprintf("Failed to attach artist ID for artist %s", artist.Name), logger.Warning)
			}
			for _, image := range artist.Images {
				thumbnail := models.NewThumbnail("artist", "", image.URL, image.Height, image.Width)
				if exists {
					thumbnail.EntityID = dbArtist.ID
				}
				thumbnails = append(thumbnails, thumbnail)
			}
		}
	}

	for _, song := range recents.Items {
		dbAlbum, exists := getAlbumBySpotifyID(dbAlbums, song.Track.Album.ID)
		if !exists {
			logger.Log(fmt.Sprintf("Failed to attach album ID for track %s", song.Track.Album.Name), logger.Warning)
		}
		for _, image := range song.Track.Album.Images {
			thumbnail := models.NewThumbnail("album", "", image.URL, image.Height, image.Width)
			if exists {
				thumbnail.EntityID = dbAlbum.ID
			}
			thumbnails = append(thumbnails, thumbnail)
		}
	}

	entityIDs := []interface{}{}
	for _, thumbnail := range thumbnails {
		entityIDs = append(entityIDs, thumbnail.EntityID)
	}

	dbThumbnails, err := spotify.Database.FetchThumbnailsByEntityID(entityIDs)
	if err != nil {
		return err
	}

	thumbnailsToInsert := []interface{}{}
	for _, thumbnail := range thumbnails {
		_, exists := getThumbnailByEntityID(dbThumbnails, thumbnail.EntityID)
		if !exists {
			thumbnailsToInsert = append(thumbnailsToInsert, thumbnail.ToSlice()...)
		}
	}

	if len(thumbnailsToInsert) == 0 {
		logger.Log("syke bitch the database already contains all necessary thumbnails", logger.Notice)
		return nil
	}

	err = spotify.createThumbnails(thumbnailsToInsert)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *spotify) PopulateAlbums(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse) ([]models.Album, error) {
	// Songs to attempt to fetch from DB
	albumSpotifyIDs := []interface{}{}
	for _, resp := range songs {
		for _, song := range resp.Items {
			albumSpotifyIDs = append(albumSpotifyIDs, song.Album.ID)
		}
	}

	for _, song := range recents.Items {
		albumSpotifyIDs = append(albumSpotifyIDs, song.Track.Album.ID)
	}

	dbAlbums, err := spotify.Database.FetchAlbumsBySpotifyID(albumSpotifyIDs)
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	albumsToFetch := []string{}
Outer:
	for _, id := range albumSpotifyIDs {
		for _, dbAlbum := range dbAlbums {
			if dbAlbum.SpotifyID == id {
				continue Outer
			}
		}
		albumsToFetch = append(albumsToFetch, id.(string))
	}

	if len(albumsToFetch) == 0 {
		logger.Log("Skipping querying API for albums as database contains all necessary data", logger.Notice)
		return dbAlbums, nil
	}

	apiAlbums, err := spotify.API.GetAlbums(albumsToFetch)
	if err != nil {
		return nil, err
	}

	for _, album := range apiAlbums {
		dbAlbums = append(dbAlbums, models.Album{Name: album.Name, SpotifyID: album.ID, ArtistID: album.Artists[0].ID, NeedsUpdate: true, ID: uuid.New().String()})
	}

	return dbAlbums, nil
}

func (spotify *spotify) Artists() (map[string]TopArtistsResponse, error) {
	artistsResp := make(map[string]TopArtistsResponse)

	for _, period := range spotify.Times {
		logger.Log(fmt.Sprintf("Processing %s_term time range for artists endpoint", period), logger.Notice)
		artists, err := spotify.API.GetTopArtists(period + "_term")
		if err != nil {
			return nil, err
		}
		artistsResp[period] = artists
	}
	return artistsResp, nil
}

func (spotify *spotify) Recents() (RecentlyPlayedResponse, error) {
	recentlyPlayed, err := spotify.API.GetRecentlyPlayed()
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}

	return recentlyPlayed, nil
}

func (spotify *spotify) Tracks() (map[string]TopTracksResponse, error) {
	topTrackResp := make(map[string]TopTracksResponse)

	for _, period := range spotify.Times {
		logger.Log(fmt.Sprintf("Processing %s_term time range for tracks endpoint", period), logger.Notice)
		tracks, err := spotify.API.GetTopTracks(period + "_term")
		if err != nil {
			return nil, err
		}

		topTrackResp[period] = tracks
	}
	return topTrackResp, nil
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

func getSongBySpotifyID(songs []models.Song, spotifyID string) (models.Song, bool) {
	for _, song := range songs {
		if song.SpotifyID == spotifyID {
			return song, true
		}
	}
	return models.Song{}, false
}

func getArtistBySpotifyID(artists []models.Artist, spotifyID string) (models.Artist, bool) {
	for _, artist := range artists {
		if artist.SpotifyID == spotifyID {
			return artist, true
		}
	}
	return models.Artist{}, false
}

func getAlbumBySpotifyID(albums []models.Album, spotifyID string) (models.Album, bool) {
	for _, album := range albums {
		if album.SpotifyID == spotifyID {
			return album, true
		}
	}
	return models.Album{}, false
}

func getThumbnailByEntityID(thumbnails []models.Thumbnail, entityID string) (models.Thumbnail, bool) {
	for _, thumbnail := range thumbnails {
		if thumbnail.EntityID == entityID {
			return thumbnail, true
		}
	}
	return models.Thumbnail{}, false
}

func BasicAuth(clientID string, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}
