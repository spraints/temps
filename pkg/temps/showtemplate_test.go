package temps

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spraints/temps/pkg/units"
)

func TestShowTemplate(t *testing.T) {
	r := func(t *testing.T, label string, c float64) string {
		var b bytes.Buffer
		if err := showHTML(&b, []temp{
			{Label: label, Temperature: units.Celsius(c)},
		}); err != nil {
			require.NoError(t, err)
		}
		return b.String()
	}

	assert.Equal(t, `
<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Temperatures around the farm</title>
    <style>
      .temp-label { text-align: left; }
      .temp { text-align: right; }
      th, td { padding: 0.5em; }
      body { font-size: x-large; font-family: sans-serif; }
    </style>
  </head>
  <body>
    <h1>Temperatures around the farm</h1>
    <table>
      <tr><th class="temp-label">Example</th><td class="temp">32°F</td><td class="temp">0°C</tr>
    </table>
  </body>
</html>
`, r(t, "Example", 0))
}
