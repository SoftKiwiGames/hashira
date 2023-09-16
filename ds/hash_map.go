package ds

import "golang.org/x/exp/constraints"

type HashMap[K constraints.Ordered, V any] struct {
	data map[K]V
}

func NewHashMap[K constraints.Ordered, V any]() *HashMap[K, V] {
	return &HashMap[K, V]{
		data: make(map[K]V),
	}
}

func (h *HashMap[K, V]) Get(key K) V {
	return h.data[key]
}

func (h *HashMap[K, V]) Set(key K, value V) {
	h.data[key] = value
}

func (h *HashMap[K, V]) Delete(key K) {
	delete(h.data, key)
}

func (h *HashMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(h.data))
	for key := range h.data {
		keys = append(keys, key)
	}
	return keys
}

func (h *HashMap[K, V]) Values() []V {
	values := make([]V, 0, len(h.data))
	for _, value := range h.data {
		values = append(values, value)
	}
	return values
}

func (h *HashMap[K, V]) Len() int {
	return len(h.data)
}

func (h *HashMap[K, V]) Clear() {
	h.data = make(map[K]V)
}

func (h *HashMap[K, V]) Has(key K) bool {
	_, ok := h.data[key]
	return ok
}

func (h *HashMap[K, V]) ForEach(fn func(key K, value V)) {
	for key, value := range h.data {
		fn(key, value)
	}
}
