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
      <tr><th>Outside</th><td>{{.Outdoor | printf "%0.2f"}}°F</td></tr>
      {{range .Sensors}}
      <tr><th>{{.Name}}</th><td>{{.Temperature | printf "%0.2f"}}°F</td></tr>
      {{end}}
    </table>
  </body>
</html>
`

var compiledShowTemplate *template.Template = template.Must(template.New("show").Parse(showTemplate))

func renderShowTemplateFahrenheit(w io.Writer, sensors []sensor, outdoorTemp fahrenheit) error {
	data := struct {
		Sensors []sensor
		Outdoor fahrenheit
	}{
		Sensors: sensors,
		Outdoor: outdoorTemp,
	}
	return compiledShowTemplate.Execute(w, &data)
}
