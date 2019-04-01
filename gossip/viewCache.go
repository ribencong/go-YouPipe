package gossip

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ViewCache struct {
	sync.RWMutex
	name       string
	pool       map[string]*viewNode
	hitCounter int
}

func newCache(n string) *ViewCache {
	return &ViewCache{
		hitCounter: 0,
		name:       n,
		pool:       make(map[string]*viewNode),
	}
}

func (v *ViewCache) Size() int {
	v.RLock()
	defer v.RUnlock()

	return len(v.pool)
}

func (v *ViewCache) Broadcast(data []byte) {
	v.Lock()
	defer v.Unlock()
	for id, node := range v.pool {
		//logger.Debugf("(%s)broad cast msg to %s nodes", v.name, id)
		if err := node.send(data); err != nil {
			delete(v.pool, id)
		}
	}
	logger.Debugf("(%s)broad cast msg to %d nodes", v.name, len(v.pool))
}

func (v *ViewCache) Add(node *viewNode, nodeId string) bool {

	if node == nil {
		return false
	}

	v.Lock()
	defer v.Unlock()

	if _, ok := v.pool[node.peerID]; ok {
		return false
	}

	_, avg := v.weightSum()
	node.probability = avg
	v.pool[node.peerID] = node

	v.hitCounter++
	if v.hitCounter >= UpdateThreshold {
		go v.broadCastNewWeight(nodeId)
	}

	logger.Debugf("new node(%s-%.2f) added to (%s) hit counter(%d)",
		node.peerID, avg, v.name, v.hitCounter)
	return true
}

func (v *ViewCache) broadCastNewWeight(nodeId string) {
	v.Lock()
	defer v.Unlock()
	v.hitCounter = 0
	v.normalizeWeight()

	logger.Debugf(" (%s)start to update weight:->", v.name)

	for id, node := range v.pool {
		data, _ := pack(UpdateWeight, nodeId, node.probability, node.dir)
		if err := node.send(data); err != nil {
			delete(v.pool, id)
			node.Close()
		}
		logger.Debugf("update node(%s) with weight:%.2f", node.peerID, node.probability)
	}

}
func (v *ViewCache) RandomSend(data []byte) error {
	err, nid := v.randomSendWithLock(data)
	if err != nil {
		v.Lock()
		delete(v.pool, nid)
		v.Unlock()
	}
	return err
}

func (v *ViewCache) randomSendWithLock(data []byte) (error, string) {
	v.RLock()
	defer v.RUnlock()

	if len(v.pool) == 0 {
		return nil, ""
	}

	idx := rand.Intn(len(v.pool))
	i := 0
	for id, item := range v.pool {
		if i != idx {
			i++
			continue
		}
		logger.Debug("rand send:->", id)
		return item.send(data), id
	}

	return nil, ""
}
func (v *ViewCache) Has(nodeId string) bool {
	v.RLock()
	defer v.RUnlock()

	_, ok := v.pool[nodeId]
	return ok
}

func (v *ViewCache) Get(nodeId string) (*viewNode, bool) {
	v.RLock()
	defer v.RUnlock()

	node, ok := v.pool[nodeId]
	return node, ok
}

func (v *ViewCache) weightSum() (float64, float64) {

	if len(v.pool) <= 1 {
		return 1.0, 1.0
	}

	sum := 0.0
	for _, node := range v.pool {
		sum += node.probability
	}

	return sum, sum / float64(len(v.pool))
}

func (v *ViewCache) normalizeWeight() {

	sum, _ := v.weightSum()

	for _, node := range v.pool {
		node.probability = node.probability / sum
		logger.Debugf("locally(%s) normalize node(%s)"+
			" weight(%.2f)", v.name, node.peerID, node.probability)
	}
}

func (v *ViewCache) ChoseByProb(data []byte) error {

	v.Lock()
	defer v.Unlock()

	v.normalizeWeight()

	rand.Seed(time.Now().UnixNano())
	randProb := rand.Float64()

	logger.Debug("random prob:->", randProb)
	sum := 0.0

	for _, node := range v.pool {

		sum += node.probability

		if randProb > sum {
			continue
		}

		logger.Debugf("found target(%s) random prob=%.2f, sum=%.2f",
			node.peerID, randProb, sum)
		return node.send(data)
	}

	logger.Warningf("random failed prob=%.2f, sum=%.2f", randProb, sum)

	return ENotFound
}

func (v *ViewCache) Remove(peerID string) {
	v.Lock()
	defer v.Unlock()
	node, ok := v.pool[peerID]
	if !ok {
		logger.Debugf(" (%s) no such node(%s) to remove", v.name, peerID)
		return
	}

	delete(v.pool, peerID)
	node.Destroy()

	logger.Debugf("remove node(%s) from  (%s)  ", peerID, v.name)
}

func (v *ViewCache) updateNodeWeight(nid string, wei float64) error {
	v.RLock()
	defer v.RUnlock()

	node, ok := v.pool[nid]
	if !ok {
		return ENotFound
	}
	node.updateWeight(wei)
	return nil
}

func (v *ViewCache) ClearAll() {

	v.Lock()
	defer v.Unlock()

	for id, node := range v.pool {
		delete(v.pool, id)
		node.Destroy()
	}
}

func (v *ViewCache) String(name string) string {
	v.RLock()
	defer v.RUnlock()

	str := fmt.Sprintf("\n\n*************************%s(%d)*************************", name, len(v.pool))
	for _, node := range v.pool {
		str += node.String()
	}
	str += fmt.Sprintf("\n\n********************************************************")
	return str
}

func (v *ViewCache) AllKeys() []string {
	v.RLock()
	defer v.RUnlock()

	keys := make([]string, len(v.pool))

	i := 0
	for id := range v.pool {
		keys[i] = id
		i++
	}

	return keys
}

func (v *ViewCache) GroupCast(ids []string, data []byte) {
	v.RLock()
	defer v.RUnlock()
	logger.Debug("group cast :->", ids)

	for _, id := range ids {
		if node, ok := v.pool[id]; ok {
			_ = node.send(data)
		}
	}
}

func (v *ViewCache) IsEmpty() bool {
	v.RLock()
	defer v.RUnlock()

	return len(v.pool) == 0
}
