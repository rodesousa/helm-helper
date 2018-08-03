## Helm-helper

Outil pour simplifier la CI/CD K8s/Helm

---
## Utilisation

`helm-helper` a besoin d'un fichier de valeur Helm `--values file` pour éxécuter l'ensemble de ces commandes. Un bloc `_metadata` doit être présent dans le fichier et décrire le déploiement helm. On peut (doit) utiliser le fichier de values d'une chart.

Exemple:

```
_metadata:
  chart: ritmx/prometheus
  name: prom
  namespace: log
  vault:
  - field: google_credentials_file
    key: gcssa
    path: secret/gcp/sandbox/thanos-sa
  version: 0.4.4
```

---

`helm-helper check_version --values file`

Vérifie que la version de la chart lors du déploiement soit la dernière disponible dans le dépôt Helm

Flag:

`--url` ou `HELM_URL` Url du dépôt Helm

Par défaut `https://kubernetes-charts.storage.googleapis.com`

`HELM_URL=http://helm.example.com helm-helper check_version --values prometheus.yaml`

---

`helm-helper command --values file`

Génération de la commande helm
