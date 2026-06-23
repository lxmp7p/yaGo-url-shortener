package pool

import "testing"

type testObject struct {
	Value       int
	ResetCalled bool
}

func (t *testObject) Reset() {
	t.ResetCalled = true
	t.Value = 0
}

func TestPoolGetCreatesObject(t *testing.T) {
	calls := 0

	p, err := New(func() *testObject {
		calls++
		return &testObject{Value: 42}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := p.Get()

	if obj == nil {
		t.Fatal("expected object")
	}

	if obj.Value != 42 {
		t.Fatalf("expected Value=42, got %d", obj.Value)
	}

	if calls != 1 {
		t.Fatalf("factory called %d times, want 1", calls)
	}
}

func TestPoolPutCallsReset(t *testing.T) {
	p, err := New(func() *testObject {
		return &testObject{Value: 100}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := p.Get()

	obj.Value = 999

	p.Put(obj)

	if !obj.ResetCalled {
		t.Fatal("Reset was not called")
	}

	if obj.Value != 0 {
		t.Fatalf("expected Value=0 after Reset, got %d", obj.Value)
	}
}

func TestPoolReusesObject(t *testing.T) {
	p, err := New(func() *testObject {
		return &testObject{}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj1 := p.Get()
	obj1.Value = 10

	p.Put(obj1)

	obj2 := p.Get()

	if obj1 != obj2 {
		t.Skip("sync.Pool is allowed to return another object")
	}

	if obj2.Value != 0 {
		t.Fatalf("expected reset value, got %d", obj2.Value)
	}
}

func TestNilFactory(t *testing.T) {
	p, err := New[*testObject](nil)

	if err == nil {
		t.Fatal("expected error")
	}

	if p != nil {
		t.Fatal("expected nil pool")
	}
}
