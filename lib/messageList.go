package lib

import (
	"sync"
)

type MessageList struct {
	lock    *sync.RWMutex
	content map[uint32]string
}

func NewMessageList() MessageList {
	return MessageList{lock: &sync.RWMutex{}, content: make(map[uint32]string)}
}

func (ml *MessageList) Insert(id uint32, msg string) bool {
	ml.lock.Lock()
	defer ml.lock.Unlock()

	if _, ok := ml.content[id]; ok {
		return false
	} else {
		ml.content[id] = msg
		return true
	}
}

func (ml *MessageList) Get(id uint32) string {
	ml.lock.RLock()
	defer ml.lock.RUnlock()

	return ml.content[id]
}

func (ml *MessageList) Possess(id uint32) bool {
	ml.lock.RLock()
	defer ml.lock.RUnlock()

	_, ok := ml.content[id]
	return ok
}
