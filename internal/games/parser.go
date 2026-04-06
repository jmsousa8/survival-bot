package games

import "survival-bot/internal/events"

type Parser interface {
	ParseLine(line string) events.Event
}

var parsers = make(map[string]func() Parser)

func Register(name string, factory func() Parser) {
	parsers[name] = factory
}

func GetParser(name string) (Parser, bool) {
	factory, ok := parsers[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}
