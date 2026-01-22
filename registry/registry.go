package registry

import (
	"net/http"
	"sync"
)

var (
	mu       sync.RWMutex
	handlers = make(map[string]http.HandlerFunc)
)

func Register(name string, h http.HandlerFunc) {
	mu.Lock()
	defer mu.Unlock()
	handlers[name] = h
}

func Get(name string) http.HandlerFunc {
	mu.RLock()
	defer mu.RUnlock()
	return handlers[name]
}
