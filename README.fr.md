# VaultCertsViewer

VaultCertsViewer (vcv) est une interface web légère qui permet de lister et de consulter les certificats stockés dans un coffre ‘pki’ d'HashiCorp Vault. Elle affiche notamment les noms communs, les SAN et surtout les dates d'expiration des certificats.

Actuellement, VaultCertsViewer (vcv) ne peut voir et afficher les certificats que d'un seul montage à la fois. Si vous avez (par exemple) 4 montages, il vous faudra déployer 4 instances de vcv.

## Quelles sont les fonctionnalités ?

- Découvre tous les certificats d’une PKI dans Vault et les affiche dans un tableau filtrable et recherchable.
- Affichage des noms communs (CN) et des SANs des certificats.
- Affiche la répartition des statuts (valide / expiré / révoqué) et les dates d'expirations à venir.
- Met en avant les certificats qui expirent bientôt (7/30 jours) et affiche les détails (CN, SAN, empreintes, émetteur, validité).
- Choix de la langue de l’UI (en, fr, es, de, it) et le thème (clair/sombre).

## Pourquoi cet outil existe-t-il ?

L'interface de Vault est trop lourde et complexe pour consulter les certificats. Elle ne permet pas **facilement** et rapidement de consulter les dates d'expiration et les détails des certificats.

VaultCertsViewer permet aux équipes plateforme / sécurité / ops une vue rapide et en **lecture seule** sur l’inventaire PKI Vault avec les seules informations nécessaires et utiles.

## À qui s'adresse-t-il ?

- Aux equipes exploitant l'outil Vault PKI qui ont besoin de visibilité sur leurs certificats.
- Aux opérateurs qui veulent une vue navigateur prête à l’emploi, à côté de la CLI ou de la Web UI de Vault.

## Comment le déployer et l'utiliser ?

Dans HashiCorp Vault, créez un rôle et un jeton en lecture seule pour l'API afin d'accéder aux certificats du moteur PKI ciblé (adaptez `pki` si vous utilisez un autre point de montage) :

```bash
vault policy write vcv - <<'EOF'
path "pki/certs"   { capabilities = ["list"] }
path "pki/cert/*"  { capabilities = ["read"] }
path "sys/health"  { capabilities = ["read"] }
EOF
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Ce jeton dédié limite les droits à la consultation des certificats, peut être renouvelé et sert de valeur `VAULT_READ_TOKEN` pour l'application.

### docker-compose

Récupérez le fichier `docker-compose.yml` et placez-le dans un répertoire de votre machine. Lancez ensuite la commande suivante :

```bash
docker compose up -d
```

Il n'y a pas besoin de stockage, sauf si vous souhaitez envoyer les journaux d'événements dans un fichier.

### docker run

Lancez rapidement le container avec cette commande:

```bash
docker run -d \
  -e "VAULT_ADDR=http://changeme:8200" \
  -e "VAULT_READ_TOKEN=changeme" \
  -e "VAULT_PKI_MOUNT=changeme" \
  -e "LOG_LEVEL=info" \
  -e "LOG_FORMAT=json" \
  -e "LOG_OUTPUT=stdout" \
  --cap-drop=ALL --read-only --security-opt no-new-privileges:true \
  -p 52000:52000 jhmmt/vcv:1.1
```

## Multilingue

L’UI est localisée en anglais, français, espagnol, allemand et italien. La langue se choisit dans l’en-tête ou via `?lang=xx`.

## Pour aller plus loin

- Documentation technique : [app/README.md](app/README.md)
- Version anglaise : [README.md](README.md)
