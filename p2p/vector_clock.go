package p2p

import (
	"encoding/json"
	"fmt"
)

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

func (vc VectorClock) String() string {
	data, _ := json.Marshal(vc)
	return string(data)
}

func ParseVectorClock(vcStr string) (VectorClock, error) {
	var vc VectorClock
	err := json.Unmarshal([]byte(vcStr), &vc)
	if err != nil {
		return nil, fmt.Errorf("error parsing vector clock: %v", err)
	}
	return vc, nil
}
