# VCV

Suivre des certificats TLS répartis sur plusieurs moteurs PKI Vault devient vite pénible : expirations oubliées, renouvellements en urgence, manque de visibilité.

**VCV (VaultCertificatesViewer)** est une petite application web auto-hébergée qui offre une vue claire et rapide de vos certificats émis par Vault — pour agir avant que les incidents n’arrivent.

## Quel problème VCV résout-il ?

Les organisations qui utilisent **Hashicorp Vault** doivent souvent gérer :

- Plusieurs instances Vault et environnements
- Plusieurs montages PKI (root, intermediate, équipes/applications dédiées)
- Des certificats qui expirent à des rythmes différents, à des endroits différents
- Le besoin d’une interface simple et unique

VCV se concentre sur l’essentiel : **visibilité et clareté**.

## Ce que vous obtenez

### Un inventaire clair des certificats

VCV liste les certificats émis par vos moteurs PKI Vault avec les informations utiles :

- Common Name (CN)
- Émetteur / montage source
- Date d’expiration (le plus important)
- Jours restants
- Statut (révoqué, expiré, valide)

L’objectif : répondre rapidement à des questions comme :

- « Quels certificats expirent bientôt ? »
- « Quel montage PKI produit le plus de certificats à courte durée de vie ? »
- « Sommes-nous tranquilles pour les X prochains jours ? »

### Des seuils d’expiration configurables

Les politiques varient selon les environnements. VCV propose des **seuils d’expiration configurables** (par exemple : warning à 30 jours, critique à 7 jours) afin de coller à vos pratiques opérationnelles sans modifier le code.

### Une interface web simple

VCV est conçu pour être :

- Facile à exécuter (auto-hébergé)
- Rapide à utiliser
- Accessible aux opérateurs

Pas de complexité inutile : juste le tableau de bord.

### Prêt pour la supervision (métriques & alerting)

VCV expose des métriques pour s’intégrer à votre stack de supervision (ex. Prometheus ou VictoriaMetrics). Vous pouvez ainsi créer des alertes telles que :

- « Des certificats expirent dans < 7 jours »
- « Problème de connectivité à Vault »
- « Dernière récupération trop ancienne »

## Comment VCV s’intègre dans votre workflow

VCV ne cherche pas à remplacer Vault, ni votre supervision. Il complète l’ensemble :

- Vault reste la source de vérité
- VCV rend les données **faciles à consommer**
- Prometheus/Alertmanager (ou équivalent) déclenche les alertes au bon moment

Concrètement, les équipes utilisent VCV comme :

- Un tableau de bord de vérification quotidien/hebdo
- Un outil de prévention d’incidents
- Une vue rapide en cas de diagnostic sur la PKI Vault

## Pour qui ?

VCV est particulièrement utile si vous :

- Exploitez Vault PKI dans une instance ou plusieurs
- Gérez plusieurs montages PKI
- Supportez de nombreux services internes avec des certificats à durée de vie courte
- Cherchez une UI légère pour repérer les risques tôt

## Images de l'application

Placeholder

## Pourquoi VCV est volontairement "petit”

Beaucoup de plateformes de gestion de certificats deviennent de véritables écosystèmes. VCV reste volontairement focalisé :

- Complexité opérationnelle minimale
- Déploiement et mises à jour simples
- UI claire et valeur immédiate

Si vous faites déjà confiance à Vault pour la PKI, VCV vous aide à **faire confiance à votre visibilité**.

## Administration de l'application

VCV est pensé pour rester facile à exploiter : les administrateurs configurent la connexion à Vault (adresses et authentification), sélectionnent les montages PKI à exposer et ajustent les seuils d’expiration selon la politique de l’organisation.

Via une page protégée par un mot de passe, la zone d'administration vous présentera des formulaires pour ajouter tous les chemins vers les instances Vault que vous gérez. Toutes les informations sont stockées dans un fichier au format JSON.

VCV ne remplace pas la gouvernance Vault : le contrôle d’accès et la gestion des secrets restent assurés par Vault ; l’enjeu côté admin est donc la configuration sûre, l’observabilité et la rotation régulière des identifiants.

## Déploiement avec docker

### Prérequis de mise en place

Les informations de configuration de l'application sont stockées dans un fichier `settings.json`. Vous devez créer ce fichier avant de lancer le conteneur.

Créez ce fichier `settings.json` dans le répertoire de travail, et saisissez ces informations en remplaçant les valeurs avec vos propres données :

```json
{
  "admin": {
    "password": "$2y$10$.changeme"
  },
  "app": {
    "env": "prod",
    "logging": {
      "level": "debug",
      "format": "json",
      "output": "both",
      "file_path": "/var/log/app/vcv.log"
    },
    "port": 52000
  },
  "certificates": {
    "expiration_thresholds": {
      "critical": 2,
      "warning": 10
    }
  },
  "metrics": {
    "per_certificate": true,
    "enhanced_metrics": false
  },
  "cors": {
    "allowed_origins": [
      "https://172.16.20.50:8443",
      "https://reproxy.vcv.local"
    ],
    "allow_credentials": true
  },
  "vaults": [
    {
      "id": "vault-main",
      "address": "http://vault:8200",
      "token": "root",
      "pki_mounts": ["pki", "pki_dev", "pki_stage", "pki_production"],
      "display_name": "Vault",
      "tls_ca_cert_base64": "BASE64_PEM_CA_BUNDLE",
      "tls_ca_cert": "",
      "tls_ca_path": "",
      "tls_server_name": "vault.service.consul",
      "tls_insecure": true,
      "enabled": true
    },
    {
      "id": "vault-dev",
      "address": "http://vault-dev:8200",
      "token": "root",
      "pki_mounts": ["pki", "pki_corporate", "pki_external", "pki_partners"],
      "display_name": "Vault dev",
      "tls_ca_cert_base64": "BASE64_PEM_CA_BUNDLE",
      "tls_ca_cert": "",
      "tls_ca_path": "",
      "tls_server_name": "vault-dev.service.consul",
      "tls_insecure": true,
      "enabled": true
    }
  ]
}
```

### Test rapide avec la commande docker run

Simplement, saisissez cette comamnde pour lancer vcv et accéder à VCV :

```bash
docker run -d \
  -v "$(pwd)/settings.json:/app/settings.json:rw" \
  -v "$(pwd)/logs:/var/log/app:rw" \
  --cap-drop=ALL --read-only --security-opt no-new-privileges:true \
  -p 52000:52000 jhmmt/vcv:1.6.1
```

### Mise en production avec un fichier docker-compose

Créez ce fichier `docker-compose.yml` et saisissez ces informations :

```yml
---
services:
  vcv:
    image: jhmmt/vcv:1.6.1
    container_name: vcv
    restart: unless-stopped
    ports:
      - "52000:52000/tcp"
    cap_drop:
      - ALL
    read_only: true
    security_opt:
      - no-new-privileges:true
    volumes:
      - ./settings.json:/app/settings.json:rw
    deploy:
      resources:
        limits:
          cpus: "0.50"
          memory: 64M
```

## Déploiement dans Kubernetes

Puisque VaultCertsViewer est une image mono-binaire, il est tout à fait possible de déployer l'application dans un cluster Kubernetes. Voici le manifest complet à utiliser :

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: vcv
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: vcv-sa
  namespace: vcv
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vcv
  namespace: vcv
  labels:
    app: vcv
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vcv
  template:
    metadata:
      labels:
        app: vcv
    spec:
      serviceAccountName: vcv-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
        - name: vcv
          image: jhmmt/vcv:1.6.1
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 52000
              protocol: TCP
          volumeMounts:
            - name: vcv-settings
              mountPath: /app/settings.json
              subPath: settings.json
              readOnly: true
          resources:
            requests:
              cpu: "100m"
              memory: "64Mi"
            limits:
              cpu: "500m"
              memory: "128Mi"
          readinessProbe:
            httpGet:
              path: /api/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /api/health
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
          securityContext:
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
      volumes:
        - name: vcv-settings
          secret:
            secretName: vcv-settings
---
apiVersion: v1
kind: Secret
metadata:
  name: vcv-settings
  namespace: vcv
type: Opaque
stringData:
  settings.json: |
    {
      "admin": {
        "password": "$2y$10$.changeme"
    },
      "app": {
        "env": "prod",
        "logging": {
        "level": "debug",
        "format": "json",
        "output": "both",
        "file_path": "/var/log/app/vcv.log"
        },
        "port": 52000
    },
      "certificates": {
        "expiration_thresholds": {
        "critical": 2,
        "warning": 10
        }
    },
      "metrics": {
        "per_certificate": true,
        "enhanced_metrics": false
    },
      "cors": {
        "allowed_origins": [
            "https://172.16.20.50:8443",
            "https://reproxy.vcv.local"
        ],
        "allow_credentials": true
    },
      "vaults": [
        {
        "id": "vault-main",
        "address": "http://vault:8200",
        "token": "root",
        "pki_mounts": [
            "pki",
            "pki_dev",
            "pki_stage",
            "pki_production"
        ],
        "display_name": "Vault",
        "tls_ca_cert_base64": "BASE64_PEM_CA_BUNDLE",
        "tls_ca_cert": "",
        "tls_ca_path": "",
        "tls_server_name": "vault.service.consul",
        "tls_insecure": true,
        "enabled": true
        },
        {
        "id": "vault-dev",
        "address": "http://vault-dev:8200",
        "token": "root",
        "pki_mounts": [
            "pki",
            "pki_corporate",
            "pki_external",
            "pki_partners"
        ],
        "display_name": "Vault dev",
        "tls_ca_cert_base64": "BASE64_PEM_CA_BUNDLE",
        "tls_ca_cert": "",
        "tls_ca_path": "",
        "tls_server_name": "vault-dev.service.consul",
        "tls_insecure": true,
        "enabled": true
        }
      ]
    }
---
apiVersion: v1
kind: Service
metadata:
  name: vcv
  namespace: vcv
  labels:
    app: vcv
spec:
  selector:
    app: vcv
  ports:
    - name: http
      port: 52000
      targetPort: http
      protocol: TCP
  type: ClusterIP
```

Suite à cette installation, vous devrez utiliser votre Gateway et créer la HTTPRoute nécessaire pour accéder à l'app depuis l'extérieur du cluster Kubernetes.

## Conclusion

L’expiration de certificats est un risque prévisible — et pourtant encore très fréquent.

VCV facilite le suivi, la priorisation et l’action proactive grâce à une vue simple des certificats sur plusieurs moteurs PKI Vault, avec des seuils configurables et des métriques prêtes pour la supervision.
