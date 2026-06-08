package logger

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type mockObserver struct {
	called bool
	event  AuditEvent
}

func (m *mockObserver) Notify(event AuditEvent) {
	m.called = true
	m.event = event
}

func TestDispatcher_Notify(t *testing.T) {
	d := &Dispatcher{}
	mock := &mockObserver{}
	d.Register(mock)

	event := AuditEvent{
		TS:     1,
		Action: "vsem",
		UserID: "privet",
		URL:    "/practicum",
	}

	d.Notify(event)
	if !mock.called {
		t.Fatal("expected observer to be called")
	}
	if mock.event.Action != "vsem" {
		t.Fatal("wrong event data")
	}
}

func TestDispatcher_Unregister(t *testing.T) {
	d := &Dispatcher{}
	mock := &mockObserver{}
	d.Register(mock)
	d.Unregister(mock)

	if len(d.observers) != 0 {
		t.Fatal("expected observer to be removed")
	}
}

func TestFileObserver_Notify(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	f := &FileObserver{
		File: tmpFile,
	}

	event := AuditEvent{
		TS:     1,
		Action: "create",
		UserID: "u1",
		URL:    "/x",
	}
	f.Notify(event)

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected file to contain data")
	}
}

func TestHTTPObserver_Notify(t *testing.T) {
	var called bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := &HTTPObserver{
		url:    server.URL,
		client: server.Client(),
	}

	event := AuditEvent{
		TS:     1,
		Action: "test",
		URL:    "/x",
	}

	h.Notify(event)

	if !called {
		t.Fatal("expected HTTP call")
	}
}

func TestDispatcher_NilSafe(t *testing.T) {
	var d *Dispatcher

	d.Notify(AuditEvent{
		TS:     1,
		Action: "x",
		URL:    "/",
	})
}
