---
title: kyma upgrade
---

Upgrades Kyma

## Synopsis

Use this command to upgrade the Kyma version on a cluster.

```bash
kyma upgrade [flags]
```

## Flags

```bash
  -c, --components string      Path to a YAML file with a component list to override.
      --custom-image string    Full image name including the registry and the tag. Required for upgrading a remote cluster from local sources.
  -d, --domain string          Domain used for the upgrade. (default "kyma.local")
      --fallback-level int     If "source=master", defines the number of commits from master branch taken into account if artifacts for newer commits do not exist yet (default 5)
  -n, --no-wait                Determines if the command should wait for the Kyma upgrade to complete.
  -o, --override stringArray   Path to a YAML file with parameters to override.
  -p, --password string        Predefined cluster password.
      --profile string         Kyma installation profile (evaluation|production). If not specified, Kyma is installed with the default chart values.
  -s, --source string          Upgrade source.
                               	- To use a specific release, write "kyma upgrade --source=1.3.0".
                               	- To use the master branch, write "kyma install --source=master".
                               	- To use a commit, write "kyma upgrade --source=34edf09a".
                               	- To use the local sources, write "kyma upgrade --source=local".
                               	- To use a custom installer image, write "kyma upgrade --source=user/my-kyma-installer:v1.4.0".
      --src-path string        Absolute path to local sources.
      --timeout duration       Timeout after which CLI stops watching the upgrade progress. (default 1h0m0s)
      --tls-cert string        TLS certificate for the domain used for the upgrade. The certificate must be a base64-encoded value.
      --tls-key string         TLS key for the domain used for the upgrade. The key must be a base64-encoded value.
```

## Flags inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems. It avoids any user interaction (such as no dialog prompts) and ensures that logs are formatted properly in log files (such as no spinners for CLI steps).
  -h, --help                Displays help for the command.
      --kubeconfig string   Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

