# ZProxy

ZProxy is a SOCKS5 proxy designed to help you access internal services deployed by Zeabur Dedicated Server.

## Installation

You can go to the [Releases](https://github.com/ImSingee/zproxy/releases) page to download the latest binary.

### Docker

You can also use the Docker image:

```bash
docker pull ghcr.io/imsingee/zproxy:latest
```

### Kubernetes

You can deploy this to your Kubernetes cluster:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zproxy
  labels:
    app: zproxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: zproxy
  template:
    metadata:
      labels:
        app: zproxy
    spec:
      containers:
      - name: zproxy
        image: ghcr.io/imsingee/zproxy:latest
        ports:
        - containerPort: 1080
        env:
        - name: CLUSTER_DOMAIN
          value: "cluster.local"
        - name: AUTH_USERNAME
          valueFrom:
            secretKeyRef:
              name: zproxy-auth
              key: username
        - name: AUTH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: zproxy-auth
              key: password
---
apiVersion: v1
kind: Service
metadata:
  name: zproxy
spec:
  selector:
    app: zproxy
  ports:
  - port: 1080
    targetPort: 1080
  type: ClusterIP
```

## Usage

```bash
./zproxy [flags]
```

### Configuration

| Flag                 | Short | Environment Variable | Description                          | Default       |
|----------------------|-------|----------------------|--------------------------------------|---------------|
| `--listen`           | `-l`  |                      | Proxy listening address              | :1080         |
| `--in-domain-suffix` | `-s`  | `IN_DOMAIN_SUFFIX`   | Domain suffix to replace             | cluster.local |
| `--cluster-domain`   | `-c`  | `CLUSTER_DOMAIN`     | Cluster domain to use as replacement | cluster.local |
| `--username`         | `-u`  | `AUTH_USERNAME`      | Authentication username              |               |
| `--password`         | `-p`  | `AUTH_PASSWORD`      | Authentication password              |               |


Additionally, you can also use `PORT` environment variable to configure listening port.

## Domain Modification Rules

ZProxy modifies domain names according to the following rules:

1. The proxy only accepts domain names (IP addresses are rejected)
2. The domain must end with the specified domain suffix (e.g., `cluster.local`)
3. The domain suffix is replaced with the specified cluster domain

For example, with default settings:
- `service.namespace.cluster.local` → `service.namespace.cluster.local` (no change if suffixes are the same)
- `service.namespace.custom.suffix` → `service.namespace.cluster.local` (if domain suffix is set to `custom.suffix`)


## Development

### CI/CD

This project uses GitHub Actions for continuous integration and delivery:

- On every push to the `main` branch, the code is built and a Docker image is published to GitHub Container Registry.
- When a tag starting with `v` is pushed (e.g., `v1.0.0`), a new release is created with binaries for multiple platforms and a Docker image with the version tag.

To create a new release:

```bash
# Tag the commit
git tag v1.0.0

# Push the tag
git push origin v1.0.0
```

## License

[MIT License](LICENSE)
