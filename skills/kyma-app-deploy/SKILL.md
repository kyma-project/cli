---
name: kyma-app-deploy
description: Build and deploy an application to an open-source Kyma cluster with kyma app push (v3) - from local source, a Dockerfile, or a pre-built image, with exposure via APIRule, Istio injection, and env/secret/config injection. Use when the user wants to push, deploy, or expose an app on Kyma. Trigger keywords - kyma app push, deploy app to kyma, push app, expose app kyma, kyma apirule app, kyma deploy application.
---

# Skill: kyma-app-deploy

Build and deploy an application to a Kyma-enabled Kubernetes cluster with the
open-source **Kyma CLI (v3)** command `kyma app push` — the only subcommand under
`kyma app`, and the most flag-heavy command in the CLI.

Kyma CLI v3 is a Go-based, Apache-2.0 tool from the open-source Kyma project
([github.com/kyma-project/cli](https://github.com/kyma-project/cli)). It is a
full rewrite of v2 — **do not assume v2 syntax.**

## Scope and guardrails

This skill covers **only the open-source Kyma project**. When using it:

- ✅ Use open-source Kyma (`kyma-project/*`), open-source modules, the Kyma CLI,
  and standard `kubectl`.
- ❌ Do **not** encode SAP-managed (SKR) specifics, BTP cockpit flows, internal
  endpoints or URLs, SKR configuration, or any SAP-internal knowledge.
- ❌ Never hardcode registry credentials or SAP-internal image registries in
  examples — use placeholders like `<registry>/<project>/my-app:latest`.

## Prerequisites

- The `kyma` CLI installed and on `PATH` (v3). Confirm with `kyma version`.
- `kubectl` installed and its current context pointing at the target cluster.
- ⚠️ **Building from source or a Dockerfile** (`--code-path` / `--dockerfile`)
  requires an **in-cluster Docker registry** — the `docker-registry` community
  module. Without it, the build fails with `failed to load the in-cluster
  registry configuration`; the CLI's own hint is to add it:

  ```bash
  kyma module add docker-registry -c experimental --default-config-cr
  ```

  On a cluster **not** managed by Kyma Lifecycle Manager, `module add` needs the
  KLM dependencies first — let the `kyma-module-lifecycle` skill drive the full
  enable sequence. Deploying a **pre-built image** (`--image`) needs no
  in-cluster registry.
- ⚠️ **`--expose`** requires the **api-gateway** module — specifically the
  **APIRule CRD in its `v2` version**. Pre-check it before pushing:

  ```bash
  kubectl get crd apirules.gateway.kyma-project.io -o jsonpath='{.spec.versions[*].name}'
  ```

  **`--istio-inject`** requires the **Istio** module. Enable both via the
  `kyma-module-lifecycle` skill or the push will fail.

Before running a command whose exact flags you are unsure of, inspect help output
rather than guessing — the installed version is the source of truth:

```bash
kyma app push --help
```

## Instructions

Image import into the in-cluster registry happens **automatically** during
`push` — there is no separate `import-image` command. Pick the build path that
matches what the user has.

### 1. Build and deploy from local source

Builds via Cloud Native Buildpacks. `--code-path` needs a real buildable source
directory (not just docs).

```bash
kyma app push --name my-app --code-path .
```

### 2. Build from a specific Dockerfile

```bash
kyma app push --name my-app --dockerfile ./Dockerfile --dockerfile-context . [--build-tag <tag>]
```

- `--dockerfile-context` sets the build context directory.
- `--build-tag` assigns a custom image tag for the build (applies only to
  `--code-path` and `--dockerfile` builds, not to `--image`).

### 3. Deploy a pre-built image (no build)

```bash
kyma app push --name my-app --image <registry>/<project>/my-app:latest
```

- `--image` — the image reference to deploy (no build step runs).
- `--image-pull-secret <secret>` — names a Kubernetes Secret holding registry
  credentials, for pulling from a **private** registry. Create the Secret first
  (`kubectl create secret docker-registry ...`); the flag references it by name
  so credentials never appear on the command line.

### 4. Expose the app and enable Istio injection

```bash
kyma app push --name my-app --code-path . --container-port 8080 --expose --istio-inject=true
```

- `--container-port` — integer; the port your container listens on.
- `--expose` — creates an **APIRule** to expose the app externally (needs the
  api-gateway module).
- `--istio-inject` — boolean; `--istio-inject=true` enables the Istio sidecar,
  `--istio-inject=false` disables it (needs the Istio module).

### 5. Inject configuration and secrets

```bash
kyma app push --name my-app --code-path . \
  --env NAME=VALUE \
  --env-from-configmap my-configmap \
  --env-from-secret my-secret \
  --env-from-file ./my-env-file \
  --mount-secret my-secret:key=/app/secret:ro \
  --mount-config my-configmap:key=/app/config \
  --mount-service-binding-secret my-sb-secret
```

- `--env` — a single `NAME=VALUE` variable (repeatable). Overrides same-named
  vars from other sources.
- `--env-from-configmap` / `--env-from-secret` — import env vars from a ConfigMap
  or Secret. Format: `RESOURCE[:ENVS_PREFIX]` to fetch **all** keys (optionally
  prefixed), or `ENV_NAME=RESOURCE:RESOURCE_KEY` for a single key — e.g.
  `--env-from-configmap my-configmap:CONFIG_` (all keys, prefixed) or
  `--env-from-secret MY_VAR=my-secret:key3` (single key).
- `--env-from-file` — load env vars from a file. Format:
  `ENV_NAME=FILE_PATH:FILE_KEY` for a single key, or `FILE_PATH[:ENVS_PREFIX]`
  for all keys.
- `--mount-secret` / `--mount-config` — mount into the container. Three accepted
  formats:
  - shorthand `NAME:KEY=MOUNT_PATH[:ro]` (`:ro` = read-only), e.g.
    `my-secret:key=/app/secret:ro`;
  - normal `name=NAME,path=MOUNT_PATH,key=KEY,ro=true` (`key`/`ro` optional);
  - legacy `NAME` — mounts the whole resource at `/bindings/secret-<NAME>` (or
    `/bindings/configmap-<NAME>`).
  Path traversal (`..`) in the mount path is prohibited.
- `--mount-service-binding-secret` — mount a service-binding secret (read-only,
  under `/bindings/secret-<NAME>`).

Other useful flags: `-n, --namespace` (target namespace, default `default`);
`--dockerfile-build-arg` (build args for Dockerfile builds, repeatable);
`--insecure` (disable the app's SecurityContext); `-q, --quiet` (print only the
exposed URL).

### 6. Verify and debug the push

```bash
# The app deploys to the namespace given by -n/--namespace on the push
# (default: "default"). Use the same namespace here.
kubectl get deployment my-app -n <namespace>
kubectl get apirule -n <namespace>          # if you used --expose
kubectl get pods -n <namespace>
```

If the app doesn't come up, first read what the Deployment is telling you:

- **`kyma app push` returned success and the Deployment exists, but Pods are not
  `Ready` (e.g. `0/1`)** — the manifest applied but the container isn't starting.
  Check `kubectl describe pod` / `kubectl logs` for the reason (image pull error,
  crash loop, missing secret/config mount), then hand off to `kyma-troubleshoot`.
  On local clusters (k3d/kind), Pods can also stick non-`Ready` due to
  environment quirks (e.g. an `AppArmor` pod status) rather than an app fault. If
  `kubectl describe pod` shows an AppArmor-related message, it is the local
  runtime, not your app — recreate the app on a cluster whose nodes allow the
  default AppArmor profile (or an `unconfined` override), rather than debugging
  the container.
- **The push itself failed** — re-read the CLI error; it usually names the missing
  prerequisite (in-cluster registry, api-gateway/Istio module).

For deeper diagnosis, hand off to the `kyma-troubleshoot` skill
(`kyma alpha diagnose logs --module ...`, `kyma alpha diagnose istio`).

## Safety

- **`kyma app push` is not destructive to the cluster**, but pushing with a name
  that already exists **overwrites** that app. Confirm the target `--name`,
  namespace, and `kubectl` context before pushing to anything shared.
- Keep credentials out of `--env`/`--image` on the command line; prefer
  `--env-from-secret` / `--mount-secret` so secrets aren't exposed in shell
  history or logs.

## Related skills

- **`kyma-module-lifecycle`** — enable the api-gateway and Istio modules that
  `--expose` and `--istio-inject` depend on.
- **`kyma-troubleshoot`** — diagnose a failing push.
- **`kyma-cli`** — general Kyma CLI reference.

## Notes

- `kyma app push` reflects Kyma CLI v3. If a flag is rejected, check
  `kyma app push --help` on the installed version — a tagged release can lag the
  latest development build.
