package workspace

const handlerJs = `module.exports = {
    main: function (event, context) {
        return 'Hello Serverless'
    }
}`

const packageJSON = `{
  "name": "{{ .Name }}",
  "version": "0.0.1",
  "dependencies": {}
}`

const (
	FileNameHandlerJs   FileName = "handler.js"
	FileNamePackageJSON FileName = "package.json"
)

var workspaceNodeJs = workspace{
	newTemplatedFile(handlerJs, FileNameHandlerJs),
	newTemplatedFile(packageJSON, FileNamePackageJSON),
}
