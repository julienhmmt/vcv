# VaultCertsViewer 🔐

![GitHub Release](https://img.shields.io/github/v/release/julienhmmt/vcv?display_name=release&style=for-the-badge) ![GitHub License](https://img.shields.io/github/license/julienhmmt/vcv?style=for-the-badge)

VaultCertsViewer (vcv) est une interface web légère qui permet de lister et de consulter les certificats stockés dans un ou plusieurs coffres PKI HashiCorp Vault ou OpenBao. Elle affiche notamment les noms communs, les SAN et surtout les dates d'expiration des certificats.

VaultCertsViewer (vcv) peut surveiller simultanément plusieurs moteurs PKI via une seule interface, avec un sélecteur modal pour choisir les montages à afficher. Grâce au fichier de configuration `settings.json`, VCV peut se connecter à plusieurs instances Vault/OpenBao et montages PKI.

**Compatible avec OpenBao** : VCV fonctionne avec HashiCorp Vault et OpenBao, car ils partagent la même API PKI. Testé avec OpenBao 2.4+ et Vault 1.20+ (au 02/2026).

## ✨ Quelles sont les fonctionnalités ?

- Découvre tous les certificats d'un ou plusieurs moteurs PKI dans Vault/OpenBao et les affiche dans un tableau filtrable et recherchable.
- Support multi-Vault : connexion simultanée à plusieurs instances Vault/OpenBao.
- Support multi-moteurs PKI : activez ou désactivez les moteurs PKI désirés.
- Affichage des noms communs (CN) et des SANs des certificats, leurs dates de création et d'**expiration**, leur statut (valide / expiré / révoqué).
- Met en avant les certificats qui expirent bientôt avec des seuils configurables (par défaut : 7 jours critique, 30 jours avertissement).
- Choix de la langue (en, fr, es, de, it) et du thème (clair/sombre).
- Panneau d'administration : gestion de la configuration via interface web (optionnel, protégé par bcrypt).
- Métriques Prometheus : voir le document [PROMETHEUS_METRICS.md](PROMETHEUS_METRICS.md).

## 🎯 Pourquoi cet outil existe-t-il ?

L'interface de Vault/OpenBao est lourde et complexe pour consulter les certificats rapidement. Elle ne permet pas **facilement** de consulter les dates d'expiration et les détails des certificats.

VaultCertsViewer permet aux équipes plateforme et sécurité d'avoir une vue rapide et en **lecture seule** sur l'inventaire des certificats dans un ou plusieurs Vault/OpenBao avec les seules informations nécessaires.

## 👥 À qui s'adresse-t-il ?

- Aux equipes exploitant l'outil Vault/OpenBao qui ont besoin d'une visibilité rapide et en lecture seule sur les certificats.
- Aux opérateurs qui veulent une vue claire et simple dans leur navigateur.

## 🚀 Comment le déployer et l'utiliser pour Hashicorp Vault ?

Dans Hashicorp Vault, créez un rôle et un jeton en lecture seule pour l'API afin d'accéder aux certificats des moteurs PKI ciblés. Pour plusieurs montages, vous pouvez spécifier chaque montage explicitement ou utiliser des motifs génériques :

```bash
# Option 1 : Montages explicites (recommandé pour la production). Remplacez 'pki' et 'pki2' par vos montages réels.
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

# Option 2 : Motif générique (pour environnements dynamiques)
vault policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"     { capabilities = ["read"] }
EOF

vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Ce jeton dédié limite les droits à la consultation des certificats, peut être renouvelé et est utilisé dans le fichier `settings.json`.

## 🚀 Comment le déployer et l'utiliser pour OpenBao

Dans OpenBao, créez un rôle et un jeton en lecture seule pour l'API afin d'accéder aux certificats des moteurs PKI ciblés. Les commandes sont similaires à Vault mais utilisent la CLI `bao` :

```bash
# Option 1 : Montages explicites (recommandé pour la production). Remplacez 'pki' et 'pki2' par vos montages réels.
bao policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

# Option 2 : Motif générique (pour environnements dynamiques)
bao policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"     { capabilities = ["read"] }
EOF

bao write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
bao token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Ce jeton dédié limite les droits à la consultation des certificats, peut être renouvelé et est utilisé dans le fichier `settings.json`.

## 🧩 Support multi-moteurs PKI

VaultCertsViewer peut surveiller simultanément plusieurs moteurs PKI via une seule interface web :

- **Sélection des montages** : Cliquez sur le bouton "Sources des certificats" dans l'en-tête pour ouvrir une fenêtre montrant tous les montages disponibles
- **Comptages en temps réel** : Chaque montage affiche un badge indiquant le nombre de certificats qu'il contient
- **Configuration flexible** : Spécifiez les montages en utilisant des valeurs séparées par des virgules dans le fichier `settings.json` (par exemple, `pki,pki2,pki-prod`) ou via l'interface d'administration.
- **Support multi-Vault** : Connexion simultanée à plusieurs instances Vault/OpenBao via le fichier `settings.json`
- **Tableau de bord** : Tous les montages sélectionnés sont agrégés dans le même tableau, tableau de bord et métriques
- **Recherche en temps réel** : Filtrage instantané pendant la saisie avec délai de 300ms
- **Filtrage par statut** : Filtres rapides pour les certificats valides/expirés/révoqués
- **Répartition** : Visualisation de la répartition des certificats par statut
- **Pagination** : Taille de page configurable (25/50/75/100/tout) avec contrôles de navigation
- **Options de tri** : Tri par vault, moteur PKI, nom commun, date de création et date d'expiration

### 🐳 docker-compose

La manière recommandée de configurer vcv est via un fichier `settings.json`.

1. Copiez le fichier d'exemple et modifiez-le :

```bash
cp settings.example.json settings.json
```

1. Montez-le dans le conteneur sous `/app/settings.json` puis lancez :

```bash
docker compose up -d
```

Si vous avez configuré `app.logging.output` pour écrire les logs dans un fichier, vous devrez monter un répertoire en lecture/écriture, par exemple :

```bash
-v "$(pwd)/logs:/var/log/app:rw"
```

### 🐳 docker run

Lancez rapidement le container avec cette commande:

```bash
docker run -d \
  -v "$(pwd)/settings.json:/app/settings.json:rw" \
  -v "$(pwd)/logs:/var/log/app:rw" \
  --cap-drop=ALL --read-only --security-opt no-new-privileges:true \
  -p 52000:52000 jhmmt/vcv:1.6
```

## 🔐 Configuration TLS Vault/OpenBao

VCV supporte la configuration TLS de Vault/OpenBao via `settings.json`.

Par instance Vault ou OpenBao (`vaults[]`), vous pouvez configurer :

- **`tls_ca_cert_base64`** : bundle CA PEM encodé en base64 (recommandé)
- **`tls_ca_cert`** : chemin vers un fichier PEM (bundle CA)
- **`tls_ca_path`** : répertoire contenant des certificats CA
- **`tls_server_name`** : surcharge du nom serveur (SNI)
- **`tls_insecure`** : désactive la vérification TLS (uniquement en développement)

Règles de priorité :

- Si `tls_ca_cert_base64` est renseigné, il est utilisé et `tls_ca_cert` / `tls_ca_path` sont ignorés.
- Sinon, `tls_ca_cert` / `tls_ca_path` sont utilisés (s’ils sont renseignés).

Notes :

- Base64 n’est pas un chiffrement. Considérez `settings.json` comme sensible.
- La valeur base64 doit encoder les bytes PEM (un ou plusieurs blocs `-----BEGIN CERTIFICATE-----`). Les encodages base64 standard et « raw » sont acceptés.
- Pour encoder un certificat en base64 : `cat path-to-cert.pem | base64 | tr -d '\n'`.

## 🛠️ Panneau d'administration

Un panneau d'administration permet de configurer plusieurs paramètres de l'application. Il est accessible via la route `/admin` et est protégé par un mot de passe. Pour activer le panneau d'administration, vous devez inclure une section `admin` dans votre fichier `settings.json` avec un hash de mot de passe bcrypt.

Les fonctionnalités du panneau d'administration sont les suivantes :

- Configuration des seuils d'expiration des certificats
- Configuration des CORS
- Configuration des instances Vault/OpenBao (adresse, port, token, TLS, montages PKI à surveiller)

Le champ `admin.password` doit contenir une valeur **hash bcrypt** (préfixe `$2a$`, `$2b$` ou `$2y$`).

Si le champ est absent ou n'est pas un hash bcrypt, la route `/admin` est désactivée et le panneau d'administration est inaccessible.

## ⏱️ Seuils d'expiration des certificats

Par défaut, VaultCertsViewer alerte sur les certificats expirant dans **7 jours** (critique) et **30 jours** (avertissement). Vous pouvez personnaliser ces seuils dans `settings.json` sous `certificates.expiration_thresholds`.

```json
"certificates": {
  "expiration_thresholds": {
    "critical": 14,
    "warning": 60
  }
}
```

Ces valeurs contrôlent :

- Le code couleur dans le tableau des certificats (rouge pour critique, jaune pour avertissement)
- Le nombre de certificats « expirant bientôt » dans le tableau de bord

## 🌍 Multilingue

L'UI est localisée en *anglais*, *français*, *espagnol*, *allemand* et *italien*. La langue se choisit dans l'en-tête via un bouton ou saisissant dans l'URL le composant `?lang=xx`.

## 📊 Exporter des métriques vers Prometheus

Les métriques sont exposées sur l’endpoint `/metrics`.

**Métriques principales:**

- vcv_certificates_total{vault_id, pki, status}
- vcv_certificates_expired_count
- vcv_certificates_expiring_soon_count{vault_id, pki, level} - Uses configured thresholds
- vcv_expiration_threshold_critical_days - Configured critical threshold
- vcv_expiration_threshold_warning_days - Configured warning threshold
- vcv_certificates_expiry_bucket{vault_id, pki, bucket} - Certificate distribution by time range
- vcv_vault_connected{vault_id}
- vcv_vault_list_certificates_success{vault_id}
- vcv_vault_list_certificates_error{vault_id}
- vcv_vault_list_certificates_duration_seconds{vault_id}
- vcv_certificates_partial_scrape{vault_id}
- vcv_vaults_configured
- vcv_pki_mounts_configured{vault_id}
- vcv_cache_size
- vcv_certificates_last_fetch_timestamp_seconds
- vcv_certificate_exporter_last_scrape_success
- vcv_certificate_exporter_last_scrape_duration_seconds

**Métriques par certificat** (haute cardinalité, désactivées par défaut):

- vcv_certificate_expiry_timestamp_seconds{certificate_id, common_name, status, vault_id, pki}
- vcv_certificate_days_until_expiry{certificate_id, common_name, status, vault_id, pki}

**Configuration avancée:**

Des métriques plus détaillées peuvent être activées dans le fichier `settings.json` ou via le panneau d'administration :

```json
{
  "metrics": {
    "per_certificate": false,
    "enhanced_metrics": true
  }
}
```

Documentation complète : [Documentation des métriques](PROMETHEUS_METRICS.md)

Exemple de résultat de métriques : [METRICS_EXAMPLE.txt](METRICS_EXAMPLE.txt)

Pour configurer le scraping côté Prometheus (exemple avec VCV exposant le port 52000) :

```yaml
scrape_configs:
  - job_name: vcv
    static_configs:
      - targets: ['<your-vcv-host>:52000']
    metrics_path: /metrics
```

N'oubliez pas de changer l'adresse et le port selon votre configuration.

## 🛎️ Alertes avec AlertManager

Si vous utilisez AlertManager, vous pouvez créer des alertes à partir de ces métriques.

Approche recommandée :

- Privilégier les métriques agrégées (`vcv_certificates_expiring_soon_count`, `vcv_certificates_total`) pour l’alerting.
- Utiliser la métrique par certificat uniquement pour investiguer / faire du drill-down (désactivée par défaut car potentiellement très cardinalisée).

Exemples de règles d’alertes (compatibles multi-vault) :

```yaml
- alert: VCVExporterScrapeFailed
  expr: vcv_certificate_exporter_last_scrape_success == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Échec du scrape de l’exporter VCV"
    description: "L’exporter n’a pas pu lister les certificats lors du dernier scrape."

- alert: VCVVaultDown_Global
  expr: vcv_vault_connected{vault_id="__all__"} == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Au moins un Vault est indisponible"
    description: "L’exporter ne parvient pas à se connecter à une ou plusieurs instances Vault."

- alert: VCVVaultDown
  expr: vcv_vault_connected{vault_id!="__all__"} == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Vault indisponible ({{ $labels.vault_id }})"
    description: "L’exporter ne parvient pas à se connecter au Vault '{{ $labels.vault_id }}'."

- alert: VCVVaultListingError
  expr: vcv_vault_list_certificates_error{vault_id!="__all__"} == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Impossible de lister les certificats ({{ $labels.vault_id }})"
    description: "Le listing des certificats a échoué pour le Vault '{{ $labels.vault_id }}'."

- alert: VCVPartialScrape
  expr: vcv_certificates_partial_scrape{vault_id="__all__"} == 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Scrape partiel VCV"
    description: "Au moins un Vault a échoué pendant le listing ; les compteurs agrégés peuvent être incomplets."

- alert: VCVStaleInventory
  expr: time() - vcv_certificates_last_fetch_timestamp_seconds > 3600
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Inventaire VCV périmé"
    description: "L’exporter n’a pas rafraîchi les certificats depuis plus d’une heure."

- alert: VCVExpiringSoonCritical
  expr: sum by (vault_id, pki) (vcv_certificates_expiring_soon_count{level="critical"}) > 0
  labels:
    severity: critical
  annotations:
    summary: "Certificats bientôt expirés (critique)"
    description: "{{ $value }} certificats expirent dans la fenêtre critique (vault={{ $labels.vault_id }}, pki={{ $labels.pki }})."

- alert: VCVExpiringSoonWarning
  expr: sum by (vault_id, pki) (vcv_certificates_expiring_soon_count{level="warning"}) > 0
  labels:
    severity: warning
  annotations:
    summary: "Certificats bientôt expirés (warning)"
    description: "{{ $value }} certificats expirent dans la fenêtre warning (vault={{ $labels.vault_id }}, pki={{ $labels.pki }})."
```

### Fonctionnalités de sécurité

- **Limitation de débit** : Activée en mode production (300 requêtes/minute, exemptant les endpoints health/ready/metrics)
- **Protection CSRF** : Toutes les requêtes modifiant l'état nécessitent des tokens CSRF
- **En-têtes de sécurité** : Inclut HSTS, X-Frame-Options, X-Content-Type-Options, CSP
- **Suivi des requêtes** : Toutes les requêtes incluent des ID uniques pour la corrélation des logs
- **Limites de taille** : Taille maximale de corps de requête de 1 Mo

## 🔎 Pour aller plus loin

- Documentation technique : [app/README.md](app/README.md)
- Version anglaise : [README.md](README.md)
- Docker Hub : [jhmmt/vcv](https://hub.docker.com/r/jhmmt/vcv)
- Code Source : [github.com/julienhmmt/vcv](https://github.com/julienhmmt/vcv)
