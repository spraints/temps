package temps

import (
	"fmt"
	"html/template"
	"io"

	"github.com/spraints/temps/pkg/units"
)

const showTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <title>Temperatures around the farm</title>
    <style>
      .temp-label { text-align: left; }
      .temp { text-align: right; }
    </style>
  </head>
  <body>
    <h1>Temperatures around the farm</h1>
    <table>
      {{range .}}
      <tr><th class="temp-label">{{.Label}}</th><td class="temp">{{.Temperature | f}}°F</td><td class="temp">{{.Temperature | c}}°C</tr>
      {{end}}
    </table>
  </body>
</html>
`

var compiledShowTemplate *template.Template = func() *template.Template {
	tempFuncs := map[string]interface{}{
		"c": func(t units.Temperature) string { return fmt.Sprintf("%0.0f", t.Celsius()) },
		"f": func(t units.Temperature) string { return fmt.Sprintf("%0.0f", t.Fahrenheit()) },
	}
	return template.Must(template.New("show").Funcs(tempFuncs).Parse(showTemplate))
}()

type temp struct {
	Label       string
	Temperature units.Temperature
}

func showHTML(w io.Writer, temps []temp) error {
	return compiledShowTemplate.Execute(w, temps)
}
