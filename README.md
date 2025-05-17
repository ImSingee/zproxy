# ZProxy

ZProxy is a SOCKS5 proxy designed to help you access internal services deployed by Zeabur Dedicated Server.

## Installation

You can go to the [Releases](https://github.com/ImSingee/zproxy/releases) page to download the latest binary.

But you generally want to deploy this to your k3s cluster directly.

(TODO)

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


## License

[MIT License](LICENSE)
