# Guide utilisateur - VaultCertsViewer (VCV)

## Qu'est-ce que VCV ?

VaultCertsViewer (VCV) est une interface web légère conçue pour visualiser et surveiller les certificats gérés par les moteurs PKI d'HashiCorp Vault (ou OpenBao). Il offre un tableau de bord centralisé pour suivre les dates d'expiration, les statuts (valide, expiré, révoqué) et les détails techniques de vos certificats sur plusieurs instances Vault et points de montage PKI.

## Fonctionnalités

- **Support multi-vault** : Connectez-vous à une ou plusieurs instances Vault simultanément.
- **Sélecteur de moteurs PKI** : Filtrez les certificats par instance Vault et point de montage PKI via une modale interactive avec recherche, sélection/désélection par vault ou globalement.
- **Tableau de bord** : Graphique en anneau avec statistiques en temps réel sur la répartition des statuts (valide, expirant, expiré, révoqué). Cliquez sur un segment ou une carte de statut pour filtrer le tableau instantanément.
- **Recherche et filtrage** : Recherche par Common Name (CN) ou Subject Alternative Names (SAN). Filtrage par statut via les cartes du tableau de bord.
- **Tri** : Triez le tableau par Common Name, date de création, date d'expiration, nom du Vault ou point de montage PKI. Cliquez sur un en-tête de colonne pour basculer entre ordre croissant/décroissant.
- **Pagination** : Pagination côté serveur avec tailles de page configurables (25, 50, 100 ou Tout).
- **Vue détaillée** : Accédez aux métadonnées complètes du certificat dans une modale organisée : identité (sujet, émetteur, numéro de série, SANs), dates de validité avec statut d'expiration, et détails techniques (algorithme de clé, utilisation de la clé, empreintes SHA-1/SHA-256).
- **Téléchargement PEM** : Téléchargez les fichiers PEM directement depuis le tableau.
- **Statut Vault** : Un indicateur dans l'en-tête (icône bouclier avec point de statut) affiche l'état de connexion en temps réel de vos instances Vault. Cliquez dessus pour ouvrir une modale détaillée avec l'état de santé par vault et un bouton de rafraîchissement.
- **Notifications d'expiration** : Une bannière en haut de la page avertit des certificats expirant dans les seuils configurés (critique / avertissement).
- **Notifications toast** : Messages toast en temps réel pour les changements de connexion Vault, les erreurs et les retours utilisateur.
- **Cache et rafraîchissement** : Les données des certificats sont mises en cache côté serveur (TTL de 15 min). Utilisez le bouton de rafraîchissement (↻) dans l'en-tête pour invalider le cache et récupérer des données fraîches.
- **Documentation intégrée** : Accédez à ce guide utilisateur et à la référence de configuration directement depuis l'interface via le bouton documentation (📖).
- **Synchronisation d'URL** : Les filtres, la recherche, l'ordre de tri, la pagination et la sélection des montages sont reflétés dans l'URL pour le partage et les favoris.
- **I18n** : Support complet de l'anglais, du français, de l'espagnol, de l'allemand et de l'italien. Changez de langue avec le menu déroulant dans l'en-tête.
- **Mode sombre** : Interface moderne avec bascule mode sombre/clair persistante.
- **Panneau d'administration** : Gérez le fichier `settings.json` visuellement (ajouter/supprimer des instances Vault, configurer les seuils, la journalisation, CORS). Nécessite la variable d'environnement `VCV_ADMIN_PASSWORD`.
- **Métriques Prometheus** : Exposez les métriques de certificats et de connexion sur `/metrics` pour la surveillance et les alertes.

## Utilisation de l'interface

### Tableau de bord

Le tableau de bord affiche un graphique en anneau et quatre cartes de statut (Valide, Expirant, Expiré, Révoqué). Cliquez sur une carte ou un segment du graphique pour filtrer le tableau des certificats par ce statut. Un bouton « Effacer le filtre » apparaît pour réinitialiser le filtre.

### Sélecteur de moteurs PKI

Cliquez sur le bouton « Moteurs PKI » dans la barre de filtres pour ouvrir la modale de sélection des montages. Les montages sont regroupés par instance Vault. Vous pouvez :

- Rechercher des montages par nom.
- Sélectionner/désélectionner des montages individuellement.
- Sélectionner/désélectionner tous les montages d'une instance Vault spécifique.
- Sélectionner/désélectionner tous les montages globalement.

Le tableau des certificats se met à jour automatiquement lorsque vous basculez des montages.

### Détails du certificat

Cliquez sur le bouton « Détails » sur n'importe quelle ligne pour ouvrir une modale avec les métadonnées complètes du certificat, organisées en trois sections : identité (sujet, émetteur, numéro de série, SANs), validité (dates de création/expiration avec compte à rebours), et détails techniques (algorithme de clé, utilisation de la clé, empreintes SHA-1/SHA-256).

### Statut Vault

L'icône bouclier dans l'en-tête indique l'état global de connexion Vault (vert = tous connectés, rouge = au moins un déconnecté). Cliquez dessus pour voir le statut par vault. Vous pouvez forcer une vérification de santé depuis la modale.

## Configuration

VCV est configuré principalement via un fichier `settings.json`. Le panneau d'administration permet de modifier ce fichier visuellement. Consultez la documentation de configuration pour tous les détails.

Tous les paramètres de l'application (instances Vault, seuils d'expiration, journalisation, CORS, etc.) sont définis dans `settings.json`. Seules deux variables d'environnement sont encore nécessaires :

- `VCV_ADMIN_PASSWORD` : Hash bcrypt pour activer le panneau d'administration (conservé en variable d'environnement pour des raisons de sécurité — il ne doit pas être stocké dans un fichier modifiable depuis l'interface).
- `SETTINGS_PATH` : Chemin vers un fichier `settings.json` personnalisé (nécessaire uniquement si le fichier n'est pas dans un emplacement par défaut).

> **Note :** Les variables d'environnement (`VAULT_ADDRS`, `LOG_LEVEL`, etc.) sont toujours supportées comme solution de repli lorsqu'aucun `settings.json` n'est trouvé, mais l'utilisation de `settings.json` est l'approche recommandée.

## Limites et ce qu'il ne fait pas

- **Lecture seule** : VCV est un outil de visualisation. Il ne permet **pas** de générer, renouveler ou révoquer des certificats.
- **Authentification** : VCV suppose que vous avez fourni des jetons valides pour les instances Vault auxquelles il se connecte.
- **Gestion de Vault** : Il ne gère pas les politiques Vault ni la configuration PKI ; il lit uniquement les données existantes.

## Support

Pour tout problème ou demande de fonctionnalité, veuillez vous référer au dépôt du projet.
