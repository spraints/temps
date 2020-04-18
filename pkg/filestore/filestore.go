package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spraints/temps/pkg/memorystore"
	"github.com/spraints/temps/pkg/types"
)

func New(file string) *FileStore {
	f := &FileStore{
		file:     file,
		snapshot: memorystore.New(),
	}
	f.tryRead()
	return f
}

type FileStore struct {
	snapshot *memorystore.MemStore

	lock sync.Mutex
	file string
}

func (f *FileStore) All() ([]types.Measurement, error) {
	return f.snapshot.All()
}

func (f *FileStore) Put(meas types.Measurement) error {
	// This doesn't assure other writers aren't accessing the file, but at least assures thread-safety here.
	f.lock.Lock()
	defer f.lock.Unlock()

	if err := f.snapshot.Put(meas); err != nil {
		return err
	}

	vals, err := f.snapshot.All()
	if err != nil {
		return err
	}

	fh, err := os.Create(f.file)
	if err != nil {
		return fmt.Errorf("opening FileStore: %s: %w", f.file, err)
	}
	defer fh.Close()

	err = json.NewEncoder(fh).Encode(&schema{
		Version:      1,
		Temperatures: ser(vals),
	})
	if err != nil {
		return fmt.Errorf("couldn't serialize measurements: %w", err)
	}

	return nil
}

func (f *FileStore) tryRead() {
	fh, err := os.Open(f.file)
	if err != nil {
		return
	}
	defer fh.Close()

	var data schema
	err = json.NewDecoder(fh).Decode(&data)
	if err != nil {
		return
	}

	if data.Version != 1 {
		return
	}

	for _, t := range data.Temperatures {
		f.snapshot.Put(t.toMeas())
	}
}

type schema struct {
	Version      int                 `json:"version"`
	Temperatures []temperatureschema `json:"temperatures"`
}

type temperatureschema struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	C    types.Celsius `json:"c"`
	At   time.Time     `json:"at"`
}

func (t *temperatureschema) toMeas() types.Measurement {
	return types.Measurement{
		ID:          t.ID,
		Name:        t.Name,
		Temperature: t.C,
		MeasuredAt:  t.At,
	}
}

func ser(all []types.Measurement) []temperatureschema {
	res := make([]temperatureschema, 0, len(all))
	for _, meas := range all {
		res = append(res, temperatureschema{
			ID:   meas.ID,
			Name: meas.Name,
			C:    meas.Temperature.Celsius(),
			At:   meas.MeasuredAt,
		})
	}
	return res
}
