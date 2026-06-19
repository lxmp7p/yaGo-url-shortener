package pool

import "sync"

// Resettable defines the interface for objects that can be reset.
type Resettable interface {
	Reset()
}

// Pool is a generic object pool for types implementing Resettable.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New creates a new Pool. It requires a factory function to create new instances when the pool is empty.
func New[T Resettable](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

// Get retrieves an object from the pool or creates a new one if the pool is empty.
func (p *Pool[T]) Get() T {
	item := p.pool.Get().(T)
	return item
}

// Put returns an object to the pool after resetting it.
func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.pool.Put(item)
}
