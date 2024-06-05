package mocks

import (
	"database/sql/driver"
	"reflect"
	"spotify/models"
)

type MockDatabase struct {
	SavedValues map[string][]interface{}
}

func NewMockDatabase() MockDatabase {
	return MockDatabase{
		SavedValues: make(map[string][]interface{}),
	}
}

func (db *MockDatabase) Create(model models.Model, values []interface{}) error {
	driverValues := make([]interface{}, len(values))
	interfaceType := reflect.TypeOf((*driver.Valuer)(nil)).Elem()

	for i, value := range values {
		canConvert := reflect.TypeOf(value).Implements(interfaceType)
		if canConvert {
			valuer := value.(driver.Valuer)
			v, _ := valuer.Value()
			driverValues[i] = v
		} else {
			driverValues[i] = value
		}
	}

	// filename := fmt.Sprintf("%s-insert.json", reflect.TypeOf(model).Elem().Name())
	// fmt.Println(filename)
	// wd, _ := os.Getwd()
	// path := path.Join(wd, filename)
	// bytes, _ := json.Marshal(driverValues)
	// os.WriteFile(path, bytes, 0644)

	db.SavedValues[reflect.TypeOf(model).Elem().Name()] = driverValues
	return nil
}

func (db *MockDatabase) FetchSongsBySpotifyID(spotifyIDs []interface{}) ([]models.Song, error) {
	return nil, nil
}

func (db *MockDatabase) FetchUsersBySpotifyIds(names []interface{}) ([]models.User, error) {
	return nil, nil
}

func (db *MockDatabase) FetchUserByName(name string) (models.User, error) {
	return models.User{}, nil
}

func (db *MockDatabase) FetchArtistBySpotifyID(spotifyID string) (models.Artist, error) {
	return models.Artist{}, nil
}

func (db *MockDatabase) FetchAlbumsBySpotifyID(spotifyIDs []interface{}) ([]models.Album, error) {
	return nil, nil
}

func (db *MockDatabase) FetchArtistsBySpotifyID(spotifyIDs []interface{}) ([]models.Artist, error) {
	return nil, nil
}

func (db *MockDatabase) FetchArtistByID(id string) (models.Artist, error) {
	return models.Artist{}, nil
}

func (db *MockDatabase) FetchRecentListensByUserIDAndTime(userID string, recentListenedToIDs []interface{}, earliestTime interface{}) ([]models.RecentListen, error) {
	return nil, nil
}

func (db *MockDatabase) FetchThumbnailsByEntityID(entityIDs []interface{}) ([]models.Thumbnail, error) {
	return nil, nil
}
