package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.Status = 200
	return r.ResponseWriter.(http.Hijacker).Hijack()
}
func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

type logger struct {
	instance *instance
	handler  http.Handler
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	recorder := &statusRecorder{
		ResponseWriter: w,
	}
	defer func() {
		l.instance.write(os.Stdout, fmt.Sprintf("%s %s %d %s", r.Method, r.URL.Path, recorder.Status, time.Since(start)))
	}()
	l.handler.ServeHTTP(recorder, r)
}
