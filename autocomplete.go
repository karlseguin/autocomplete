// Thread-safe autocomplete for english
package autocomplete

import (
	"bytes"
	"sync"
)

// what could go wrong?
const NULL uint = 18446744073709551615

var empty = []uint{}

type processor func(id uint, value []byte, front bool)

// The root of the trie
type Root struct {
	pool *Pool
	head *Level
	sync.RWMutex
	maxLength int
	lookup    map[uint]string
}

// A node within the trie
type Level struct {
	ids    []uint
	prefix map[byte]*Level
	set    map[uint]struct{}
}

// Creates a new trie, specifying the maximum length of input
// we must handle
func New(maxLength int) *Root {
	return &Root{
		head:      newLevel(),
		maxLength: maxLength,
		pool:      newPool(1024, maxLength),
		lookup:    make(map[uint]string),
	}
}

// Create a new inner level
func newLevel() *Level {
	return &Level{
		ids:    make([]uint, 0),
		set: make(map[uint]struct{}),
		prefix: make(map[byte]*Level),
	}
}

// Find the ids that match the specified value
func (root *Root) Find(value string) []uint {
	root.RLock()
	defer root.RUnlock()
	node := root.head
	buffer := root.normalize(value, false)
	defer buffer.Close()
	for _, b := range buffer.Bytes() {
		var exists bool
		node, exists = node.prefix[b]
		if exists == false {
			return empty
		}
	}
	return node.ids
}

// Insert the id=>value into the tree
func (root *Root) Insert(id uint, value string) {
	root.Lock()
	oldValue, exists := root.lookup[id]
	root.lookup[id] = value
	root.Unlock()
	if exists {
		root.process(id, oldValue, root.remove)
	}
	root.process(id, value, root.insert)
}

// Removes the id for the given title
func (root *Root) Remove(id uint) {
	root.Lock()
	value, exists := root.lookup[id]
	root.Unlock()
	if exists == false {
		return
	}
	root.process(id, value, root.remove)
}

func (root *Root) insert(id uint, value []byte, front bool) {
	root.Lock()
	defer root.Unlock()
	node := root.head
	for _, b := range value {
		sub, exists := node.prefix[b]
		if exists == false {
			sub = newLevel()
			node.prefix[b] = sub
		}
		if _, exists := sub.set[id]; exists == false {
			ids := make([]uint, len(sub.ids)+1)
			if front {
				ids[0] = id
				copy(ids[1:], sub.ids)

			} else {
				copy(ids, sub.ids)
				ids[len(sub.ids)] = id
			}
			sub.ids = ids
			sub.set[id] = struct{}{}
		}
		node = node.prefix[b]
	}
}

func (root *Root) remove(id uint, value []byte, front bool) {
	root.Lock()
	defer root.Unlock()
	node := root.head
	for _, b := range value {
		sub, exists := node.prefix[b]
		if exists == false {
			return
		}
		sub.ids = delete(sub.ids, id)
		node = node.prefix[b]
	}
}

func (root *Root) process(id uint, value string, processor processor) {
	valueBuffer := root.normalize(value, true)
	defer valueBuffer.Close()
	full := valueBuffer.Bytes()

	parts := bytes.Split(full, []byte{32})
	partials := make([][]byte, 0, len(parts))

	for i := 0; i < len(parts); i++ {
		partial := bytes.Join(parts[i:], []byte{})
		for _, seen := range partials {
			if bytes.Index(seen, partial) == 0 {
				goto next
			}
		}
		processor(id, partial, i == 0)
		partials = append(partials, partial)
	next:
	}
}

// Lowercase A-Z and strip out none a-z characters, preserving spaces if requested
func (root *Root) normalize(value string, preserveSpaces bool) *Buffer {
	buffer := root.pool.Checkout()
	for _, b := range []byte(value) {
		if (b >= 97 && b <= 122) || (b >= 48 && b <= 57) || (preserveSpaces && b == 32) {
			if buffer.Write(b) == root.maxLength {
				break
			}
		} else if b >= 65 && b <= 90 {
			if buffer.Write(b+32) == root.maxLength {
				break
			}
		}
	}
	return buffer
}

// Removes an item from a slice.
func delete(array []uint, target uint) []uint {
	for index, value := range array {
		if value != target {
			continue
		}
		l := len(array)
		if l == 0 {
			return make([]uint, 0, 1)
		}
		newArray := make([]uint, l-1)
		copy(newArray, array[0:index])
		copy(newArray[index:], array[index+1:])
		return newArray
	}
	return array
}
