package valheim

import "survival-bot/internal/games"

func init() {
	games.Register(games.Valheim, func() games.Parser { return New() })
}
