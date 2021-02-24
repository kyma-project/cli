---
title: Debug function
type: Tutorials
---

This tutorial shows how to use external IDE to debug function in Kyma

## Steps

Follow these steps:

Follows these steps:

<div tabs name="steps" group="expose-function">
  <details>
  <summary label="vsc">
  Visual Studio Code
  </summary>

1. Open VSC in location where you have a file with definition Function. 
2. Create directory `.vscode`
3. In the `.vscode` create file `lunch.json` 
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
Goland
</summary>

1. Open Goland in location where you have a file with definition Function.
2. Choose option Add Configuration...
3. Add new configuration `Attach to Node.js/Chrome` with options:
- host: `localhost`
- port: `9229`

    </details>
</div>