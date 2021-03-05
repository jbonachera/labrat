package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

var (
	BuiltVersion string = "snapshot"
)

type responseMessage struct {
	InstanceID string    `json:"instance_id"`
	StartedAt  time.Time `json:"started_at"`
	Timestamp  time.Time `json:"timestamp"`
	Version    string    `json:"version"`
	Message    string    `json:"message"`
}

type instanceConfig struct {
	slashCode int
}

func defaultConfig() *instanceConfig {
	return &instanceConfig{
		slashCode: http.StatusOK,
	}
}

type instance struct {
	id        string
	config    *instanceConfig
	startedAt time.Time
	httpc     *http.Client
}

func newInstance() *instance {
	c := http.DefaultClient
	c.Timeout = 3 * time.Second
	return &instance{
		id:        uuid.New().String(),
		config:    defaultConfig(),
		startedAt: time.Now(),
		httpc:     c,
	}
}

func (i *instance) write(w io.Writer, message string) {
	json.NewEncoder(w).Encode(responseMessage{
		Message:    message,
		InstanceID: i.id,
		Timestamp:  time.Now(),
		StartedAt:  i.startedAt,
		Version:    BuiltVersion,
	})
}

func main() {
	instance := newInstance()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		panic(err)
	}
	srv := &http.Server{}

	http.Handle("/proxy/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			instance.write(w, "Usage: POST {\"url\": \"url_to_get\"}")
			return
		}
		req := map[string]interface{}{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req["url"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			instance.write(w, "failed to decode json")
			return
		}
		resp, err := instance.httpc.Get(req["url"].(string))
		if err != nil || req["url"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			instance.write(w, err.Error())
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}))
	http.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.WriteHeader(instance.config.slashCode)
		instance.write(w, "hello.")
	}))
	http.Handle("/closeListener", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		instance.write(w, "listener will close.")
		go func() {
			srv.Shutdown(context.Background())
			instance.write(os.Stdout, "http client connections closed.")
		}()
	}))
	http.Handle("/toggleHealth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		switch instance.config.slashCode {
		case http.StatusOK:
			instance.config.slashCode = http.StatusInternalServerError
		case http.StatusInternalServerError:
			instance.config.slashCode = http.StatusOK
		}
		instance.write(w, fmt.Sprintf("/health will now return http status code %d", instance.config.slashCode))
	}))

	http.Handle("/panic", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		instance.write(w, "panic requested")
		go os.Exit(1)
	}))
	http.Handle("/err500", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		instance.write(w, "This response is an error 500, as requested.")
	}))

	type proxyPayload struct {
		Method  string `json:"method"`
		URL     string `json:"url"`
		Payload string `json:"payload"`
	}

	http.Handle("/", http.FileServer(http.Dir("/usr/share/www/")))

	go func() {
		srv.Handler = &logger{instance: instance, handler: http.DefaultServeMux}
		instance.write(os.Stdout, fmt.Sprintf("http listener running on port %s", port))
		srv.Serve(ln)
		instance.write(os.Stdout, "listener stopped.")
	}()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
	instance.write(os.Stdout, "exited.")
}
