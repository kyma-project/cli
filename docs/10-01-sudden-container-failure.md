---
title: Sudden container failure
type: Troubleshooting
---

The container can suddenly fail when you use the `kyma run function` command with these flags:
- runtime=Nodejs12 or runtime=Nodejs10
- `debug=true`
- `hot-deploy=true`

the container can sudden failure. In container's logs you can see a message:
`[nodemon] app crashed`

If you use kyma in Kubernetes, K8S itself should run function in container.
If you use kyma without Kubernetes, you have to run the container yourself. 
