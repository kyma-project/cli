---
title: kyma upgrade
---

Upgrades Kyma to match  the CLI version.

## Synopsis

Use this command to upgrade the Kyma version on a cluster so that it matches the CLI version

```bash
kyma upgrade [flags]
```

## Options

```bash
  -c, --components string      Path to a YAML file with a component list to override.
  -d, --domain string          Domain used for the upgrade. (default "kyma.local")
  -n, --noWait                 Determines if the command should wait for the Kyma upgrade to complete.
  -o, --override stringArray   Path to a YAML file with parameters to override.
  -p, --password string        Predefined cluster password.
      --timeout duration       Timeout after which CLI stops watching the upgrade progress. (default 1h0m0s)
      --tlsCert string         TLS certificate for the domain used for the upgrade. The certificate must be a base64-encoded value.
      --tlsKey string          TLS key for the domain used for the upgrade. The key must be a base64-encoded value.
```

## Options inherited from parent commands

```bash
      --ci                  Enables the CI mode to run on CI/CD systems.
  -h, --help                Displays help for the command.
      --kubeconfig string   Specifies the path to the kubeconfig file. By default, Kyma CLI uses the KUBECONFIG environment variable or "/$HOME/.kube/config" if the variable is not set.
      --non-interactive     Enables the non-interactive shell mode.
  -v, --verbose             Displays details of actions triggered by the command.
```

## See also

* [kyma](#kyma-kyma)	 - Controls a Kyma cluster.

