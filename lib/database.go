package lib

import (
	"errors"
	"sync"
)

type Database struct {
	lock    *sync.RWMutex
	entries map[string]*MessageList
}

func NewDatabase() Database {
	return Database{lock: &sync.RWMutex{}, entries: make(map[string]*MessageList)}
}

func (db *Database) PossessRumorMessage(msg *RumorMessage) bool {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, ok := db.entries[msg.Origin]; ok {
		return entry.Possess(msg.ID)
	}
	return false
}

func (db *Database) InsertRumorMessage(msg *RumorMessage) {
	db.lock.Lock()
	defer db.lock.Unlock()
	if msg.Text == "" {
		panic(errors.New("oupsi"))
	}
	if entry, ok := db.entries[msg.Origin]; ok {
		entry.Insert(msg.ID, msg.Text)
	} else {
		entry = NewMessageList()
		entry.Insert(msg.ID, msg.Text)
		db.entries[msg.Origin] = entry
	}
}

func (db *Database) GetMessageContent(name string, id uint32) string {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return db.entries[name].Get(id)
}

func (db *Database) GetMinNotPresent(name string) uint32 {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if _, ok := db.entries[name]; !ok {
		return uint32(1)
	}

	return db.entries[name].NextId()
}

func (db *Database) GetPeerStatus() []PeerStatus {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var status []PeerStatus

	for name, entry := range db.entries {
		next := entry.NextId()
		status = append(status, PeerStatus{Identifier: name, NextID: next})
	}
	return status
}
