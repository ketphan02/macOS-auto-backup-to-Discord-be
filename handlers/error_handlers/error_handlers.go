package error_handlers

import (
	"log"
	"net/http"
)

type LogWriter struct {
	http.ResponseWriter
}

func (w LogWriter) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
	return
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("405 Method Not Allowed"))
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}
