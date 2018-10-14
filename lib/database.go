package lib

import (
	"sync"
)

/* A database is a structure holding together the list
of all messages stored until now.
It is basically a two dimensional hashmap, the first one
being indexed by the peers name and the second one by the
message ids */

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

func auxGetMinNotPresent(m map[uint32]string) uint32 {
	return uint32(len(m) + 1)
}

func (db *Database) GetMinNotPresent(name string) uint32 {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if _, ok := db.entries[name]; !ok {
		return uint32(1)
	}

	return auxGetMinNotPresent(db.entries[name])
}

func (db *Database) GetPeerStatus() []PeerStatus {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var status []PeerStatus

	for name, entry := range db.entries {
		status = append(status,
			PeerStatus{Identifier: name,
				NextID: auxGetMinNotPresent(entry)})
	}
	return status
}
