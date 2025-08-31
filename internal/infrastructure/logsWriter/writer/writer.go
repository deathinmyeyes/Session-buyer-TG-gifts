package writer

import (
	"errors"
	"fmt"
	"gift-buyer/internal/infrastructure/logsWriter/logTypes"
	"gift-buyer/internal/infrastructure/logsWriter/logWriterInterface"
	"gift-buyer/pkg/logger"
	"os"
	"sync"
)

type writerImpl struct {
	File      *os.File
	mu        sync.Mutex
	level     string
	formatter logWriterInterface.LogFormatter
}

func NewLogsWriter(level string, formatter logWriterInterface.LogFormatter) *writerImpl {
	file, err := os.OpenFile(fmt.Sprintf("%s_logs.jsonl", level), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.GlobalLogger.Fatalf("Failed to open log file: %v", err)
	}

	writer := &writerImpl{
		File:      file,
		level:     level,
		formatter: formatter,
	}

	return writer
}

func (l *writerImpl) WriteToFile(entry *logTypes.LogEntry) (err error) {
	bytes, err := l.formatter.Format(entry)
	if err != nil {
		return errors.New("failed to marshal to json")
	}

	if err := l.write(bytes); err != nil {
		return errors.New("failed to write logs to file")
	}

	return nil
}

func (l *writerImpl) write(bytes []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.File.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
