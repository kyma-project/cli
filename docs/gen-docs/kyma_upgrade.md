---
title: kyma upgrade
---

[DEPRECATED] Upgrades Kyma

## Synopsis

[DEPRECATED: The "upgrade" command works only when upgrading to Kyma 1.x.x. To upgrade to Kyma 2.x.x, use the "deploy" command.]

Use this command to upgrade the Kyma version on a cluster.

```bash
kyma upgrade [flags]
```

## Flags

```bash
  -c, --components string      Path to a YAML file with a component list to override.
      --custom-image string    Full image name including the registry and the tag. Required for upgrading a remote cluster from local sources.
  -d, --domain string          Domain used for the upgrade. (default "kyma.local")
      --fallback-level int     If "source=main", defines the number of commits from main branch taken into account if artifacts for newer commits do not exist yet (default 5)
  -n, --no-wait                Determines if the command should wait for the Kyma upgrade to complete.
  -o, --override stringArray   Path to a YAML file with parameters to override.
  -p, --password string        Predefined cluster password.
      --profile string         Kyma installation profile (evaluation|production). If not specified, Kyma is installed with the default chart values.
  -s, --source string          Upgrade source.
                               	- To use a specific release, write "kyma upgrade --source=1.3.0".
                               	- To use the main branch, write "kyma install --source=main".
                               	- To use a commit, write "kyma upgrade --source=34edf09a".
                               	- To use the local sources, write "kyma upgrade --source=local".
                               	- To use a custom installer image, write "kyma upgrade --source=user/my-kyma-installer:v1.4.0". (default "1.24.9")
      --src-path string        Absolute path to local sources.
      --timeout duration       Timeout after which CLI stops watching the upgrade progress. (default 1h0m0s)
      --tls-cert string        TLS certificate for the domain used for the upgrade. The certificate must be a base64-encoded value.
      --tls-key string         TLS key for the domain used for the upgrade. The key must be a base64-encoded value.
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Provides command help.
      --kubeconfig string   Path to the kubeconfig file. If undefined, Kyma CLI uses the KUBECONFIG environment variable, or falls back "/$HOME/.kube/config".
      --non-interactive     Enables the non-interactive shell mode (no colorized output, no spinner)
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](kyma.md)	 - Controls a Kyma cluster.

