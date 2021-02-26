---
title: Sudden container failure
type: Troubleshooting
---

The container can suddenly fail when you use the `kyma run function` command with these flags:
- runtime=Nodejs12 or runtime=Nodejs10
- `debug=true`
- `hot-deploy=true`

In such a case, you can see the `[nodemon] app crashed` message in the container's logs.
`[nodemon] app crashed`

If you use kyma in Kubernetes, K8S itself should run function in container.
If you use kyma without Kubernetes, you have to run the container yourself. 
