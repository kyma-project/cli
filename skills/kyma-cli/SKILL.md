---
name: kyma-cli
description: Use the open-source Kyma CLI (v3) to manage Kyma modules, deploy applications, provision clusters, and run diagnostics on a Kubernetes cluster. Use when the user wants to install Kyma, enable or disable a module, push an app, open the Kyma dashboard, or troubleshoot a Kyma cluster. Trigger keywords - kyma cli, kyma module, kyma app push, kyma deploy, enable kyma module, kyma dashboard, kyma diagnose, kyma troubleshoot.
---

# Skill: kyma-cli

Use the open-source **Kyma CLI (v3)** to interact with a Kyma-enabled Kubernetes
cluster: manage modules, deploy applications, authorize subjects with RBAC,
generate cluster access, open the dashboard, and run diagnostics.

Kyma CLI v3 is a Go-based, Apache-2.0 tool from the open-source Kyma project
([github.com/kyma-project/cli](https://github.com/kyma-project/cli)). It is a
full rewrite of v2 — the v2 top-level `provision`, `deploy`, and `import`
commands no longer exist. **Do not assume v2 syntax.**

## Scope and guardrails

This skill covers **only the open-source Kyma project**. When using it:

- ✅ Use open-source Kyma (`kyma-project/*`), open-source modules, the Kyma CLI,
  and standard `kubectl`.
- ❌ Do **not** encode SAP-managed (SKR) specifics, BTP cockpit flows, internal
  endpoints or URLs, SKR configuration, or any SAP-internal knowledge.
- ⚠️ `kyma alpha provision` and `kyma alpha hana map` interact with SAP BTP. You
  may mention that they exist, but never hardcode credentials, subaccount
  details, or internal provisioning parameters.

## Prerequisites

- The `kyma` CLI installed and on `PATH` (v3).
- `kubectl` installed and its current context pointing at the target cluster.

### Install

```bash
# Homebrew (stable)
brew install kyma-cli

# Official install scripts (pick one)
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_stable.sh" | sh -
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_latest.sh" | sh -
curl -sL "https://raw.githubusercontent.com/kyma-project/cli/refs/heads/main/hack/install_cli_nightly.sh" | sh -
```

Confirm the version:

```bash
kyma version
```

## Instructions

### 1. Orient before acting

The installed CLI version is the source of truth. Before running a command whose
exact flags you are unsure of, inspect help output rather than guessing:

```bash
kyma --help
kyma <group> --help      # e.g. kyma module --help
kyma help <command>
```

### 2. Manage modules

Modules live under the **non-alpha** `kyma module` group (7 subcommands: `catalog`,
`list`, `add`, `delete`, `manage`, `unmanage`, `pull`).

```bash
# Discover
kyma module catalog                 # all available modules (-o table|json|yaml)
kyma module list                    # modules installed on the cluster

# Enable / disable
kyma module add <module> -c <channel> [--default-config-cr | --config-cr-path <path>] [--auto-approve]
kyma module delete <module> [--auto-approve]

# Managed state
kyma module manage <module> [--policy CreateAndDelete|Ignore]   # let Kyma reconcile the module
kyma module unmanage <module>       # stop reconciling; leave resources in place

# Community modules
kyma module pull <module-name> [--namespace <ns>] [--version <v>]
kyma module add <namespace>/<module-template-name>   # community module format
```

> ⚠️ **Important:** `kyma alpha module` only exposes `catalog` and `pull`. The
> `add`, `delete`, `list`, `manage`, and `unmanage` subcommands exist **only**
> under the non-alpha `kyma module` group — never prefix them with `alpha`.

> ℹ️ **Cluster prerequisite (Kyma Lifecycle Manager):** `add`, `list`, `delete`,
> `manage`, and `unmanage` require a cluster managed by Kyma Lifecycle Manager
> (KLM). On a plain cluster they fail with a "not managed by KLM" error. Only
> `catalog` and `pull` work without KLM (use `kyma module pull --force` to
> approve dependency installation on a non-KLM cluster) — `pull` is the CLI's
> own suggested recovery path when `add`/`list`/`delete` fail with that error.

> ℹ️ **Catalog is cluster-dependent:** available module names and channels depend
> on the catalog the target cluster exposes. Run `kyma module catalog` to confirm
> what is available before `kyma module add <module> -c <channel>` — a given name
> (e.g. `istio`) may not appear on every cluster.

To inspect the underlying `ModuleTemplate` custom resource directly with
`kubectl` (not a `kyma` command):

```bash
kubectl get crd moduletemplates.operator.kyma-project.io -o yaml
```

### 3. Deploy an application

`kyma app push` is the only subcommand under `kyma app`. Image import into the
in-cluster registry happens automatically during `push` — there is no separate
`import-image` command.

```bash
# Build and deploy from local source
kyma app push --name my-app --code-path .

# Build from a specific Dockerfile
kyma app push --name my-app --dockerfile ./Dockerfile --dockerfile-context .

# Deploy a pre-built image (no build)
kyma app push --name my-app --image <registry>/<project>/my-app:latest

# Expose the app via an APIRule and enable Istio sidecar injection
kyma app push --name my-app --code-path . --container-port 8080 --expose --istio-inject=true

# Inject configuration and secrets
kyma app push --name my-app --code-path . \
  --env NAME=VALUE \
  --env-from-configmap my-configmap \
  --env-from-secret my-secret \
  --mount-secret my-secret:key=/app/secret:ro \
  --mount-config my-configmap:key=/app/config
```

`kyma app push` uses Kubernetes Server-Side Apply — running it again with the
same `--name` **updates the existing Deployment, Service, and APIRule in-place**.
There is no `--replace` flag. The new spec fully replaces the previous one, so
env vars, mounts, and image from the previous run are replaced, not merged.

### 4. Authorize RBAC subjects

`kyma alpha authorize` creates RoleBindings and ClusterRoleBindings for users,
groups, service accounts, and GitHub Actions workflows. Use this when a developer
or pipeline hits a `403 Forbidden` error and needs access to a namespace or the
whole cluster.

```bash
# Grant a user cluster-wide access (ClusterRoleBinding)
kyma alpha authorize user --name alice --clusterrole kyma-admin --cluster-wide

# Grant multiple users namespaced access (RoleBinding)
kyma alpha authorize user --name alice,bob --role view --namespace dev

# Grant a group namespaced access
kyma alpha authorize group --name developers --role edit --namespace dev

# Grant a service account cluster-wide access
kyma alpha authorize serviceaccount --name ci-bot --clusterrole cluster-admin --cluster-wide

# Bind a service account in a different namespace
kyma alpha authorize serviceaccount --name remote-sa --sa-namespace tools --clusterrole view --namespace dev

# Authorize a GitHub Actions workflow via OIDC trust (no long-lived credentials)
kyma alpha authorize repository --repository kyma-project/cli --client-id repo-ci-client \
  --clusterrole kyma-admin --cluster-wide

# Preview without applying (dry-run)
kyma alpha authorize user --name bob --role operator --namespace ops --dry-run -o yaml
```

Key flags:
- `--role` / `--clusterrole` — exactly one required; `--role` creates a
  RoleBinding, `--clusterrole` creates a RoleBinding referencing a ClusterRole
  (or a ClusterRoleBinding when combined with `--cluster-wide`).
- `--cluster-wide` — creates a ClusterRoleBinding; requires `--clusterrole`.
- `--namespace` — required with `--role`; the namespace for the RoleBinding.
- `--force` — overwrites an existing binding of the same name without prompting.
- `--dry-run -o yaml|json` — prints the manifest without applying it.
- `--binding-name` — custom name for the generated binding; auto-generated from
  role and subject name if omitted.

> ⚠️ `kyma alpha authorize` is under the `alpha` group — its API may change
> between releases. Check `kyma alpha authorize --help` on the installed version.

### 5. Access the cluster and open the dashboard

```bash
# Run the Kyma dashboard locally in Docker
kyma alpha dashboard start -o -p 8000     # -o opens the browser
kyma alpha dashboard stop

# Generate a kubeconfig (Service Account or OIDC based)
kyma alpha kubeconfig generate --serviceaccount <sa> --namespace <ns> --clusterrole <role>

# Shell completion — load for the current session
source <(kyma completion zsh)

# Install permanently (macOS, Homebrew):
kyma completion zsh > $(brew --prefix)/share/zsh/site-functions/_kyma
# Start a new shell after running this.

# Install permanently (Linux):
kyma completion zsh > "${fpath[1]}/_kyma"
# Start a new shell after running this.
# (Replace zsh with bash, fish, or powershell as needed.)
```

### 6. Troubleshoot

The `kyma alpha diagnose` group is the diagnostic workhorse:

```bash
kyma alpha diagnose cluster                 # cluster health and configuration
kyma alpha diagnose istio                   # Istio configuration checks
kyma alpha diagnose logs --module <module> [--since 1h] [--lines 200] [--timeout 30s] [--strict]
```

Combine with `kyma module list` (to see module state) and standard `kubectl`
inspection of the relevant namespaces.

## Safety: destructive operations

Module deletion and unmanaging are destructive. Follow these rules:

- **Always confirm before deleting.** Run `kyma module list` first, show the user
  the exact module(s) that would be affected, and get explicit confirmation.
- **There is no bulk "delete all modules" command** — deletion is per-module. If
  a user says "delete everything, don't ask," do **not** loop `delete` blindly:
  list the modules and confirm the specific set to remove, one by one.
- **`--auto-approve` bypasses the CLI's built-in approval prompt** (and the
  community-module SLA prompt). Use it only non-interactively (e.g. CI) after the
  user has explicitly authorized the exact operation.
- **`kyma module unmanage`** stops reconciliation but leaves resources running;
  make sure the user understands the module will no longer be reconciled.

## Notes

- Commands under `kyma alpha` use unstable APIs and may change between releases.
- Flag signatures here reflect Kyma CLI v3. If a flag is rejected, check
  `kyma <command> --help` on the installed version — a tagged release can lag the
  latest development build.
