package types

type IMetricHandler interface {
	AddApiRequestIndex(method string, url string, reqBody string) error
	AddNewSongIndex(spotifyId string, songName string) error
	Close() error
}
