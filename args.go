package main

import (
	"flag"
	"log"
	"spotify/ingest"
)

// interface Args {

// }

func parseArgs() ingest.SpotifyIngestOptions {
	recentListen := flag.Bool("r", false, "Parse and ingest user data regarding a users recently listened tracks")
	topSongs := flag.Bool("t", false, "Parse and ingest user data regarding a users top songs")
	topArtists := flag.Bool("a", false, "Parse and ingest user data regarding a users top artists")
	user := flag.String("u", "", "Username to query the spotify API for, must have relevant refresh_token in env")
	flag.Parse()

	if *user == "" {
		log.Fatalf("UserID must be specified!")
	}

	return ingest.SpotifyIngestOptions{
		RecentListen: *recentListen,
		TopSongs:     *topSongs,
		TopArtists:   *topArtists,
		UserID:       *user,
	}
}
