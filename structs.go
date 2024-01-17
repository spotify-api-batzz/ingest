package main

import "time"

type MeResponse struct {
	Country      string `json:"country"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href interface{} `json:"href"`
	} `json:"followers"`
	Href    string  `json:"href"`
	ID      string  `json:"id"`
	Images  []Image `json:"images"`
	Product string  `json:"product"`
	Type    string  `json:"type"`
	URI     string  `json:"uri"`
}

type RecentlyPlayedResponse struct {
	Items []struct {
		Track struct {
			Album struct {
				AlbumType string `json:"album_type"`
				Artists   []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"artists"`
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href                 string  `json:"href"`
				ID                   string  `json:"id"`
				Images               []Image `json:"images"`
				Name                 string  `json:"name"`
				ReleaseDate          string  `json:"release_date"`
				ReleaseDatePrecision string  `json:"release_date_precision"`
				TotalTracks          float64 `json:"total_tracks"`
				Type                 string  `json:"type"`
				URI                  string  `json:"uri"`
			} `json:"album"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       float64  `json:"disc_number"`
			DurationMs       float64  `json:"duration_ms"`
			Explicit         bool     `json:"explicit"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string  `json:"href"`
			ID          string  `json:"id"`
			IsLocal     bool    `json:"is_local"`
			Name        string  `json:"name"`
			Popularity  float64 `json:"popularity"`
			PreviewURL  string  `json:"preview_url"`
			TrackNumber float64 `json:"track_number"`
			Type        string  `json:"type"`
			URI         string  `json:"uri"`
		} `json:"track"`
		PlayedAt time.Time `json:"played_at"`
		Context  struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"context"`
	} `json:"items"`
	Next    string `json:"next"`
	Cursors struct {
		After  string `json:"after"`
		Before string `json:"before"`
	} `json:"cursors"`
	Limit float64 `json:"limit"`
	Href  string  `json:"href"`
}

type TopArtistsResponse struct {
	Items    []Artist    `json:"items"`
	Limit    float64     `json:"limit"`
	Offset   float64     `json:"offset"`
	Href     string      `json:"href"`
	Previous interface{} `json:"previous"`
	Next     string      `json:"next"`
}

type Artist struct {
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href interface{} `json:"href"`
	} `json:"followers"`
	Genres     []string `json:"genres"`
	Href       string   `json:"href"`
	ID         string   `json:"id"`
	Images     []Image  `json:"images"`
	Name       string   `json:"name"`
	Popularity float64  `json:"popularity"`
	Type       string   `json:"type"`
	URI        string   `json:"uri"`
}

type TopTracksResponse struct {
	Items []Song `json:"items"`

	Limit    float64     `json:"limit"`
	Offset   float64     `json:"offset"`
	Href     string      `json:"href"`
	Previous interface{} `json:"previous"`
	Next     string      `json:"next"`
}

type TracksResponse struct {
	Tracks []Song `json:"tracks"`
}

type ArtistsResponse struct {
	Artists []Artist `json:"artists"`
}

type Image struct {
	Height float64 `json:"height"`
	URL    string  `json:"url"`
	Width  float64 `json:"width"`
}

type Song struct {
	Album struct {
		AlbumType string `json:"album_type"`
		Artists   []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href                 string  `json:"href"`
		ID                   string  `json:"id"`
		Images               []Image `json:"images"`
		Name                 string  `json:"name"`
		ReleaseDate          string  `json:"release_date"`
		ReleaseDatePrecision string  `json:"release_date_precision"`
		TotalTracks          float64 `json:"total_tracks"`
		Type                 string  `json:"type"`
		URI                  string  `json:"uri"`
	} `json:"album"`
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`
	AvailableMarkets []string `json:"available_markets"`
	DiscNumber       float64  `json:"disc_number"`
	DurationMs       float64  `json:"duration_ms"`
	Explicit         bool     `json:"explicit"`
	ExternalIds      struct {
		Isrc string `json:"isrc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href        string  `json:"href"`
	ID          string  `json:"id"`
	IsLocal     bool    `json:"is_local"`
	Name        string  `json:"name"`
	Popularity  float64 `json:"popularity"`
	PreviewURL  string  `json:"preview_url"`
	TrackNumber float64 `json:"track_number"`
	Type        string  `json:"type"`
	URI         string  `json:"uri"`
}

type AlbumResponse struct {
	Albums []Album `json:"albums"`
}

type Album struct {
	AlbumType string `json:"album_type"`
	Artists   []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`
	AvailableMarkets []string `json:"available_markets"`
	Copyrights       []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"copyrights"`
	ExternalIds struct {
		Upc string `json:"upc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Genres               []interface{} `json:"genres"`
	Href                 string        `json:"href"`
	ID                   string        `json:"id"`
	Images               []Image       `json:"images"`
	Name                 string        `json:"name"`
	Popularity           float64       `json:"popularity"`
	ReleaseDate          string        `json:"release_date"`
	ReleaseDatePrecision string        `json:"release_date_precision"`
	Tracks               struct {
		Href  string `json:"href"`
		Items []struct {
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       float64  `json:"disc_number"`
			DurationMs       float64  `json:"duration_ms"`
			Explicit         bool     `json:"explicit"`
			ExternalUrls     struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string  `json:"href"`
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			PreviewURL  string  `json:"preview_url"`
			TrackNumber float64 `json:"track_number"`
			Type        string  `json:"type"`
			URI         string  `json:"uri"`
		} `json:"items"`
		Limit    float64     `json:"limit"`
		Next     interface{} `json:"next"`
		Offset   float64     `json:"offset"`
		Previous interface{} `json:"previous"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}
