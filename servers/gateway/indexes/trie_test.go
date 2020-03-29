package indexes

import "testing"

//TODO: implement automated tests for your trie data structure
func TestAddLen(t *testing.T) {
	cases := []struct {
		name     string
		keys     []string
		values   []int64
		expected int
	}{
		{
			"Single Entry",
			[]string{"Eric"},
			[]int64{1},
			1,
		},
		{
			"Duplicate Entries",
			[]string{"Eric", "Eric"},
			[]int64{1, 2},
			2,
		},
		{
			"Distinct Values",
			[]string{"Eric", "Guo", "Ziyu"},
			[]int64{1, 2, 3},
			3,
		},
		{
			"Distinct Values Then Duplicates",
			[]string{"Eric", "Guo", "Eric"},
			[]int64{1, 2, 3},
			3,
		},
		{
			"Duplicates Then Distinct Values",
			[]string{"Eric", "Eric", "Ziyu"},
			[]int64{1, 2, 3},
			3,
		},
	}

	for _, c := range cases {
		testTrie := NewTrieNode()
		emptyLen := testTrie.Len()
		if emptyLen != 0 {
			t.Errorf("Len() func returns incorrect len")
		}
		for idx, k := range c.keys {
			v := c.values[idx]
			testTrie.Add(k, v)
		}

		l := testTrie.Len()
		if l != c.expected {
			t.Errorf("case %s: incorrect length after add into trie: expected %d but got %d", c.name, c.expected, l)
		}
	}
}

func TestAddFind(t *testing.T) {
	cases := []struct {
		name     string
		keys     []string
		values   []int64
		prefix   string
		max      int
		expected []int64
	}{
		{
			"Single Entry",
			[]string{"Eric"},
			[]int64{1},
			"Eri",
			10,
			[]int64{1},
		},
		{
			"Duplicate Entries",
			[]string{"Eric", "Eric"},
			[]int64{1, 2},
			"Eric",
			10,
			[]int64{1, 2},
		},
		{
			"Distinct Values",
			[]string{"Eric", "Guo", "Ziyu"},
			[]int64{1, 2, 3},
			"Eric",
			10,
			[]int64{1},
		},
		{
			"Distinct Values Then Duplicates",
			[]string{"Eric", "Guo", "Eric"},
			[]int64{1, 2, 3},
			"Eric",
			10,
			[]int64{1, 3},
		},
		{
			"Duplicates Then Distinct Values",
			[]string{"Eric", "Eric", "Ziyu"},
			[]int64{1, 2, 3},
			"Eric",
			10,
			[]int64{1, 2},
		},
		{
			"Empty Trie",
			[]string{},
			[]int64{},
			"Eric",
			10,
			[]int64{},
		},
		{
			"Empty prefix",
			[]string{"Eric", "Eric", "Ziyu"},
			[]int64{1, 2, 3},
			"",
			10,
			[]int64{},
		},
		{
			"Max is zero",
			[]string{"Eric", "Eric", "Ziyu"},
			[]int64{1, 2, 3},
			"Eric",
			0,
			[]int64{},
		},
		{
			"Prefix not found",
			[]string{"Eric", "Eric", "Ziyu"},
			[]int64{1, 2, 3},
			"K",
			10,
			[]int64{},
		},
		{
			"Common prefix",
			[]string{"John", "Johnson", "Johnny", "Johseph", "Kate"},
			[]int64{1, 2, 3, 4, 5},
			"Joh",
			10,
			[]int64{1, 3, 2, 4},
		},
	}

	for _, c := range cases {
		testTrie := NewTrieNode()
		emptyLen := testTrie.Len()
		if emptyLen != 0 {
			t.Errorf("Len() func returns incorrect len")
		}
		for idx, k := range c.keys {
			v := c.values[idx]
			testTrie.Add(k, v)
		}
		slice := testTrie.Find(c.prefix, c.max)
		if len(slice) != len(c.expected) {
			t.Errorf("case %s: incorrect search result length: expected %d but got %d",
				c.name, len(c.expected), len(slice))
		}
		for idx, v := range c.expected {
			rv := slice[idx]
			if rv != v {
				t.Errorf("case %s: incorrect search result: expected %d but got %d", c.name, v, rv)
			}
		}
	}
}

func TestAddRemove(t *testing.T) {
	cases := []struct {
		name      string
		keys      []string
		values    []int64
		remkeys   []string
		remvalues []int64
		expected  int
	}{
		{
			"Empty Trie After Removal",
			[]string{"Eric"},
			[]int64{1},
			[]string{"Eric"},
			[]int64{1},
			0,
		},
		{
			"Non-Empty Trie After Removal",
			[]string{"Eric", "Guo", "Ziyu"},
			[]int64{1, 2, 3},
			[]string{"Eric"},
			[]int64{1},
			2,
		},
		{
			"Remove One of the Duplicates",
			[]string{"Eric", "Eric"},
			[]int64{1, 2},
			[]string{"Eric"},
			[]int64{1},
			1,
		},
		{
			"Remove Second One of the Duplicates",
			[]string{"Eric", "Eric"},
			[]int64{1, 2},
			[]string{"Eric"},
			[]int64{2},
			1,
		},
	}

	for _, c := range cases {
		testTrie := NewTrieNode()
		for idx, k := range c.keys {
			v := c.values[idx]
			testTrie.Add(k, v)
		}
		for idx, rk := range c.remkeys {
			rv := c.remvalues[idx]
			testTrie.Remove(rk, rv)
		}
		l := testTrie.Len()
		if l != c.expected {
			t.Errorf("case %s: incorrect length after remove from trie: expected %d but got %d", c.name, c.expected, l)
		}
	}
}
