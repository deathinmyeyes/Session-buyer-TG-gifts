package logWriterInterface

import "gift-buyer/internal/infrastructure/logsWriter/logTypes"

type LogFormatter interface {
	Format(entry *logTypes.LogEntry) ([]byte, error)
}

type LogWriter interface {
	WriteToFile(entry *logTypes.LogEntry) (err error)
}
