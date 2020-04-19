package memorystore

import (
	"sync"

	"github.com/spraints/temps/pkg/types"
)

func New() *MemStore {
	return &MemStore{
		current: map[string]types.Measurement{},
	}
}

type MemStore struct {
	lock    sync.RWMutex
	current map[string]types.Measurement
}

func (m *MemStore) All() ([]types.Measurement, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	res := make([]types.Measurement, 0, len(m.current))
	for _, t := range m.current {
		res = append(res, t)
	}
	return res, nil
}

func (m *MemStore) Put(meas types.Measurement) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.current[meas.ID] = meas

	return nil
}
