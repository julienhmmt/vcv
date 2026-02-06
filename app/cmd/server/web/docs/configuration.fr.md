# Référence de configuration

## 📋 Vue d'ensemble

VaultCertsViewer (VCV) est configuré principalement via un fichier `settings.json`. Le panneau d'administration vous permet de gérer ce fichier directement depuis l'interface web. Les variables d'environnement sont supportées comme solution de repli lorsqu'aucun `settings.json` n'est trouvé.

VCV utilise une architecture de rendu côté serveur propulsée par [HTMX](https://htmx.org/). Tous les filtrages, tris et paginations sont gérés côté serveur pour des performances optimales.

> **⚠️ Important :** Après avoir enregistré les modifications, un redémarrage du serveur peut être nécessaire pour que tous les changements prennent effet.

## 🔐 Accès au panneau d'administration

### VCV_ADMIN_PASSWORD

Variable d'environnement requise pour activer le panneau d'administration. Doit être un **hash bcrypt**.

```bash
# Générer un hash bcrypt (exemple avec htpasswd)
htpasswd -nbBC 10 admin VotreMotDePasseSecurise | cut -d: -f2

# Ou avec Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'VotreMotDePasse', bcrypt.gensalt()).decode())"

# Définir la variable d'environnement
export VCV_ADMIN_PASSWORD='$2a$10$...'
```

Vous pouvez également utiliser le service 'bcrypt' de <https://tools.hommet.net/bcrypt> pour générer un hash bcrypt (aucune donnée n'est stockée).

**Nom d'utilisateur par défaut :** `admin` (non modifiable, à titre indicatif)
**Durée de session :** 12 heures (non modifiable, à titre indicatif)
**Limitation de débit de connexion :** 10 tentatives par 5 minutes (non modifiable, à titre indicatif)

## 📁 Paramètres de l'application

### Environnement (app.env)

Définit l'environnement de l'application. Affecte les fonctionnalités de sécurité et le comportement des logs.

- `dev` - Mode développement (logs verbeux, pas de limitation de débit)
- `prod` - Mode production (cookies sécurisés, limitation de débit activée)

**Par défaut :** `prod`

### Port (app.port)

Port d'écoute du serveur HTTP.

**Par défaut :** `52000`

### Chemin du fichier de configuration

La variable d'environnement `SETTINGS_PATH` spécifie le chemin vers le fichier `settings.json`. Si non définie, VCV recherche les fichiers dans cet ordre :

1. `settings.<env>.json` (ex : `settings.dev.json`)
2. `settings.json`
3. `./settings.json`
4. `/etc/vcv/settings.json`

### Journalisation (app.logging)

Configurer le comportement de la journalisation :

- **level** : `debug`, `info`, `warn`, `error`
- **format** : `json` ou `text`
- **output** : `stdout`, `file`, ou `both`
- **file_path** : Chemin du fichier de log quand output est `file` ou `both`

**Par défaut :**

- level: `info`
- format: `json`
- output: `stdout`
- file_path: `/var/log/app/vcv.log`

## 📜 Paramètres des certificats

### Seuils d'expiration (certificates.expiration_thresholds)

Configurer quand les certificats sont signalés comme expirant bientôt :

- **critical** : Jours avant expiration pour afficher une alerte critique
- **warning** : Jours avant expiration pour afficher un avertissement

Ces seuils contrôlent :

- La bannière de notification en haut de la page
- Le code couleur dans le tableau des certificats (rouge pour critique, jaune pour avertissement)
- La visualisation de la timeline sur le tableau de bord
- Les métriques Prometheus (`vcv_certificates_expiring_critical`, `vcv_certificates_expiring_warning`)

**Par défaut :**

- critical: `7`
- warning: `30`

## 🌐 Paramètres CORS (cors)

### Origines autorisées (cors.allowed_origins)

Tableau des origines CORS autorisées. Utilisez `["*"]` pour autoriser toutes les origines (non recommandé en production).

**Exemple :**

```json
"allowed_origins": ["https://example.com", "https://app.example.com"]
```

### Autoriser les credentials (cors.allow_credentials)

Booléen pour autoriser les credentials dans les requêtes CORS.

**Par défaut :** `false`

**Note :** CORS est principalement utile si vous intégrez VCV dans une autre application web ou y accédez depuis un domaine différent.

## 🔐 Configuration Vault

### Instances Vault multiples

VaultCertsViewer prend en charge la surveillance de plusieurs instances Vault simultanément. Chaque instance Vault nécessite :

- **ID** : Identifiant unique pour cette instance Vault (requis)
- **Display name** : Nom lisible affiché dans l'interface (optionnel)
- **Address** : URL du serveur Vault (ex : `https://vault.example.com:8200`)
- **Token** : Token Vault en lecture seule avec accès PKI (requis)
- **PKI mounts** : Tableau de chemins de montage PKI (ex : `["pki", "pki2", "pki-prod"]`)
- **Enabled** : Si cette instance Vault est active

### Configuration TLS

Pour les Vaults utilisant des certificats CA personnalisés ou auto-signés :

- **TLS CA cert (Base64)** : Bundle CA PEM encodé en base64 (méthode préférée)
- **TLS CA cert path** : Chemin du fichier vers le bundle CA PEM
- **TLS CA path** : Répertoire contenant les certificats CA
- **TLS server name** : Remplacement du nom de serveur SNI
- **TLS insecure** : Ignorer la vérification TLS (⚠️ développement uniquement, non recommandé)

```bash
# Encoder un certificat en base64
cat chemin-vers-cert.pem | base64 | tr -d '\n'
```

**Précédence :** Si `tls_ca_cert_base64` est défini, il a la priorité sur les chemins de fichiers.

### Permissions du token vault

Le token Vault doit avoir un accès en lecture seule aux montages PKI. Exemple de politique :

```hcl
# Montages explicites (recommandé pour la production)
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

# Pattern avec wildcard (pour les environnements dynamiques)
vault policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"    { capabilities = ["read"] }
EOF

# Créer le token
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Vous devez remplacer 'pki' et 'pki2' par les chemins de montage PKI de votre Vault. En plus, ajoutez autant de chemins de montage PKI que vous avez dans votre Vault.

## ⚡ Optimisations de performance

### Cache

VaultCertsViewer implémente un cache pour améliorer les performances :

- **TTL du cache des certificats :** 15 minutes (réduit les appels API Vault)
- **Cache des vérifications de santé :** 30 secondes (pour l'indicateur de statut dans l'en-tête)
- **Récupération parallèle :** Plusieurs Vaults sont interrogés simultanément
- **Invalidation du cache :** Utilisez le bouton de rafraîchissement (↻) dans l'en-tête ou `POST /api/cache/invalidate` pour vider le cache des certificats

Avec plusieurs Vaults, la récupération parallèle offre des temps de chargement **3 à 10× plus rapides** par rapport aux requêtes séquentielles.

## 📊 Surveillance & Métriques

### Métriques Prometheus

Disponibles sur l'endpoint `/metrics` :

- `vcv_certificates_total` - Nombre total de certificats
- `vcv_certificates_valid` - Nombre de certificats valides
- `vcv_certificates_expired` - Nombre de certificats expirés
- `vcv_certificates_revoked` - Nombre de certificats révoqués
- `vcv_certificates_expiring_critical` - Certificats expirant dans le seuil critique
- `vcv_certificates_expiring_warning` - Certificats expirant dans le seuil d'avertissement
- `vcv_vault_connected` - Statut de connexion Vault (1=connecté, 0=déconnecté)
- `vcv_cache_size` - Nombre d'entrées en cache
- `vcv_last_fetch_timestamp` - Timestamp Unix de la dernière récupération de certificats

Toutes les métriques incluent les labels : `vault_id`, `vault_name`, `pki_mount`

### Endpoints santé & API

- `/api/health` - Vérification de santé basique (retourne toujours 200 OK)
- `/api/ready` - Sonde de disponibilité (vérifie l'état de l'application)
- `/api/status` - Statut détaillé incluant toutes les connexions Vault
- `/api/version` - Informations de version de l'application
- `/api/config` - Configuration de l'application (seuils d'expiration, liste des vaults)
- `/api/i18n` - Traductions pour la langue courante
- `/api/certs` - Liste des certificats (JSON)
- `/api/certs/{id}/details` - Détails d'un certificat (JSON)
- `/api/certs/{id}/pem` - Contenu PEM d'un certificat (JSON)
- `/api/certs/{id}/pem/download` - Téléchargement du fichier PEM d'un certificat
- `POST /api/cache/invalidate` - Invalidation du cache des certificats

### Limitation de débit

En mode `prod`, la limitation de débit de l'API est activée à **300 requêtes par minute** par client. Les chemins suivants sont exemptés :

- `/api/health`, `/api/ready`, `/metrics`
- `/assets/*` (fichiers statiques)

## 🔒 Bonnes pratiques de sécurité

- Toujours utiliser l'environnement `prod` en production
- Protéger le fichier `settings.json` (contient des tokens sensibles)
- Utiliser des tokens Vault en lecture seule avec permissions minimales
- La limitation de débit est automatique en mode `prod` (300 req/min)
- La protection CSRF est activée sur toutes les requêtes modifiant l'état
- Les en-têtes de sécurité (X-Content-Type-Options, X-Frame-Options, etc.) sont définis automatiquement
- Exécuter le conteneur avec `--read-only` et `--cap-drop=ALL`

## 📝 Exemple settings.json

```json
{
  "app": {
    "env": "prod",
    "port": 52000,
    "logging": {
      "level": "info",
      "format": "json",
      "output": "stdout",
      "file_path": "/var/log/app/vcv.log"
    }
  },
  "certificates": {
    "expiration_thresholds": {
      "critical": 7,
      "warning": 30
    }
  },
  "cors": {
    "allowed_origins": ["https://example.com"],
    "allow_credentials": false
  },
  "vaults": [
    {
      "id": "vault-prod",
      "display_name": "Vault Production",
      "address": "https://vault.example.com:8200",
      "token": "hvs.xxx",
      "pki_mounts": ["pki", "pki-intermediate"],
      "enabled": true,
      "tls_insecure": false,
      "tls_ca_cert_base64": "LS0tLS1CRUdJTi...",
      "tls_server_name": "vault.example.com"
    },
    {
      "id": "vault-dev",
      "display_name": "Vault Développement",
      "address": "https://vault-dev.example.com:8200",
      "token": "hvs.yyy",
      "pki_mounts": ["pki_dev"],
      "enabled": true,
      "tls_insecure": true
    }
  ]
}
```

> **💡 Astuce :** Utilisez le panneau d'administration pour éditer ces paramètres visuellement. Les modifications sont enregistrées dans le fichier `settings.json`.
