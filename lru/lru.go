package lru

import (
	"container/list"
	"errors"
	"sync"
)

type Elementer interface {
	Element(key, value interface{})
}

type ElementerFunc func(key, value interface{})

func (f ElementerFunc) Element(key, value interface{}) {
	f(key, value)
}

type lruEntry struct {
	key   interface{}
	value interface{}
}

type LRU struct {
	sync.RWMutex
	size      int
	evictList *list.List
	data      map[interface{}]*list.Element
	onEvict   Elementer
}

func New(size int, onEvict Elementer) *LRU {
	if size <= 0 {
		panic(errors.New("lru must have a positive size"))
	}

	return &LRU{
		size:      size,
		evictList: list.New(),
		data:      map[interface{}]*list.Element{},
		onEvict:   onEvict,
	}
}

func (l *LRU) Each(e Elementer) {
	l.Lock()
	defer l.Unlock()

	for k, elem := range l.data {
		e.Element(k, elem.Value.(*lruEntry).value)
	}
}

func (l *LRU) Remove(k interface{}) {
	l.Lock()
	defer l.Unlock()

	l.RemoveNoLock(k)
}

func (l *LRU) RemoveNoLock(k interface{}) {
	if elem, ok := l.data[k]; ok {
		l.removeElement(elem)
	}
}

func (l *LRU) removeElement(elem *list.Element) {
	l.evictList.Remove(elem)

	ent := elem.Value.(*lruEntry)
	delete(l.data, ent.key)
	if l.onEvict != nil {
		l.onEvict.Element(ent.key, ent.value)
	}
}

func (l *LRU) removeOldest() {
	if elem := l.evictList.Back(); elem != nil {
		l.removeElement(elem)
	}
}

func (l *LRU) Get(k interface{}) (interface{}, bool) {
	l.Lock()
	defer l.Unlock()

	if ent, ok := l.data[k]; ok {
		l.evictList.MoveToFront(ent)
		return ent.Value.(*lruEntry).value, true
	}

	return nil, false
}

func (l *LRU) AddIfNotContains(k, value interface{}) bool {
	l.Lock()
	defer l.Unlock()

	if _, ok := l.data[k]; ok {
		return false
	}

	ent := &lruEntry{k, value}
	lruEntry := l.evictList.PushFront(ent)
	l.data[k] = lruEntry

	if l.evictList.Len() > l.size {
		l.removeOldest()
	}

	return true
}

func (l *LRU) Add(k, value interface{}) {
	l.Lock()
	defer l.Unlock()

	if ent, ok := l.data[k]; ok {
		l.evictList.MoveToFront(ent)
		ent.Value.(*lruEntry).value = value
		return
	}

	lruEntry := l.evictList.PushFront(&lruEntry{
		key:   k,
		value: value,
	})
	l.data[k] = lruEntry

	if l.evictList.Len() > l.size {
		l.removeOldest()
	}
}

func (l *LRU) Purge() {
	l.Lock()
	defer l.Unlock()

	for k, v := range l.data {
		if l.onEvict != nil {
			l.onEvict.Element(k, v.Value.(*lruEntry).value)
		}
		delete(l.data, k)
	}
	l.evictList.Init()
}
