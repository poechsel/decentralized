package lib

import (
	"sync"
)

type Stack struct {
	items []interface{}
	lock  *sync.Mutex
}

func NewStack() *Stack {
	o := Stack{
		items: make([]interface{}, 0),
		lock:  &sync.Mutex{},
	}
	return &o
}

func (stack *Stack) Empty() bool {
	stack.lock.Lock()
	defer stack.lock.Unlock()
	return len(stack.items) == 0
}

func (stack *Stack) Push(el interface{}) {
	stack.lock.Lock()
	defer stack.lock.Unlock()
	stack.items = append(stack.items, el)
}

func (stack *Stack) Pop() interface{} {
	stack.lock.Lock()
	defer stack.lock.Unlock()
	item := stack.items[len(stack.items)-1]
	stack.items = stack.items[0 : len(stack.items)-1]
	return item
}
