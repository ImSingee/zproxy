package main

import (
	"sync"
)

// ZeaburDnsStore provides thread-safe access to the Zeabur DNS mapping
type ZeaburDnsStore struct {
	dnsMap map[string]string
	lock   sync.RWMutex
}

// NewZeaburDnsStore creates a new ZeaburDnsStore instance
func NewZeaburDnsStore() *ZeaburDnsStore {
	return &ZeaburDnsStore{
		dnsMap: make(map[string]string),
	}
}

// Get retrieves a value from the DNS map by key
func (s *ZeaburDnsStore) Get(key string) (string, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, exists := s.dnsMap[key]
	return value, exists
}

// Set updates the entire DNS map
func (s *ZeaburDnsStore) Set(newMap map[string]string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.dnsMap = newMap
}

// Size returns the number of entries in the DNS map
func (s *ZeaburDnsStore) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.dnsMap)
}