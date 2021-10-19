package logging

import (
	"log"
	"net/http"
)

type Logger interface {
	Debug(interface{}, *http.Request)
	Info(interface{}, *http.Request)
	Warning(interface{}, *http.Request)
	Error(interface{}, *http.Request)
	Close()
}

type DefaultLogger struct {
}

func (DefaultLogger) Debug(payload interface{}, r *http.Request) {
	log.Printf("DEBUG %v\n", payload)
}

func (DefaultLogger) Info(payload interface{}, r *http.Request) {
	log.Printf("INFO %v\n", payload)
}

func (DefaultLogger) Warning(payload interface{}, r *http.Request) {
	log.Printf("WARN %v\n", payload)
}

func (DefaultLogger) Error(payload interface{}, r *http.Request) {
	log.Printf("ERROR %v\n", payload)
}

func (DefaultLogger) Close() {}

var logger Logger

func SetLogger(l Logger) {
	logger = l
}

func Debug(payload interface{}, r *http.Request) {
	logger.Debug(payload, r)
}

func Info(payload interface{}, r *http.Request) {
	logger.Info(payload, r)
}

func Warning(payload interface{}, r *http.Request) {
	logger.Warning(payload, r)
}

func Error(payload interface{}, r *http.Request) {
	logger.Error(payload, r)
}
