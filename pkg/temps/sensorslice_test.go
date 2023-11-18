package temps

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spraints/temps/pkg/types"
)

func TestSortableSensors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		slice := sensorSlice(nil)
		sort.Sort(slice)
		assert.Equal(t, sensorSlice(nil), slice)
	})

	t.Run("sort by name", func(t *testing.T) {
		slice := sensorSlice([]types.Measurement{
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
	})

	t.Run("with numbers in names", func(t *testing.T) {
		slice := sensorSlice([]types.Measurement{
			{Name: "1 C", ID: "01"},
			{Name: "10 A", ID: "zzyy"},
			{Name: "2 B", ID: "ab"},
			{Name: "Z", ID: outsideID},
		})
		sort.Sort(slice)
		// outsideID should still be first, other IDs should be sorted numerically.
		assert.Equal(t, sensorSlice([]types.Measurement{
			{Name: "Z", ID: outsideID},
			{Name: "1 C", ID: "01"},
			{Name: "2 B", ID: "ab"},
			{Name: "10 A", ID: "zzyy"},
		}), slice)
	})
}

func TestGetIntFromName(t *testing.T) {
	shouldWork := []struct {
		name     string
		expected int
	}{
		{"1", 1},
		{"11", 11},
		{"1 hi", 1},
		{"99999 @#$%^&*", 99999},
	}

	shouldNotWork := []string{
		"ok1",
		"no",
	}

	for _, ex := range shouldWork {
		slice := sensorSlice([]types.Measurement{{Name: ex.name}})
		n, ok := slice.getIntFromName(0)
		if assert.Truef(t, ok, "should find number in %q", ex.name) {
			assert.Equalf(t, ex.expected, n, "number in %q", ex.name)
		}
	}

	for _, ex := range shouldNotWork {
		slice := sensorSlice([]types.Measurement{{Name: ex}})
		n, ok := slice.getIntFromName(0)
		assert.Falsef(t, ok, "should not find number in %q but found %d", ex, n)
	}
}
