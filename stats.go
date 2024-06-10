package main

import "spotify/models"

// func
// OnNewEntity *func(model models.Model)
// OnFinish    *func(stats SpotifyIngestStats)

type StatsHandler struct {
	entities map[string]int
}

func NewStatsHandler() StatsHandler {
	return StatsHandler{
		entities: make(map[string]int),
	}
}

func (s *StatsHandler) OnNewEntity(model models.Model) {
	s.entities[model.TableName()] += 1
}
