// Thread-safe autocomplete for english
package autocomplete

import (
  "sync"
  "bytes"
)

var empty = []string{}
type processor func(id string, value []byte)

// The root of the trie
type Root struct {
  pool *Pool
  head *Level
  sync.RWMutex
  maxLength int
  lookup map[string]string
}

// A node within the trie
type Level struct {
  ids []string
  prefix map[byte]*Level
}

// Creates a new trie, specifying the maximum length of input
// we must handle
func New(maxLength int) *Root {
  return &Root{
    head: newLevel(),
    maxLength: maxLength,
    pool: newPool(1024, maxLength),
    lookup: make(map[string]string),
  }
}

// Create a new inner level
func newLevel() *Level {
  return &Level{
    ids: make([]string, 0, 1),
    prefix: make(map[byte]*Level),
  }
}

// Find the ids that match the specified value
func (root *Root) Find(value string) []string {
  root.RLock()
  defer root.RUnlock()
  node := root.head
  buffer := root.normalize(value, false)
  defer buffer.Close()
  for _, b := range buffer.Bytes() {
    var exists bool
    node, exists = node.prefix[b]
    if exists == false { return empty }
  }
  return node.ids
}

// Insert the id=>value into the tree
func (root *Root) Insert(id string, value string) {
  root.Lock()
  oldValue, exists := root.lookup[id]
  root.lookup[id] = value
  root.Unlock()
  if exists { root.process(id, oldValue, root.remove) }
  root.process(id, value, root.insert)
}

// Removes the id for the given title
func (root *Root) Remove(id string) {
  root.Lock()
  value, exists := root.lookup[id]
  root.Unlock()
  if exists == false { return }
  root.process(id, value, root.remove)
}

func (root *Root) insert(id string, value []byte) {
  root.Lock()
  defer root.Unlock()
  node := root.head
  for _, b := range value {
    sub, exists := node.prefix[b]
    if exists == false {
      sub = newLevel()
      node.prefix[b] = sub
    }
    sub.ids = append(sub.ids, id)
    node = node.prefix[b]
  }
}

func (root *Root) remove(id string, value []byte) {
  root.Lock()
  defer root.Unlock()
  node := root.head
  for _, b := range value {
    sub, exists := node.prefix[b]
    if exists == false { return }
    sub.ids = delete(sub.ids, id)
    node = node.prefix[b]
  }
}

func (root *Root) process(id string, value string, processor processor) {
  valueBuffer := root.normalize(value, true)
  defer valueBuffer.Close()
  full := valueBuffer.Bytes()

  parts := bytes.Split(full, []byte{32})
  partials := make([][]byte, len(parts))

  for i := 0; i < len(parts); i++ {
    partial := bytes.Join(parts[i:], []byte{})
    for _, seen := range partials {
      if bytes.Index(seen, partial) == 0 { goto next }
    }
    processor(id, partial)
    partials = append(partials, partial)
next:
  }
}

// Lowercase A-Z and strip out none a-z characters, preserving spaces if requested
func (root *Root) normalize(value string, preserveSpaces bool) *Buffer {
  buffer := root.pool.Checkout()
  for _, b := range []byte(value) {
    if (b >= 97 && b <= 122) || (b >= 48 && b <= 57) || (preserveSpaces && b == 32) {
      if buffer.Write(b) == root.maxLength { break }
    } else if b >= 65 && b <= 90 {
      if buffer.Write(b + 32) == root.maxLength { break }
    }
  }
  return buffer
}

// Removes an item from a slice.
func delete(array []string, target string) []string {
  for index, value := range array {
    if value != target { continue }
    l := len(array)
    if l == 0 { return make([]string, 0, 1) }
    newArray := make([]string, l-1)
    copy(newArray, array[0:index])
    copy(newArray[index:], array[index+1:])
    return newArray
  }
  return array
}
