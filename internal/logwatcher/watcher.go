package logwatcher

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"survival-bot/internal/events"
	"survival-bot/internal/games"
)

type Watcher struct {
	file   *os.File
	reader *bufio.Reader
	parser games.Parser
}

type IWatcher interface {
	ReadNewLines() ([]events.Event, error)
	Close() error
}

func NewWatcher(filepath string, gameName string) (IWatcher, error) {
	parser, ok := games.GetParser(gameName)
	if !ok {
		return nil, fmt.Errorf("unsupported game: %s", gameName)
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to seek to end of log file: %w", err)
	}

	return &Watcher{
		file:   file,
		reader: bufio.NewReader(file),
		parser: parser,
	}, nil
}

func (lw *Watcher) ReadNewLines() ([]events.Event, error) {
	var newEvents []events.Event

	for {
		line, err := lw.reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSuffix(line, "\r\n")
		line = strings.TrimSuffix(line, "\n")

		if line == "" {
			continue
		}

		event := lw.parser.ParseLine(line)
		if event != nil {
			newEvents = append(newEvents, event)
		}
	}

	return newEvents, nil
}

func (lw *Watcher) Close() error {
	if lw.file != nil {
		return lw.file.Close()
	}
	return nil
}
