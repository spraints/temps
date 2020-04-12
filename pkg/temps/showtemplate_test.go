package temps

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spraints/temps/pkg/units"
)

func TestShowHTML(t *testing.T) {
	var buf bytes.Buffer
	temps := []temp{
		{Label: "Example", Temperature: units.Celsius(0)},
	}
	require.NoError(t, showHTML(&buf, "ws://thing/live", temps))
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
    <div class="js-temp-table" data-ws-url="ws://thing/live">
      <table>
        <tr><th class="temp-label">Example</th><td class="temp">32°F</td><td class="temp">0°C</tr>
      </table>
    </div>
    <script defer type="text/javascript" src="/app.js" charset="utf-8"></script>
  </body>
</html>
`, buf.String())
}

func TestShowFrag(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, showFrag(&buf, []temp{
		{Label: "Example", Temperature: units.Celsius(0)},
	}))
	assert.Equal(t, `<table>
        <tr><th class="temp-label">Example</th><td class="temp">32°F</td><td class="temp">0°C</tr>
      </table>`, buf.String())
}