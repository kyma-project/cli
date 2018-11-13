# Kymactl

## Overview

A command line tool to support Kyma developers

## Available Commands

- `version`: Shows the kymactl version. The version is set at compile time passing it to the go linker as a flag:

    ```bash
    go build -o kymactl -ldflags "-X github.com/kyma-incubator/kymactl/cmd.Version=1.5.0"
    ```
- `status`: Tracks the status of a Kyma cluster (replaces the `is-installed.sh` script)
- `help`: Displays usage for the given command (e.g. `kymactl help`, `kymactl help status`, etc...)

## kymactl as a kubectl plugin

To follow this section a kubectl version of 1.12.0 or later is recommended.

A plugin is nothing more than a standalone executable file, whose name begins with kubectl- . To install a plugin, simply move this executable file to anywhere on your PATH.

Rename a `kymactl` binary to `kubectl-kymactl` and place it anywhere in your PATH:

```bash
sudo mv ./kymactl /usr/local/bin/kubectl-kyma
```

Run `kubectl plugin list` command and you will see your plugin in the list of available plugins.

You may now invoke your plugin as a kubectl command:

```bash
$ kubectl kymactl status
Kyma is running!
```

To know more about extending kubectl with plugins read [kubernetes documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).
