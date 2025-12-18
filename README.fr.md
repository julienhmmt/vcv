# VaultCertsViewer 🔐

VaultCertsViewer (vcv) est une interface web légère qui permet de lister et de consulter les certificats stockés dans un ou plusieurs coffres 'pki' d'HashiCorp Vault. Elle affiche notamment les noms communs, les SAN et surtout les dates d'expiration des certificats.

VaultCertsViewer (vcv) peut surveiller simultanément plusieurs moteurs PKI via une seule interface, avec un sélecteur modal pour choisir les montages à afficher. Grâce au fichier de configuration `settings.json`, VCV peut se connecter à plusieurs instances Vault et montages PKI.

## ✨ Quelles sont les fonctionnalités ?

- Découvre tous les certificats d'une ou plusieurs moteurs PKI dans Vault et les affiche dans un tableau filtrable et recherchable.
- Support multi-moteurs PKI : Sélectionnez les montages à afficher via une interface modale intuitive avec des badges de comptage de certificats en temps réel.
- Affichage des noms communs (CN) et des SANs des certificats.
- Affiche la répartition des statuts (valide / expiré / révoqué) et les dates d'expirations à venir.
- Met en avant les certificats qui expirent bientôt (7/30 jours) et affiche les détails (CN, SAN, empreintes, émetteur, validité).
- Choix de la langue de l'UI (en, fr, es, de, it) et le thème (clair/sombre).
- Surveillance en temps réel de la connexion Vault avec notifications toast en cas de perte/rétablissement.

## 🎯 Pourquoi cet outil existe-t-il ?

L'interface de Vault est trop lourde et complexe pour consulter les certificats. Elle ne permet pas **facilement** et rapidement de consulter les dates d'expiration et les détails des certificats.

VaultCertsViewer permet aux équipes plateforme / sécurité / ops une vue rapide et en **lecture seule** sur l'inventaire PKI Vault avec les seules informations nécessaires et utiles.

## 👥 À qui s'adresse-t-il ?

- Aux equipes exploitant l'outil Vault PKI qui ont besoin de visibilité sur leurs certificats.
- Aux opérateurs qui veulent une vue navigateur prête à l’emploi, à côté de la CLI ou de la Web UI de Vault.

## 🚀 Comment le déployer et l'utiliser ?

Dans HashiCorp Vault, créez un rôle et un jeton en lecture seule pour l'API afin d'accéder aux certificats des moteurs PKI ciblés. Pour plusieurs montages, vous pouvez spécifier chaque montage explicitement ou utiliser des motifs génériques :

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

Ce jeton dédié limite les droits à la consultation des certificats, peut être renouvelé et sert de valeur `VAULT_READ_TOKEN` pour l'application.

## 🧩 Support multi-moteurs PKI

VaultCertsViewer peut surveiller simultanément plusieurs moteurs PKI via une seule interface web :

- **Sélection des montages** : Cliquez sur le bouton de sélecteur de montage dans l'en-tête pour ouvrir une fenêtre modale montrant tous les moteurs PKI disponibles
- **Comptages en temps réel** : Chaque montage affiche un badge indiquant le nombre de certificats qu'il contient
- **Configuration flexible** : Spécifiez les montages en utilisant des valeurs séparées par des virgules dans `VAULT_PKI_MOUNTS` (par exemple, `pki,pki2,pki-prod`)
- **Vues indépendantes** : Sélectionnez ou désélectionnez n'importe quelle combinaison de montages pour personnaliser votre vue des certificats
- **Tableau de bord** : Tous les montages sélectionnés sont agrégés dans le même tableau, tableau de bord et métriques
- **Recherche en temps réel** : Filtrage instantané pendant la saisie avec délai de 300ms
- **Filtrage par statut** : Filtres rapides pour les certificats valides/expirés/révoqués
- **Timeline d'expiration** : Visualisation temporelle de la distribution des expirations
- **Pagination** : Taille de page configurable (25/50/75/100/tout) avec contrôles de navigation
- **Options de tri** : Tri par nom commun, date d'expiration ou numéro de série

Cette approche élimine le besoin de déployer plusieurs instances vcv lorsque vous avez plusieurs moteurs PKI à surveiller.

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
  -p 52000:52000 jhmmt/vcv:1.4
```

## 🔐 Configuration TLS Vault

VCV supporte la configuration TLS de Vault via `settings.json` (recommandé) ou via des variables d’environnement (fallback historique).

Par instance Vault (`vaults[]`), vous pouvez configurer :

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
- Pour mettre un certificat en base64, faites la commande `cat chemin-vers-cert.pem | base64 | tr -d '\n'`, copiez le résultat et collez-le dans le champ.

Le panneau d'administration (`/admin`, activé via `VCV_ADMIN_PASSWORD`) permet de définir ces champs TLS par Vault.

## ⏱️ Seuils d'expiration des certificats

Par défaut, VaultCertsViewer alerte sur les certificats expirant dans **7 jours** (critique) et **30 jours** (avertissement). Vous pouvez personnaliser ces seuils dans `settings.json` sous `certificates.expiration_thresholds`.

```text
"certificates": {
  "expiration_thresholds": {
    "critical": 14,
    "warning": 60
  }
}
```

Les variables d'environnement historiques (`VCV_EXPIRE_CRITICAL`, `VCV_EXPIRE_WARNING`) restent supportées en fallback.

Ces valeurs contrôlent :

- La banneau de notification en haut de la page
- Le code couleur dans le tableau des certificats (rouge pour critique, jaune pour avertissement)
- La visualisation de la chronologie sur le tableau de bord
- Le nombre de certificats « expirant bientôt » dans le tableau de bord

## 🌍 Multilingue

L'UI est localisée en *anglais*, *français*, *espagnol*, *allemand* et *italien*. La langue se choisit dans l'en-tête via un bouton ou saisissant dans l'URL le composant `?lang=xx`.

## 📊 Exporter des métriques vers Prometheus

Les métriques sont exposées sur l’endpoint `/metrics`.

- vcv_cache_size
- vcv_certificate_exporter_last_scrape_duration_seconds
- vcv_certificate_expiry_timestamp_seconds{certificate_id, common_name, status, vault_id, pki} (optionnel)
- vcv_certificate_exporter_last_scrape_success
- vcv_certificates_expired_count
- vcv_certificates_expiring_soon_count{vault_id, pki, level}
- vcv_certificates_last_fetch_timestamp_seconds
- vcv_certificates_total{vault_id, pki, status}
- vcv_vault_connected{vault_id}
- vcv_vault_list_certificates_success{vault_id}
- vcv_vault_list_certificates_error{vault_id}
- vcv_vault_list_certificates_duration_seconds{vault_id}
- vcv_certificates_partial_scrape{vault_id}
- vcv_vaults_configured
- vcv_pki_mounts_configured{vault_id}

Pour configurer le scraping côté Prometheus :

```yaml
scrape_configs:
  - job_name: vcv
    static_configs:
      - targets: ['localhost:52000']
    metrics_path: /metrics
```

Example scrape output (truncated):

```bash
$ curl -v http://localhost:52000/metrics
...
# HELP vcv_cache_size Number of items currently cached
# TYPE vcv_cache_size gauge
vcv_cache_size 0
# HELP vcv_certificate_exporter_last_scrape_duration_seconds Duration of the last certificate scrape in seconds
# TYPE vcv_certificate_exporter_last_scrape_duration_seconds gauge
vcv_certificate_exporter_last_scrape_duration_seconds 0.000118208
# HELP vcv_certificate_exporter_last_scrape_success Whether the last scrape succeeded (1) or failed (0)
# TYPE vcv_certificate_exporter_last_scrape_success gauge
vcv_certificate_exporter_last_scrape_success 1
# HELP vcv_certificates_expired_count Number of expired certificates
# TYPE vcv_certificates_expired_count gauge
vcv_certificates_expired_count 30
# HELP vcv_certificates_expiring_soon_count Number of certificates expiring soon within threshold window
# TYPE vcv_certificates_expiring_soon_count gauge
vcv_certificates_expiring_soon_count{level="critical",pki="__all__",vault_id="__all__"} 17
vcv_certificates_expiring_soon_count{level="critical",pki="pki",vault_id="vault-main"} 3
vcv_certificates_expiring_soon_count{level="critical",pki="pki_blockchain",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_cloud",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_corporate",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_dev",vault_id="vault-main"} 1
vcv_certificates_expiring_soon_count{level="critical",pki="pki_dmz",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_edge",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_external",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_internal",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_iot",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_lab",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_partners",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_perf",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_production",vault_id="vault-main"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_qa",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_shared",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_stage",vault_id="vault-main"} 1
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault2",vault_id="vault-dev-2"} 2
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault3",vault_id="vault-dev-3"} 2
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault4",vault_id="vault-dev-4"} 4
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault5",vault_id="vault-dev-5"} 4
vcv_certificates_expiring_soon_count{level="warning",pki="__all__",vault_id="__all__"} 45
vcv_certificates_expiring_soon_count{level="warning",pki="pki",vault_id="vault-main"} 7
vcv_certificates_expiring_soon_count{level="warning",pki="pki_blockchain",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_cloud",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_corporate",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_dev",vault_id="vault-main"} 2
vcv_certificates_expiring_soon_count{level="warning",pki="pki_dmz",vault_id="vault-dev-5"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_edge",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_external",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_internal",vault_id="vault-dev-5"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_iot",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_lab",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_partners",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_perf",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_production",vault_id="vault-main"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_qa",vault_id="vault-dev-4"} 6
vcv_certificates_expiring_soon_count{level="warning",pki="pki_shared",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_stage",vault_id="vault-main"} 2
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault2",vault_id="vault-dev-2"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault3",vault_id="vault-dev-3"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault4",vault_id="vault-dev-4"} 4
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault5",vault_id="vault-dev-5"} 4
# HELP vcv_certificates_last_fetch_timestamp_seconds Timestamp of last successful certificates fetch
# TYPE vcv_certificates_last_fetch_timestamp_seconds gauge
vcv_certificates_last_fetch_timestamp_seconds 1.765985686e+09
# HELP vcv_certificates_total Total certificates grouped by status
# TYPE vcv_certificates_total gauge
vcv_certificates_total{pki="__all__",status="expired",vault_id="__all__"} 30
vcv_certificates_total{pki="__all__",status="revoked",vault_id="__all__"} 14
vcv_certificates_total{pki="__all__",status="valid",vault_id="__all__"} 85
vcv_certificates_total{pki="pki",status="expired",vault_id="vault-main"} 3
vcv_certificates_total{pki="pki",status="revoked",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki",status="valid",vault_id="vault-main"} 12
vcv_certificates_total{pki="pki_blockchain",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_blockchain",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_blockchain",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_cloud",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_cloud",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_cloud",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_corporate",status="expired",vault_id="vault-dev-2"} 0
vcv_certificates_total{pki="pki_corporate",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_corporate",status="valid",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_dev",status="expired",vault_id="vault-main"} 1
vcv_certificates_total{pki="pki_dev",status="revoked",vault_id="vault-main"} 2
vcv_certificates_total{pki="pki_dev",status="valid",vault_id="vault-main"} 5
vcv_certificates_total{pki="pki_dmz",status="expired",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_dmz",status="revoked",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_dmz",status="valid",vault_id="vault-dev-5"} 6
vcv_certificates_total{pki="pki_edge",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_edge",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_edge",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_external",status="expired",vault_id="vault-dev-2"} 0
vcv_certificates_total{pki="pki_external",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_external",status="valid",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_internal",status="expired",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_internal",status="revoked",vault_id="vault-dev-5"} 1
vcv_certificates_total{pki="pki_internal",status="valid",vault_id="vault-dev-5"} 6
vcv_certificates_total{pki="pki_iot",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_iot",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_iot",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_lab",status="expired",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_lab",status="revoked",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_lab",status="valid",vault_id="vault-dev-4"} 7
vcv_certificates_total{pki="pki_partners",status="expired",vault_id="vault-dev-2"} 0
vcv_certificates_total{pki="pki_partners",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_partners",status="valid",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_perf",status="expired",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_perf",status="revoked",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_perf",status="valid",vault_id="vault-dev-4"} 1
vcv_certificates_total{pki="pki_production",status="expired",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki_production",status="revoked",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki_production",status="valid",vault_id="vault-main"} 1
vcv_certificates_total{pki="pki_qa",status="expired",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_qa",status="revoked",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_qa",status="valid",vault_id="vault-dev-4"} 7
vcv_certificates_total{pki="pki_shared",status="expired",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_shared",status="revoked",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_shared",status="valid",vault_id="vault-dev-5"} 6
vcv_certificates_total{pki="pki_stage",status="expired",vault_id="vault-main"} 1
vcv_certificates_total{pki="pki_stage",status="revoked",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki_stage",status="valid",vault_id="vault-main"} 5
vcv_certificates_total{pki="pki_vault2",status="expired",vault_id="vault-dev-2"} 5
vcv_certificates_total{pki="pki_vault2",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_vault2",status="valid",vault_id="vault-dev-2"} 6
vcv_certificates_total{pki="pki_vault3",status="expired",vault_id="vault-dev-3"} 5
vcv_certificates_total{pki="pki_vault3",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_vault3",status="valid",vault_id="vault-dev-3"} 6
vcv_certificates_total{pki="pki_vault4",status="expired",vault_id="vault-dev-4"} 7
vcv_certificates_total{pki="pki_vault4",status="revoked",vault_id="vault-dev-4"} 1
vcv_certificates_total{pki="pki_vault4",status="valid",vault_id="vault-dev-4"} 5
vcv_certificates_total{pki="pki_vault5",status="expired",vault_id="vault-dev-5"} 8
vcv_certificates_total{pki="pki_vault5",status="revoked",vault_id="vault-dev-5"} 1
vcv_certificates_total{pki="pki_vault5",status="valid",vault_id="vault-dev-5"} 5
# HELP vcv_vault_connected Vault connection status (1=connected,0=disconnected)
# TYPE vcv_vault_connected gauge
vcv_vault_connected{vault_id="__all__"} 0
vcv_vault_connected{vault_id="vault-dev-2"} 1
vcv_vault_connected{vault_id="vault-dev-3"} 1
vcv_vault_connected{vault_id="vault-dev-4"} 1
vcv_vault_connected{vault_id="vault-dev-5"} 1
vcv_vault_connected{vault_id="vault-dev-6"} 0
vcv_vault_connected{vault_id="vault-main"} 1
```

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

Pour activer la métrique par certificat `vcv_certificate_expiry_timestamp_seconds`, définissez `VCV_METRICS_PER_CERTIFICATE=true`.

Si vous l’activez, vous pouvez adapter librement la fenêtre « bientôt » (ex. 14 jours) directement en PromQL, sans modifier l’exporter.

## 🔐 Admin

Si vous définissez `VCV_ADMIN_PASSWORD`, un panneau d’administration est activé sur `/admin`.

- Le mot de passe peut être fourni en clair ou sous forme de **hash bcrypt**.
- Le panneau admin modifie le fichier de settings configuré, donc `settings.json` doit être monté en écriture.

Le panneau d'administration vous permet d'afficher la liste des vaults et des moteurs PKI associés. En plus de l'affichage, vous pourrez modifier, ajouter et supprimer des points de connexion à tous les vaults dont vous disposez.

## 🔎 Pour aller plus loin

- Documentation technique : [app/README.md](app/README.md)
- Version anglaise : [README.md](README.md)
- Docker Hub : [jhmmt/vcv](https://hub.docker.com/r/jhmmt/vcv)
- Code Source : [github.com/julienhmmt/vcv](https://github.com/julienhmmt/vcv)

## 🖼️ Picture of the app

![VaultCertsViewer v1.4](img/VaultCertsViewer-v1.4.png)

![VaultCertsViewer v1.4 - Light Mode](img/VaultCertsViewer-v1.4-light.png)

![VaultCertsViewer v1.4 - Dark Mode](img/VaultCertsViewer-v1.4-dark.png)
