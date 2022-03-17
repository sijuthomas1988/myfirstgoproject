package main

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"sync"
)

// URLStore common map
type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	file *os.File
}

// record object to store in file
type record struct {
	Key, Url string
}

func NewURLStore(filename string) *URLStore {
	s := &URLStore{urls: make(map[string]string)}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("Error opening UrlStore :", err)
	}
	s.file = f
	if err := s.Load(); err != nil {
		log.Println("Error loading url store", err)
	}
	return s
}

func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[key]
}

func (s *URLStore) Set(key, url string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, present := s.urls[key]; present {
		return false
	}
	s.urls[key] = url
	return true
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func (s *URLStore) Put(url string) string {
	for {
		key := genKey(s.Count())
		if ok := s.Set(key, url); ok {
			if err := s.Save(key, url); err != nil {
				log.Println("Error saving to UrlStore", err)
			}
			return key
		} else {
			return ""
		}
	}
}

func (s *URLStore) Load() error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}
	d := gob.NewDecoder(s.file)
	var err error
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.Url)
		}
	}
	if err == io.EOF {
		return nil
	}
	return err
}

func (s *URLStore) Save(key, url string) error {
	e := gob.NewEncoder(s.file)
	return e.Encode(record{key, url})
}
