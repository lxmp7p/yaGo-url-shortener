package logger

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
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
		client: retryablehttp.NewClient(),
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

func TestDispatcher_MultipleObservers(t *testing.T) {
	d := &Dispatcher{}

	m1 := &mockObserver{}
	m2 := &mockObserver{}

	d.Register(m1)
	d.Register(m2)

	event := AuditEvent{TS: 1, Action: "multi", URL: "/x"}

	d.Notify(event)

	if !m1.called || !m2.called {
		t.Fatal("expected all observers to be called")
	}
}

func TestDispatcher_Unregister_NonExisting(t *testing.T) {
	d := &Dispatcher{}

	m1 := &mockObserver{}
	m2 := &mockObserver{}

	d.Register(m1)
	d.Unregister(m2)
	if len(d.observers) != 1 {
		t.Fatal("unexpected observers slice mutation")
	}
}

func TestDispatcher_Close_FileObserver(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-close")
	if err != nil {
		t.Fatal(err)
	}

	f := &FileObserver{File: tmpFile}

	d := &Dispatcher{}
	d.Register(f)

	err = d.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmpFile.WriteString("x")
	if err == nil {
		t.Fatal("expected write to closed file to fail")
	}
}

func TestDispatcher_ConcurrentNotify(t *testing.T) {
	d := &Dispatcher{}

	m := &mockObserver{}
	d.Register(m)

	event := AuditEvent{TS: 1, Action: "concurrent", URL: "/x"}

	done := make(chan struct{})

	for i := 0; i < 50; i++ {
		go func() {
			d.Notify(event)
			done <- struct{}{}
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}

	if !m.called {
		t.Fatal("expected observer to be called under concurrency")
	}
}

func TestFileObserver_ConcurrentWrites(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-concurrent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	f := &FileObserver{File: tmpFile}

	event := AuditEvent{TS: 1, Action: "c", URL: "/x"}

	done := make(chan struct{})

	for i := 0; i < 30; i++ {
		go func() {
			f.Notify(event)
			done <- struct{}{}
		}()
	}

	for i := 0; i < 30; i++ {
		<-done
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("expected concurrent writes to persist data")
	}
}
