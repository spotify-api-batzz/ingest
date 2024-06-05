package ingest

import "spotify/models"

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
