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
    <div class="js-temp-table" {{.WSAttr}}>
      {{define "table"}}<table>{{range .}}
        <tr><th class="temp-label">{{.Label}}</th><td class="temp">{{.Temperature | f}}°F</td><td class="temp">{{.Temperature | c}}°C</tr>{{end}}
      </table>{{end}}{{template "table" .Temps}}
    </div>
    <script defer type="text/javascript" src="/app.js" charset="utf-8"></script>
  </body>
</html>
`

type showData struct {
	WSAttr template.HTMLAttr
	Temps  []temp
}

type temp struct {
	Label       string
	Temperature units.Temperature
}

var compiledShowTemplate *template.Template = func() *template.Template {
	tempFuncs := map[string]interface{}{
		"c": func(t units.Temperature) string { return fmt.Sprintf("%0.0f", t.Celsius()) },
		"f": func(t units.Temperature) string { return fmt.Sprintf("%0.0f", t.Fahrenheit()) },
	}
	return template.Must(template.New("show").Funcs(tempFuncs).Parse(showTemplate))
}()

func showHTML(w io.Writer, wsURL string, temps []temp) error {
	return compiledShowTemplate.Execute(w, &showData{
		WSAttr: template.HTMLAttr("data-ws-url=\"" + wsURL + "\""),
		Temps:  temps,
	})
}

func showFrag(w io.Writer, temps []temp) error {
	return compiledShowTemplate.ExecuteTemplate(w, "table", temps)
}
