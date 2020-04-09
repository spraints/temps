package temps

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortableSensors(t *testing.T) {
	slice := sensorSlice(nil)
	sort.Sort(slice)
	assert.Equal(t, sensorSlice(nil), slice)

	slice = sensorSlice([]*sensor{
		{Name: "C"},
		{Name: "A"},
		{Name: "B"},
	})
	sort.Sort(slice)
	assert.Equal(t, sensorSlice([]*sensor{
		{Name: "A"},
		{Name: "B"},
		{Name: "C"},
	}), slice)
}
