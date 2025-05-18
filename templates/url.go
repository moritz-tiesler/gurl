package templates

import (
	"bytes"
	"html/template"
)

var templateString string = `
    <input id="url-result" class="copy-link-input" type="text" value={{ .Value }} readonly>
`

type URL struct {
	Value string
}

func (u URL) Render() []byte {
	tmpl := template.Must(
		template.New("url").Parse(templateString),
	)

	var out bytes.Buffer

	err := tmpl.Execute(&out, u)
	if err != nil {
		panic(err)
	}

	return out.Bytes()
}
