package soulmask

import "survival-bot/internal/games"

func init() {
	games.Register(games.Soulmask, func() games.Parser { return New() })
}
