package typedsyncmap

import "sync"

// Map is a typed version of sync.Map.
// All methods behave the same way as those on sync.Map.
// See: https://pkg.go.dev/sync#Map
type Map[K, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	rawValue, ok := m.m.Load(key)
	if !ok {
		return value, false
	}

	return rawValue.(V), true
}

func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	rawValue, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return value, false
	}

	return rawValue.(V), true
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	rawActual, loaded := m.m.LoadOrStore(key, value)
	if !loaded {
		return value, false
	}

	return rawActual.(V), true
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(rawKey, rawValue any) bool {
		key := rawKey.(K)
		value := rawValue.(V)

		return f(key, value)
	})
}
