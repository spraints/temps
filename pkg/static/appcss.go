package static

const AppCSS = "/app.css"

const appCSS = `
.temp-label {
  text-align: left;
}
.temp {
  text-align: right;
}
.temp-date {
  color: #ccc;
}
@media only screen and (max-width: 580px) {
  .temp-date {
    display: none;
  }
}
th, td {
  padding: 0.5em;
}

body {
  background-color: white;
  color: black;
  font-size: x-large;
  font-family: sans-serif;
}
`
