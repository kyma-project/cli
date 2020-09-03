package workspace

import "io"

type emptyFile struct {
	name FileName
}

func (e emptyFile) write(_ io.Writer, _ interface{}) error {
	return nil
}

func (e emptyFile) fileName() string {
	return string(e.name)
}

func newEmptyFile(name FileName) file {
	return emptyFile{
		name: name,
	}
}
