# Documentation de configuration

## 📋 Vue d'ensemble

VaultCertsViewer peut être configuré via un fichier `settings.json` ou des variables d'environnement. Le fichier de configuration a la priorité sur les variables d'environnement. Ce panneau d'administration vous permet de gérer le fichier `settings.json` directement depuis l'interface web.

> **⚠️ Important :** Après avoir enregistré les modifications, un redémarrage du serveur peut être nécessaire pour que tous les changements prennent effet.

## 🔐 Accès au panneau d'administration

### VCV_ADMIN_PASSWORD

Variable d'environnement requise pour activer le panneau d'administration. Doit être un **hash bcrypt** (préfixe `$2a$`, `$2b$`, ou `$2y$`).

```bash
# Générer un hash bcrypt (exemple avec htpasswd)
htpasswd -nbBC 10 admin VotreMotDePasseSecurise | cut -d: -f2

# Ou avec Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'VotreMotDePasse', bcrypt.gensalt()).decode())"

# Définir la variable d'environnement
export VCV_ADMIN_PASSWORD='$2a$10$...'
```

**Nom d'utilisateur par défaut :** `admin`  
**Durée de session :** 12 heures  
**Limitation de débit :** 10 tentatives par 5 minutes (production uniquement)

## 📁 Paramètres de l'application

### Environnement (app.env)

Définit l'environnement de l'application. Affecte les fonctionnalités de sécurité et le comportement des logs.

- `dev` - Mode développement (logs verbeux, pas de limitation de débit)
- `stage` - Environnement de staging
- `prod` - Mode production (cookies sécurisés, limitation de débit activée)

```bash
# Variable d'environnement (fallback)
export APP_ENV=prod
```

### Port (app.port)

Port d'écoute du serveur HTTP. Par défaut : `52000`

```bash
# Variable d'environnement (fallback)
export PORT=52000
```

### Journalisation (app.logging)

Configurer le comportement de la journalisation :

- **level** : `debug`, `info`, `warn`, `error` (par défaut : `info`)
- **format** : `json` ou `text` (par défaut : `json`)
- **output** : `stdout`, `file`, ou `both` (par défaut : `stdout`)
- **file_path** : Chemin du fichier de log quand output est `file` ou `both` (par défaut : `/var/log/app/vcv.log`)

```bash
# Variables d'environnement (fallback)
export LOG_LEVEL=info
export LOG_FORMAT=json
export LOG_OUTPUT=stdout
export LOG_FILE_PATH=/var/log/app/vcv.log
```

## 📜 Paramètres des certificats

### Seuils d'expiration

Configurer quand les certificats sont signalés comme expirant bientôt :

- **critical** : Jours avant expiration pour afficher une alerte critique (par défaut : `7`)
- **warning** : Jours avant expiration pour afficher un avertissement (par défaut : `30`)

Ces seuils contrôlent :

- La bannière de notification en haut de la page
- Le code couleur dans le tableau des certificats (rouge pour critique, jaune pour avertissement)
- La visualisation de la timeline sur le tableau de bord
- Les métriques Prometheus (`vcv_certificates_expiring_critical`, `vcv_certificates_expiring_warning`)

```bash
# Variables d'environnement (fallback)
export VCV_EXPIRE_CRITICAL=7
export VCV_EXPIRE_WARNING=30
```

## 🌐 Paramètres CORS

### Origines autorisées

Liste séparée par des virgules des origines CORS autorisées. Utilisez `*` pour autoriser toutes les origines (non recommandé en production).

```text
# Exemple
https://example.com,https://app.example.com
```

**Note :** CORS est principalement utile si vous intégrez VCV dans une autre application web ou y accédez depuis un domaine différent.

## 🔐 Configuration Vault

### Instances Vault multiples

VaultCertsViewer prend en charge la surveillance de plusieurs instances Vault simultanément. Chaque instance Vault nécessite :

- **ID** : Identifiant unique pour cette instance Vault (requis)
- **Display Name** : Nom lisible affiché dans l'interface (optionnel)
- **Address** : URL du serveur Vault (ex : `https://vault.example.com:8200`)
- **Token** : Token Vault en lecture seule avec accès PKI (requis)
- **PKI Mounts** : Liste séparée par des virgules des chemins de montage PKI (ex : `pki,pki2,pki-prod`)
- **Enabled** : Si cette instance Vault est active

### Configuration TLS

Pour les Vaults utilisant des certificats CA personnalisés ou auto-signés :

- **TLS CA Cert (Base64)** : Bundle CA PEM encodé en base64 (méthode préférée)
- **TLS CA Cert Path** : Chemin du fichier vers le bundle CA PEM
- **TLS CA Path** : Répertoire contenant les certificats CA
- **TLS Server Name** : Remplacement du nom de serveur SNI
- **TLS Insecure** : Ignorer la vérification TLS (⚠️ développement uniquement, non recommandé)

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

## ⚡ Optimisations de performance

### Cache

VaultCertsViewer implémente un cache intelligent pour améliorer les performances :

- **TTL du cache des certificats :** 15 minutes (réduit les appels API Vault)
- **Cache des vérifications de santé :** 30 secondes (pour les indicateurs de statut du footer)
- **Récupération parallèle :** Plusieurs Vaults sont interrogés simultanément

Avec plusieurs Vaults, la récupération parallèle offre des temps de chargement **3 à 10× plus rapides** par rapport aux requêtes séquentielles.

## 📊 Surveillance & Métriques

### Métriques prometheus

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

### Endpoints de Santé

- `/api/health` - Vérification de santé basique (retourne toujours 200 OK)
- `/api/ready` - Sonde de disponibilité (vérifie l'état de l'application)
- `/api/status` - Statut détaillé incluant toutes les connexions Vault
- `/api/version` - Informations de version de l'application

## 🔒 Bonnes pratiques de sécurité

- Toujours utiliser l'environnement `prod` en production
- Utiliser des mots de passe hashés bcrypt pour l'accès admin
- Ne jamais utiliser `tls_insecure: true` en production
- Protéger le fichier `settings.json` (contient des tokens sensibles)
- Utiliser des tokens Vault en lecture seule avec permissions minimales
- Activer la limitation de débit en production (automatique en mode `prod`)
- Exécuter le conteneur avec `--read-only` et `--cap-drop=ALL`
- Monter le répertoire de logs en lecture-écriture si vous utilisez la journalisation fichier

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

> **💡 Astuce :** Utilisez le panneau d'administration pour éditer ces paramètres visuellement. Les modifications sont enregistrées automatiquement dans `settings.json`.
