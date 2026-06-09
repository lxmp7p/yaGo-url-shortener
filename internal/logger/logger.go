package logger

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
)

type AuditEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

type Observer interface {
	Notify(event AuditEvent)
}

type Dispatcher struct {
	mu        sync.RWMutex
	observers []Observer
}

type FileObserver struct {
	File *os.File
	mu   sync.Mutex
}

type HTTPObserver struct {
	url    string
	client *retryablehttp.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5

	return &HTTPObserver{
		url:    url,
		client: retryClient,
	}
}

func NewDispatcher(cfg config.Config) (*Dispatcher, error) {
	dispatcher := &Dispatcher{}

	if cfg.FileLogging.Enable {
		file, err := os.OpenFile(cfg.FileLogging.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return dispatcher, err
		}
		dispatcher.Register(&FileObserver{File: file})
	}

	if cfg.RemoteLogging.Enable {
		dispatcher.Register(NewHTTPObserver(cfg.RemoteLogging.Path))
	}

	return dispatcher, nil
}

func (d *Dispatcher) Register(o Observer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.observers = append(d.observers, o)
}

func (d *Dispatcher) Unregister(target Observer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, o := range d.observers {
		if o == target {
			d.observers = append(d.observers[:i], d.observers[i+1:]...)
			break
		}
	}
}

func (d *Dispatcher) Notify(event AuditEvent) {
	if d == nil {
		return
	}

	d.mu.RLock()
	observers := append([]Observer(nil), d.observers...)
	d.mu.RUnlock()

	for _, o := range observers {
		o.Notify(event)
	}
}

func (f *FileObserver) Notify(event AuditEvent) {
	data, _ := json.Marshal(event)

	f.mu.Lock()
	f.File.WriteString(string(data) + "\n")
	f.mu.Unlock()
}

func (h *HTTPObserver) Notify(event AuditEvent) {
	data, _ := json.Marshal(event)

	req, _ := retryablehttp.NewRequest(
		"POST",
		h.url,
		data,
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func (f *FileObserver) Close() error {
	return f.File.Close()
}

func (d *Dispatcher) Close() error {
	var firstErr error

	for _, obs := range d.observers {
		if c, ok := obs.(io.Closer); ok {
			if err := c.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
