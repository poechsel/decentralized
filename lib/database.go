package lib

import (
	"sync"
)

type Database struct {
	lock    *sync.RWMutex
	entries map[string](map[uint32]string)
}

func NewDatabase() Database {
	return Database{lock: &sync.RWMutex{}, entries: make(map[string](map[uint32]string))}
}

func (db *Database) PossessRumorMessage(msg *RumorMessage) bool {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, ok := db.entries[msg.Origin]; ok {
		_, ok := entry[msg.ID]
		return ok
	}
	return false
}

func (db *Database) InsertRumorMessage(msg *RumorMessage) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if entry, ok := db.entries[msg.Origin]; ok {
		entry[msg.ID] = msg.Text
	} else {
		entry = make(map[uint32]string)
		entry[msg.ID] = msg.Text
		db.entries[msg.Origin] = entry
	}
}

func (db *Database) GetMessageContent(name string, id uint32) string {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return db.entries[name][id]
}

func (db *Database) GetMinNotPresent(name string) uint32 {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if _, ok := db.entries[name]; !ok {
		return uint32(1)
	}

	return uint32(len(db.entries[name]) + 1)
}

func (db *Database) GetPeerStatus() []PeerStatus {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var status []PeerStatus

	for name, entry := range db.entries {
		status = append(status, PeerStatus{Identifier: name, NextID: uint32(len(entry) + 1)})
	}
	return status
}
