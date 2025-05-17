package lru_cache

import (
	"testing"
)

func TestAddWithoutEvict(t *testing.T) {

	type testdata struct {
		key string
		val string
	}
	adds := []testdata{
		{key: "a", val: "a"},
		{key: "b", val: "b"},
	}

	c := New[string, string](2)

	for _, tt := range adds {
		c.Add(tt.key, tt.val)
	}

	testLenghts(t, c, 2)
}

func TestAddWithEvict(t *testing.T) {

	type testdata struct {
		key string
		val string
	}
	adds := []testdata{
		{key: "a", val: "a"},
		{key: "b", val: "b"},
		{key: "c", val: "c"},
	}

	c := New[string, string](2)

	for _, tt := range adds {
		c.Add(tt.key, tt.val)
	}

	testLenghts(t, c, 2)

	testKeys(t, c)
}

func TestGetPresent(t *testing.T) {

	type testdata struct {
		key string
		val string
	}
	adds := []testdata{
		{key: "a", val: "a"},
		{key: "b", val: "b"},
		{key: "c", val: "c"},
	}

	c := New[string, string](3)

	for _, tt := range adds {
		c.Add(tt.key, tt.val)
	}

	frontKey := c.evictList.Front().Value.(ListEntry[string, string]).Key
	if frontKey != "c" {
		t.Errorf("wrong evictList front: expected=%v, got=%v", "c", frontKey)
	}

	_, ok := c.Get("a")
	if !ok {
		t.Errorf("expected key=%v to be in cache", "a")
	}

	frontKey = c.evictList.Front().Value.(ListEntry[string, string]).Key
	if frontKey != "a" {
		t.Errorf("wrong evictList front: expected=%v, got=%v", "a", frontKey)
	}

	_, ok = c.Get("b")
	if !ok {
		t.Errorf("expected key=%v to be in cache", "b")
	}

	frontKey = c.evictList.Front().Value.(ListEntry[string, string]).Key
	if frontKey != "b" {
		t.Errorf("wrong evictList front: expected=%v, got=%v", "b", frontKey)
	}
	testLenghts(t, c, 3)

	testKeys(t, c)
}
func TestGetAbsent(t *testing.T) {

	type testdata struct {
		key string
		val string
	}
	adds := []testdata{
		{key: "a", val: "a"},
		{key: "b", val: "b"},
		{key: "c", val: "c"},
	}

	c := New[string, string](3)

	for _, tt := range adds {
		c.Add(tt.key, tt.val)
	}

	_, ok := c.Get("d")
	if ok {
		t.Errorf("expected key=%v to not be in cache", "d")
	}

	testLenghts(t, c, 3)

	testKeys(t, c)
}

func TestGetEmptyCache(t *testing.T) {

	c := New[string, string](3)

	_, ok := c.Get("d")
	if ok {
		t.Errorf("expected key=%v to not be in cache", "d")
	}

	testLenghts(t, c, 0)

	testKeys(t, c)

	c.Add("d", "d")

	testLenghts(t, c, 1)

	testKeys(t, c)
}

func testLenghts[K comparable, V any](t *testing.T, c *Cache[K, V], expected int) {
	if len(c.data) != expected {
		t.Errorf("wrong len cache map, expected=%d, got=%d", expected, len(c.data))
	}
	if c.evictList.Len() != expected {
		t.Errorf("wrong len cache map, expected=%d, got=%d", expected, c.evictList.Len())
	}
	if c.evictList.Len() != len(c.data) {
		t.Errorf("evict list len is not equal map len")
	}
}

func testKeys[K comparable, V any](t *testing.T, c *Cache[K, V]) {

	for e := c.evictList.Front(); e != nil; e = e.Next() {
		key := e.Value.(ListEntry[K, V]).Key
		if _, ok := c.data[key]; !ok {
			t.Errorf("evictList key %v not in cache data %v", key, c.data)
		}
	}
}
