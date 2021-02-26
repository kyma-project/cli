---
title: Debugger in strange location
type: Troubleshooting
---

If you debug your Function (in `runtime=Nodejs12` or `runtime=Nodejs10`) and you set a breakpoint in the first line of the main Function, the debugger can stop at dependencies.

You shouldn't debug first line, you can add for example comment in first line. 
