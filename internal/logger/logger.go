package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
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
	Path string
	mu   sync.Mutex
}

type HTTPObserver struct {
	URL    string
	Client *http.Client
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
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(string(data) + "\n")
}

func (h *HTTPObserver) Notify(event AuditEvent) {
	data, _ := json.Marshal(event)

	req, _ := http.NewRequest("POST", h.URL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
