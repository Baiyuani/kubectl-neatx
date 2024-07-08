# kubectl-neatx

> based on https://github.com/itaysk/kubectl-neat, add subcommand 'export' and 'migrate'(todo)

If you just want yaml readability, please use [kubectl-neat](https://github.com/itaysk/kubectl-neat). If you want to back up yaml or create it in another cluster, you can use kubectl-neatx.

## Installation

- From [Releases](https://github.com/Baiyuani/kubectl-neatx/releases) download the latest version tarball

- Unzip the tarball

    ```bash
    tar -xvf kubectl-neatx_linux_amd64.tar.gz
    ```

- Install the kubectl-neatx

    ```shell
    sudo install -o root -g root -m 0755 kubectl-neatx /usr/local/bin/kubectl-neatx
    ```

## Usage

```shell
$ kubectl neatx -h
Usage:
  kubectl-neatx [flags]
  kubectl-neatx [command]

Examples:
kubectl get pod mypod -o yaml | kubectl neatx
kubectl neatx -f - <./my-pod.json
kubectl neatx -f ./my-pod.json
kubectl neatx -f ./my-pod.json --output yaml

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  export      Batch export of specified resource manifests
  get         Print specific resource manifest
  help        Help about any command
  version     Print kubectl-neatx version

Flags:
  -f, --file string     file path to neat, or - to read from stdin (default "-")
  -h, --help            help for kubectl-neatx
  -o, --output string   output format: yaml or json (default "yaml")

Use "kubectl-neatx [command] --help" for more information about a command.
```
