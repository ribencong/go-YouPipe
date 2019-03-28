package gossip

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

var messageID = time.Now().Unix()
var messageLock sync.RWMutex

type counter struct {
	sync.RWMutex
	expired time.Time
	cache   map[string]int
}

func GenMsgID(suffix string) string {
	messageLock.Lock()
	defer messageLock.Unlock()
	messageID++
	return fmt.Sprintf("%d-%s", messageID, suffix)
}

func newCounter() *counter {
	return &counter{
		cache:   make(map[string]int),
		expired: time.Now(),
	}
}

func (c *counter) IsOverFlood(s string) bool {
	c.RLock()
	defer c.RUnlock()
	return c.cache[s] >= ForwardThreshold
}

func (c *counter) Inc(s string) {
	c.Lock()
	defer c.Unlock()
	c.cache[s]++

	if len(c.cache) >= MsgCacheSize {
		go c.halfTheCache()
	}
}

//TODO:: to be tested
func (c *counter) halfTheCache() {
	c.Lock()
	defer c.Unlock()

	var keys []string
	for k := range c.cache {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	delKey := keys[:len(keys)/2]

	for _, k := range delKey {
		delete(c.cache, k)
	}
}

func (c *counter) Has(s string) bool {
	c.RLock()
	defer c.RUnlock()

	_, ok := c.cache[s]
	return ok
}
