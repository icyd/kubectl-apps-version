
## Usage
The following assumes you have the plugin installed via

```shell
kubectl krew install kubectl-apps-version
```

### Scan images in your current kubecontext

```shell
kubectl appsversion
```

### Scan images in another kubecontext

```shell
kubectl appsversion --context=context-name
```

## How it works

Retrieves the data of Deployments, Statefulsets (if flag `--statefulsets` is given
and Daemonsets (if flag `--daemonsets` is given), and print information of the application
name, and the images for both _initcontainers_ and _containers_. Also retrieve application's
labels `app.kubernetes.io/managed-by`, `helm.sh/chart` and `argocd.argoproj.io/instance` to
identify how the specific application is managed.
