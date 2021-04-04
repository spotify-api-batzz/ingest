package main

import "fmt"

type ErrRefresh struct {
	Tries int
}

func (r *ErrRefresh) Error() string {
	return fmt.Sprintf("API request failed when attempting to reset token, tried %d times", r.Tries)
}
