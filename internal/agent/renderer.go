package agent

import (
	"bytes"
	"text/template"
)

func RenderTemplate(body map[string]string, params map[string]string) (map[string]string, error) {
	out := make(map[string]string)

	for k, v := range body {
		tpl, err := template.New("body").Parse(v)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, params); err != nil {
			return nil, err
		}
		out[k] = buf.String()
	}

	return out, nil
}
