package p2p

type VectorClock map[string]int

func (vc VectorClock) Increment(nodeID string) {
	vc[nodeID]++
}

func (vc VectorClock) Update(other VectorClock) {
	for node, ts := range other {
		if current, found := vc[node]; !found || ts > current {
			vc[node] = ts
		}
	}
}

func (vc VectorClock) Copy() VectorClock {
	newVC := make(VectorClock)
	for node, ts := range vc {
		newVC[node] = ts
	}
	return newVC
}
