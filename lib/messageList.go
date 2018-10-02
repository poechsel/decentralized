package lib

import (
	"sync"
)

type MessageList struct {
	lock    *sync.RWMutex
	content map[uint32]string
}

func NewMessageList() MessageList {
	return MessageList{lock: &sync.RWMutex{}}
}

func (ml *MessageList) Insert(id uint32, msg string) {
	ml.lock.Lock()
	defer ml.lock.Unlock()

	ml.content[id] = msg
}

func (ml *MessageList) Get(id uint32) string {
	ml.lock.RLock()
	defer ml.lock.RUnlock()

	return ml.content[id]
}
