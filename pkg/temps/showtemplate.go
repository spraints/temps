package temps

import (
	"html/template"
	"io"
)

const showTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <title>Temperatures around the farm</title>
  </head>
  <body>
    <h1>Temperatures (°F) around the farm</h1>
    <table>
      {{range .}}
      <tr><th>{{.Label}}</th><td>{{.Temperature}}°F</td></tr>
      {{end}}
    </table>
  </body>
</html>
`

var compiledShowTemplate *template.Template = template.Must(template.New("show").Parse(showTemplate))

type temp struct {
	Label       string
	Temperature fahrenheit
}

func renderShowTemplateFahrenheit(w io.Writer, temps []temp) error {
	return compiledShowTemplate.Execute(w, temps)
}
