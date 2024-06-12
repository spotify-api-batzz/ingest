package ingest

import (
	"fmt"
	"spotify/api"
	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"
)

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
