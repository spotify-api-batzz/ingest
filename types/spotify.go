package types

type API interface {
	Authorize(code string) error
	Options() APIOptions
	Refresh() error
}

type APIOptions struct {
	RefreshRetries int
}

type SpotifyAPI interface {
	Me() (MeResponse, error)
	RecentlyPlayedByUser() (RecentlyPlayedResponse, error)
	TopArtistsForUser(period string) (TopArtistsResponse, error)
	TopTracksForUser(period string) (TopTracksResponse, error)
	TracksBySpotifyID(ids []string) ([]Song, error)
	ArtistsBySpotifyID(ids []string) ([]Artist, error)
	AlbumsBySpotifyID(ids []string) ([]Album, error)
}

type IPreIngest interface {
	EnsureBaseDataExists() (string, error)
	GetUserUUID(usernameToReturn string, user MeResponse) (string, error)
}
