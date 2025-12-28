# Documentation utilisateur - VaultCertsViewer (VCV)

## Qu'est-ce que VCV ?

VaultCertsViewer (VCV) est une interface web légère conçue pour visualiser et surveiller les certificats gérés par les moteurs PKI d'HashiCorp Vault. Il offre un tableau de bord centralisé pour suivre les dates d'expiration, les statuts (valide, expiré, révoqué) et les détails techniques de vos certificats sur plusieurs instances Vault et points de montage PKI.

## Capacités

- **Support multi-vault** : Connectez-vous à une ou plusieurs instances Vault.
- **Découverte des moteurs PKI** : Découvre automatiquement les points de montage PKI auxquels vous avez accès.
- **Tableau de bord** : Statistiques en temps réel sur la répartition des statuts et la chronologie des expirations.
- **Recherche et filtrage** : Recherche par Common Name (CN) ou Subject Alternative Names (SAN). Filtrage par Vault, moteur PKI, statut ou seuil d'expiration.
- **Vue détaillée** : Accès aux métadonnées complètes du certificat, y compris l'émetteur, les empreintes numériques et le contenu PEM.
- **Exportation** : Téléchargement direct des fichiers PEM depuis l'interface.
- **I18n** : Support complet de l'anglais, du français, de l'espagnol, de l'allemand et de l'italien.
- **Mode sombre** : Interface moderne avec bascule mode sombre/clair.

## Configuration

VCV est configuré principalement via des variables d'environnement ou un fichier `settings.json`.

### Principales variables d'environnement

- `VAULT_ADDRS` : Liste des adresses Vault séparées par des virgules.
- `VCV_EXPIRE_WARNING` : Seuil en jours pour les notifications d'avertissement (défaut : 30).
- `VCV_EXPIRE_CRITICAL` : Seuil en jours pour les notifications critiques (défaut : 7).
- `LOG_LEVEL` : Niveau de détail des logs (info, debug, error).

## Limites et ce qu'il ne fait pas

- **Lecture seule** : VCV est actuellement un outil de visualisation. Il ne permet **pas** de générer, renouveler ou révoquer des certificats.
- **Authentification** : VCV assumes que vous avez fourni des jetons valides ou configuré l'authentification pour les instances Vault auxquelles il se connecte.
- **Gestion de Vault** : Il ne gère pas les politiques Vault ni la configuration PKI ; il lit uniquement les données existantes.

## Support

Pour tout problème ou demande de fonctionnalité, veuillez vous référer au dépôt du projet.
