# Local Development Setup for O-RAN Near-RT RIC

This guide provides instructions for setting up a local development environment for the O-RAN Near-RT RIC project. This includes spinning up a local Kubernetes cluster, building the Go backend, and serving the Angular dashboards.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Go](https://golang.org/doc/install) (version 1.17+)
- [Node.js](https://nodejs.org/en/download/) (version 16.14.2+)
- [Angular CLI](https://angular.io/cli) (version 13.3.3)
- [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (Kubernetes in Docker)

## 1. Local Kubernetes Cluster with KIND

KIND allows you to run a local Kubernetes cluster using Docker containers as nodes.

### Create a Cluster

1.  **Create a cluster configuration file `kind-cluster.yaml`:**

    ```yaml
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
    - role: control-plane
    - role: worker
    - role: worker
    ```

2.  **Create the cluster:**

    ```bash
    kind create cluster --config kind-cluster.yaml
    ```

3.  **Verify the cluster is running:**

    ```bash
    kubectl cluster-info --context kind-kind
    ```

### Accessing the Cluster

Your `kubeconfig` file will be automatically updated with the new cluster context. You can switch to the KIND context with:

```bash
kubectl config use-context kind-kind
```

## 2. Build the Go Backend

The Go backend provides the API server for the main dashboard.

1.  **Navigate to the backend directory:**

    ```bash
    cd dashboard-master/dashboard-master
    ```

2.  **Build the backend:**

    ```bash
    make build
    ```

    This will compile the Go code and create the necessary binaries.

3.  **Deploy to your KIND cluster:**

    ```bash
    make deploy
    ```

    This will apply the Kubernetes manifests to your local cluster.

## 3. Serve the Angular Dashboards

There are two Angular dashboards: the main dashboard and the xApp dashboard.

### Main Dashboard

1.  **Navigate to the frontend directory:**

    ```bash
    cd dashboard-master/dashboard-master/src/app/frontend
    ```

2.  **Install dependencies:**

    ```bash
    npm install
    ```

3.  **Serve the dashboard:**

    ```bash
    npm start
    ```

    The dashboard will be available at `http://localhost:4200`.

### xApp Dashboard

1.  **Navigate to the xApp dashboard directory:**

    ```bash
    cd xAPP_dashboard-master
    ```

2.  **Install dependencies:**

    ```bash
    npm install
    ```

3.  **Serve the dashboard:**

    ```bash
    npm start
    ```

    The xApp dashboard will be available at `http://localhost:4201` (or another port if 4200 is in use).

## 4. Troubleshooting

### RBAC Errors

Role-Based Access Control (RBAC) errors are common when interacting with the Kubernetes API.

**Symptom:** You receive errors like `User "system:serviceaccount:default:default" cannot list resource "pods" in API group "" in the namespace "default"`.

**Solution:**

1.  **Identify the ServiceAccount:** Check the `serviceAccountName` in the `deployment.yaml` or `pod.yaml` for the component that is failing.

2.  **Check ClusterRole and ClusterRoleBinding:**
    - A `ClusterRole` defines permissions.
    - A `ClusterRoleBinding` grants those permissions to a user or a `ServiceAccount`.

3.  **Create or Update a ClusterRoleBinding:**

    Ensure a `ClusterRoleBinding` exists that binds the `ServiceAccount` to a `ClusterRole` with the necessary permissions. For example, to grant `cluster-admin` rights (use with caution in production):

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: my-app-cluster-admin
    subjects:
    - kind: ServiceAccount
      name: <your-service-account-name>
      namespace: <your-namespace>
    roleRef:
      kind: ClusterRole
      name: cluster-admin
      apiGroup: rbac.authorization.k8s.io
    ```

    Apply this with `kubectl apply -f <filename>.yaml`.

### Image Pull Errors

If you are using a local Docker registry with KIND, you may need to configure KIND to trust it. Refer to the [KIND documentation on local registries](https://kind.sigs.k8s.io/docs/user/local-registry/).
