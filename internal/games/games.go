package games

import "slices"

const (
	Soulmask = "soulmask"
	Valheim  = "valheim"
)

func IsValid(gameInput string) bool {
	return slices.Contains([]string{Soulmask, Valheim}, gameInput)
}
