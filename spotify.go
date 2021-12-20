package main

import (
	"fmt"
	"spotify/models"
	"spotify/utils"
	"time"

	"github.com/batzz-00/goutils/logger"

	"github.com/google/uuid"
)

type SpotifyIngestOptions struct {
	RecentListen bool
	TopSongs     bool
	TopArtists   bool
	UserID       string
}

type Spotify struct {
	Database *Database
	Options  SpotifyIngestOptions
	API      SpotifyAPI
	Times    []string
}

type APIData struct {
	Songs   map[string]TopTracksResponse
	Artists map[string]TopArtistsResponse
	Recents RecentlyPlayedResponse
}

type DBData struct {
	Songs             []models.Song
	Albums            []models.Album
	Artists           []models.Artist
	RecentListensData []models.RecentListenData
}

func newSpotify(database *Database, api SpotifyAPI, options SpotifyIngestOptions) Spotify {
	return Spotify{
		Database: database,
		API:      api,

		Times:   []string{"short", "medium", "long"},
		Options: options,
	}
}
func (spotify *Spotify) Ingest() error {
	logger.Log("Fetching user data from spotify API", logger.Info)
	APIData, err := spotify.FetchAPIData()
	if err != nil {
		return err
	}

	logger.Log("Fetching related data for songs, albums etc from database and then the spotify API if we don't already have it", logger.Info)
	relatedData, err := spotify.FetchRelated(APIData)
	if err != nil {
		return err
	}

	logger.Log("Attaching appropriate UUIDs based on spotify ID, and inserting any missed data freshly gathered from spotify API", logger.Info)
	relatedData, err = spotify.AttachAndInsertFreshData(APIData, relatedData)
	if err != nil {
		return err
	}

	logger.Log("Fetching all existing recently listened to songs", logger.Info)
	relatedData.RecentListensData, err = spotify.FetchExistingRecentListens(APIData.Recents)
	if err != nil {
		logger.Log("Failed to fetch recently listened to songs from the database", logger.Error)
		return err
	}

	logger.Log("Prefetch and insert phase complete, moving on to user data insert", logger.Info)
	err = spotify.InsertUserData(APIData, relatedData)
	if err != nil {
		logger.Log("Failed to insert user data", logger.Error)
		return err
	}

	return nil
}

func (spotify *Spotify) FetchAPIData() (APIData, error) {
	logger.Log("Attempting to fetch users top tracks", logger.Info)
	songs, err := spotify.Tracks()
	if err != nil {
		logger.Log("Failed to fetch users top tracks!", logger.Error)
		return APIData{}, err
	}

	logger.Log("Attempting to fetch users top artists", logger.Info)
	artists, err := spotify.Artists()
	if err != nil {
		logger.Log("Failed to fetch users top artists!", logger.Error)
		return APIData{}, err
	}

	logger.Log("Attempting to fetch users recently played tracks", logger.Info)
	recents, err := spotify.Recents()
	if err != nil {
		logger.Log("Failed to fetch users recently played tracks!", logger.Error)
		return APIData{}, err
	}

	return APIData{
		Songs:   songs,
		Artists: artists,
		Recents: recents,
	}, nil
}

func (spotify *Spotify) FetchRelated(APIData APIData) (DBData, error) {
	logger.Log("Fetching tracks from database, then API if missing", logger.Info)
	dbSongs, err := spotify.PopulateTracks(APIData.Songs, APIData.Recents)
	if err != nil {
		logger.Log("Failed to fetch tracks from database", logger.Error)
		return DBData{}, err
	}

	logger.Log("Fetching artists from database, then API if missing", logger.Info)
	dbArtists, err := spotify.PopulateArtists(APIData.Songs, APIData.Artists, APIData.Recents)
	if err != nil {
		logger.Log("Failed to fetch spotify artists!", logger.Error)
		return DBData{}, err
	}

	logger.Log("Fetching albums from database, then API if missing", logger.Info)
	dbAlbums, err := spotify.PopulateAlbums(APIData.Songs, APIData.Recents)
	if err != nil {
		logger.Log("Failed to fetch spotify recently played albums!", logger.Error)
		return DBData{}, err
	}

	return DBData{
		Songs:   dbSongs,
		Artists: dbArtists,
		Albums:  dbAlbums,
	}, err
}

func (spotify *Spotify) AttachAndInsertFreshData(APIData APIData, dbData DBData) (DBData, error) {
	logger.Log("Artists dont need related data, simply inserting", logger.Info)
	err := spotify.InsertArtists(dbData.Artists)
	if err != nil {
		logger.Log("Failed to insert artists into the database", logger.Error)
		return dbData, err
	}

	logger.Log("Attaching appropriate UUIDs to albums to be inserted, then inserting", logger.Info)
	err = spotify.AttachAlbumUUIDs(dbData.Albums, dbData.Artists)
	if err != nil {
		logger.Log("Failed to attach and insert albums into the database", logger.Error)
		return dbData, err
	}

	logger.Log("Attaching appropriate UUIDs to tracks to be inserted, then inserting", logger.Info)
	dbSongs, err := spotify.AttachTrackUUIDs(dbData.Songs, dbData.Artists, dbData.Albums)
	if err != nil {
		logger.Log("Failed to attach and insert songs into the database", logger.Error)
		return dbData, err
	}
	dbData.Songs = dbSongs

	logger.Log("Inserting all relevant thumbnails into DB", logger.Info)
	err = spotify.InsertThumbnails(APIData.Songs, APIData.Recents, APIData.Artists, dbData.Artists, dbData.Albums)
	if err != nil {
		logger.Log("Failed to insert thumbnails!", logger.Error)
		return dbData, err
	}

	return dbData, nil
}

func (spotify *Spotify) InsertUserData(APIData APIData, dbData DBData) error {

	if spotify.Options.RecentListen {
		logger.Log("Inserting all recently listened to songs", logger.Info)
		err := spotify.InsertRecentListens(APIData.Recents, dbData.Songs, dbData.RecentListensData)
		if err != nil {
			logger.Log("Failed to insert recentlistens into the database", logger.Error)
			return err
		}
	}

	if spotify.Options.TopSongs {
		logger.Log("Inserting all top songs", logger.Info)
		err := spotify.InsertTopSongs(APIData.Songs, dbData.Songs)
		if err != nil {
			logger.Log("Failed to insert top songs into the database", logger.Error)
			return err
		}
	}

	if spotify.Options.TopSongs {
		logger.Log("Inserting all top artists", logger.Info)
		err := spotify.InsertTopArtists(APIData.Artists, dbData.Artists)
		if err != nil {
			logger.Log("Failed to insert top artists into the database", logger.Error)
			return err
		}
	}

	return nil
}

func (spotify *Spotify) InsertTopSongs(songs map[string]TopTracksResponse, dbSongs []models.Song) error {
	topSongDataValues := []interface{}{}
	topSongValues := []interface{}{}

	topSong := models.NewTopSong(spotify.Options.UserID)

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
		logger.Log("No top song data to ingest.", logger.Info)
		return nil
	}

	logger.Log("Inserting new top song record", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_song records", len(topSongValues)/len((&models.TopSong{}).TableColumns())), logger.Debug)
	err := spotify.Database.Create(&models.TopSong{}, topSongValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new top song data", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_song_data records", len(topSongDataValues)/len((&models.TopSongData{}).TableColumns())), logger.Debug)
	err = spotify.Database.Create(&models.TopSongData{}, topSongDataValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) InsertTopArtists(songs map[string]TopArtistsResponse, dbArtists []models.Artist) error {
	topArtistDataValues := []interface{}{}
	topArtistValues := []interface{}{}
	newTopArtist := models.NewTopArtist(spotify.Options.UserID)

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
		logger.Log("No top artist data to ingest.", logger.Info)
		return nil
	}

	logger.Log("Inserting new top artist record", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_artist records", len(topArtistValues)/len((&models.TopArtist{}).TableColumns())), logger.Debug)
	err := spotify.Database.Create(&models.TopArtist{}, topArtistValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new top artist data", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_artist_data records", len(topArtistDataValues)/len((&models.TopArtistData{}).TableColumns())), logger.Debug)
	err = spotify.Database.Create(&models.TopArtistData{}, topArtistDataValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) InsertRecentListens(recents RecentlyPlayedResponse, songs []models.Song, existingRecentListens []models.RecentListenData) error {

	logger.Log(spotify.Options.UserID, logger.Warning)
	logger.Log(spotify.Options.UserID, logger.Warning)
	logger.Log(spotify.Options.UserID, logger.Warning)
	logger.Log(spotify.Options.UserID, logger.Warning)
	logger.Log(spotify.Options.UserID, logger.Warning)
	logger.Log(spotify.Options.UserID, logger.Warning)
	logger.Log(spotify.Options.UserID, logger.Warning)
	newRecentListen := models.NewRecentListen(spotify.Options.UserID)
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
		logger.Log("No recent listen data to ingest", logger.Info)
		return nil
	}

	logger.Log("Inserting new recently listened to record", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d new recent_listen records", len(recentListenValues)/len((&models.RecentListen{}).TableColumns())), logger.Debug)
	err := spotify.Database.Create(&models.RecentListen{}, recentListenValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new recently listened to songs", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d new recent_listen_data records", len(recentListenDataValues)/len((&models.RecentListenData{}).TableColumns())), logger.Debug)
	err = spotify.Database.Create(&models.RecentListenData{}, recentListenDataValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) FetchExistingRecentListens(recents RecentlyPlayedResponse) ([]models.RecentListenData, error) {
	recentPlayedAtList := []interface{}{}
	for _, recent := range recents.Items {
		recentPlayedAtList = append(recentPlayedAtList, recent.PlayedAt.Format(time.RFC3339))
	}

	if len(recentPlayedAtList) == 0 {
		logger.Log("User has no recently played songs", logger.Debug)
		return []models.RecentListenData{}, nil
	}

	recentListens, err := spotify.Database.FetchRecentListensByUserID(spotify.Options.UserID)
	if err != nil {
		return nil, err
	}

	if len(recentListens) == 0 {
		logger.Log("Database contains no recently listened to songs for this user", logger.Debug)
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
		logger.Log("No new song data to ingest", logger.Debug)
		return songs, nil
	}

	logger.Log(fmt.Sprintf("Inserting %d new song records", len(songValues)/len((&models.Song{}).TableColumns())), logger.Debug)
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
		logger.Log("No new artist data to ingest", logger.Debug)
		return nil
	}

	logger.Log("Inserting new artists", logger.Debug)
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
			// hardcoded 'various artists'
			if albums[i].ArtistID == "0LyfQWJT6nXafLPZqxe9Of" {
				albums[i].ArtistID = "7e1339ee-0d2a-4de6-9187-78d6874ae044"
			}
			if albums[i].ArtistID == artist.SpotifyID {
				albums[i].ArtistID = artist.ID
				break
			}
		}
		albumValues = append(albumValues, albums[i].ToSlice()...)
	}

	if len(albumValues) == 0 {
		logger.Log("No new album data to ingest", logger.Debug)
		return nil
	}

	logger.Log("Inserting new albums", logger.Debug)
	err := spotify.Database.CreateAlbum(albumValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) PopulateTracks(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse) ([]models.Song, error) {
	// Songs to attempt to fetch from DB
	songSpotifyIDs := utils.NewStringArgs()
	for _, resp := range songs {
		for _, song := range resp.Items {
			songSpotifyIDs.Add(song.ID)
		}
	}

	for _, song := range recents.Items {
		songSpotifyIDs.Add(song.Track.ID)
	}

	logger.Log(fmt.Sprintf("Querying database for %d songs", len(songSpotifyIDs.Args())), logger.Debug)
	dbSongs, err := spotify.Database.FetchSongsBySpotifyID(songSpotifyIDs.Args())
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	songsToFetch := []string{}
Outer:
	for id := range songSpotifyIDs.UniqueMap {
		for _, song := range dbSongs {
			if id == song.SpotifyID {
				continue Outer
			}
		}
		songsToFetch = append(songsToFetch, id)
	}

	if len(songsToFetch) == 0 {
		logger.Log("Database contains already contains all ingested song references, not querying api", logger.Debug)
		return dbSongs, nil
	}

	apiSongs, err := spotify.API.TracksBySpotifyID(songsToFetch)
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

	artistSpotifyIDs := utils.NewStringArgs()
	// artistSpotifyIDs := []interface{}{}
	for _, resp := range songs {
		for _, song := range resp.Items {
			artistSpotifyIDs.Add(song.Artists[0].ID)
		}
	}

	for _, resp := range artists {
		for _, artist := range resp.Items {
			artistSpotifyIDs.Add(artist.ID)
		}
	}

	for _, song := range recents.Items {
		artistSpotifyIDs.Add(song.Track.Album.Artists[0].ID)
	}

	logger.Log(fmt.Sprintf("Querying database for %d artists", len(artistSpotifyIDs.Args())), logger.Debug)
	dbArtists, err := spotify.Database.FetchArtistsBySpotifyID(artistSpotifyIDs.Args())
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	artistsToFetch := []string{}
Outer:
	for id := range artistSpotifyIDs.UniqueMap {
		for _, dbArtist := range dbArtists {
			if dbArtist.SpotifyID == id {
				continue Outer
			}
		}
		artistsToFetch = append(artistsToFetch, id)
	}

	if len(artistsToFetch) == 0 {
		logger.Log("Database contains already contains all ingested artist references, not querying api", logger.Debug)
		return dbArtists, nil
	}

	apiArtists, err := spotify.API.ArtistsBySpotifyID(artistsToFetch)
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
	thumbnails := make(map[string]models.Thumbnail)
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
				thumbnails[thumbnail.UniqueID()] = thumbnail
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
				thumbnails[thumbnail.UniqueID()] = thumbnail
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
			thumbnails[thumbnail.UniqueID()] = thumbnail
		}
	}

	entityIDs := make([]interface{}, len(thumbnails))
	for _, thumbnail := range thumbnails {
		entityIDs = append(entityIDs, thumbnail.EntityID)
	}

	dbThumbnails, err := spotify.Database.FetchThumbnailsByEntityID(entityIDs)
	if err != nil {
		return err
	}

	thumbnailsToInsert := []interface{}{}
	for _, thumbnail := range thumbnails {
		_, exists := getThumbnailByEntityIDAndDimensions(dbThumbnails, thumbnail.EntityID, thumbnail.Height, thumbnail.Width)
		if !exists {
			thumbnailsToInsert = append(thumbnailsToInsert, thumbnail.ToSlice()...)
		}
	}

	if len(thumbnailsToInsert) == 0 {
		logger.Log("No new thumbnails to ingest", logger.Debug)
		return nil
	}

	logger.Log(fmt.Sprintf("Inserting %d new thumbnail records", len(thumbnailsToInsert)), logger.Debug)
	err = spotify.Database.CreateThumbnail(thumbnailsToInsert)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *Spotify) PopulateAlbums(songs map[string]TopTracksResponse, recents RecentlyPlayedResponse) ([]models.Album, error) {
	albumSpotifyIDs := utils.NewStringArgs()
	for _, resp := range songs {
		for _, song := range resp.Items {
			albumSpotifyIDs.Add(song.Album.ID)
		}
	}

	for _, song := range recents.Items {
		albumSpotifyIDs.Add(song.Track.Album.ID)
	}

	logger.Log(fmt.Sprintf("Querying database for %d albums", len(albumSpotifyIDs.Args())), logger.Debug)
	dbAlbums, err := spotify.Database.FetchAlbumsBySpotifyID(albumSpotifyIDs.Args())
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	albumsToFetch := []string{}
Outer:
	for id := range albumSpotifyIDs.UniqueMap {
		for _, dbAlbum := range dbAlbums {
			if dbAlbum.SpotifyID == id {
				continue Outer
			}
		}
		albumsToFetch = append(albumsToFetch, id)
	}

	if len(albumsToFetch) == 0 {
		logger.Log("Database contains already contains all ingested album references, not querying api", logger.Debug)
		return dbAlbums, nil
	}

	apiAlbums, err := spotify.API.AlbumsBySpotifyID(albumsToFetch)
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
		logger.Log(fmt.Sprintf("Processing %s_term time range for artists endpoint", period), logger.Debug)
		artists, err := spotify.API.TopArtistsForUser(period + "_term")
		if err != nil {
			return nil, err
		}
		artistsResp[period] = artists
	}
	return artistsResp, nil
}

func (spotify *Spotify) Recents() (RecentlyPlayedResponse, error) {
	recentlyPlayed, err := spotify.API.RecentlyPlayedByUser()
	if err != nil {
		return RecentlyPlayedResponse{}, err
	}

	return recentlyPlayed, nil
}

func (spotify *Spotify) Tracks() (map[string]TopTracksResponse, error) {
	topTrackResp := make(map[string]TopTracksResponse)

	for _, period := range spotify.Times {
		logger.Log(fmt.Sprintf("Processing %s_term time range for tracks endpoint", period), logger.Debug)
		tracks, err := spotify.API.TopTracksForUser(period + "_term")
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

func getThumbnailByEntityIDAndDimensions(thumbnails []models.Thumbnail, entityID string, entityHeight int, entityWidth int) (models.Thumbnail, bool) {
	for _, thumbnail := range thumbnails {
		if thumbnail.EntityID == entityID && thumbnail.Height == entityHeight && thumbnail.Width == entityWidth {
			return thumbnail, true
		}
	}
	return models.Thumbnail{}, false
}
