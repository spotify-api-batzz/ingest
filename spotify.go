package main

import (
	"fmt"
	"spotify/logger"
	"spotify/models"
	"time"

	"github.com/google/uuid"
)

type Spotify struct {
	Database *Database
	UserID   string
	API      *API
	Times    []string
}

func newSpotify(database *Database, api *API, userID string) Spotify {
	return Spotify{
		Database: database,
		API:      api,

		Times:  []string{"short", "medium", "long"},
		UserID: userID,
	}
}

func (spotify *Spotify) DataInserts() error {
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
		logger.Log("Failed to insert recentlistens into the database", logger.Error)
		return err
	}

	logger.Log("Inserting all top songs", logger.Notice)
	err = spotify.InsertTopSongs(songs, dbSongs)
	if err != nil {
		logger.Log("Failed to insert top songs into the database", logger.Error)
		return err
	}

	logger.Log("Inserting all top artists", logger.Notice)
	err = spotify.InsertTopArtists(artists, dbArtists)
	if err != nil {
		logger.Log("Failed to insert top artists into the database", logger.Error)
		return err
	}

	return nil
}

func (spotify *Spotify) InsertTopSongs(songs map[string]TopTracksResponse, dbSongs []models.Song) error {
	topSongDataValues := []interface{}{}
	topSongValues := []interface{}{}

	topSong := models.NewTopSong(spotify.UserID)

	for term, resp := range songs {
		for i, song := range resp.Items {
			dbSong, exists := getSongBySpotifyID(dbSongs, song.ID)
			newTopSong := models.NewTopSongData(topSong.ID.String(), "", i+1, term)
			if exists {
				newTopSong.SongID = dbSong.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach song ID for song %s", song.Name), logger.Warning)
			}
			topSongDataValues = append(topSongDataValues, newTopSong.ToSlice()...)
		}
	}

	topSongValues = append(topSongValues, topSong.ToSlice()...)

	if len(topSongDataValues) == 0 {
		logger.Log("syke bitch no top song data to insert", logger.Notice)
		return nil
	}

	logger.Log("Inserting new top song record", logger.Notice)
	err := spotify.Database.Create(&models.TopSong{}, topSongValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new top song data", logger.Notice)
	err = spotify.Database.Create(&models.TopSongData{}, topSongDataValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) InsertTopArtists(songs map[string]TopArtistsResponse, dbArtists []models.Artist) error {
	topArtistDataValues := []interface{}{}
	topArtistValues := []interface{}{}
	newTopArtist := models.NewTopArtist(spotify.UserID)

	for term, resp := range songs {
		for i, artist := range resp.Items {
			dbArtist, exists := getArtistBySpotifyID(dbArtists, artist.ID)
			newTopArtist := models.NewTopArtistData(artist.Name, "", i+1, term, newTopArtist.ID.String())
			if exists {
				newTopArtist.ArtistID = dbArtist.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach artist ID for artist %s", artist.Name), logger.Warning)
			}
			topArtistDataValues = append(topArtistDataValues, newTopArtist.ToSlice()...)
		}
	}

	topArtistValues = append(topArtistValues, newTopArtist.ToSlice()...)

	if len(topArtistDataValues) == 0 {
		logger.Log("syke bitch no top artist data to insert", logger.Notice)
		return nil
	}

	logger.Log("Inserting new top artist record", logger.Notice)
	err := spotify.Database.Create(&models.TopArtist{}, topArtistValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new top artist data", logger.Notice)
	err = spotify.Database.Create(&models.TopArtistData{}, topArtistDataValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) InsertRecentListens(recents RecentlyPlayedResponse, songs []models.Song, existingRecentListens []models.RecentListenData) error {

	newRecentListen := models.NewRecentListen(spotify.UserID)
	recentListenValues := []interface{}{}
	recentListenDataValues := []interface{}{}
Outer:
	for _, recentListen := range recents.Items {
		for _, existingRecentListen := range existingRecentListens {
			if recentListen.PlayedAt.Format(time.RFC3339) == existingRecentListen.PlayedAt.Format(time.RFC3339) {
				continue Outer
			}
		}

		newRecentListenData := models.NewRecentListenData("", newRecentListen.ID.String(), recentListen.PlayedAt)
		song, exists := getSongBySpotifyID(songs, recentListen.Track.ID)
		if exists {
			newRecentListenData.SongID = song.ID
		} else {
			fmt.Printf("Failed to find %s\n", recentListen.Track.ID)
		}
		recentListenDataValues = append(recentListenDataValues, newRecentListenData.ToSlice()...)
	}

	recentListenValues = append(recentListenValues, newRecentListen.ToSlice()...)

	if len(recentListenDataValues) == 0 {
		logger.Log("syke bitch the database already contains recent listens for this user", logger.Notice)
		return nil
	}

	logger.Log("Inserting new recently listened to record", logger.Notice)
	err := spotify.Database.Create(&models.RecentListen{}, recentListenValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new recently listened to songs", logger.Notice)
	err = spotify.Database.Create(&models.RecentListenData{}, recentListenDataValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) FetchExistingRecentListens(recents RecentlyPlayedResponse) ([]models.RecentListenData, error) {
	recentPlayedAtList := []interface{}{}
	for _, recent := range recents.Items {
		recentPlayedAtList = append(recentPlayedAtList, recent.PlayedAt)
	}

	if len(recentPlayedAtList) == 0 {
		logger.Log("syke bitch no recently played data to fetch", logger.Notice)
		return []models.RecentListenData{}, nil
	}

	recentListens, err := spotify.Database.FetchRecentListensByUserID(spotify.UserID)
	if err != nil {
		return nil, err
	}

	if len(recentListens) == 0 {
		logger.Log("syke bitch we have no recently played data for this user in the database", logger.Notice)
		return []models.RecentListenData{}, nil
	}

	recentListenIDs := []interface{}{}
	for _, recentListen := range recentListens {
		recentListenIDs = append(recentListenIDs, recentListen.ID.String())
	}

	recentListenData, err := spotify.Database.FetchRecentListenDataByTime(recentPlayedAtList, recentListenIDs)
	if err != nil {
		return nil, err
	}

	return recentListenData, nil
}

func (spotify *Spotify) AttachTrackUUIDs(songs []models.Song, artists []models.Artist, albums []models.Album) ([]models.Song, error) {
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
	err := spotify.Database.CreateSong(songValues)
	if err != nil {
		return nil, err
	}

	return songs, nil
}

func (spotify *Spotify) InsertArtists(artists []models.Artist) error {
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
	err := spotify.Database.CreateArtist(artistValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) AttachAlbumUUIDs(albums []models.Album, artists []models.Artist) error {
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
	err := spotify.Database.CreateAlbum(albumValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) PopulateTracks(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse) ([]models.Song, error) {
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

func (spotify *Spotify) PopulateArtists(songs map[string]TopTracksResponse, artists map[string]TopArtistsResponse, recents RecentlyPlayedResponse) ([]models.Artist, error) {
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
func (spotify *Spotify) InsertThumbnails(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse, artists map[string]TopArtistsResponse, dbArtists []models.Artist, dbAlbums []models.Album) error {
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

	err = spotify.Database.CreateThumbnail(thumbnailsToInsert)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *Spotify) PopulateAlbums(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse) ([]models.Album, error) {
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

func (spotify *Spotify) Artists() (map[string]TopArtistsResponse, error) {
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

func (spotify *Spotify) Recents() (RecentlyPlayedResponse, error) {
	recentlyPlayed, err := spotify.API.GetRecentlyPlayed()
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}

	return recentlyPlayed, nil
}

func (spotify *Spotify) Tracks() (map[string]TopTracksResponse, error) {
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
