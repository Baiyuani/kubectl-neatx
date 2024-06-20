# kubectl-neatx

> based on https://github.com/itaysk/kubectl-neat, add subcommand 'export' and 'migrate'(todo)

Remove clutter from Kubernetes manifests to make them more readable.


## Why

When you create a Kubernetes resource, let's say a Pod, Kubernetes adds a whole bunch of internal system information to the yaml or json that you originally authored. This includes:

- Metadata such as creation timestamp, or some internal IDs
- Fill in for missing attributes with default values
- Additional system attributes created by admission controllers, such as service account token
- Status information

If you try to `kubectl get` resources you have created, they will no longer look like what you originally authored, and will be unreadably verbose.   
`kubectl-neatx` cleans up that redundant information for you.

## Installation

```bash

```

or just download the binary if you prefer.

When used as a kubectl plugin the command is `kubectl neatx`, and when used as a standalone executable it's `kubectl-neatx`.

## Usage

There are two modes of operation that specify where to get the input document from: a local file or from  Kubernetes.

### Local - file or Stdin

This is the default mode if you run just `kubectl neatx`. This command accepts an optional flag `-f/--file` which specifies the file to neatx. It can be a path to a local file, or `-` to read the file from stdin. If omitted, it will default to `-`. The file must be a yaml or json file and a valid Kubernetes resource.

There's another optional optional flag, `-o/--output` which specifies the format for the output. If omitted it will default to the same format of the input (auto-detected).

Examples:
```bash
kubectl get pod mypod -o yaml | kubectl neatx

kubectl get pod mypod -oyaml | kubectl neatx -o json

kubectl neatx -f - <./my-pod.json

kubectl neatx -f ./my-pod.json

kubectl neatx -f ./my-pod.json --output yaml
```

### Kubernetes - kubectl get wrapper

This mode is invoked by calling the `get` subcommand, i.e `kubectl neatx get ...`. It is a convenience to run `kubectl get` and then `kubectl neatx` the output in a single command. It accepts any argument that `kubectl get` accepts and passes those arguments as is to `kubectl get`. Since it executes `kubectl`, it need to be able to find it in the path.

Examples:
```bash
kubectl neatx get -- pod mypod -oyaml
kubectl neatx get -- svc -n default myservice --output json
```
