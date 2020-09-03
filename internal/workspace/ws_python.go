package workspace

const handlerPython = `def foo(event, context):
    return "hello world"`

const (
	FileNameHandlerPy       FileName = "handler.py"
	FileNameRequirementsTxt FileName = "requirements.txt"
)

var workspacePython = workspace{
	newTemplatedFile(handlerPython, FileNameHandlerPy),
	newEmptyFile(FileNameRequirementsTxt),
}
