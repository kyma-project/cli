---
name: kyma-troubleshoot
description: Diagnose a misbehaving open-source Kyma cluster or module using the Kyma CLI (v3) and kubectl. Use when a module is stuck or not ready, an app is failing, Istio or networking is broken, or the user wants to troubleshoot, diagnose, or inspect a Kyma cluster. Also use when kyma module add/delete itself is failing with an error. Trigger keywords - kyma diagnose, kyma alpha diagnose, kyma module not ready, istio broken, kyma logs, troubleshoot kyma, kyma cluster health. Do NOT use for pure module lifecycle tasks (add/enable/remove) where no problem is reported — e.g. "add the Istio module" with no error context → use kyma-module-lifecycle instead; "add the Istio module because sidecar injection is broken" → this skill applies for the diagnosis step.
---

# Skill: kyma-troubleshoot

Diagnose problems on a Kyma-enabled Kubernetes cluster with the open-source
**Kyma CLI (v3)** and `kubectl`, following a **read-only-first** ladder that
narrows from cluster health, to the affected component, to individual resources —
before proposing any fix.

Kyma CLI v3 is a Go-based, Apache-2.0 tool from the open-source Kyma project
([github.com/kyma-project/cli](https://github.com/kyma-project/cli)). It is a
full rewrite of v2 — **do not assume v2 syntax.**

## Scope and guardrails

This skill covers **only the open-source Kyma project**. When using it:

- ✅ Use open-source Kyma (`kyma-project/*`), open-source modules, the Kyma CLI,
  and standard `kubectl`.
- ❌ Do **not** encode SAP-managed (SKR) specifics, BTP cockpit flows, internal
  endpoints or URLs, SKR configuration, or any SAP-internal knowledge.
- ⚠️ **Diagnosis is read-only.** Never `delete` or `unmanage` a module as a
  troubleshooting step without showing the user the exact impact and getting
  explicit confirmation (see [Safety](#safety-destructive-operations)).

## Prerequisites

- The `kyma` CLI installed and on `PATH` (v3). Confirm with `kyma version`.
- `kubectl` installed and its current context pointing at the **affected** cluster.

**Verify your kubectl context before starting** — all diagnose commands fail silently
with a connection error if no context is set:

```bash
kubectl config current-context     # confirm which cluster you are targeting
kubectl cluster-info               # verify connectivity
# If wrong or missing, switch context:
kubectl config get-contexts
kubectl config use-context <name>
# Or pass it per-command: kyma alpha diagnose cluster --kubeconfig <path>
```

Before running a command whose exact flags you are unsure of, inspect help output
rather than guessing — the installed version is the source of truth:

```bash
kyma alpha diagnose cluster --help
kyma help <command>
```

## Instructions

Work top-down. Do not skip to a fix before you have observed the actual state.

### 1. Establish overall state (read-only)

Run the cluster diagnose command first — it is the single most informative first step:

```bash
kyma alpha diagnose cluster              # full cluster snapshot (YAML output)
kyma alpha diagnose cluster -f json      # same, JSON format — pipe to jq for filtering
kyma alpha diagnose cluster --verbose    # include collection-error details (cluster subcommand only)
# Add --skip-extensions to suppress the harmless extensions warning that clutters output
```

`kyma alpha diagnose cluster` output has four sections:

| Section | What it shows |
|---|---|
| `metadata` | Cluster ID, domain, shoot name, region, provider, K8s version, NAT gateway IPs, installed Gardener extensions |
| `warnings` | All Kubernetes warning events cluster-wide (source, involved object, message, count) |
| `nodes` | Per-node: machine info, capacity, actual CPU/memory usage + percentages, topology (region/zone) |
| `kymaModulesErrors` | Only modules in `Error` or `Warning` state, with their full `.status.conditions` |

> **`metadata` is only populated on Gardener/SAP-managed clusters.** On plain clusters
> (k3d, kind, EKS, GKE, etc.) all metadata fields will be empty — this is expected, not an error.

Then get the module list to see every module and its state:

```bash
kyma module list                 # which modules exist and their Ready/Processing/Error/etc. state
kyma module list --show-errors   # same, with inline error details per module (KLM-managed clusters only)
```

> **If `kyma module list` returns "not managed by KLM"** (plain cluster without Lifecycle Manager),
> fall back to kubectl to discover what's installed:
> ```bash
> kubectl api-resources --verbs=list | grep kyma-project.io  # which module CRDs are present
> kubectl get pods -n kyma-system | grep operator             # which operator pods are running
> ```

> **`kymaModulesErrors: []` in diagnose cluster output is ambiguous** — an empty list means
> both "all installed modules are healthy" AND "no modules are installed at all".
> Always cross-check with `kyma module list` (or the kubectl fallback above) to distinguish the two.

> **`kyma alpha diagnose cluster` is always the first command.** It surfaces cluster-wide
> warning events, node pressure, and all module errors in a single call — before you
> narrow by symptom.

### 2. Narrow by symptom

Pick the diagnostic that matches what the user reports:

#### 2a. Networking, gateway, mTLS, sidecar issues

```bash
# First-pass mesh analysis — catches misconfigured AuthorizationPolicy,
# PeerAuthentication, DestinationRule, Gateway, and injection issues:
istioctl analyze [-n <namespace> | --all-namespaces]

# Then run the Kyma-level Istio diagnostic:
kyma alpha diagnose istio [-A|--all-namespaces] [--level warning|error|info] [-n <namespace>] [--timeout 30s]
```

> **Confirm Istio is installed first** — `kyma alpha diagnose istio` exits silently with
> zero bytes when the Istio module is absent, which is indistinguishable from "no issues found".
> Run `kubectl get istio -n kyma-system` before running the diagnose command.
> - `"no resources found"` → CRD is present, but no Istio CR was created (partial install); check operator pod logs
> - `"the server doesn't have a resource type 'istio'"` → CRD was never applied; enable the Istio module first

`istioctl analyze` is your quickest signal — it names the problematic resource and
namespace and calls out common IST codes (e.g. `IST0133` for an overly broad
`AuthorizationPolicy`).

> **`istioctl` not on PATH?** Skip `istioctl analyze` and go directly to
> `kyma alpha diagnose istio -n <namespace> --level warning` as a standalone
> alternative — it covers the same injection and policy checks without requiring
> `istioctl` to be installed.

#### 2b. App or module errors

> **`--strict` silently drops ALL non-JSON lines**, including plain-text error messages
> from containers that don't emit structured logs. If you see fewer logs than expected,
> drop `--strict` to see all output.

```bash
# Aggregates JSON-formatted error-level container logs.
# --since and --lines are independent; combine to apply both limits:
kyma alpha diagnose logs --module <module> --since 1h [--timeout 30s] [--strict]
kyma alpha diagnose logs --module <module> --lines 200 [--timeout 30s] [--strict]
# --lines defaults to 200 when omitted
```

`kyma alpha diagnose logs` accepts `--module` repeatedly to aggregate across
several modules at once. Append `--format json --output <file>` to save results.
Add `--strict` to suppress non-JSON log lines and see only structured error entries.

> **`--since` and `--lines` are independent flags** — use either or both.
> `--since` bounds by time window; `--lines` caps line count; combine them to apply both limits simultaneously.

> **Empty output behaviour differs by subcommand:**
> - `kyma alpha diagnose logs` returns `logs: {}` when a module is absent — non-empty output, but empty map.
> - `kyma alpha diagnose istio` exits silently (zero bytes) when Istio is not installed.
> In both cases, check `kyma module list` to confirm the module is present before concluding there is no problem.

### 3. Drill down with kubectl

Once you know the affected component, inspect its resources directly.

First, discover the module's CR kind and operator pod:

```bash
# Find the CRD name for a module (e.g. "istio", "serverless")
kubectl api-resources --verbs=list | grep kyma-project.io

# Find the operator pod for a module
kubectl get pods -n kyma-system | grep <module>
# or by label:
kubectl get pods -n kyma-system -l app.kubernetes.io/name=<module>-operator
```

Then inspect:

```bash
kubectl get <module-cr> -n kyma-system -o yaml        # inspect CR .status conditions
kubectl describe <resource> -n <namespace>            # events + conditions
kubectl get events -n <namespace> --sort-by=.lastTimestamp
kubectl logs <operator-pod> -n kyma-system            # module operator logs
```

Focus on the CR's `.status.conditions` — Kyma modules report their health there.
Module CRs are in the `kyma-system` namespace. Common module CR names:

> **Verify the CRD is installed first** — if `kubectl get <module-cr>` returns
> `"no resource type"`, the module's CRD was never applied (e.g. install failed mid-way).
> Run `kubectl api-resources --verbs=list | grep kyma-project.io` to see which CRDs exist
> before assuming a CR is simply missing.

| Module | CR command |
|---|---|
| Istio | `kubectl get istio -n kyma-system -o yaml` |
| API Gateway | `kubectl get apigateway -n kyma-system -o yaml` |
| Serverless | `kubectl get serverless -n kyma-system -o yaml` |
| Keda | `kubectl get keda -n kyma-system -o yaml` |
| Telemetry | `kubectl get telemetry -n kyma-system -o yaml` |
| BTP Manager | `kubectl get btpoperator -n kyma-system -o yaml` |

To extract just the conditions without the full YAML:

```bash
kubectl get <module-cr> -n kyma-system -o jsonpath='{.status.conditions[*]}' | jq .
# Example for serverless:
kubectl get serverless -n kyma-system -o jsonpath='{.status.conditions[*]}' | jq .
```

**For Istio/service-mesh issues, also run:**

```bash
# Sidecar version overview — shows proxy version vs control plane version:
istioctl proxy-status

# Check istiod and api-gateway logs for post-change or post-upgrade errors:
kubectl logs -n istio-system deployments/istiod
kubectl logs -n kyma-system deployments/api-gateway
```

**For network connectivity issues:**

```bash
# Check for NetworkPolicies that may be blocking traffic:
kubectl get networkpolicies -A

# List Istio resources that affect connectivity:
kubectl get authorizationpolicies,peerauthentications,destinationrules,gateways -A
```

### 4. Match against common failure patterns

#### General module failures

| Symptom | Likely cause | Next step |
|---|---|---|
| `not managed by KLM` error on `module list/add/delete` | Cluster isn't managed by Kyma Lifecycle Manager (KLM) | Expected on plain clusters — `kyma module catalog` (list available modules) and `kyma module pull <name>` (install without KLM) still work; see [upstream docs](https://github.com/kyma-project/lifecycle-manager) for full lifecycle management |
| Module stuck in `Processing` | Operator not reconciling / waiting on dependency | Check operator pod logs and the CR `.status`; if operator pod is healthy but CR is stuck, propose (with explicit confirmation): `kubectl rollout restart deployment/<module>-operator -n kyma-system` |
| Module CR in `Error` state | See `kymaModulesErrors` in `kyma alpha diagnose cluster` output | Inspect `.status.conditions` on the module CR; check operator pod logs |
| `--expose` on `kyma app push` fails | `api-gateway` module (APIRule CRD) not installed | Enable the api-gateway module first |
| App pods `CrashLoopBackOff` | App-level error | `kyma alpha diagnose logs --module <module>` + `kubectl logs` |

#### Istio / networking failures

| Symptom | Likely cause | Next step |
|---|---|---|
| Istio sidecar injection fails | Istio module not installed | Enable the Istio module first |
| Some pods missing sidecar, others have it | Injection label changed after workloads were deployed | Run `istioctl analyze`; restart pods in namespace to re-inject |
| `Connection reset by peer` / `GOAWAY` response | mTLS enabled; non-mesh service talking to mesh service | Create `DestinationRule` (disable mTLS for that service) or `PeerAuthentication` in `PERMISSIVE` mode |
| 503 on Kyma endpoint | Istio Ingress Gateway configuration error | Restart ingress gateway pods (confirm with user first): `kubectl delete pod -l app=istio-ingressgateway -n istio-system` |
| 404 on Kyma endpoint | Duplicate `Gateway` resources pointing to same host | Check for duplicate Gateways; Istio cannot reliably route when two Gateways share a host |
| 403 Forbidden | Misconfigured `AuthorizationPolicy` blocking traffic | Run `istioctl analyze`; check `AuthorizationPolicy` selector and rules — a deny policy without a `ports` field blocks all TCP, not just HTTP |
| 403 with port number in Host header | Istio host-matching includes port; AuthorizationPolicy only defines hostname | Add port to `AuthorizationPolicy` hosts, or fix client to omit default port |
| Jobs / Pods created by Jobs stuck in `NotReady` | `istio-proxy` running as regular container, not native sidecar | Add annotation `sidecar.istio.io/nativeSidecar: "true"` to pod/job template |
| Init containers can't access network | Same as above — `istio-proxy` is regular container, not native sidecar | Add annotation `sidecar.istio.io/nativeSidecar: "true"` |
| Sidecar version mismatch after Istio upgrade | Jobs/standalone Pods couldn't be auto-restarted | Check `istioctl proxy-status`; manually restart listed workloads |
| `istio-proxy` is declared as container instead of initContainer | `sidecar.istio.io/nativeSidecar` annotation set to `"false"` | Remove or change annotation to `"true"` |
| Istiod deployment failing / reconciliation error | Mutating webhook injecting a broken container into Istio system deployments | `kubectl describe deployment istiod -n istio-system`; check for injected containers; remove via `kubectl edit` |
| `x509: certificate signed by unknown authority` in istiod logs; JWT `Jwt verification fails` for valid tokens | JWKS endpoint uses a CA not trusted by istiod | Create ConfigMap with CA PEM in `istio-system`; patch `istiod` Deployment to mount it under `/cacerts` |
| Ingress Gateway metrics missing (`promscrape.maxScrapeSize` exceeded) | Istio Ingress Gateway never clears in-memory metrics buffer | Apply `EnvoyFilter` with `rotation_interval` and `graceful_deletion_interval`; verify with `istioctl proxy-config listeners <pod>.istio-system --port 8080 -o json` |
| Client IP not in `X-Forwarded-For` / `X-Envoy-External-Address` after setting `numTrustedProxies` | Cloud provider requires additional configuration for client IP forwarding | AWS: add `EnvoyFilter` for Proxy Protocol; Azure/GCP: `kubectl patch svc istio-ingressgateway -n istio-system -p '{"spec":{"externalTrafficPolicy":"Local"}}'` — **note: Kyma Istio Operator reverts this change every ~10 hours** |
| Unintentional Istio removal — Istio CR in `Warning`, condition `IstioCustomResourcesDangling` | Istio deletion blocked because VirtualService/other resources still exist | Remove finalizer from Istio CR (`kubectl edit istio -n kyma-system default`); then re-add module |

#### API Gateway failures

| Symptom | Likely cause | Next step |
|---|---|---|
| 401 Unauthorized | Missing or expired JWT | Decode JWT; check `exp` and `scp` fields; generate a new JWT |
| 403 Forbidden | JWT scope mismatch, or `AuthorizationPolicy` blocking | Check scopes; run `istioctl analyze` for AuthorizationPolicy issues |
| APIRule in `ERROR` status | Misconfigured APIRule spec (e.g. missing `issuer`) | `kubectl get apirules -n <ns> <name> -o=jsonpath='{.status.description}'`; fix the error and re-apply |
| Gateway not reachable / APIRule traffic not working | Duplicate `kyma-gateway` resources | Identify the duplicate with `kubectl get gateway -n kyma-system`; confirm with user, then delete: `kubectl -n kyma-system delete gateway <duplicate-name>` |
| Requests timeout after 180s | Default APIRule timeout applies | Add `spec.timeout` (or `spec.rules[].timeout`) to the APIRule |
| In-cluster communication blocked with `RBAC: access denied` after APIRule v1beta1→v2 migration | APIRule v2 blocks internal traffic by default (secure-by-default) | Create an ALLOW-type `AuthorizationPolicy` for internal traffic excluding `istio-ingressgateway` |
| `Connection reset by peer` on Kyma Gateway; istiod logs TLS errors | `kyma-gateway-certs` Secret missing in `istio-system` | Certificate restores automatically on next reconciliation; trigger immediately with `kubectl -n kyma-system rollout restart deployment api-gateway-controller-manager` |
| Unintentional API Gateway removal — APIGateway CR `Warning`, condition `DeletionBlockedExistingResources` | Deletion blocked by existing APIRule resources | Remove finalizers from APIGateway CR; then re-add module |

#### Lifecycle Manager (KLM) failures

| Symptom | Likely cause | Next step |
|---|---|---|
| Kyma CR stuck in `Processing` during module deletion, module CR in `Deleting` | Module deleted but user CRs (e.g. `ScaledObject`) still exist — finalizer blocks deletion | Remove finalizer from module CR; re-add module to restore state |
| Kyma CR stuck with `deletionTimestamp` set, never fully deleted | `skip-reconciliation: "true"` label on Kyma CR, or broken connection to cluster | Check `status.lastOperation.operation`; if `"failed to ensure remote namespace exists"`, verify the cluster kubeconfig/access is still valid |
| Module silently stops reconciling | `skip-reconciliation` label on Manifest CR or Kyma CR, or `state: unmanaged` | Check both CRs for the `operator.kyma-project.io/skip-reconciliation` label; check `status.modules[].state` |
| Kyma/Manifest CR not reconciling — need to trigger immediately | Normal requeue is up to 5 minutes | Add any label to the CR (e.g. `operator.kyma-project.io/requeue: "true"`) to trigger immediate requeue; remove after done |
| Kyma CR in `Error` with `"unable to create new content in namespace kyma-system because it is being terminated"` | `kyma-system` namespace in terminating state | Remove finalizers from stuck resources in `kyma-system` (check `ServiceInstances` — customer cleanup may be required); KLM re-creates the namespace automatically |
| Two Manifest CRs for the same module | Module FQDN prefix changed between versions | Label old Manifest with `skip-reconciliation=true`; delete with `--wait=false`; patch finalizers to null |

#### Serverless failures

| Symptom | Likely cause | Next step |
|---|---|---|
| Serverless controller restarts every ~10 minutes | Large or multi-repository Git sources cause CPU/IO spike during polling | Avoid large/multi repositories; use dedicated single-Function repos |
| Function not starting after customizing OpenTelemetry deps | OTel dependency version conflict | Align all OTel packages (`api`, `sdk`, `instrumentation`) to compatible versions; reference official runtime deps |
| Function stuck in `SourceUpdateFailed` with `certificate signed by unknown authority` | Git server uses private/internal CA not trusted by cluster | Contact landscape admin to add CA to cluster trust bundle |
| Serverless CR in `ConfigurationErr` | User misconfiguration in Serverless CR | Review and correct the Serverless CR configuration |
| Serverless CR in `DeletionErr` | Attempting to delete module while Function CRs still exist | Delete all Function CRs before removing the Serverless module |
| Serverless CR in `ServerlessDuplicated` | Multiple Serverless instances running concurrently | Redundant instances are removed automatically; if not, manually delete duplicate instances |

### 5. Report and propose

Summarize for the user: what you observed, the likely root cause, and a proposed
fix. **Do not apply destructive fixes unprompted** — deletion, unmanaging, or
re-adding a module all need explicit confirmation.

## Safety: destructive operations

Operations fall into four tiers — do not skip to a lower tier:

| Tier | Examples | Gate |
|---|---|---|
| **Read-only** | `kubectl get/describe/logs`, `kyma alpha diagnose *`, `istioctl analyze` | No gate — run freely |
| **Non-destructive with side effects** | `kubectl rollout restart deployment/<op> -n kyma-system` | Show exact command + expected impact; get explicit confirmation before running |
| **Semi-destructive** | Removing a CR finalizer | Show exact impact (unblocks deletion immediately); confirm user intent (recovery vs. intentional deletion) |
| **Destructive** | `kyma module delete`, `kyma module unmanage`, `kubectl delete` on resources | Run `kyma module list` first; show exact module(s) affected; get explicit confirmation |

- **`kyma module unmanage`** stops reconciliation but leaves resources running — ensure the user understands the module will no longer be reconciled.
- **Diagnosis is read-only.** The commands in steps 1–3 belong to the first tier; keep it that way until the user approves a fix.

## Related skills

- **`kyma-module-lifecycle`** — enable, verify, or remove a module once you've
  diagnosed the problem.
- **`kyma-app-deploy`** — deploy and re-push applications.
- **`kyma-cli`** — general Kyma CLI reference.

## Notes

- Commands under `kyma alpha` use unstable APIs and may change between releases.
- If a flag is rejected, check `kyma <command> --help` on the installed version —
  a tagged release can lag the latest development build.
- `istioctl` commands (`proxy-status`, `analyze`, `proxy-config`) require the
  Istio module to be installed and `istioctl` to be on `PATH`.
  Install: `brew install istioctl` (macOS/Linux) or see
  [istio.io/latest/docs/ops/diagnostic-tools/istioctl/](https://istio.io/latest/docs/ops/diagnostic-tools/istioctl/).
- Every `kyma alpha` command may emit an extensions warning (`failed to fetch all extensions`).
  This is harmless. Suppress it with `--skip-extensions`; use `--show-extensions-error`
  to see the full error detail.
