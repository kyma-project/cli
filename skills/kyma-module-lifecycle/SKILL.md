---
name: kyma-module-lifecycle
description: Manage an open-source Kyma module through its full lifecycle with the Kyma CLI (v3) - discover, enable, configure, verify, change managed state, and remove. Use when the user wants to enable, install, add, configure, or remove a Kyma module (e.g. Serverless, Istio, API Gateway), pick a module channel, or manage/unmanage a module. Trigger keywords - enable kyma module, install serverless, add kyma module, module channel, managed unmanaged, remove kyma module, kyma module catalog.
---

# Skill: kyma-module-lifecycle

Take an open-source Kyma module through its full lifecycle with the **Kyma CLI
(v3)**: discover what's available, enable it on the right channel, verify it
becomes Ready, adjust its managed state, and remove it cleanly.

Kyma CLI v3 is a Go-based, Apache-2.0 tool from the open-source Kyma project
([github.com/kyma-project/cli](https://github.com/kyma-project/cli)). It is a
full rewrite of v2 — **do not assume v2 syntax.**

## Scope and guardrails

This skill covers **only the open-source Kyma project**. When using it:

- ✅ Use open-source Kyma (`kyma-project/*`), open-source modules, the Kyma CLI,
  and standard `kubectl`.
- ❌ Do **not** encode SAP-managed (SKR) specifics, BTP cockpit flows, internal
  endpoints or URLs, SKR configuration, or any SAP-internal knowledge.

> ℹ️ **Cluster prerequisite (Kyma Lifecycle Manager):** `add`, `list`, `delete`,
> `manage`, and `unmanage` require a cluster managed by Kyma Lifecycle Manager
> (KLM). Error messages differ by subcommand on a bare cluster:
> - `add`/`list`: `unknown module name or channel / failed to list module catalog from the target Kyma environment`
> - `manage`/`unmanage`/`delete`: `This cluster is not managed by Kyma Lifecycle Manager (KLM).`
>
> Only `catalog` and `pull` work without KLM.

## Prerequisites

- The `kyma` CLI installed and on `PATH` (v3). Confirm with `kyma version`.
- `kubectl` installed and its current context pointing at the target cluster.
- For the mutating stages (enable/verify/manage/remove): a **KLM-managed cluster**.

Before running a command whose exact flags you are unsure of, inspect help output
rather than guessing — the installed version is the source of truth:

```bash
kyma module --help
kyma help module add
```

## Instructions

Modules live under the **non-alpha** `kyma module` group (7 subcommands:
`catalog`, `list`, `add`, `delete`, `manage`, `unmanage`, `pull`).

> ⚠️ **Important:** `kyma alpha module` only exposes `catalog` and `pull`. The
> `add`, `delete`, `list`, `manage`, and `unmanage` subcommands exist **only**
> under the non-alpha `kyma module` group — never prefix them with `alpha`.

### 1. Discover

```bash
kyma module catalog       # all available modules (-o table|json|yaml)
kyma module list          # modules already installed on the cluster
```

> ℹ️ **Catalog is cluster-dependent:** available module names and channels depend
> on the catalog the target cluster exposes. Always run `kyma module catalog`
> first — a given name (e.g. `istio`) may not appear on every cluster.

> ℹ️ **Catalog output format:** columns are `NAME | AVAILABLE VERSIONS | ORIGIN`. Channel
> information is encoded inside the `AVAILABLE VERSIONS` cell as `<version>(<channel>)`
> tuples — e.g. `3.11.0(regular), 3.12.0(fast)`. Not all modules publish to every
> channel; check the AVAILABLE VERSIONS cell before specifying `-c <channel>`.

> ℹ️ **`kyma module list` output:** columns are `NAME | VERSION | CR POLICY | MANAGED | MODULE STATUS | INSTALLATION STATUS`. A module is fully installed only when `MODULE STATUS` is `Ready`. Other observed states: `Processing`, `Error`, `Warning`. Supports `-o table|json|yaml`. The `--show-errors` flag is accepted but may not add visible detail in table view; use `kyma alpha diagnose cluster` to see full error conditions for modules in `Error` or `Warning` state.

### 2. Enable a module

```bash
kyma module add <module> -c <channel> [--default-config-cr | --config-cr-path <path>] [--auto-approve]
```

- `-c/--channel` — pick the release channel from the catalog. Confirm the channel
  exists for that module in step 1.
- `--default-config-cr` — install with the module's default configuration CR.
- `--config-cr-path <path>` — install with a custom configuration CR from a file.
- Choose **default-CR** for a standard install; use **config-cr-path** when the
  user needs non-default configuration.

> ℹ️ **Channel semantics:** channel names are defined by the Kyma release process
> and visible in `kyma module catalog` output. Typical values are:
> - `regular` — stable, less-frequently-updated; recommended for production.
> - `fast` — newer module versions released sooner; suitable for testing or when a
>   specific new feature is needed before it reaches `regular`.
>
> Not all modules publish to every channel. Always confirm available channels
> from the catalog before running `module add`.

### 3. Community modules

Community modules must be pulled from their remote repository before they can be
added. The module name must appear in `kyma module catalog` output before `pull`
will succeed — `pull` does not fetch from an arbitrary upstream independently of
the local catalog.

```bash
kyma module pull <module-name> [--namespace <ns>] [--version <v>] [--force]
kyma module add <namespace>/<module-template-name>   # community-module name format
```

- `--force` — automatically approves the installation of KLM dependencies when
  prompted. Use this flag when pulling a community module onto a cluster where
  KLM dependency installation requires approval. Note: `--force` only handles
  the approval prompt; `pull` will still fail with a CRD-discovery error
  (`failed to discover API resource`) on a fully bare cluster that has no KLM
  operator installed at all — in that case, install KLM first.

Community modules trigger an SLA/approval prompt on `add`. Only pass
`--auto-approve` non-interactively after the user has authorized the exact module.

### 4. Verify it becomes Ready

```bash
kyma module list                                      # check MODULE STATUS column
kubectl get <module-cr> -n <namespace> -o yaml        # inspect .status.conditions
```

Wait for `MODULE STATUS` to show `Ready` in `kyma module list` before treating
the module as installed.
If it sticks in `Processing` or `Error`, hand off to the `kyma-troubleshoot` skill.

### 5. Adjust managed state

```bash
kyma module manage <module> [--policy CreateAndDelete|Ignore]   # let Kyma reconcile it
kyma module unmanage <module>                                   # stop reconciling; leave resources
```

- `manage` — Kyma actively reconciles the module (default policy
  `CreateAndDelete`).
- `unmanage` — Kyma stops reconciling but leaves the module's resources running.

### 6. Remove a module

```bash
kyma module delete <module> [--auto-approve]
```

Follow the [Safety](#safety-destructive-operations) rules before deleting.

To inspect the underlying `ModuleTemplate` custom resource directly with
`kubectl` (not a `kyma` command):

```bash
kubectl get crd moduletemplates.operator.kyma-project.io -o yaml
```

## Safety: destructive operations

- **Always confirm before deleting.** Run `kyma module list` first, show the user
  the exact module(s) that would be affected, and get explicit confirmation.
- **There is no bulk "delete all modules" command** — deletion is per-module. If a
  user says "delete everything, don't ask," do **not** loop `delete` blindly: list
  the modules and confirm the specific set to remove, one by one.
- **`--auto-approve` bypasses the CLI's built-in approval prompt** (and the
  community-module SLA prompt). Use it only non-interactively (e.g. CI) after the
  user has explicitly authorized the exact operation.
- **`kyma module unmanage`** stops reconciliation but leaves resources running;
  make sure the user understands the module will no longer be reconciled.

## Related skills

- **`kyma-troubleshoot`** — if a module won't reach Ready.
- **`kyma-app-deploy`** — deploy apps that depend on modules (e.g. api-gateway,
  Istio).
- **`kyma-cli`** — general Kyma CLI reference.

## Notes

- Commands under `kyma alpha` use unstable APIs and may change between releases.
- If a flag is rejected, check `kyma <command> --help` on the installed version —
  a tagged release can lag the latest development build.
