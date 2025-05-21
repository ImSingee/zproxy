package main

import (
	"fmt"
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

// Set updates the entire DNS map and prints the changes
func (s *ZeaburDnsStore) Set(newMap map[string]string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	const prefix = "[zeabur-dns]"

	// Compare old and new maps to print changes
	// Print removals for keys that are in old map but not in new map
	for key, oldValue := range s.dnsMap {
		if newValue, exists := newMap[key]; !exists {
			// Key was removed
			fmt.Printf("%s - %s \t -> \t %s\n", prefix, key, oldValue)
		} else if oldValue != newValue {
			// Key exists but value changed - print removal first
			fmt.Printf("%s - %s \t -> \t %s\n", prefix, key, oldValue)
			fmt.Printf("%s + %s \t -> \t %s\n", prefix, key, newValue)
		}
	}

	// Print additions for keys that are in new map but not in old map
	for key, newValue := range newMap {
		if _, exists := s.dnsMap[key]; !exists {
			// New key was added
			fmt.Printf("%s + %s \t -> \t %s\n", prefix, key, newValue)
		}
	}

	// Update the map
	s.dnsMap = newMap
}

// Size returns the number of entries in the DNS map
func (s *ZeaburDnsStore) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.dnsMap)
}
