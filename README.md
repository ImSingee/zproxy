# ZProxy

ZProxy is a SOCKS5 proxy designed to help you access internal services deployed by Zeabur Dedicated Server.


> [!NOTE]
> This project is primarily completed by AI. I am only responsible for code review and detail modifications. The code quality of this project does not represent my personal code quality.


## Installation

[![Deploy on Zeabur](https://zeabur.com/button.svg)](https://zeabur.com/templates/ZL5OKK?referralCode=ImSingee)


### Docker

```bash
docker pull ghcr.io/imsingee/zproxy:latest
```

### Kubernetes

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
        # Zeabur DNS mapping configuration (optional)
        - name: ZEABUR_API_KEY
          valueFrom:
            secretKeyRef:
              name: zeabur-credentials
              key: api-key
        - name: ZEABUR_SERVER_ID
          value: "server-xxxxxx"
        - name: ZEABUR_UPDATE_INTERVAL
          value: "5m"
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

| Flag                       | Short | Environment Variable     | Description                          | Default       |
|----------------------------|-------|--------------------------|--------------------------------------|---------------|
| `--listen`                 | `-l`  | `PORT` (port only)       | Proxy listening address              | :1080         |
| `--in-domain-suffix`       | `-s`  | `IN_DOMAIN_SUFFIX`       | Domain suffix to replace             | cluster.local |
| `--cluster-domain`         | `-c`  | `CLUSTER_DOMAIN`         | Cluster domain to use as replacement | cluster.local |
| `--username`               | `-u`  | `AUTH_USERNAME`          | Authentication username              |               |
| `--password`               | `-p`  | `AUTH_PASSWORD`          | Authentication password              |               |
| `--zeabur-api-key`         |       | `ZEABUR_API_KEY`         | Zeabur API key                       |               |
| `--zeabur-server-id`       |       | `ZEABUR_SERVER_ID`       | Zeabur server ID                     |               |
| `--zeabur-update-interval` |       | `ZEABUR_UPDATE_INTERVAL` | Interval for updating Zeabur DNS map | 5m            |


Additionally, you can also use `PORT` environment variable to configure listening port.

## Domain Modification Rules

ZProxy modifies domain names according to the following rules:

1. The proxy only accepts domain names (IP addresses are rejected)
2. The domain must end with the specified domain suffix (e.g., `cluster.local`, configured with `--in-domain-suffix`)
3. The domain suffix is replaced with the specified cluster domain

For example, with default settings:
- `service.namespace.svc.cluster.local` → `service.namespace.svc.cluster.local` (no change if suffixes are the same)

And if you set a custom suffix (e.g. `custom.suffix`)
- `service.namespace.svc.custom.suffix` → `service.namespace.svc.cluster.local` 

### Zeabur DNS Mapping

ZProxy now supports special handling for Zeabur services. When the Zeabur API key and server ID are provided, ZProxy will:

1. Fetch service mappings from the Zeabur API
2. Allow access to Zeabur services using a special domain format
3. Periodically update the service mappings based on the configured interval

For domains with the format `{service-name}.{project-name}.zeabur.{domain-suffix}`, ZProxy will:
1. Look up the service name in the Zeabur DNS store
2. Map it to the corresponding Kubernetes service address

For example:
- `my-service.my-prject.zeabur.cluster.local` → the `my-service` service in `my-project` project

To enable this feature, you need to provide:
- `ZEABUR_API_KEY`: Your Zeabur API key
- `ZEABUR_SERVER_ID`: Your Zeabur server ID
- `ZEABUR_UPDATE_INTERVAL` (optional): How often to update the DNS mappings (default: 5m)

## License

[MIT License](LICENSE)
