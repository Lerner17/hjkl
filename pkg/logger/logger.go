package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	file          *os.File
	warningLogger *log.Logger
	infoLogger    *log.Logger
	errorLogger   *log.Logger
}

func (l Logger) Warn(msg string) {
	l.warningLogger.Println(msg)
}

func (l Logger) Info(msg string) {
	l.infoLogger.Println(msg)
}

func (l Logger) Error(msg string) {
	l.errorLogger.Println(msg)
}

func (l Logger) Close() error {
	l.Info("---CLOSE---")
	return l.file.Close()
}

func New(path string) (Logger, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return Logger{}, fmt.Errorf("Cannot open logs file: %v", err)
	}

	return Logger{
		file:          file,
		infoLogger:    log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:   log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}
