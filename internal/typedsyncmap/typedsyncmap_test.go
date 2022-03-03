package typedsyncmap_test

import (
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/eventbus/internal/typedsyncmap"
)

func TestStore(t *testing.T) {
	ensure := ensure.New(t)

	m := typedsyncmap.Map[string, int]{}

	m.Store("hello", 123)

	v, ok := m.Load("hello")
	ensure(v).Equals(123)
	ensure(ok).IsTrue()
}

func TestLoad(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("when value is present", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		v, ok := m.Load("hello")
		ensure(v).Equals(123)
		ensure(ok).IsTrue()
	})

	ensure.Run("when value is not present", func(ensure ensurepkg.Ensure) {
		ensure.Run("returns 0 when the value type is int", func(ensure ensurepkg.Ensure) {
			m := typedsyncmap.Map[string, int]{}

			m.Store("hello", 123)
			m.Store("world", 456)

			v, ok := m.Load("another")
			ensure(v).Equals(0)
			ensure(ok).IsFalse()
		})

		ensure.Run("returns empty string when the value type is string", func(ensure ensurepkg.Ensure) {
			m := typedsyncmap.Map[int, string]{}

			m.Store(123, "hello")
			m.Store(456, "world")

			v, ok := m.Load(404)
			ensure(v).Equals("")
			ensure(ok).IsFalse()
		})

		ensure.Run("returns nil when the value type is a pointer", func(ensure ensurepkg.Ensure) {
			type User struct{ID int}
			m := typedsyncmap.Map[int, *User]{}

			m.Store(123, &User{123})
			m.Store(456, &User{456})

			v, ok := m.Load(404)
			ensure(v).Equals(nil)
			ensure(ok).IsFalse()
		})
	})
}

func TestDelete(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("when value is present", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		v1, ok := m.Load("hello")
		ensure(v1).Equals(123)
		ensure(ok).IsTrue()

		m.Delete("hello")

		v2, ok := m.Load("hello")
		ensure(v2).Equals(0)
		ensure(ok).IsFalse()

		v3, ok := m.Load("world")
		ensure(v3).Equals(456)
		ensure(ok).IsTrue()
	})

	ensure.Run("when value is not present", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		m.Delete("another")

		v1, ok := m.Load("another")
		ensure(v1).Equals(0)
		ensure(ok).IsFalse()

		v2, ok := m.Load("hello")
		ensure(v2).Equals(123)
		ensure(ok).IsTrue()

		v3, ok := m.Load("world")
		ensure(v3).Equals(456)
		ensure(ok).IsTrue()
	})
}

func TestLoadAndDelete(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("when value is present", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		v1, ok := m.LoadAndDelete("hello")
		ensure(v1).Equals(123)
		ensure(ok).IsTrue()

		v2, ok := m.Load("hello")
		ensure(v2).Equals(0)
		ensure(ok).IsFalse()
	})

	ensure.Run("when value is not present", func(ensure ensurepkg.Ensure) {
		ensure.Run("returns 0 when the value type is int", func(ensure ensurepkg.Ensure) {
			m := typedsyncmap.Map[string, int]{}

			m.Store("hello", 123)
			m.Store("world", 456)

			v, ok := m.LoadAndDelete("another")
			ensure(v).Equals(0)
			ensure(ok).IsFalse()
		})

		ensure.Run("returns empty string when the value type is string", func(ensure ensurepkg.Ensure) {
			m := typedsyncmap.Map[int, string]{}

			m.Store(123, "hello")
			m.Store(456, "world")

			v, ok := m.LoadAndDelete(404)
			ensure(v).Equals("")
			ensure(ok).IsFalse()
		})

		ensure.Run("returns nil when the value type is a pointer", func(ensure ensurepkg.Ensure) {
			type User struct{ID int}
			m := typedsyncmap.Map[int, *User]{}

			m.Store(123, &User{123})
			m.Store(456, &User{456})

			v, ok := m.LoadAndDelete(404)
			ensure(v).Equals(nil)
			ensure(ok).IsFalse()
		})
	})
}

func TestLoadOrStore(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("when value is present", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		v1, ok := m.LoadOrStore("hello", 789)
		ensure(v1).Equals(123)
		ensure(ok).IsTrue()

		v2, ok := m.Load("hello")
		ensure(v2).Equals(123)
		ensure(ok).IsTrue()
	})

	ensure.Run("when value is not present", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		v1, ok := m.LoadOrStore("another", 789)
		ensure(v1).Equals(789)
		ensure(ok).IsFalse()

		v2, ok := m.Load("another")
		ensure(v2).Equals(789)
		ensure(ok).IsTrue()
	})
}

func TestRange(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("when empty", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		visited := map[string]int{}

		m.Range(func(key string, value int) bool {
			visited[key] = value
			return true
		})

		ensure(visited).IsEmpty()
	})

	ensure.Run("when non-empty", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		visited := map[string]int{}

		m.Range(func(key string, value int) bool {
			visited[key] = value
			return true
		})

		ensure(visited).Equals(map[string]int{
			"hello": 123,
			"world": 456,
		})
	})

	ensure.Run("when false is returned from the function", func(ensure ensurepkg.Ensure) {
		m := typedsyncmap.Map[string, int]{}

		m.Store("hello", 123)
		m.Store("world", 456)

		visited := map[string]int{}

		m.Range(func(key string, value int) bool {
			visited[key] = value
			return false
		})

		ensure(len(visited)).Equals(1)
	})
}
