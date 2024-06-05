package main

import (
	"fmt"
	"spotify/types"

	"github.com/batzz-00/goutils/logger"
)

type Refreshable interface {
	Refresh() error
}

func Refresh(refreshable types.API) error {
	opts := refreshable.Options()
	logger.Log("Refreshing user access token", logger.Info)
	err := refresh(refreshable, 0, opts.RefreshRetries)
	if err != nil {
		return &ErrRefresh{Tries: opts.RefreshRetries}
	}
	return nil
}

func refresh(refreshable Refreshable, tries int, maxTries int) error {
	err := refreshable.Refresh()
	if err != nil {
		if maxTries != 0 && tries < maxTries {
			logger.Log(fmt.Sprintf("Failed to refresh API token, tried %d/%d times.", tries+1, maxTries), logger.Warning)
			return refresh(refreshable, tries+1, maxTries)
		}
		return err
	}
	return nil
}
