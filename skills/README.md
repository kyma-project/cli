# AI agent skills

AI coding-agent skills for the open-source **Kyma CLI (v3)**. Each skill is a
`SKILL.md` file with `name` and `description` frontmatter in the
[SAP AI Skills Library](https://github.com/SAP/ai-skills-library) format, which is
also compatible with Claude Code, Codex, Cursor, and other agents that support it.

Skills give an AI agent accurate, current knowledge of the `v3` CLI — command
groups, correct flag signatures, and the important gotchas — so it follows the
CLI's actual behavior instead of guessing (for example, assuming deprecated `v2`
syntax).

## Scope

These skills cover the **open-source Kyma project only** (`kyma-project/*`).
SAP-managed Kyma (SKR), BTP cockpit flows, and SAP-internal configuration are out
of scope. The BTP-touching commands (`kyma alpha provision`, `kyma alpha hana map`)
are mentioned but never carry hardcoded credentials or subaccount details.

## Skills

| Skill | Purpose |
|-------|---------|
| [`kyma-cli`](./kyma-cli/SKILL.md) | General orientation: install, command groups (`kyma module`, `kyma app`, `kyma alpha`), dashboard, kubeconfig, completion, diagnostics. |
| [`kyma-module-lifecycle`](./kyma-module-lifecycle/SKILL.md) | Full module lifecycle: discover, enable, configure, verify, manage/unmanage, and delete, including Kyma Lifecycle Manager (KLM) prerequisites. |
| [`kyma-app-deploy`](./kyma-app-deploy/SKILL.md) | Deploying applications with `kyma app push` (source, Dockerfile, or image), `--expose`, Istio injection, and env/secret injection. |
| [`kyma-troubleshoot`](./kyma-troubleshoot/SKILL.md) | Read-only-first diagnosis via `kyma alpha diagnose` and `kubectl`, with a common-failure reference table. |

## Usage

With an AI agent that supports the skills format, open this repository and
describe your task (for example, "enable the Istio module" or "why is my app not
reachable?"). The relevant skill loads on demand.
