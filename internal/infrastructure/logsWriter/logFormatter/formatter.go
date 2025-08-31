package logFormatter

import (
	"encoding/json"
	"gift-buyer/internal/infrastructure/logsWriter/logTypes"
	"time"
)

type logFormatterImpl struct {
	level string
}

func NewLogFormatter(level string) *logFormatterImpl {
	return &logFormatterImpl{
		level: level,
	}
}

func (l *logFormatterImpl) Format(entry *logTypes.LogEntry) ([]byte, error) {
	entryCopy := *entry
	entryCopy.Timestamp = time.Now().Format(time.RFC3339)

	if entryCopy.Level == "" {
		entryCopy.Level = l.level
	}

	jsonBytes, err := json.Marshal(entryCopy)
	if err != nil {
		return nil, err
	}

	jsonBytes = append(jsonBytes, '\n')
	return jsonBytes, nil
}
