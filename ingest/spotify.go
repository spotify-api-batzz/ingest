package ingest

import (
	"fmt"
	"spotify/api"
	"spotify/models"
	"spotify/utils"
	"time"

	"github.com/batzz-00/goutils/logger"

	"github.com/google/uuid"
)

type EntityMetric struct {
	ProcessedCount int
	NewCount       int
}
type SpotifyIngestStats struct {
	Songs     EntityMetric
	Albums    EntityMetric
	Artists   EntityMetric
	StartTime time.Time
	EndTime   time.Time
}

type SpotifyIngestEvents struct {
	OnNewEntity *func(model models.Model)
	OnFinish    *func(stats SpotifyIngestStats)
}

func (e *EntityMetric) IncrementCount(new bool) {
	e.ProcessedCount += 1
	if new {
		e.NewCount += 1
	}
}

type MetricHandler interface {
	AddNewSongIndex(spotifyId string, songName string) error
	AddNewAlbumIndex(spotifyId string, albumName string) error
	// AddNewFailure(failureType string, err error) error
	AddNewThumbnailIndex(entity string, name string, url string) error
}

func NewEntityMetric() EntityMetric {
	return EntityMetric{
		ProcessedCount: 0,
		NewCount:       0,
	}
}

func NewSpotifyIngestStats() SpotifyIngestStats {
	return SpotifyIngestStats{
		Songs:     NewEntityMetric(),
		Albums:    NewEntityMetric(),
		Artists:   NewEntityMetric(),
		StartTime: time.Now(),
	}
}

type SpotifyIngestOptions struct {
	RecentListen       bool
	TopSongs           bool
	TopArtists         bool
	UserID             string
	VariousArtistsUUID string
	EnvUsers           []string
	Events             SpotifyIngestEvents
}

type SpotifyIngestContext struct {
	Options SpotifyIngestOptions
	Id      string
}

type IngestDatabase interface {
	Create(model models.Model, values []interface{}) error
	// FetchUsersBySpotifyIds(names []interface{}) ([]models.User, error)
	FetchArtistsBySpotifyID(spotifyIDs []interface{}) ([]models.Artist, error)
	FetchArtistBySpotifyID(spotifyID string) (models.Artist, error)
	FetchSongsBySpotifyID(spotifyIDs []interface{}) ([]models.Song, error)
	FetchUserByName(name string) (models.User, error)
	FetchAlbumsBySpotifyID(spotifyIDs []interface{}) ([]models.Album, error)
	FetchArtistByID(id string) (models.Artist, error)
	FetchRecentListensByUserIDAndTime(userID string, recentListenedToIDs []interface{}, earliestTime interface{}) ([]models.RecentListen, error)
	FetchThumbnailsByEntityID(entityIDs []interface{}) ([]models.Thumbnail, error)
}

type API interface {
	Me() (api.MeResponse, error)
	RecentlyPlayedByUser() (api.RecentlyPlayedResponse, error)
	TopArtistsForUser(period string) (api.TopArtistsResponse, error)
	TopTracksForUser(period string) (api.TopTracksResponse, error)
	ArtistsBySpotifyID(ids []string) ([]api.Artist, error)
	TracksBySpotifyID(ids []string) ([]api.Song, error)
	AlbumsBySpotifyID(ids []string) ([]api.Album, error)
	Authorize(code string) error
	Refresh() error
}

type SpotifyIngest struct {
	Database      IngestDatabase
	Options       SpotifyIngestOptions
	API           API
	MetricHandler MetricHandler
	Times         []string

	Stats SpotifyIngestStats
}

type APIData struct {
	Songs   map[string]api.TopTracksResponse
	Artists map[string]api.TopArtistsResponse
	Recents api.RecentlyPlayedResponse
}

type DBData struct {
	Songs         []models.Song
	Albums        []models.Album
	Artists       []models.Artist
	RecentListens []models.RecentListen
}

func NewIngestContext(options SpotifyIngestOptions) SpotifyIngestContext {
	return SpotifyIngestContext{
		Id:      uuid.NewString(),
		Options: options,
	}
}

func NewSpotifyIngest(database IngestDatabase, api API, options SpotifyIngestOptions, metricHandler MetricHandler) SpotifyIngest {
	return SpotifyIngest{
		Database:      database,
		API:           api,
		MetricHandler: metricHandler,

		Stats:   NewSpotifyIngestStats(),
		Times:   []string{"short", "medium", "long"},
		Options: options,
	}
}

func (spotify *SpotifyIngest) OnNewEntityEvent(model models.Model) {
	if spotify.Options.Events.OnNewEntity != nil {
		(*spotify.Options.Events.OnNewEntity)(model)
	}
}

func (spotify *SpotifyIngest) OnNewAlbum(model *models.Album, new bool) {
	spotify.OnNewEntityEvent(model)
	spotify.Stats.Albums.IncrementCount(new)
}

func (spotify *SpotifyIngest) OnNewArtist(model *models.Artist, new bool) {
	spotify.OnNewEntityEvent(model)
	spotify.Stats.Artists.IncrementCount(new)
}

func (spotify *SpotifyIngest) OnNewSong(model *models.Song, new bool) {
	spotify.OnNewEntityEvent(model)
	spotify.Stats.Songs.IncrementCount(new)
}

func (spotify *SpotifyIngest) Ingest() error {
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
	relatedData.RecentListens, err = spotify.FetchExistingRecentListens(APIData.Recents)
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

	spotify.Stats.EndTime = time.Now()
	return nil
}

func (spotify *SpotifyIngest) FetchAPIData() (APIData, error) {
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

func (spotify *SpotifyIngest) FetchRelated(APIData APIData) (DBData, error) {
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

func (spotify *SpotifyIngest) AttachAndInsertFreshData(APIData APIData, dbData DBData) (DBData, error) {
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

func (spotify *SpotifyIngest) InsertUserData(APIData APIData, dbData DBData) error {
	if spotify.Options.RecentListen {
		logger.Log("Inserting all recently listened to songs", logger.Info)
		err := spotify.InsertRecentListens(APIData.Recents, dbData.Songs, dbData.RecentListens)
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

func (spotify *SpotifyIngest) InsertTopSongs(songs map[string]api.TopTracksResponse, dbSongs []models.Song) error {
	topSongDataValues := []interface{}{}
	topSongValues := []interface{}{}

	topSong := models.NewTopSong(spotify.Options.UserID)
	spotify.OnNewEntityEvent(&topSong)
	for term, resp := range songs {
		for i, song := range resp.Items {
			dbSong, exists := getSongBySpotifyID(dbSongs, song.ID)
			newTopSongData := models.NewTopSongData(topSong.ID, "", i+1, term)
			spotify.OnNewEntityEvent(&newTopSongData)
			if exists {
				newTopSongData.SongID = dbSong.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach song ID for song %s", song.Name), logger.Warning)
			}

			topSongDataValues = append(topSongDataValues, utils.ReflectValues(newTopSongData)...)
		}
	}

	topSongValues = append(topSongValues, utils.ReflectValues(topSong)...)

	if len(topSongDataValues) == 0 {
		logger.Log("No top song data to ingest.", logger.Info)
		return nil
	}

	topSongRecordsToInsert := len(topSongValues) / len(utils.ReflectColumns(topSong))
	logger.Log("Inserting new top song record", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_song records", topSongRecordsToInsert), logger.Debug)
	err := spotify.Database.Create(&models.TopSong{}, topSongValues)
	if err != nil {
		return err
	}

	topSongDataRecordsToInsert := len(topSongDataValues) / len(utils.ReflectColumns(models.TopSongData{}))
	logger.Log("Inserting new top song data", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_song_data records", topSongDataRecordsToInsert), logger.Debug)
	err = spotify.Database.Create(&models.TopSongData{}, topSongDataValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *SpotifyIngest) InsertTopArtists(songs map[string]api.TopArtistsResponse, dbArtists []models.Artist) error {
	topArtistDataValues := []interface{}{}
	topArtistValues := []interface{}{}
	newTopArtist := models.NewTopArtist(spotify.Options.UserID)
	spotify.OnNewEntityEvent(&newTopArtist)

	for term, resp := range songs {
		for i, artist := range resp.Items {
			dbArtist, exists := getArtistBySpotifyID(dbArtists, artist.ID)
			newTopArtistData := models.NewTopArtistData(artist.Name, "", i+1, term, newTopArtist.ID)
			spotify.OnNewEntityEvent(&newTopArtistData)
			if exists {
				newTopArtistData.ArtistID = dbArtist.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach artist ID for artist %s", artist.Name), logger.Warning)
			}
			topArtistDataValues = append(topArtistDataValues, utils.ReflectValues(newTopArtistData)...)
		}
	}

	topArtistValues = append(topArtistValues, utils.ReflectValues(newTopArtist)...)
	if len(topArtistDataValues) == 0 {
		logger.Log("No top artist data to ingest.", logger.Info)
		return nil
	}

	logger.Log("Inserting new top artist record", logger.Info)
	topArtistRecords := len(topArtistValues) / len(utils.ReflectColumns(&models.TopArtist{}))
	logger.Log(fmt.Sprintf("Inserting %d top_artist records", topArtistRecords), logger.Debug)
	err := spotify.Database.Create(&models.TopArtist{}, topArtistValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new top artist data", logger.Info)
	topArtistDataRecords := len(topArtistDataValues) / len(utils.ReflectColumns(&models.TopArtistData{}))
	logger.Log(fmt.Sprintf("Inserting %d top_artist_data records", topArtistDataRecords), logger.Debug)
	err = spotify.Database.Create(&models.TopArtistData{}, topArtistDataValues)
	if err != nil {
		return err
	}

	return nil
}

// too many args should use struct tbh
func (spotify *SpotifyIngest) InsertThumbnails(songs map[string]api.TopTracksResponse, recents api.RecentlyPlayedResponse, artists map[string]api.TopArtistsResponse, dbArtists []models.Artist, dbAlbums []models.Album) error {
	thumbnails := make(map[string]models.Thumbnail)
	// TODO: normalize then iterate over one loop plss
	for _, key := range utils.MapOrderedKeys(songs) {
		for _, song := range songs[key].Items {
			dbAlbum, exists := getAlbumBySpotifyID(dbAlbums, song.Album.ID)
			if !exists {
				logger.Log(fmt.Sprintf("Failed to attach album ID for album %s", song.Album.Name), logger.Warning)
			}
			for _, image := range song.Album.Images {
				spotify.MetricHandler.AddNewThumbnailIndex("Album", song.Album.Name, image.URL)
				thumbnail := models.NewThumbnail("Album", "", image.URL, image.Height, image.Width)
				spotify.OnNewEntityEvent(&thumbnail)
				if exists {
					thumbnail.EntityID = dbAlbum.ID
				}
				thumbnails[thumbnail.UniqueID()] = thumbnail
			}
		}
	}

	for _, key := range utils.MapOrderedKeys(artists) {
		for _, artist := range artists[key].Items {
			dbArtist, exists := getArtistBySpotifyID(dbArtists, artist.ID)
			if !exists {
				logger.Log(fmt.Sprintf("Failed to attach artist ID for artist %s", artist.Name), logger.Warning)
			}
			for _, image := range artist.Images {
				spotify.MetricHandler.AddNewThumbnailIndex("Artist", artist.Name, image.URL)
				thumbnail := models.NewThumbnail("Artist", "", image.URL, image.Height, image.Width)
				spotify.OnNewEntityEvent(&thumbnail)
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
			spotify.MetricHandler.AddNewThumbnailIndex("Album", song.Track.Name, image.URL)
			thumbnail := models.NewThumbnail("Album", "", image.URL, image.Height, image.Width)
			spotify.OnNewEntityEvent(&thumbnail)
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
	for _, key := range utils.MapOrderedKeys(thumbnails) {
		thumbnail := thumbnails[key]
		_, exists := getThumbnailByEntityIDAndDimensions(dbThumbnails, thumbnail.EntityID, thumbnail.Height, thumbnail.Width)
		if !exists {
			thumbnailsToInsert = append(thumbnailsToInsert, utils.ReflectValues(thumbnail)...)
		}
	}

	if len(thumbnailsToInsert) == 0 {
		logger.Log("No new thumbnails to ingest", logger.Debug)
		return nil
	}

	logger.Log(fmt.Sprintf("Inserting %d new thumbnail records", len(thumbnailsToInsert)), logger.Debug)
	err = spotify.Database.Create(&models.Thumbnail{}, thumbnailsToInsert)
	if err != nil {
		return err
	}
	return nil
}

func (spotify *SpotifyIngest) Recents() (api.RecentlyPlayedResponse, error) {
	recentlyPlayed, err := spotify.API.RecentlyPlayedByUser()
	if err != nil {
		return api.RecentlyPlayedResponse{}, err
	}

	return recentlyPlayed, nil
}

func BootstrapSpotifyingest(database IngestDatabase, api API, args SpotifyIngestOptions, preingest *PreIngest, metricHandler MetricHandler) SpotifyIngest {
	me, err := api.Me()
	if err != nil {
		logger.Log("Failed to fetch Me endpoint", logger.Error)
		panic(err)
	}

	logger.Log("Handling base user data", logger.Info)
	userId, err := preingest.GetUserUUID(args.UserID, me)
	if err != nil {
		logger.Log("Failed when handling base user routine", logger.Error)
		panic(err)
	}

	logger.Log("Ensure base data exists", logger.Info)
	variousArtistsId, err := preingest.EnsureBaseDataExists()
	if err != nil {
		logger.Log("Failed when ensuring base data exists", logger.Error)
		panic(err)
	}

	options := SpotifyIngestOptions{
		RecentListen:       args.RecentListen,
		TopSongs:           args.TopSongs,
		TopArtists:         args.TopArtists,
		UserID:             userId,
		VariousArtistsUUID: variousArtistsId,
	}

	return NewSpotifyIngest(database, api, options, metricHandler)
}
