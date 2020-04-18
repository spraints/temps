package filestore

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spraints/temps/pkg/types"
)

func TestFileStore(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-file-store-")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	m1 := types.Measurement{
		ID:          "aaa",
		Name:        "A",
		MeasuredAt:  time.Unix(1587250000, 0),
		Temperature: types.Celsius(10),
	}

	m2 := types.Measurement{
		ID:          "aaa",
		Name:        "A2",
		MeasuredAt:  time.Unix(1587251000, 0),
		Temperature: types.Celsius(15),
	}

	m3 := types.Measurement{
		ID:          "bbb",
		Name:        "B",
		MeasuredAt:  time.Unix(1587252000, 0),
		Temperature: types.Celsius(20),
	}

	t.Run("Put and All work", func(t *testing.T) {
		var all []types.Measurement
		var err error

		fs := New(tmpfile.Name())

		require.NoError(t, fs.Put(m1))

		all, err = fs.All()
		require.NoError(t, err)
		assert.Equal(t, []types.Measurement{m1}, all, "m1 was added")

		require.NoError(t, fs.Put(m2))

		all, err = fs.All()
		require.NoError(t, err)
		assert.ElementsMatch(t, []types.Measurement{m2}, all, "m2 replaced an earlier entry")

		require.NoError(t, fs.Put(m3))

		all, err = fs.All()
		require.NoError(t, err)
		assert.ElementsMatch(t, []types.Measurement{m2, m3}, all, "m3 was added")
	})

	t.Run("read an old file", func(t *testing.T) {
		var all []types.Measurement
		var err error

		fs := New(tmpfile.Name())

		all, err = fs.All()
		require.NoError(t, err)
		assert.ElementsMatch(t, []types.Measurement{m2, m3}, all, "file was reread")
	})
}
