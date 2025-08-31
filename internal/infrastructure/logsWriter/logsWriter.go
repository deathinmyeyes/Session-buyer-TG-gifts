package logsWriter

import (
	"fmt"
	"gift-buyer/internal/infrastructure/logsWriter/logTypes"
	"gift-buyer/internal/infrastructure/logsWriter/logWriterInterface"
	"gift-buyer/pkg/logger"
)

type logsWriterImpl struct {
	writer  logWriterInterface.LogWriter
	logFlag bool
}

func NewLogger(
	writer logWriterInterface.LogWriter,
	logFlag bool,
) *logsWriterImpl {
	return &logsWriterImpl{
		writer:  writer,
		logFlag: logFlag,
	}
}

func (l *logsWriterImpl) LogInfo(message string) {
	l.writer.WriteToFile(&logTypes.LogEntry{
		Message: message,
	})
	l.logInfoToTerminal(message)
}

func (l *logsWriterImpl) LogError(message string) {
	l.writer.WriteToFile(&logTypes.LogEntry{
		Message: message,
	})
	l.logErrorToTerminal(message)
}

func (l *logsWriterImpl) LogErrorf(format string, args ...interface{}) {
	l.LogError(fmt.Sprintf(format, args...))
	l.logErrorToTerminal(fmt.Sprintf(format, args...))
}

func (l *logsWriterImpl) logInfoToTerminal(message string) {
	if l.logFlag {
		logger.GlobalLogger.Infof(message)
	}
}

func (l *logsWriterImpl) logErrorToTerminal(message string) {
	if l.logFlag {
		logger.GlobalLogger.Errorf(message)
	}
}
