package types

import "spotify/models"

type IDatabase interface {
	Create(model models.Model, values []interface{}) error
	FetchSongsBySpotifyID(spotifyIDs []interface{}) ([]models.Song, error)
	FetchUsersBySpotifyIds(names []interface{}) ([]models.User, error)
	FetchUserByName(name string) (models.User, error)
	FetchArtistBySpotifyID(spotifyID string) (models.Artist, error)
	FetchAlbumsBySpotifyID(spotifyIDs []interface{}) ([]models.Album, error)
	FetchArtistsBySpotifyID(spotifyIDs []interface{}) ([]models.Artist, error)
	FetchArtistByID(id string) (models.Artist, error)
	FetchRecentListensByUserIDAndTime(userID string, recentListenedToIDs []interface{}, earliestTime interface{}) ([]models.RecentListen, error)
	FetchThumbnailsByEntityID(entityIDs []interface{}) ([]models.Thumbnail, error)
}
