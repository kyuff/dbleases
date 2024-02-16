package tmpl

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

func Parse(fsys embed.FS, data any) (map[string]string, error) {
	archive, err := template.ParseFS(fsys, "**/*.tmpl")
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, template := range archive.Templates() {
		var buf bytes.Buffer
		err := template.Execute(&buf, data)
		if err != nil {
			return nil, fmt.Errorf("parse template %q: %w", template.Name(), err)
		}

		result[template.Name()] = buf.String()
	}

	return result, nil
}
