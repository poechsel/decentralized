package lib

import (
	"sync"
)

type Entry struct {
	messageList    *MessageList
	sparseSequence *SparseSequence
}

func NewEntry() Entry {
	msglist := NewMessageList()
	sparse := NewSparseSequence()
	// no messages can have id 0, so we are always gonna insert it
	sparse.Insert(0)
	e := Entry{messageList: &msglist,
		sparseSequence: &sparse}
	return e
}

func (entry *Entry) Insert(id uint32, msg string) bool {
	entry.sparseSequence.Insert(id)
	return entry.messageList.Insert(id, msg)
}

type Database struct {
	lock    *sync.RWMutex
	entries map[string]Entry
}

func NewDatabase() Database {
	return Database{lock: &sync.RWMutex{}}
}

func (db *Database) InsertRumorMessage(msg *RumorMessage) {
	db.lock.Lock()
	defer db.lock.Unlock()
	if entry, ok := db.entries[msg.Origin]; ok {
		entry.Insert(msg.ID, msg.Text)
	} else {
		entry = NewEntry()
		entry.Insert(msg.ID, msg.Text)
		db.entries[msg.Origin] = entry
	}
}

func (db *Database) GetMessageContent(name string, id uint32) string {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return db.entries[name].messageList.Get(id)
}

func (db *Database) GetMinNotPresent(name string) uint32 {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return db.entries[name].sparseSequence.GetMinNotPresent()
}

func (db *Database) GetPeerStatus() []PeerStatus {
	db.lock.RLock()
	defer db.lock.RUnlock()

	status := make([]PeerStatus, len(db.entries))

	for name, entry := range db.entries {
		next := entry.sparseSequence.GetMinNotPresent()
		status = append(status, PeerStatus{Identifier: name, NextID: next})
	}
	return status
}
