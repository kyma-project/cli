---
title: Debug a Function
type: Tutorials
---

This tutorial shows how to use an external IDE to debug a Function in Kyma CLI.

## Steps


Follows these steps:

<div tabs name="steps" group="expose-function">
  <details>
  <summary label="vsc">
  Visual Studio Code
  </summary>

1. In VSC, navigate to the location of the file with the Function definition.
2. Create the `.vscode` directory.
3. In the `.vscode` directory, create the `lunch.json` file with this content:
  ```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "attach",
      "type": "node",
      "request": "attach",
      "port": 9229,
      "address": "localhost",
      "localRoot": "${workspaceFolder}/kubeless",
      "remoteRoot": "/kubeless",
      "restart": true,
      "protocol": "inspector",
      "timeout": 1000
    }
  ]
}
  ```


</details>
<details>
<summary label="goland">
GoLand
</summary>

1. In GoLand, navigate to the location of the file with the Function definition.
2. Choose the **Add Configuration...** option.
3. Add new **Attach to Node.js/Chrome** configuration with these options:
- Host: `localhost`
- port: `9229`

    </details>
</div>
