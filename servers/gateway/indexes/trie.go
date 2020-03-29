package indexes

import (
	"sort"
	"sync"
)

//PRO TIP: if you are having troubles and want to see
//what your trie structure looks like at various points,
//either use the debugger, or try this package:
//https://github.com/davecgh/go-spew

//TrieNode implements a trie data structure mapping strings to int64s
//that is safe for concurrent use.
type TrieNode struct {
	children map[rune]*TrieNode // Each letter will point to a new subTrieNode
	values   int64set           // IDs for the same name will be stored in set
	mx       sync.RWMutex       // Read/write mutex
}

//NewTrieNode constructs a new TrieNode.
func NewTrieNode() *TrieNode {
	return &TrieNode{}
}

//Len returns the number of entries in the trie.
func (t *TrieNode) Len() int {
	t.mx.RLock()
	defer t.mx.RUnlock()
	return t.len()
}

func (t *TrieNode) len() int {
	entryCount := len(t.values)     // Find the length of self
	for child := range t.children { // Find length of children
		entryCount += t.children[child].len()
	}
	return entryCount
}

//Add adds a key and value to the trie.
func (t *TrieNode) Add(key string, value int64) {
	t.mx.Lock()
	defer t.mx.Unlock()
	runes := []rune(key)
	t.add(runes, value)
}

// add is a private helper method that adds a key and value to the trie
func (t *TrieNode) add(key []rune, value int64) {
	// if children do not exist, make sure it is an empty map
	if len(t.children) == 0 {
		t.children = make(map[rune]*TrieNode)
	}
	// if the child does not exist, create a new trie node and store it there
	if t.children[key[0]] == nil {
		t.children[key[0]] = NewTrieNode()
	}
	if len(key) == 1 {
		if len(t.children[key[0]].values) == 0 {
			t.children[key[0]].values = make(map[int64]struct{})
		}
		// add the value and then return
		t.children[key[0]].values.add(value)
		return
	}
	// otherwise, call the add method again on the child node (recursively)
	t.children[key[0]].add(key[1:len(key)], value)
}

//Find finds `max` values matching `prefix`. If the trie
//is entirely empty, or the prefix is empty, or max == 0,
//or the prefix is not found, this returns a nil slice.
func (t *TrieNode) Find(prefix string, max int) []int64 {
	t.mx.RLock()
	defer t.mx.RUnlock()

	if len(t.children) == 0 || prefix == "" || max <= 0 {
		return nil
	}

	// iterate through trie until at end of prefix | O(1)
	prefixRunes := []rune(prefix)
	triePointer := t
	for _, s := range prefixRunes {
		if triePointer.children[s] == nil {
			return nil
		}
		triePointer = triePointer.children[s]
	}
	// create int64 slice
	var returnSlice []int64
	triePointer.findDFS(&returnSlice, max)
	return returnSlice
}

//findDFS is a private function that recursively goes down the trie to return a slice of results
func (t *TrieNode) findDFS(list *[]int64, max int) {
	// add all current values in node to list (or until hit max)
	values := t.values.all()
	canGet := max - len(*list)
	if len(values) > canGet {
		*list = append(*list, values[0:canGet]...)
		return
	}
	*list = append(*list, values...)
	// if max reached or no children, just return.
	if len(*list) == max || len(t.children) == 0 {
		return
	}
	// sort children
	children := make([]rune, 0, len(t.children))
	for k := range t.children {
		children = append(children, k)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i] < children[j]
	})
	// for every child, recurse and add to list and check for max
	for _, child := range children {
		t.children[child].findDFS(list, max)
		if len(*list) == max {
			return
		}
	}
	return
}

//Remove removes a key/value pair from the trie
//and trims branches with no values.
func (t *TrieNode) Remove(key string, value int64) {
	t.mx.Lock()
	defer t.mx.Unlock()
	runes := []rune(key)
	t.remove(runes, value)
}

func (t *TrieNode) remove(key []rune, value int64) {
	if len(key) == 0 {
		t.values.remove(value)
		return
	}
	focusChild := t.children[key[0]]
	focusChild.remove(key[1:], value)
	if len(focusChild.children) == 0 && len(focusChild.values) == 0 {
		delete(t.children, key[0])
	}
}
