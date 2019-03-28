# charts

_[Helm](https://helm.sh) charts for [`api`][api]._

## Installation

You can install these charts from the repository located at `https://charts.stevenxie.me`.

```bash
## Add the repository.
helm repo add stevenxie https://charts.stevenxie.me

## Install the chart.
helm install -f values.yaml -n api stevenxie/api
```

## Configuration

See
[`api/values.yaml`](https://github.com/stevenxie/api/blob/master/deployment/charts/api/values.yaml)
for an the default `values.yaml` configuration.

To install `api` for production, one should have an Ingress controller in
the target namespace, and configure a `values.yaml` with an appropriate
`ingress.host` value:

```yaml
ingress:
  host: api.stevenxie.me # example
```

[api]: https://github.com/stevenxie/merlin
