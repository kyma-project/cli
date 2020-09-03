package workspace

import (
	"io"
	"text/template"
)

var _ file = templatedFile{}

type templatedFile struct {
	name FileName
	tpl  string
}

func (t templatedFile) fileName() string {
	return string(t.name)
}

func (t templatedFile) write(writer io.Writer, cfg interface{}) error {
	tpl, err := template.New("templatedFile").Parse(t.tpl)
	if err != nil {
		return err
	}

	return tpl.Execute(writer, cfg)
}

func newTemplatedFile(tpl string, name FileName) file {
	return &templatedFile{
		tpl:  tpl,
		name: name,
	}
}
