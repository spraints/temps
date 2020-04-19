package temps

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spraints/temps/pkg/types"
)

func TestSortableSensors(t *testing.T) {
	slice := sensorSlice(nil)
	sort.Sort(slice)
	assert.Equal(t, sensorSlice(nil), slice)

	slice = sensorSlice([]types.Measurement{
		{Name: "C", ID: "01"},
		{Name: "A", ID: "zzyy"},
		{Name: "B", ID: "ab"},
		{Name: "Z", ID: outsideID},
	})
	sort.Sort(slice)
	// outsideID should be first, other IDs shouldn't matter.
	assert.Equal(t, sensorSlice([]types.Measurement{
		{Name: "Z", ID: outsideID},
		{Name: "A", ID: "zzyy"},
		{Name: "B", ID: "ab"},
		{Name: "C", ID: "01"},
	}), slice)
}
