package i18n

import (
	"net/http"
	"net/url"
	"strings"
	"vcv/internal/logger"
)

// Language represents a supported UI language.
type Language string

const (
	LanguageEnglish Language = "en"
	LanguageFrench  Language = "fr"
	LanguageGerman  Language = "de"
	LanguageItalian Language = "it"
	LanguageSpanish Language = "es"
)

// Messages contains all translatable UI strings used by the web interface.
type Messages struct {
	AppTitle                    string `json:"appTitle"`
	AppSubtitle                 string `json:"appSubtitle"`
	ButtonToggleTheme           string `json:"buttonToggleTheme"`
	ButtonClose                 string `json:"buttonClose"`
	ButtonDetails               string `json:"buttonDetails"`
	ButtonDownloadPEM           string `json:"buttonDownloadPEM"`
	ButtonExport                string `json:"buttonExport"`
	ExportCSV                   string `json:"exportCSV"`
	ExportJSON                  string `json:"exportJSON"`
	ExportEmpty                 string `json:"exportEmpty"`
	ExportSuccess               string `json:"exportSuccess"`
	CommandPaletteTitle         string `json:"commandPaletteTitle"`
	CommandPaletteHint          string `json:"commandPaletteHint"`
	CommandPalettePlaceholder   string `json:"commandPalettePlaceholder"`
	CommandPaletteEmpty         string `json:"commandPaletteEmpty"`
	CommandPaletteCertsGroup    string `json:"commandPaletteCertsGroup"`
	CommandPaletteFiltersGroup  string `json:"commandPaletteFiltersGroup"`
	CommandPaletteActionsGroup  string `json:"commandPaletteActionsGroup"`
	CommandPaletteOpenAdmin     string `json:"commandPaletteOpenAdmin"`
	ButtonRefresh               string `json:"buttonRefresh"`
	ButtonViewCA                string `json:"buttonViewCA"`
	CacheInvalidateFailed       string `json:"cacheInvalidateFailed"`
	CacheInvalidated            string `json:"cacheInvalidated"`
	CertificateInformationTitle string `json:"certificateInformationTitle"`
	ColumnCommonName            string `json:"columnCommonName"`
	ColumnCreatedAt             string `json:"columnCreatedAt"`
	ColumnExpiresAt             string `json:"columnExpiresAt"`
	ColumnSAN                   string `json:"columnSan"`
	ColumnStatus                string `json:"columnStatus"`
	DashboardCertsLabel         string `json:"dashboardCertsLabel"`
	DashboardOverviewLabel      string `json:"dashboardOverviewLabel"`
	DaysRemaining               string `json:"daysRemaining"`
	DaysRemainingShort          string `json:"daysRemainingShort"`
	DaysRemainingSingular       string `json:"daysRemainingSingular"`
	ExpiredDays                 string `json:"expiredDays"`
	ExpiredDaysSingular         string `json:"expiredDaysSingular"`
	ExpiringToday               string `json:"expiringToday"`
	DeselectAll                 string `json:"deselectAll"`
	DownloadPEMSuccess          string `json:"downloadPEMSuccess"`
	AdminDocsTitle              string `json:"adminDocsTitle"`
	AdminDocsError              string `json:"adminDocsError"`
	CertTypeFilterAll           string `json:"certTypeFilterAll"`
	CertTypeFilterBoth          string `json:"certTypeFilterBoth"`
	CertTypeFilterMachine       string `json:"certTypeFilterMachine"`
	CertTypeFilterUnknown       string `json:"certTypeFilterUnknown"`
	CertTypeFilterUser          string `json:"certTypeFilterUser"`
	LabelCertificateType        string `json:"labelCertificateType"`
	FooterVaultSummary          string `json:"footerVaultSummary"`
	LabelFingerprintSHA1        string `json:"labelFingerprintSHA1"`
	LabelFingerprintSHA256      string `json:"labelFingerprintSHA256"`
	LabelIssuer                 string `json:"labelIssuer"`
	LabelKeyAlgorithm           string `json:"labelKeyAlgorithm"`
	LabelLanguage               string `json:"labelLanguage"`
	LabelLoading                string `json:"labelLoading"`
	LabelPEM                    string `json:"labelPem"`
	LabelRootCA                 string `json:"labelRootCA"`
	LabelIntermediateCA         string `json:"labelIntermediateCA"`
	LabelSerialNumber           string `json:"labelSerialNumber"`
	LabelSubject                string `json:"labelSubject"`
	LabelUsage                  string `json:"labelUsage"`
	LabelVault                  string `json:"labelVault"`
	LabelPKI                    string `json:"labelPki"`
	LoadDetailsNetworkError     string `json:"loadDetailsNetworkError"`
	LoadNetworkError            string `json:"loadNetworkError"`
	LoadSuccess                 string `json:"loadSuccess"`
	ModalVaultStatusTitle       string `json:"modalVaultStatusTitle"`
	MountSearchPlaceholder      string `json:"mountSearchPlaceholder"`
	MountSelectorTitle          string `json:"mountSelectorTitle"`
	MountSelectorTooltip        string `json:"mountSelectorTooltip"`
	MountStatsSelected          string `json:"mountStatsSelected"`
	MountStatsVaults            string `json:"mountStatsVaults"`
	NotificationCritical        string `json:"notificationCritical"`
	NotificationWarning         string `json:"notificationWarning"`
	PaginationInfo              string `json:"paginationInfo"`
	PaginationNext              string `json:"paginationNext"`
	PaginationPageSizeLabel     string `json:"paginationPageSizeLabel"`
	PaginationPrev              string `json:"paginationPrev"`
	SearchPlaceholder           string `json:"searchPlaceholder"`
	SearchShortcutHint          string `json:"searchShortcutHint"`
	SelectAll                   string `json:"selectAll"`
	FilterChipSearch            string `json:"filterChipSearch"`
	FilterChipStatus            string `json:"filterChipStatus"`
	FilterChipCertType          string `json:"filterChipCertType"`
	FilterChipSources           string `json:"filterChipSources"`
	FilterChipReset             string `json:"filterChipReset"`
	SourcesButtonAll            string `json:"sourcesButtonAll"`
	SourcesButtonPartial        string `json:"sourcesButtonPartial"`
	StatusLabelExpired          string `json:"statusLabelExpired"`
	StatusLabelRevoked          string `json:"statusLabelRevoked"`
	StatusLabelValid            string `json:"statusLabelValid"`
	VaultConnectionLost         string `json:"vaultConnectionLost"`
	VaultConnectionRestored     string `json:"vaultConnectionRestored"`
	AdminTitle                  string `json:"adminTitle"`
	AdminBackToVCV              string `json:"adminBackToVCV"`
	AdminSettingsSaved          string `json:"adminSettingsSaved"`
	AdminLogout                 string `json:"adminLogout"`
	AdminLogin                  string `json:"adminLogin"`
	AdminPassword               string `json:"adminPassword"`
	AdminCriticalThreshold      string `json:"adminCriticalThreshold"`
	AdminWarningThreshold       string `json:"adminWarningThreshold"`
	AdminCORSOrigins            string `json:"adminCORSOrigins"`
	AdminVaults                 string `json:"adminVaults"`
	AdminVaultsHint             string `json:"adminVaultsHint"`
	AdminAddVault               string `json:"adminAddVault"`
	AdminSaveSettings           string `json:"adminSaveSettings"`
	AdminRestartNote            string `json:"adminRestartNote"`
	AdminVaultID                string `json:"adminVaultID"`
	AdminVaultDisplayName       string `json:"adminVaultDisplayName"`
	AdminVaultAddress           string `json:"adminVaultAddress"`
	AdminVaultPKIMounts         string `json:"adminVaultPKIMounts"`
	AdminVaultToken             string `json:"adminVaultToken"`
	AdminVaultTokenReveal       string `json:"adminVaultTokenReveal"`
	AdminVaultTokenHide         string `json:"adminVaultTokenHide"`
	AdminVaultTLSCABase64       string `json:"adminVaultTLSCABase64"`
	AdminVaultTLSCAFile         string `json:"adminVaultTLSCAFile"`
	AdminVaultTLSCAPath         string `json:"adminVaultTLSCAPath"`
	AdminVaultTLSServerName     string `json:"adminVaultTLSServerName"`
	AdminVaultTLSInsecure       string `json:"adminVaultTLSInsecure"`
	AdminVaultEnabled           string `json:"adminVaultEnabled"`
	AdminVaultConnected         string `json:"adminVaultConnected"`
	AdminVaultDisconnected      string `json:"adminVaultDisconnected"`
	AdminVaultDisabled          string `json:"adminVaultDisabled"`
	AdminVaultRemove            string `json:"adminVaultRemove"`
	AdminVaultTLSTip            string `json:"adminVaultTLSTip"`
	AdminLoginHint              string `json:"adminLoginHint"`
	AdminVaultsEmpty            string `json:"adminVaultsEmpty"`
	AdminCORSOriginsHint        string `json:"adminCORSOriginsHint"`
	AdminVaultTokenHint         string `json:"adminVaultTokenHint"`
	AdminThresholdsHint         string `json:"adminThresholdsHint"`
	AdminMetrics                string `json:"adminMetrics"`
	AdminMetricsHint            string `json:"adminMetricsHint"`
	AdminMetricsPerCertificate  string `json:"adminMetricsPerCertificate"`
	AdminMetricsEnhanced        string `json:"adminMetricsEnhanced"`

	// Added for the Svelte UI: status labels/descriptions, pagination,
	// banners, copy actions and admin strings the rewrite previously hardcoded.
	StatusLabelCritical      string `json:"statusLabelCritical"`
	StatusLabelWarning       string `json:"statusLabelWarning"`
	StatusDescValid          string `json:"statusDescValid"`
	StatusDescWarning        string `json:"statusDescWarning"`
	StatusDescCritical       string `json:"statusDescCritical"`
	StatusDescExpired        string `json:"statusDescExpired"`
	StatusDescRevoked        string `json:"statusDescRevoked"`
	LabelValidity            string `json:"labelValidity"`
	LabelCopy                string `json:"labelCopy"`
	LabelCopied              string `json:"labelCopied"`
	LabelCopyPEM             string `json:"labelCopyPem"`
	ButtonDone               string `json:"buttonDone"`
	MountNoMatch             string `json:"mountNoMatch"`
	CAIssuerCertificate      string `json:"caIssuerCertificate"`
	AdminUsername            string `json:"adminUsername"`
	AdminSigningIn           string `json:"adminSigningIn"`
	AdminInvalidateCache     string `json:"adminInvalidateCache"`
	AdminThresholdsTitle     string `json:"adminThresholdsTitle"`
	AdminSaving              string `json:"adminSaving"`
	AdminVaultUnknown        string `json:"adminVaultUnknown"`
	NavAdmin                 string `json:"navAdmin"`
	ToastRefreshing          string `json:"toastRefreshing"`
	ToastRefreshFailed       string `json:"toastRefreshFailed"`
	SkipToContent            string `json:"skipToContent"`
	VaultsUnreachable        string `json:"vaultsUnreachable"`
	VaultsUnreachableHint    string `json:"vaultsUnreachableHint"`
	TableNoMatch             string `json:"tableNoMatch"`
	TableEmpty               string `json:"tableEmpty"`
	TableEmptyHint           string `json:"tableEmptyHint"`
	FooterMoreInfo           string `json:"footerMoreInfo"`
	FooterLicense            string `json:"footerLicense"`
	StatusConnecting         string `json:"statusConnecting"`
	StatusNoVaults           string `json:"statusNoVaults"`
	StatusNoVaultsConfigured string `json:"statusNoVaultsConfigured"`
	PaginationRange          string `json:"paginationRange"`
	PaginationResults        string `json:"paginationResults"`
	PaginationPageSizeAll    string `json:"paginationPageSizeAll"`
}

// Response is the payload returned by the /api/i18n endpoint.
type Response struct {
	Language Language `json:"language"`
	Messages Messages `json:"messages"`
}

var englishMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	AppSubtitle:                 "Monitor Vault/OpenBao certificate expirations",
	SearchShortcutHint:          "Press / to focus search",
	ButtonToggleTheme:           "Toggle theme",
	ButtonClose:                 "Close",
	ButtonDetails:               "Details",
	ButtonDownloadPEM:           "Download PEM",
	ButtonExport:                "Export",
	ExportCSV:                   "Export CSV",
	ExportJSON:                  "Export JSON",
	ExportEmpty:                 "Nothing to export",
	ExportSuccess:               "Exported {count} certificate(s)",
	CommandPaletteTitle:         "Command palette",
	CommandPaletteHint:          "Jump to a certificate or run a command",
	CommandPalettePlaceholder:   "Search certificates or commands…",
	CommandPaletteEmpty:         "No results found.",
	CommandPaletteCertsGroup:    "Certificates",
	CommandPaletteFiltersGroup:  "Filter by status",
	CommandPaletteActionsGroup:  "Actions",
	CommandPaletteOpenAdmin:     "Open admin panel",
	ButtonRefresh:               "Refresh",
	ButtonViewCA:                "View intermediate CA",
	CacheInvalidateFailed:       "Failed to clear cache",
	CacheInvalidated:            "Cache cleared and data refreshed",
	CertificateInformationTitle: "Certificate information",
	ColumnCommonName:            "Common name",
	ColumnCreatedAt:             "Created at",
	ColumnExpiresAt:             "Expires at",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Status",
	DashboardCertsLabel:         "certs",
	DashboardOverviewLabel:      "Certificate status overview",
	DaysRemaining:               "{{days}} days remaining",
	DaysRemainingShort:          "{{days}}d",
	DaysRemainingSingular:       "{{days}} day remaining",
	ExpiredDays:                 "Expired {{days}} days ago",
	ExpiredDaysSingular:         "Expired {{days}} day ago",
	ExpiringToday:               "Expires today",
	DeselectAll:                 "Deselect all",
	DownloadPEMSuccess:          "Certificate PEM downloaded successfully",
	AdminDocsTitle:              "Admin documentation",
	AdminDocsError:              "Failed to load documentation",
	CertTypeFilterAll:           "All types",
	CertTypeFilterBoth:          "Machine + user",
	CertTypeFilterMachine:       "Machine",
	CertTypeFilterUnknown:       "Unknown type",
	CertTypeFilterUser:          "User",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} up",
	LabelFingerprintSHA1:        "SHA-1 Fingerprint",
	LabelFingerprintSHA256:      "SHA-256 Fingerprint",
	LabelIssuer:                 "Issuer",
	LabelKeyAlgorithm:           "Key Algorithm",
	LabelLanguage:               "Language",
	LabelLoading:                "Loading...",
	LabelCertificateType:        "Certificate type",
	LabelPEM:                    "PEM Certificate",
	LabelRootCA:                 "Root CA",
	LabelIntermediateCA:         "Intermediate CA",
	LabelSerialNumber:           "Serial Number",
	LabelSubject:                "Subject",
	LabelUsage:                  "Usage",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsNetworkError:     "Network error loading certificate details. Please try again.",
	LoadNetworkError:            "Network error loading certificates. Please try again.",
	LoadSuccess:                 "Certificates loaded successfully",
	ModalVaultStatusTitle:       "Vault status",
	MountSearchPlaceholder:      "Search vaults or PKI engines...",
	MountSelectorTitle:          "Certificate sources",
	MountSelectorTooltip:        "Select which certificate sources to display",
	MountStatsSelected:          "Selected",
	MountStatsVaults:            "Vaults",
	NotificationCritical:        "{{count}} certificate(s) expiring within {{threshold}} days or less!",
	NotificationWarning:         "{{count}} certificate(s) expiring within {{threshold}} days or less",
	PaginationInfo:              "Page {{current}} of {{total}}",
	PaginationNext:              "Next",
	PaginationPageSizeLabel:     "Results per page",
	PaginationPrev:              "Previous",
	SearchPlaceholder:           "Search by Serial Number, Common Name (CN) or SAN",
	SelectAll:                   "Select all",
	FilterChipSearch:            "Search",
	FilterChipStatus:            "Status",
	FilterChipCertType:          "Type",
	FilterChipSources:           "Sources",
	FilterChipReset:             "Reset filters",
	SourcesButtonAll:            "Sources: {total}/{total}",
	SourcesButtonPartial:        "Sources: {selected}/{total}",
	StatusLabelExpired:          "Expired",
	StatusLabelRevoked:          "Revoked",
	StatusLabelValid:            "Valid",
	VaultConnectionLost:         "Vault connection lost",
	VaultConnectionRestored:     "Vault connection restored",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Back to VCV",
	AdminSettingsSaved:          "Settings saved",
	AdminLogout:                 "Logout",
	AdminLogin:                  "Login",
	AdminPassword:               "Password",
	AdminCriticalThreshold:      "Critical threshold (days)",
	AdminWarningThreshold:       "Warning threshold (days)",
	AdminCORSOrigins:            "Allowed origins (comma-separated)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Manage configured Vault instances.",
	AdminAddVault:               "Add vault",
	AdminSaveSettings:           "Save settings.json",
	AdminRestartNote:            "Changes are persisted to the settings file. A server restart may be required for all changes to take effect.",
	AdminVaultID:                "ID",
	AdminVaultDisplayName:       "Display name",
	AdminVaultAddress:           "Address",
	AdminVaultPKIMounts:         "PKI mounts (comma-separated)",
	AdminVaultToken:             "Token",
	AdminVaultTokenReveal:       "Reveal",
	AdminVaultTokenHide:         "Hide",
	AdminVaultTLSCABase64:       "TLS CA cert (base64)",
	AdminVaultTLSCAFile:         "TLS CA cert (file path)",
	AdminVaultTLSCAPath:         "TLS CA path (directory)",
	AdminVaultTLSServerName:     "TLS server name (SNI)",
	AdminVaultTLSInsecure:       "TLS insecure",
	AdminVaultEnabled:           "Enabled",
	AdminVaultConnected:         "Connected",
	AdminVaultDisconnected:      "Disconnected",
	AdminVaultDisabled:          "Disabled",
	AdminVaultRemove:            "Remove",
	AdminVaultTLSTip:            "TLS tip: Provide the CA bundle either inline as base64 (preferred) or via a PEM file path / CA directory. If \"TLS CA cert (base64)\" is set, it takes precedence and the file/path fields are ignored. Encode a PEM bundle with: cat /path/to/ca.pem | base64 | tr -d '\\n'. \"TLS server name\" overrides SNI. \"TLS insecure\" disables verification (development only).",
	AdminLoginHint:              "Use the bcrypt-hashed password configured in settings.json.",
	AdminVaultsEmpty:            "No Vault instances configured yet. Click \"Add vault\" to get started.",
	AdminCORSOriginsHint:        "e.g. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Vault access token. This value is stored in the settings file.",
	AdminThresholdsHint:         "Certificates expiring within these thresholds are flagged in the dashboard.",
	AdminMetrics:                "Metrics",
	AdminMetricsHint:            "Configure Prometheus metrics collection behavior.",
	AdminMetricsPerCertificate:  "Per-certificate metrics (⚠️ high cardinality)",
	AdminMetricsEnhanced:        "Enhanced metrics (categorize certificats by expiration time ranges)",

	StatusLabelCritical:      "Critical",
	StatusLabelWarning:       "Warning",
	StatusDescValid:          "All good",
	StatusDescWarning:        "≤ {days} days",
	StatusDescCritical:       "≤ {days} days",
	StatusDescExpired:        "Past expiry",
	StatusDescRevoked:        "Revoked by CA",
	LabelValidity:            "Validity",
	LabelCopy:                "Copy",
	LabelCopied:              "Copied!",
	LabelCopyPEM:             "Copy PEM",
	ButtonDone:               "Done",
	MountNoMatch:             "No mount matches.",
	CAIssuerCertificate:      "Issuer certificate",
	AdminUsername:            "Username",
	AdminSigningIn:           "Signing in…",
	AdminInvalidateCache:     "Invalidate cache",
	AdminThresholdsTitle:     "Expiration thresholds (days)",
	AdminSaving:              "Saving…",
	AdminVaultUnknown:        "Unknown",
	NavAdmin:                 "Admin",
	ToastRefreshing:          "Refreshing…",
	ToastRefreshFailed:       "Refresh failed",
	SkipToContent:            "Skip to main content",
	VaultsUnreachable:        "{count} vault(s) unreachable",
	VaultsUnreachableHint:    "Showing partial results.",
	TableNoMatch:             "No certificates match the current filters.",
	TableEmpty:               "No certificates found.",
	TableEmptyHint:           "No PKI mount returned any certificates yet.",
	FooterMoreInfo:           "More info",
	FooterLicense:            "License",
	StatusConnecting:         "connecting…",
	StatusNoVaults:           "no vaults",
	StatusNoVaultsConfigured: "No vaults configured.",
	PaginationRange:          "{start}–{end} of {total}",
	PaginationResults:        "{count} results",
	PaginationPageSizeAll:    "All",
}

var frenchMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	AppSubtitle:                 "Surveillez les expirations de certificats Vault/OpenBao",
	SearchShortcutHint:          "Appuyez sur / pour cibler la recherche",
	ButtonToggleTheme:           "Changer de thème",
	ButtonClose:                 "Fermer",
	ButtonDetails:               "Détails",
	ButtonDownloadPEM:           "Télécharger PEM",
	ButtonExport:                "Exporter",
	ExportCSV:                   "Exporter CSV",
	ExportJSON:                  "Exporter JSON",
	ExportEmpty:                 "Rien à exporter",
	ExportSuccess:               "{count} certificat(s) exporté(s)",
	CommandPaletteTitle:         "Palette de commandes",
	CommandPaletteHint:          "Accéder à un certificat ou lancer une commande",
	CommandPalettePlaceholder:   "Rechercher des certificats ou des commandes…",
	CommandPaletteEmpty:         "Aucun résultat.",
	CommandPaletteCertsGroup:    "Certificats",
	CommandPaletteFiltersGroup:  "Filtrer par statut",
	CommandPaletteActionsGroup:  "Actions",
	CommandPaletteOpenAdmin:     "Ouvrir le panneau d'administration",
	ButtonRefresh:               "Rafraîchir",
	ButtonViewCA:                "Voir l'autorité intermédiaire",
	CacheInvalidateFailed:       "Échec du vidage du cache",
	CacheInvalidated:            "Cache vidé et données actualisées",
	CertificateInformationTitle: "Informations du certificat",
	ColumnCommonName:            "Nom commun",
	ColumnCreatedAt:             "Créé le",
	ColumnExpiresAt:             "Expire le",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Statut",
	DashboardCertsLabel:         "certs",
	DashboardOverviewLabel:      "Vue d'ensemble du statut des certificats",
	DaysRemaining:               "{{days}} jours restants",
	DaysRemainingShort:          "{{days}}j",
	DaysRemainingSingular:       "{{days}} jour restant",
	ExpiredDays:                 "Expiré il y a {{days}} jours",
	ExpiredDaysSingular:         "Expiré il y a {{days}} jour",
	ExpiringToday:               "Expire aujourd'hui",
	DeselectAll:                 "Tout désélectionner",
	DownloadPEMSuccess:          "Certificat PEM téléchargé avec succès",
	AdminDocsTitle:              "Documentation admin",
	AdminDocsError:              "Échec du chargement de la documentation",
	CertTypeFilterAll:           "Tous les types",
	CertTypeFilterBoth:          "Machine + utilisateur",
	CertTypeFilterMachine:       "Machine",
	CertTypeFilterUnknown:       "Type inconnu",
	CertTypeFilterUser:          "Utilisateur",
	FooterVaultSummary:          "Vaults : {{up}}/{{total}} OK",
	LabelFingerprintSHA1:        "Empreinte SHA-1",
	LabelFingerprintSHA256:      "Empreinte SHA-256",
	LabelIssuer:                 "Émetteur",
	LabelKeyAlgorithm:           "Algorithme de clé",
	LabelLanguage:               "Langue",
	LabelLoading:                "Chargement...",
	LabelCertificateType:        "Type de certificat",
	LabelPEM:                    "Certificat PEM",
	LabelRootCA:                 "Autorité racine",
	LabelIntermediateCA:         "Autorité intermédiaire",
	LabelSerialNumber:           "Numéro de série",
	LabelSubject:                "Sujet",
	LabelUsage:                  "Utilisation",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsNetworkError:     "Erreur réseau lors du chargement des détails du certificat. Veuillez réessayer.",
	LoadNetworkError:            "Erreur réseau lors du chargement des certificats. Veuillez réessayer.",
	LoadSuccess:                 "Certificats chargés avec succès",
	ModalVaultStatusTitle:       "Statut Vaults",
	MountSearchPlaceholder:      "Rechercher des vaults ou moteurs PKI...",
	MountSelectorTitle:          "Sources des certificats",
	MountSelectorTooltip:        "Choisir les sources de certificats à afficher",
	MountStatsSelected:          "Sélectionnés",
	MountStatsVaults:            "Vaults",
	NotificationCritical:        "{{count}} certificat(s) expirant dans {{threshold}} jours ou moins !",
	NotificationWarning:         "{{count}} certificat(s) expirant dans {{threshold}} jours ou moins",
	PaginationInfo:              "Page {{current}} sur {{total}}",
	PaginationNext:              "Suivant",
	PaginationPageSizeLabel:     "Résultats par page",
	PaginationPrev:              "Précédent",
	SearchPlaceholder:           "Rechercher par numéro de série, nom commun (CN) ou SAN",
	SelectAll:                   "Tout sélectionner",
	FilterChipSearch:            "Recherche",
	FilterChipStatus:            "Statut",
	FilterChipCertType:          "Type",
	FilterChipSources:           "Sources",
	FilterChipReset:             "Réinitialiser les filtres",
	SourcesButtonAll:            "Sources : {total}/{total}",
	SourcesButtonPartial:        "Sources : {selected}/{total}",
	StatusLabelExpired:          "Expiré",
	StatusLabelRevoked:          "Révoqué",
	StatusLabelValid:            "Valide",
	VaultConnectionLost:         "Connexion à Vault perdue",
	VaultConnectionRestored:     "Connexion à Vault rétablie",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Retour à VCV",
	AdminSettingsSaved:          "Paramètres enregistrés",
	AdminLogout:                 "Déconnexion",
	AdminLogin:                  "Connexion",
	AdminPassword:               "Mot de passe",
	AdminCriticalThreshold:      "Seuil critique (jours)",
	AdminWarningThreshold:       "Seuil d'avertissement (jours)",
	AdminCORSOrigins:            "Origines autorisées (séparées par des virgules)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Gérer les instances Vault configurées.",
	AdminAddVault:               "Ajouter un vault",
	AdminSaveSettings:           "Enregistrer settings.json",
	AdminRestartNote:            "Les modifications sont enregistrées dans le fichier de paramètres. Un redémarrage du serveur peut être nécessaire pour que tous les changements prennent effet.",
	AdminVaultID:                "ID",
	AdminVaultDisplayName:       "Nom d'affichage",
	AdminVaultAddress:           "Adresse",
	AdminVaultPKIMounts:         "Montages PKI (séparés par des virgules)",
	AdminVaultToken:             "Jeton",
	AdminVaultTokenReveal:       "Révéler",
	AdminVaultTokenHide:         "Masquer",
	AdminVaultTLSCABase64:       "Certificat CA TLS (base64)",
	AdminVaultTLSCAFile:         "Certificat CA TLS (chemin du fichier)",
	AdminVaultTLSCAPath:         "Chemin CA TLS (répertoire)",
	AdminVaultTLSServerName:     "Nom du serveur TLS (SNI)",
	AdminVaultTLSInsecure:       "TLS non sécurisé",
	AdminVaultEnabled:           "Activé",
	AdminVaultConnected:         "Connecté",
	AdminVaultDisconnected:      "Déconnecté",
	AdminVaultDisabled:          "Désactivé",
	AdminVaultRemove:            "Supprimer",
	AdminVaultTLSTip:            "Astuce TLS : Fournissez le bundle CA soit en ligne en base64 (préféré) soit via un chemin de fichier PEM / répertoire CA. Si \"Certificat CA TLS (base64)\" est défini, il a la priorité et les champs fichier/chemin sont ignorés. Encodez un bundle PEM avec : cat /chemin/vers/ca.pem | base64 | tr -d '\\n'. \"Nom du serveur TLS\" remplace SNI. \"TLS non sécurisé\" désactive la vérification (développement uniquement).",
	AdminLoginHint:              "Utilisez le mot de passe haché en bcrypt configuré dans settings.json.",
	AdminVaultsEmpty:            "Aucune instance Vault configurée. Cliquez sur « Ajouter un vault » pour commencer.",
	AdminCORSOriginsHint:        "ex. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Jeton d'accès Vault. Cette valeur est stockée dans le fichier de paramètres.",
	AdminThresholdsHint:         "Les certificats expirant dans ces seuils sont signalés dans le tableau de bord.",
	AdminMetrics:                "Métriques",
	AdminMetricsHint:            "Configurer le comportement de collecte des métriques Prometheus.",
	AdminMetricsPerCertificate:  "Métriques par certificat (⚠️ haute cardinalité)",
	AdminMetricsEnhanced:        "Métriques améliorées (catégoriser les certificats par intervalles de date d'expiration)",

	StatusLabelCritical:      "Critique",
	StatusLabelWarning:       "Avertissement",
	StatusDescValid:          "Tout va bien",
	StatusDescWarning:        "≤ {days} jours",
	StatusDescCritical:       "≤ {days} jours",
	StatusDescExpired:        "Date dépassée",
	StatusDescRevoked:        "Révoqué par l'autorité",
	LabelValidity:            "Validité",
	LabelCopy:                "Copier",
	LabelCopied:              "Copié !",
	LabelCopyPEM:             "Copier le PEM",
	ButtonDone:               "Terminé",
	MountNoMatch:             "Aucun montage correspondant.",
	CAIssuerCertificate:      "Certificat émetteur",
	AdminUsername:            "Identifiant",
	AdminSigningIn:           "Connexion…",
	AdminInvalidateCache:     "Vider le cache",
	AdminThresholdsTitle:     "Seuils d'expiration (jours)",
	AdminSaving:              "Enregistrement…",
	AdminVaultUnknown:        "Inconnu",
	NavAdmin:                 "Administration",
	ToastRefreshing:          "Actualisation…",
	ToastRefreshFailed:       "Échec de l'actualisation",
	SkipToContent:            "Aller au contenu principal",
	VaultsUnreachable:        "{count} vault(s) injoignable(s)",
	VaultsUnreachableHint:    "Résultats partiels affichés.",
	TableNoMatch:             "Aucun certificat ne correspond aux filtres actuels.",
	TableEmpty:               "Aucun certificat trouvé.",
	TableEmptyHint:           "Aucun montage PKI n'a encore renvoyé de certificat.",
	FooterMoreInfo:           "En savoir plus",
	FooterLicense:            "Licence",
	StatusConnecting:         "connexion…",
	StatusNoVaults:           "aucun vault",
	StatusNoVaultsConfigured: "Aucun vault configuré.",
	PaginationRange:          "{start}–{end} sur {total}",
	PaginationResults:        "{count} résultats",
	PaginationPageSizeAll:    "Tous",
}

var spanishMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	AppSubtitle:                 "Monitor Vault/OpenBao certificate expirations",
	SearchShortcutHint:          "Press / to focus search",
	ButtonToggleTheme:           "Cambiar tema",
	ButtonClose:                 "Cerrar",
	ButtonDetails:               "Detalles",
	ButtonDownloadPEM:           "Descargar PEM",
	ButtonExport:                "Exportar",
	ExportCSV:                   "Exportar CSV",
	ExportJSON:                  "Exportar JSON",
	ExportEmpty:                 "Nada que exportar",
	ExportSuccess:               "{count} certificado(s) exportado(s)",
	CommandPaletteTitle:         "Paleta de comandos",
	CommandPaletteHint:          "Ir a un certificado o ejecutar un comando",
	CommandPalettePlaceholder:   "Buscar certificados o comandos…",
	CommandPaletteEmpty:         "No se encontraron resultados.",
	CommandPaletteCertsGroup:    "Certificados",
	CommandPaletteFiltersGroup:  "Filtrar por estado",
	CommandPaletteActionsGroup:  "Acciones",
	CommandPaletteOpenAdmin:     "Abrir el panel de administración",
	ButtonRefresh:               "Actualizar",
	ButtonViewCA:                "Ver CA intermedia",
	CacheInvalidateFailed:       "Error al borrar el caché",
	CacheInvalidated:            "Caché borrado y datos actualizados",
	CertificateInformationTitle: "Información del certificado",
	ColumnCommonName:            "Nombre común",
	ColumnCreatedAt:             "Creado el",
	ColumnExpiresAt:             "Caduca el",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Estado",
	DashboardCertsLabel:         "certs",
	DashboardOverviewLabel:      "Resumen del estado de los certificados",
	DaysRemaining:               "{{days}} días restantes",
	DaysRemainingShort:          "{{days}}d",
	DaysRemainingSingular:       "{{days}} día restante",
	ExpiredDays:                 "Vencido hace {{days}} días",
	ExpiredDaysSingular:         "Vencido hace {{days}} día",
	ExpiringToday:               "Vence hoy",
	DeselectAll:                 "Deseleccionar todo",
	DownloadPEMSuccess:          "Certificado PEM descargado exitosamente",
	AdminDocsTitle:              "Documentación admin",
	AdminDocsError:              "Error al cargar la documentación",
	CertTypeFilterAll:           "Todos los tipos",
	CertTypeFilterBoth:          "Máquina + usuario",
	CertTypeFilterMachine:       "Máquina",
	CertTypeFilterUnknown:       "Tipo desconocido",
	CertTypeFilterUser:          "Usuario",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} OK",
	LabelFingerprintSHA1:        "Huella SHA-1",
	LabelFingerprintSHA256:      "Huella SHA-256",
	LabelIssuer:                 "Emisor",
	LabelKeyAlgorithm:           "Algoritmo de clave",
	LabelLanguage:               "Idioma",
	LabelLoading:                "Cargando...",
	LabelCertificateType:        "Tipo de certificado",
	LabelPEM:                    "Certificado PEM",
	LabelRootCA:                 "CA raíz",
	LabelIntermediateCA:         "CA intermedia",
	LabelSerialNumber:           "Número de serie",
	LabelSubject:                "Sujeto",
	LabelUsage:                  "Uso",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsNetworkError:     "Error de red al cargar los detalles del certificado. Por favor intente nuevamente.",
	LoadNetworkError:            "Error de red al cargar los certificados. Por favor intente nuevamente.",
	LoadSuccess:                 "Certificados cargados exitosamente",
	ModalVaultStatusTitle:       "Estado Vault",
	MountSearchPlaceholder:      "Buscar vaults o motores PKI...",
	MountSelectorTitle:          "Fuentes de certificados",
	MountSelectorTooltip:        "Seleccionar las fuentes de certificados a mostrar",
	MountStatsSelected:          "Seleccionados",
	MountStatsVaults:            "Vaults",
	NotificationCritical:        "{{count}} certificado(s) caducando en {{threshold}} días o menos!",
	NotificationWarning:         "{{count}} certificado(s) caducando en {{threshold}} días o menos",
	PaginationInfo:              "Página {{current}} de {{total}}",
	PaginationNext:              "Siguiente",
	PaginationPageSizeLabel:     "Resultados por página",
	PaginationPrev:              "Anterior",
	SearchPlaceholder:           "Buscar por Número de Serie, Nombre Común (CN) o SAN",
	SelectAll:                   "Seleccionar todo",
	FilterChipSearch:            "Search",
	FilterChipStatus:            "Status",
	FilterChipCertType:          "Type",
	FilterChipSources:           "Sources",
	FilterChipReset:             "Reset filters",
	SourcesButtonAll:            "Sources: {total}/{total}",
	SourcesButtonPartial:        "Sources: {selected}/{total}",
	StatusLabelExpired:          "Caducado",
	StatusLabelRevoked:          "Revocado",
	StatusLabelValid:            "Válido",
	VaultConnectionLost:         "Conexión a Vault perdida",
	VaultConnectionRestored:     "Conexión a Vault restablecida",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Volver a VCV",
	AdminSettingsSaved:          "Configuración guardada",
	AdminLogout:                 "Cerrar sesión",
	AdminLogin:                  "Iniciar sesión",
	AdminPassword:               "Contraseña",
	AdminCriticalThreshold:      "Umbral crítico (días)",
	AdminWarningThreshold:       "Umbral de advertencia (días)",
	AdminCORSOrigins:            "Orígenes permitidos (separados por comas)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Administrar instancias de Vault configuradas.",
	AdminAddVault:               "Agregar vault",
	AdminSaveSettings:           "Guardar settings.json",
	AdminRestartNote:            "Los cambios se guardan en el archivo de configuración. Es posible que se requiera reiniciar el servidor para que todos los cambios surtan efecto.",
	AdminVaultConnected:         "Conectado",
	AdminVaultDisconnected:      "Desconectado",
	AdminVaultDisabled:          "Deshabilitado",
	AdminVaultID:                "ID",
	AdminVaultDisplayName:       "Nombre para mostrar",
	AdminVaultAddress:           "Dirección",
	AdminVaultPKIMounts:         "Montajes PKI (separados por comas)",
	AdminVaultToken:             "Token",
	AdminVaultTokenReveal:       "Revelar",
	AdminVaultTokenHide:         "Ocultar",
	AdminVaultTLSCABase64:       "Certificado CA TLS (base64)",
	AdminVaultTLSCAFile:         "Certificado CA TLS (ruta del archivo)",
	AdminVaultTLSCAPath:         "Ruta CA TLS (directorio)",
	AdminVaultTLSServerName:     "Nombre del servidor TLS (SNI)",
	AdminVaultTLSInsecure:       "TLS inseguro",
	AdminVaultEnabled:           "Habilitado",
	AdminVaultRemove:            "Eliminar",
	AdminVaultTLSTip:            "Consejo TLS: Proporcione el paquete CA en línea como base64 (preferido) o mediante una ruta de archivo PEM / directorio CA. Si se establece \"Certificado CA TLS (base64)\", tiene prioridad y se ignoran los campos de archivo/ruta. Codifique un paquete PEM con: cat /ruta/a/ca.pem | base64 | tr -d '\\n'. \"Nombre del servidor TLS\" anula SNI. \"TLS inseguro\" deshabilita la verificación (solo desarrollo).",
	AdminLoginHint:              "Use la contraseña hasheada con bcrypt configurada en settings.json.",
	AdminVaultsEmpty:            "No hay instancias de Vault configuradas. Haga clic en \"Agregar vault\" para comenzar.",
	AdminCORSOriginsHint:        "ej. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Token de acceso de Vault. Este valor se almacena en el archivo de configuración.",
	AdminThresholdsHint:         "Los certificados que expiran dentro de estos umbrales se señalan en el panel.",
	AdminMetrics:                "Métricas",
	AdminMetricsHint:            "Configurar el comportamiento de recolección de métricas de Prometheus.",
	AdminMetricsPerCertificate:  "Métricas por certificado (⚠️ alta cardinalidad)",
	AdminMetricsEnhanced:        "Métricas mejoradas (categorizar certificados por intervalos de fecha de expiración)",

	StatusLabelCritical:      "Crítico",
	StatusLabelWarning:       "Advertencia",
	StatusDescValid:          "Todo correcto",
	StatusDescWarning:        "≤ {days} días",
	StatusDescCritical:       "≤ {days} días",
	StatusDescExpired:        "Fecha vencida",
	StatusDescRevoked:        "Revocado por la autoridad",
	LabelValidity:            "Validez",
	LabelCopy:                "Copiar",
	LabelCopied:              "¡Copiado!",
	LabelCopyPEM:             "Copiar PEM",
	ButtonDone:               "Hecho",
	MountNoMatch:             "Ningún montaje coincide.",
	CAIssuerCertificate:      "Certificado emisor",
	AdminUsername:            "Usuario",
	AdminSigningIn:           "Iniciando sesión…",
	AdminInvalidateCache:     "Vaciar caché",
	AdminThresholdsTitle:     "Umbrales de expiración (días)",
	AdminSaving:              "Guardando…",
	AdminVaultUnknown:        "Desconocido",
	NavAdmin:                 "Administración",
	ToastRefreshing:          "Actualizando…",
	ToastRefreshFailed:       "Error al actualizar",
	SkipToContent:            "Ir al contenido principal",
	VaultsUnreachable:        "{count} vault(s) inaccesible(s)",
	VaultsUnreachableHint:    "Mostrando resultados parciales.",
	TableNoMatch:             "Ningún certificado coincide con los filtros actuales.",
	TableEmpty:               "No se encontraron certificados.",
	TableEmptyHint:           "Ningún montaje PKI ha devuelto certificados todavía.",
	FooterMoreInfo:           "Más información",
	FooterLicense:            "Licencia",
	StatusConnecting:         "conectando…",
	StatusNoVaults:           "sin vaults",
	StatusNoVaultsConfigured: "Ningún vault configurado.",
	PaginationRange:          "{start}–{end} de {total}",
	PaginationResults:        "{count} resultados",
	PaginationPageSizeAll:    "Todos",
}

var germanMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	AppSubtitle:                 "Monitor Vault/OpenBao certificate expirations",
	SearchShortcutHint:          "Press / to focus search",
	ButtonToggleTheme:           "Design umschalten",
	ButtonClose:                 "Schließen",
	ButtonDetails:               "Details",
	ButtonDownloadPEM:           "PEM herunterladen",
	ButtonExport:                "Exportieren",
	ExportCSV:                   "CSV exportieren",
	ExportJSON:                  "JSON exportieren",
	ExportEmpty:                 "Nichts zu exportieren",
	ExportSuccess:               "{count} Zertifikat(e) exportiert",
	CommandPaletteTitle:         "Befehlspalette",
	CommandPaletteHint:          "Zu einem Zertifikat springen oder einen Befehl ausführen",
	CommandPalettePlaceholder:   "Zertifikate oder Befehle suchen…",
	CommandPaletteEmpty:         "Keine Ergebnisse gefunden.",
	CommandPaletteCertsGroup:    "Zertifikate",
	CommandPaletteFiltersGroup:  "Nach Status filtern",
	CommandPaletteActionsGroup:  "Aktionen",
	CommandPaletteOpenAdmin:     "Admin-Bereich öffnen",
	ButtonRefresh:               "Aktualisieren",
	ButtonViewCA:                "Zwischen-CA anzeigen",
	CacheInvalidateFailed:       "Cache konnte nicht geleert werden",
	CacheInvalidated:            "Cache geleert und Daten aktualisiert",
	CertificateInformationTitle: "Zertifikatsinformationen",
	ColumnCommonName:            "Allgemeiner Name",
	ColumnCreatedAt:             "Erstellt am",
	ColumnExpiresAt:             "Gültig bis",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Status",
	DashboardCertsLabel:         "Zert.",
	DashboardOverviewLabel:      "Übersicht über den Zertifikatsstatus",
	DaysRemaining:               "{{days}} verbleibende Tage",
	DaysRemainingShort:          "{{days}}T",
	DaysRemainingSingular:       "{{days}} verbleibender Tag",
	ExpiredDays:                 "Vor {{days}} Tagen abgelaufen",
	ExpiredDaysSingular:         "Vor {{days}} Tag abgelaufen",
	ExpiringToday:               "Läuft heute ab",
	DeselectAll:                 "Alle abwählen",
	DownloadPEMSuccess:          "Zertifikat-PEM erfolgreich heruntergeladen",
	AdminDocsTitle:              "Admin-dokumentation",
	AdminDocsError:              "Dokumentation konnte nicht geladen werden",
	CertTypeFilterAll:           "Alle Typen",
	CertTypeFilterBoth:          "Maschine + Benutzer",
	CertTypeFilterMachine:       "Maschine",
	CertTypeFilterUnknown:       "Unbekannter Typ",
	CertTypeFilterUser:          "Benutzer",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} OK",
	LabelFingerprintSHA1:        "SHA-1-Fingerabdruck",
	LabelFingerprintSHA256:      "SHA-256-Fingerabdruck",
	LabelIssuer:                 "Aussteller",
	LabelKeyAlgorithm:           "Schlüsselalgorithmus",
	LabelLanguage:               "Sprache",
	LabelLoading:                "Wird geladen...",
	LabelCertificateType:        "Zertifikatstyp",
	LabelPEM:                    "PEM-Zertifikat",
	LabelRootCA:                 "Stamm-CA",
	LabelIntermediateCA:         "Zwischen-CA",
	LabelSerialNumber:           "Seriennummer",
	LabelSubject:                "Betreff",
	LabelUsage:                  "Verwendung",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsNetworkError:     "Netzwerkfehler beim Laden der Zertifikatsdetails. Bitte versuchen Sie es erneut.",
	LoadNetworkError:            "Netzwerkfehler beim Laden der Zertifikate. Bitte versuchen Sie es erneut.",
	LoadSuccess:                 "Zertifikate erfolgreich geladen",
	ModalVaultStatusTitle:       "Vault-Status",
	MountSearchPlaceholder:      "Vaults oder PKI-Motoren suchen...",
	MountSelectorTitle:          "Zertifikatsquellen",
	MountSelectorTooltip:        "Zertifikatsquellen zur Anzeige auswählen",
	MountStatsSelected:          "Ausgewählt",
	MountStatsVaults:            "Vaults",
	NotificationCritical:        "{{count}} Zertifikat(e) laufen in {{threshold}} Tagen oder weniger ab!",
	NotificationWarning:         "{{count}} Zertifikat(e) laufen in {{threshold}} Tagen oder weniger ab",
	PaginationInfo:              "Seite {{current}} von {{total}}",
	PaginationNext:              "Weiter",
	PaginationPageSizeLabel:     "Ergebnisse pro Seite",
	PaginationPrev:              "Zurück",
	SearchPlaceholder:           "Suche nach Seriennummer, Common Name (CN) oder SAN",
	SelectAll:                   "Alle auswählen",
	FilterChipSearch:            "Search",
	FilterChipStatus:            "Status",
	FilterChipCertType:          "Type",
	FilterChipSources:           "Sources",
	FilterChipReset:             "Reset filters",
	SourcesButtonAll:            "Sources: {total}/{total}",
	SourcesButtonPartial:        "Sources: {selected}/{total}",
	StatusLabelExpired:          "Abgelaufen",
	StatusLabelRevoked:          "Widerrufen",
	StatusLabelValid:            "Gültig",
	VaultConnectionLost:         "Verbindung zu Vault unterbrochen",
	VaultConnectionRestored:     "Verbindung zu Vault wiederhergestellt",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Zurück zu VCV",
	AdminSettingsSaved:          "Einstellungen gespeichert",
	AdminLogout:                 "Abmelden",
	AdminLogin:                  "Anmelden",
	AdminPassword:               "Passwort",
	AdminCriticalThreshold:      "Kritischer Schwellenwert (Tage)",
	AdminWarningThreshold:       "Warnschwellenwert (Tage)",
	AdminCORSOrigins:            "Erlaubte Ursprünge (durch Kommas getrennt)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Konfigurierte Vault-Instanzen verwalten.",
	AdminAddVault:               "Vault hinzufügen",
	AdminSaveSettings:           "settings.json speichern",
	AdminRestartNote:            "Änderungen werden in der Einstellungsdatei gespeichert. Ein Neustart des Servers kann erforderlich sein, damit alle Änderungen wirksam werden.",
	AdminVaultConnected:         "Verbindung hergestellt",
	AdminVaultDisconnected:      "Verbindung unterbrochen",
	AdminVaultDisabled:          "Deaktiviert",
	AdminVaultID:                "ID",
	AdminVaultDisplayName:       "Anzeigename",
	AdminVaultAddress:           "Adresse",
	AdminVaultPKIMounts:         "PKI-Mounts (durch Kommas getrennt)",
	AdminVaultToken:             "Token",
	AdminVaultTokenReveal:       "Anzeigen",
	AdminVaultTokenHide:         "Verbergen",
	AdminVaultTLSCABase64:       "TLS-CA-Zertifikat (base64)",
	AdminVaultTLSCAFile:         "TLS-CA-Zertifikat (Dateipfad)",
	AdminVaultTLSCAPath:         "TLS-CA-Pfad (Verzeichnis)",
	AdminVaultTLSServerName:     "TLS-Servername (SNI)",
	AdminVaultTLSInsecure:       "TLS unsicher",
	AdminVaultEnabled:           "Aktiviert",
	AdminVaultRemove:            "Entfernen",
	AdminVaultTLSTip:            "TLS-Tipp: Geben Sie das CA-Bundle entweder inline als base64 (bevorzugt) oder über einen PEM-Dateipfad / CA-Verzeichnis an. Wenn \"TLS-CA-Zertifikat (base64)\" gesetzt ist, hat es Vorrang und die Datei-/Pfadfelder werden ignoriert. Kodieren Sie ein PEM-Bundle mit: cat /pfad/zu/ca.pem | base64 | tr -d '\\n'. \"TLS-Servername\" überschreibt SNI. \"TLS unsicher\" deaktiviert die Überprüfung (nur Entwicklung).",
	AdminLoginHint:              "Verwenden Sie das bcrypt-gehashte Passwort aus settings.json.",
	AdminVaultsEmpty:            "Noch keine Vault-Instanzen konfiguriert. Klicken Sie auf \"Vault hinzufügen\", um zu beginnen.",
	AdminCORSOriginsHint:        "z.B. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Vault-Zugriffstoken. Dieser Wert wird in der Einstellungsdatei gespeichert.",
	AdminThresholdsHint:         "Zertifikate, die innerhalb dieser Schwellenwerte ablaufen, werden im Dashboard markiert.",
	AdminMetrics:                "Metriken",
	AdminMetricsHint:            "Prometheus-Metrikensammelverhalten konfigurieren.",
	AdminMetricsPerCertificate:  "Metriken pro Zertifikat (⚠️ hohe Kardinalität)",
	AdminMetricsEnhanced:        "Erweiterte Metriken (Zertifikate nach Ablaufzeit kategorisieren)",

	StatusLabelCritical:      "Kritisch",
	StatusLabelWarning:       "Warnung",
	StatusDescValid:          "Alles in Ordnung",
	StatusDescWarning:        "≤ {days} Tage",
	StatusDescCritical:       "≤ {days} Tage",
	StatusDescExpired:        "Abgelaufen",
	StatusDescRevoked:        "Von der CA widerrufen",
	LabelValidity:            "Gültigkeit",
	LabelCopy:                "Kopieren",
	LabelCopied:              "Kopiert!",
	LabelCopyPEM:             "PEM kopieren",
	ButtonDone:               "Fertig",
	MountNoMatch:             "Kein Mount gefunden.",
	CAIssuerCertificate:      "Aussteller-Zertifikat",
	AdminUsername:            "Benutzername",
	AdminSigningIn:           "Anmeldung…",
	AdminInvalidateCache:     "Cache leeren",
	AdminThresholdsTitle:     "Ablaufschwellen (Tage)",
	AdminSaving:              "Speichern…",
	AdminVaultUnknown:        "Unbekannt",
	NavAdmin:                 "Verwaltung",
	ToastRefreshing:          "Aktualisierung…",
	ToastRefreshFailed:       "Aktualisierung fehlgeschlagen",
	SkipToContent:            "Zum Hauptinhalt springen",
	VaultsUnreachable:        "{count} Vault(s) nicht erreichbar",
	VaultsUnreachableHint:    "Teilergebnisse werden angezeigt.",
	TableEmpty:               "Keine Zertifikate gefunden.",
	TableEmptyHint:           "Noch hat kein PKI-Mount Zertifikate zurückgegeben.",
	TableNoMatch:             "Keine Zertifikate entsprechen den aktuellen Filtern.",
	FooterMoreInfo:           "Mehr Infos",
	FooterLicense:            "Lizenz",
	StatusConnecting:         "verbinde…",
	StatusNoVaults:           "keine Vaults",
	StatusNoVaultsConfigured: "Keine Vaults konfiguriert.",
	PaginationRange:          "{start}–{end} von {total}",
	PaginationResults:        "{count} Ergebnisse",
	PaginationPageSizeAll:    "Alle",
}

var italianMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	AppSubtitle:                 "Monitor Vault/OpenBao certificate expirations",
	SearchShortcutHint:          "Press / to focus search",
	ButtonToggleTheme:           "Cambia tema",
	ButtonClose:                 "Chiudi",
	ButtonDetails:               "Dettagli",
	ButtonDownloadPEM:           "Scarica PEM",
	ButtonExport:                "Esporta",
	ExportCSV:                   "Esporta CSV",
	ExportJSON:                  "Esporta JSON",
	ExportEmpty:                 "Niente da esportare",
	ExportSuccess:               "{count} certificato/i esportato/i",
	CommandPaletteTitle:         "Palette comandi",
	CommandPaletteHint:          "Vai a un certificato o esegui un comando",
	CommandPalettePlaceholder:   "Cerca certificati o comandi…",
	CommandPaletteEmpty:         "Nessun risultato.",
	CommandPaletteCertsGroup:    "Certificati",
	CommandPaletteFiltersGroup:  "Filtra per stato",
	CommandPaletteActionsGroup:  "Azioni",
	CommandPaletteOpenAdmin:     "Apri il pannello di amministrazione",
	ButtonRefresh:               "Aggiorna",
	ButtonViewCA:                "Visualizza CA intermedia",
	CacheInvalidateFailed:       "Impossibile cancellare la cache",
	CacheInvalidated:            "Cache cancellata e dati aggiornati",
	CertificateInformationTitle: "Informazioni sul certificato",
	ColumnCommonName:            "Nome comune",
	ColumnCreatedAt:             "Creato il",
	ColumnExpiresAt:             "Scade il",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Stato",
	DashboardCertsLabel:         "cert.",
	DashboardOverviewLabel:      "Panoramica dello stato dei certificati",
	DaysRemaining:               "{{days}} giorni rimanenti",
	DaysRemainingShort:          "{{days}}g",
	DaysRemainingSingular:       "{{days}} giorno rimanente",
	ExpiredDays:                 "Scaduto {{days}} giorni fa",
	ExpiredDaysSingular:         "Scaduto {{days}} giorno fa",
	ExpiringToday:               "Scade oggi",
	DeselectAll:                 "Deseleziona tutto",
	DownloadPEMSuccess:          "Certificato PEM scaricato con successo",
	AdminDocsTitle:              "Documentazione admin",
	AdminDocsError:              "Impossibile caricare la documentazione",
	CertTypeFilterAll:           "Tutti i tipi",
	CertTypeFilterBoth:          "Macchina + utente",
	CertTypeFilterMachine:       "Macchina",
	CertTypeFilterUnknown:       "Tipo sconosciuto",
	CertTypeFilterUser:          "Utente",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} OK",
	LabelFingerprintSHA1:        "Impronta SHA-1",
	LabelFingerprintSHA256:      "Impronta SHA-256",
	LabelIssuer:                 "Emittente",
	LabelKeyAlgorithm:           "Algoritmo della chiave",
	LabelLanguage:               "Lingua",
	LabelLoading:                "Caricamento...",
	LabelCertificateType:        "Tipo di certificato",
	LabelPEM:                    "Certificato PEM",
	LabelRootCA:                 "CA radice",
	LabelIntermediateCA:         "CA intermedia",
	LabelSerialNumber:           "Numero di serie",
	LabelSubject:                "Soggetto",
	LabelUsage:                  "Utilizzo",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsNetworkError:     "Errore di rete durante il caricamento dei dettagli del certificato. Riprova.",
	LoadNetworkError:            "Errore di rete durante il caricamento dei certificati. Riprova.",
	LoadSuccess:                 "Certificati caricati correttamente",
	ModalVaultStatusTitle:       "Stato Vault",
	MountSearchPlaceholder:      "Cerca vaults o motori PKI...",
	MountSelectorTitle:          "Fonti dei certificati",
	MountSelectorTooltip:        "Seleziona le fonti di certificati da visualizzare",
	MountStatsSelected:          "Selezionati",
	MountStatsVaults:            "Vaults",
	NotificationCritical:        "{{count}} certificato/i in scadenza entro {{threshold}} giorni o meno!",
	NotificationWarning:         "{{count}} certificato/i in scadenza entro {{threshold}} giorni o meno",
	PaginationInfo:              "Pagina {{current}} di {{total}}",
	PaginationNext:              "Successivo",
	PaginationPageSizeLabel:     "Risultati per pagina",
	PaginationPrev:              "Precedente",
	SearchPlaceholder:           "Cerca per Numero di Serie, Nome Comune (CN) o SAN",
	SelectAll:                   "Seleziona tutto",
	FilterChipSearch:            "Search",
	FilterChipStatus:            "Status",
	FilterChipCertType:          "Type",
	FilterChipSources:           "Sources",
	FilterChipReset:             "Reset filters",
	SourcesButtonAll:            "Sources: {total}/{total}",
	SourcesButtonPartial:        "Sources: {selected}/{total}",
	StatusLabelExpired:          "Scaduto",
	StatusLabelRevoked:          "Revocato",
	StatusLabelValid:            "Valido",
	VaultConnectionLost:         "Connessione al Vault interrotta",
	VaultConnectionRestored:     "Connessione al Vault ripristinata",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Torna a VCV",
	AdminSettingsSaved:          "Impostazioni salvate",
	AdminLogout:                 "Disconnetti",
	AdminLogin:                  "Accedi",
	AdminPassword:               "Password",
	AdminCriticalThreshold:      "Soglia critica (giorni)",
	AdminWarningThreshold:       "Soglia di avviso (giorni)",
	AdminCORSOrigins:            "Origini consentite (separate da virgole)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Gestisci le istanze Vault configurate.",
	AdminAddVault:               "Aggiungi vault",
	AdminSaveSettings:           "Salva settings.json",
	AdminRestartNote:            "Le modifiche vengono salvate nel file delle impostazioni. Potrebbe essere necessario riavviare il server affinché tutte le modifiche abbiano effetto.",
	AdminVaultConnected:         "Connesso",
	AdminVaultDisconnected:      "Disconnesso",
	AdminVaultDisabled:          "Disabilitato",
	AdminVaultID:                "ID",
	AdminVaultDisplayName:       "Nome visualizzato",
	AdminVaultAddress:           "Indirizzo",
	AdminVaultPKIMounts:         "Mount PKI (separati da virgole)",
	AdminVaultToken:             "Token",
	AdminVaultTokenReveal:       "Mostra",
	AdminVaultTokenHide:         "Nascondi",
	AdminVaultTLSCABase64:       "Certificato CA TLS (base64)",
	AdminVaultTLSCAFile:         "Certificato CA TLS (percorso file)",
	AdminVaultTLSCAPath:         "Percorso CA TLS (directory)",
	AdminVaultTLSServerName:     "Nome server TLS (SNI)",
	AdminVaultTLSInsecure:       "TLS non sicuro",
	AdminVaultEnabled:           "Abilitato",
	AdminVaultRemove:            "Rimuovi",
	AdminVaultTLSTip:            "Suggerimento TLS: Fornire il bundle CA in linea come base64 (preferito) o tramite un percorso file PEM / directory CA. Se \"Certificato CA TLS (base64)\" è impostato, ha la precedenza e i campi file/percorso vengono ignorati. Codificare un bundle PEM con: cat /percorso/a/ca.pem | base64 | tr -d '\\n'. \"Nome server TLS\" sovrascrive SNI. \"TLS non sicuro\" disabilita la verifica (solo sviluppo).",
	AdminLoginHint:              "Utilizzare la password hashata con bcrypt impostata in settings.json.",
	AdminVaultsEmpty:            "Nessuna istanza Vault configurata. Fai clic su \"Aggiungi vault\" per iniziare.",
	AdminCORSOriginsHint:        "es. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Token di accesso Vault. Questo valore viene memorizzato nel file delle impostazioni.",
	AdminThresholdsHint:         "I certificati in scadenza entro queste soglie vengono segnalati nella dashboard.",
	AdminMetrics:                "Metriche",
	AdminMetricsHint:            "Configura il comportamento di raccolta delle metriche di Prometheus.",
	AdminMetricsPerCertificate:  "Metriche per certificato (⚠️ alta cardinalità)",
	AdminMetricsEnhanced:        "Metriche avanzate (categorizzare certificati per intervallo di data di scadenza)",

	StatusLabelCritical:      "Critico",
	StatusLabelWarning:       "Avviso",
	StatusDescValid:          "Tutto a posto",
	StatusDescWarning:        "≤ {days} giorni",
	StatusDescCritical:       "≤ {days} giorni",
	StatusDescExpired:        "Data superata",
	StatusDescRevoked:        "Revocato dalla CA",
	LabelValidity:            "Validità",
	LabelCopy:                "Copia",
	LabelCopied:              "Copiato!",
	LabelCopyPEM:             "Copia PEM",
	ButtonDone:               "Fatto",
	MountNoMatch:             "Nessun mount corrispondente.",
	CAIssuerCertificate:      "Certificato emittente",
	AdminUsername:            "Nome utente",
	AdminSigningIn:           "Accesso…",
	AdminInvalidateCache:     "Svuota cache",
	AdminThresholdsTitle:     "Soglie di scadenza (giorni)",
	AdminSaving:              "Salvataggio…",
	AdminVaultUnknown:        "Sconosciuto",
	NavAdmin:                 "Amministrazione",
	ToastRefreshing:          "Aggiornamento…",
	ToastRefreshFailed:       "Aggiornamento non riuscito",
	SkipToContent:            "Vai al contenuto principale",
	VaultsUnreachable:        "{count} vault non raggiungibile/i",
	VaultsUnreachableHint:    "Risultati parziali mostrati.",
	TableNoMatch:             "Nessun certificato corrisponde ai filtri attuali.",
	TableEmpty:               "Nessun certificato trovato.",
	TableEmptyHint:           "Nessun mount PKI ha ancora restituito certificati.",
	FooterMoreInfo:           "Maggiori informazioni",
	FooterLicense:            "Licenza",
	StatusConnecting:         "connessione…",
	StatusNoVaults:           "nessun vault",
	StatusNoVaultsConfigured: "Nessun vault configurato.",
	PaginationRange:          "{start}–{end} di {total}",
	PaginationResults:        "{count} risultati",
	PaginationPageSizeAll:    "Tutti",
}

// MessagesForLanguage returns the translations for a given language code.
func MessagesForLanguage(language Language) Messages {
	if language == LanguageFrench {
		return frenchMessages
	}
	if language == LanguageSpanish {
		return spanishMessages
	}
	if language == LanguageGerman {
		return germanMessages
	}
	if language == LanguageItalian {
		return italianMessages
	}
	return englishMessages
}

// FromQueryLanguage parses a short language code coming from a query parameter.
func FromQueryLanguage(value string) (Language, bool) {
	return GetLanguage(value)
}

// Translations maps language codes to their message sets.
var Translations = map[string]Messages{
	string(LanguageEnglish): englishMessages,
	string(LanguageFrench):  frenchMessages,
	string(LanguageSpanish): spanishMessages,
	string(LanguageGerman):  germanMessages,
	string(LanguageItalian): italianMessages,
}

// GetLanguage returns the Language constant for a given code.
func GetLanguage(code string) (Language, bool) {
	code = strings.ToLower(strings.TrimSpace(code))
	switch code {
	case string(LanguageEnglish):
		return LanguageEnglish, true
	case string(LanguageFrench):
		return LanguageFrench, true
	case string(LanguageSpanish):
		return LanguageSpanish, true
	case string(LanguageGerman):
		return LanguageGerman, true
	case string(LanguageItalian):
		return LanguageItalian, true
	}
	return "", false
}

// FromAcceptLanguage parses the Accept-Language header and returns the best match.
func FromAcceptLanguage(headerValue string) Language {
	if headerValue == "" {
		return LanguageEnglish
	}

	// Split by comma: fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7
	parts := strings.Split(headerValue, ",")
	for _, part := range parts {
		// Extract the language code before any semicolon
		langCode := strings.Split(strings.TrimSpace(part), ";")[0]
		langCode = strings.ToLower(langCode)

		// Try exact match first
		if lang, ok := GetLanguage(langCode); ok {
			return lang
		}

		// Try primary language (e.g., "fr" from "fr-FR")
		if len(langCode) > 2 {
			primary := langCode[:2]
			if lang, ok := GetLanguage(primary); ok {
				return lang
			}
		}
	}

	return LanguageEnglish
}

// ResolveLanguage resolves the language from the request using query param, cookie, or Accept-Language header.
func ResolveLanguage(r *http.Request) Language {
	// 1. Check query parameter
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if l, ok := GetLanguage(lang); ok {
			logger.Get().Debug().Str("source", "query_param").Str("lang", lang).Msg("language resolved")
			return l
		}
	}

	// 2. Check HX-Current-URL header
	currentURL := r.Header.Get("hx-current-url")
	if currentURL != "" {
		parsed, err := url.Parse(currentURL)
		if err == nil {
			headerLanguage := parsed.Query().Get("lang")
			if headerLanguage != "" {
				if language, ok := GetLanguage(headerLanguage); ok {
					logger.Get().Debug().Str("source", "hx_current_url").Str("lang", headerLanguage).Msg("language resolved")
					return language
				}
			}
		}
	}

	// 3. Check cookie
	if cookie, err := r.Cookie("lang"); err == nil {
		if l, ok := GetLanguage(cookie.Value); ok {
			logger.Get().Debug().Str("source", "cookie").Str("lang", cookie.Value).Msg("language resolved")
			return l
		}
	}

	// 4. Use Accept-Language header
	acceptLang := r.Header.Get("Accept-Language")
	lang := FromAcceptLanguage(acceptLang)
	logger.Get().Debug().Str("source", "accept_language").Str("header", acceptLang).Str("lang", string(lang)).Msg("language resolved")
	return lang
}
