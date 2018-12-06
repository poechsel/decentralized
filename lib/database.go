package lib

import (
	"sync"
)

/* A database is a structure holding together the list
of all messages stored until now.
It is basically a two dimensional hashmap, the first one
being indexed by the peers name and the second one by the
message ids */

type Entry struct {
	min_not_present  uint32
	min_not_present2 uint32
	messages         map[uint32]string
}

func NewEntry() *Entry {
	return &Entry{min_not_present: 1, min_not_present2: 1, messages: make(map[uint32]string)}
}

func (entry *Entry) Insert(text string, id uint32) {
	entry.messages[id] = text
	entry.min_not_present = uint32(len(entry.messages))
	entry.min_not_present2 = max(entry.min_not_present2, id+1)
}

type Database struct {
	lock    *sync.Mutex
	entries map[string](*Entry)
}

func NewDatabase() Database {
	return Database{lock: &sync.Mutex{}, entries: make(map[string](*Entry))}
}

func (db *Database) PossessRumorMessage(msg *RumorMessage) bool {
	db.lock.Lock()
	defer db.lock.Unlock()

	if entry, ok := db.entries[msg.Origin]; ok {
		_, ok := entry.messages[msg.ID]
		return ok
	}
	return false
}

func (db *Database) InsertRumorMessage(msg *RumorMessage) {
	db.lock.Lock()
	defer db.lock.Unlock()

	if entry, ok := db.entries[msg.Origin]; ok {
		entry.Insert(msg.Text, msg.ID)
	} else {
		entry = NewEntry()
		entry.Insert(msg.Text, msg.ID)
		db.entries[msg.Origin] = entry
	}
}

func (db *Database) GetMessageContent(name string, id uint32) string {
	db.lock.Lock()
	defer db.lock.Unlock()

	return db.entries[name].messages[id]
}

func auxGetMinNotPresent(m *Entry) uint32 {
	return m.min_not_present2
	return uint32(len(m.messages) + 1)
}

func (db *Database) GetMinNotPresent(name string) uint32 {
	db.lock.Lock()
	defer db.lock.Unlock()

	if _, ok := db.entries[name]; !ok {
		db.entries[name] = NewEntry()
	}
	return auxGetMinNotPresent(db.entries[name])
}

func (db *Database) GetPeerStatus() []PeerStatus {
	db.lock.Lock()
	defer db.lock.Unlock()

	var status []PeerStatus

	for name, entry := range db.entries {
		status = append(status,
			PeerStatus{Identifier: name,
				NextID: auxGetMinNotPresent(entry)})
	}
	return status
}
