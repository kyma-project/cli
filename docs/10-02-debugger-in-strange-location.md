---
title: Debugger in strange location
type: Troubleshooting
---

If you debug your function (runtime=Nodejs12 or runtime=Nodejs10) and you set breakpoint in first line of main function, the debugger can try stop in dependencies.

You shouldn't debug first line, you can add for example comment in first line. 