## Helm-helper

Make easier Helm Command

---
## Use

helm-helper is a tool that build a helm command with metadata

In your values file, you can put

```
# prometheus.yaml
_metadata:
  chart: ritmx/prometheus
  name: prom
  namespace: log
  vault:
  - field: google_credentials_file
    key: gcssa
    path: secret/gcp/credentials
  version: 0.4.4
...
```
---

```
helm-helper command -f prometheus.yaml | sh
```


## Commands

`helm-helper check_version --values file`

Check last version of the chart

Flag:

`--url` ou `HELM_URL` Helm repo URL
> Default `https://kubernetes-charts.storage.googleapis.com`

`HELM_URL=http://helm.example.com helm-helper check_version --values prometheus.yaml`

---

`helm-helper command --values file`

Build helm command line

In `_metadata:`

```
chart: chart_name
name: release_name
namespace: namespace
version: release version
vault:
- field: credentials_field
  key: credentials_key
  path: credentials_field
```

```
helm-helper command -f prometheus.yaml
helm upgrade --install ritmx/prometheus prom --version 0.4.4 --namespace log --values prometheus.yaml --set gcssa=$(shell vault read -field google_credentials_file secret/gcp/credentials)
```
