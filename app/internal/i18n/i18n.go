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
	ButtonToggleTheme           string `json:"buttonToggleTheme"`
	ButtonClose                 string `json:"buttonClose"`
	ButtonDetails               string `json:"buttonDetails"`
	ButtonDocumentation         string `json:"buttonDocumentation"`
	ButtonDownloadPEM           string `json:"buttonDownloadPEM"`
	ButtonRefresh               string `json:"buttonRefresh"`
	ButtonToggleFilters         string `json:"buttonToggleFilters"`
	CacheInvalidateFailed       string `json:"cacheInvalidateFailed"`
	CacheInvalidated            string `json:"cacheInvalidated"`
	CertificateInformationTitle string `json:"certificateInformationTitle"`
	ColumnActions               string `json:"columnActions"`
	ColumnCommonName            string `json:"columnCommonName"`
	ColumnCreatedAt             string `json:"columnCreatedAt"`
	ColumnExpiresAt             string `json:"columnExpiresAt"`
	ColumnSAN                   string `json:"columnSan"`
	ColumnStatus                string `json:"columnStatus"`
	DashboardCertsLabel         string `json:"dashboardCertsLabel"`
	DashboardClearFilter        string `json:"dashboardClearFilter"`
	DashboardCritical           string `json:"dashboardCritical"`
	DashboardCriticalDesc       string `json:"dashboardCriticalDesc"`
	DashboardExpired            string `json:"dashboardExpired"`
	DashboardExpiredDesc        string `json:"dashboardExpiredDesc"`
	DashboardWarning            string `json:"dashboardWarning"`
	DashboardWarningDesc        string `json:"dashboardWarningDesc"`
	DashboardFilterHint         string `json:"dashboardFilterHint"`
	DashboardRevoked            string `json:"dashboardRevoked"`
	DashboardRevokedDesc        string `json:"dashboardRevokedDesc"`
	DashboardTotal              string `json:"dashboardTotal"`
	DashboardValid              string `json:"dashboardValid"`
	DashboardValidDesc          string `json:"dashboardValidDesc"`
	DaysRemaining               string `json:"daysRemaining"`
	DaysRemainingShort          string `json:"daysRemainingShort"`
	DaysRemainingSingular       string `json:"daysRemainingSingular"`
	ExpiredSince                string `json:"expiredSince"`
	ExpiredToday                string `json:"expiredToday"`
	ExpiredDays                 string `json:"expiredDays"`
	ExpiredDaysSingular         string `json:"expiredDaysSingular"`
	ExpiringToday               string `json:"expiringToday"`
	DeselectAll                 string `json:"deselectAll"`
	DownloadPEMFailed           string `json:"downloadPEMFailed"`
	DownloadPEMNetworkError     string `json:"downloadPEMNetworkError"`
	DownloadPEMSuccess          string `json:"downloadPEMSuccess"`
	AdminDocsTitle              string `json:"adminDocsTitle"`
	ExpiryFilter30Days          string `json:"expiryFilter30Days"`
	ExpiryFilter7Days           string `json:"expiryFilter7Days"`
	ExpiryFilter90Days          string `json:"expiryFilter90Days"`
	ExpiryFilterAll             string `json:"expiryFilterAll"`
	FooterVaultConnected        string `json:"footerVaultConnected"`
	FooterVaultDisconnected     string `json:"footerVaultDisconnected"`
	FooterVaultLoading          string `json:"footerVaultLoading"`
	FooterVaultNotConfigured    string `json:"footerVaultNotConfigured"`
	FooterVaultSummary          string `json:"footerVaultSummary"`
	FooterVersion               string `json:"footerVersion"`
	LabelFingerprintSHA1        string `json:"labelFingerprintSHA1"`
	LabelFingerprintSHA256      string `json:"labelFingerprintSHA256"`
	LabelIssuer                 string `json:"labelIssuer"`
	LabelKeyAlgorithm           string `json:"labelKeyAlgorithm"`
	LabelLanguage               string `json:"labelLanguage"`
	LabelLoading                string `json:"labelLoading"`
	LabelPEM                    string `json:"labelPem"`
	LabelSerialNumber           string `json:"labelSerialNumber"`
	LabelSubject                string `json:"labelSubject"`
	LabelUsage                  string `json:"labelUsage"`
	LegendExpiredText           string `json:"legendExpiredText"`
	LegendExpiredTitle          string `json:"legendExpiredTitle"`
	LegendRevokedText           string `json:"legendRevokedText"`
	LegendRevokedTitle          string `json:"legendRevokedTitle"`
	LegendValidText             string `json:"legendValidText"`
	LegendValidTitle            string `json:"legendValidTitle"`
	LabelVault                  string `json:"labelVault"`
	LabelPKI                    string `json:"labelPki"`
	LoadDetailsFailed           string `json:"loadDetailsFailed"`
	LoadDetailsNetworkError     string `json:"loadDetailsNetworkError"`
	LoadFailed                  string `json:"loadFailed"`
	LoadNetworkError            string `json:"loadNetworkError"`
	LoadSuccess                 string `json:"loadSuccess"`
	LoadUnexpectedFormat        string `json:"loadUnexpectedFormat"`
	LoadingDetails              string `json:"loadingDetails"`
	ModalDetailsTitle           string `json:"modalDetailsTitle"`
	ModalVaultStatusTitle       string `json:"modalVaultStatusTitle"`
	MountSearchPlaceholder      string `json:"mountSearchPlaceholder"`
	MountSelectorTitle          string `json:"mountSelectorTitle"`
	MountSelectorTooltip        string `json:"mountSelectorTooltip"`
	MountStatsSelected          string `json:"mountStatsSelected"`
	MountStatsTotal             string `json:"mountStatsTotal"`
	NoData                      string `json:"noData"`
	NotificationCritical        string `json:"notificationCritical"`
	NotificationWarning         string `json:"notificationWarning"`
	PaginationAll               string `json:"paginationAll"`
	PaginationInfo              string `json:"paginationInfo"`
	PaginationNext              string `json:"paginationNext"`
	PaginationPageSizeLabel     string `json:"paginationPageSizeLabel"`
	PaginationPrev              string `json:"paginationPrev"`
	SearchPlaceholder           string `json:"searchPlaceholder"`
	SelectAll                   string `json:"selectAll"`
	StatusFilterAll             string `json:"statusFilterAll"`
	StatusFilterExpired         string `json:"statusFilterExpired"`
	StatusFilterExpiring        string `json:"statusFilterExpiring"`
	StatusFilterRevoked         string `json:"statusFilterRevoked"`
	StatusFilterTitle           string `json:"statusFilterTitle"`
	StatusFilterValid           string `json:"statusFilterValid"`
	StatusLabelExpired          string `json:"statusLabelExpired"`
	StatusLabelRevoked          string `json:"statusLabelRevoked"`
	StatusLabelValid            string `json:"statusLabelValid"`
	SummaryAll                  string `json:"summaryAll"`
	SummaryNoCertificates       string `json:"summaryNoCertificates"`
	SummarySome                 string `json:"summarySome"`
	TechnicalDetailsTitle       string `json:"technicalDetailsTitle"`
	VaultConnectionLost         string `json:"vaultConnectionLost"`
	VaultConnectionRestored     string `json:"vaultConnectionRestored"`
	AdminTitle                  string `json:"adminTitle"`
	AdminBackToVCV              string `json:"adminBackToVCV"`
	AdminSettings               string `json:"adminSettings"`
	AdminSettingsSaved          string `json:"adminSettingsSaved"`
	AdminLogout                 string `json:"adminLogout"`
	AdminLogin                  string `json:"adminLogin"`
	AdminPassword               string `json:"adminPassword"`
	AdminCertificates           string `json:"adminCertificates"`
	AdminCriticalThreshold      string `json:"adminCriticalThreshold"`
	AdminWarningThreshold       string `json:"adminWarningThreshold"`
	AdminCORS                   string `json:"adminCORS"`
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
	AdminVaultRemove            string `json:"adminVaultRemove"`
	AdminVaultTLSTip            string `json:"adminVaultTLSTip"`
	AdminToggleEnable           string `json:"adminToggleEnable"`
	AdminLoginHint              string `json:"adminLoginHint"`
	AdminVaultsEmpty            string `json:"adminVaultsEmpty"`
	AdminCORSOriginsHint        string `json:"adminCORSOriginsHint"`
	AdminVaultTokenHint         string `json:"adminVaultTokenHint"`
	AdminThresholdsHint         string `json:"adminThresholdsHint"`
}

// Response is the payload returned by the /api/i18n endpoint.
type Response struct {
	Language Language `json:"language"`
	Messages Messages `json:"messages"`
}

var englishMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonToggleTheme:           "Toggle theme",
	ButtonClose:                 "Close",
	ButtonDetails:               "Details",
	ButtonDocumentation:         "Documentation",
	ButtonDownloadPEM:           "Download PEM",
	ButtonRefresh:               "Refresh",
	ButtonToggleFilters:         "Toggle filters",
	CacheInvalidateFailed:       "Failed to clear cache",
	CacheInvalidated:            "Cache cleared and data refreshed",
	CertificateInformationTitle: "Certificate information",
	ColumnActions:               "Actions",
	ColumnCommonName:            "Common name",
	ColumnCreatedAt:             "Created at",
	ColumnExpiresAt:             "Expires at",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Status",
	DashboardCertsLabel:         "certs",
	DashboardClearFilter:        "Clear filter",
	DashboardCritical:           "Expires very soon",
	DashboardCriticalDesc:       "Valid certificates expiring within the critical threshold",
	DashboardExpired:            "Expired",
	DashboardExpiredDesc:        "Certificates past their expiration date",
	DashboardWarning:            "Expiring soon",
	DashboardWarningDesc:        "Valid certificates expiring within the warning threshold",
	DashboardFilterHint:         "Click a status to filter the table",
	DashboardRevoked:            "Revoked",
	DashboardRevokedDesc:        "Certificates explicitly revoked in Vault",
	DashboardTotal:              "Total certificates",
	DashboardValid:              "Valid",
	DashboardValidDesc:          "Certificates not expiring within the warning threshold",
	DaysRemaining:               "{{days}} days remaining",
	DaysRemainingShort:          "{{days}}d",
	DaysRemainingSingular:       "{{days}} day remaining",
	ExpiredSince:                "Expired since the {{date}}",
	ExpiredToday:                "Expired today",
	ExpiredDays:                 "Expired {{days}} days ago",
	ExpiredDaysSingular:         "Expired {{days}} day ago",
	ExpiringToday:               "Expires today",
	DeselectAll:                 "Deselect all",
	DownloadPEMFailed:           "Failed to download certificate PEM ({{status}})",
	DownloadPEMNetworkError:     "Network error downloading certificate PEM. Please try again.",
	DownloadPEMSuccess:          "Certificate PEM downloaded successfully",
	AdminDocsTitle:              "Admin documentation",
	ExpiryFilter30Days:          "≤ 30 days",
	ExpiryFilter7Days:           "≤ 7 days",
	ExpiryFilter90Days:          "≤ 90 days",
	ExpiryFilterAll:             "All dates",
	FooterVaultConnected:        "Vault: connected",
	FooterVaultDisconnected:     "Vault: disconnected",
	FooterVaultLoading:          "Vault: …",
	FooterVaultNotConfigured:    "Vault: not configured",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} up",
	FooterVersion:               "VCV v{{version}}",
	LabelFingerprintSHA1:        "SHA-1 Fingerprint",
	LabelFingerprintSHA256:      "SHA-256 Fingerprint",
	LabelIssuer:                 "Issuer",
	LabelKeyAlgorithm:           "Key Algorithm",
	LabelLanguage:               "Language",
	LabelLoading:                "Loading...",
	LabelPEM:                    "PEM Certificate",
	LabelSerialNumber:           "Serial Number",
	LabelSubject:                "Subject",
	LabelUsage:                  "Usage",
	LegendExpiredText:           "Past the expiration date.",
	LegendExpiredTitle:          "Expired",
	LegendRevokedText:           "Explicitly revoked in Vault.",
	LegendRevokedTitle:          "Revoked",
	LegendValidText:             "Not expired and not revoked.",
	LegendValidTitle:            "Valid",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsFailed:           "Failed to load certificate details ({{status}})",
	LoadDetailsNetworkError:     "Network error loading certificate details. Please try again.",
	LoadFailed:                  "Failed to load certificates ({{status}})",
	LoadNetworkError:            "Network error loading certificates. Please try again.",
	LoadSuccess:                 "Certificates loaded successfully",
	LoadUnexpectedFormat:        "Unexpected response format from server",
	LoadingDetails:              "Loading certificate details...",
	ModalDetailsTitle:           "Certificate details",
	ModalVaultStatusTitle:       "Vault status",
	MountSearchPlaceholder:      "Search vaults or PKI engines...",
	MountSelectorTitle:          "Vaults & PKI mounts",
	MountSelectorTooltip:        "Filter certificates by Vault instance and PKI mount",
	MountStatsSelected:          "Selected",
	MountStatsTotal:             "Total",
	NoData:                      "No data",
	NotificationCritical:        "{{count}} certificate(s) expiring within {{threshold}} days or less!",
	NotificationWarning:         "{{count}} certificate(s) expiring within {{threshold}} days or less",
	PaginationAll:               "All results",
	PaginationInfo:              "Page {{current}} of {{total}}",
	PaginationNext:              "Next",
	PaginationPageSizeLabel:     "Results per page",
	PaginationPrev:              "Previous",
	SearchPlaceholder:           "Search by Common Name or SAN",
	SelectAll:                   "Select all",
	StatusFilterAll:             "All",
	StatusFilterExpired:         "Expired",
	StatusFilterExpiring:        "Expiring soon",
	StatusFilterRevoked:         "Revoked",
	StatusFilterTitle:           "Status filter",
	StatusFilterValid:           "Valid",
	StatusLabelExpired:          "Expired",
	StatusLabelRevoked:          "Revoked",
	StatusLabelValid:            "Valid",
	SummaryAll:                  "{{total}} certificates",
	SummaryNoCertificates:       "No certificates.",
	SummarySome:                 "{{visible}} of {{total}} certificates shown",
	TechnicalDetailsTitle:       "Technical details",
	VaultConnectionLost:         "Vault connection lost",
	VaultConnectionRestored:     "Vault connection restored",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Back to VCV",
	AdminSettings:               "Settings",
	AdminSettingsSaved:          "Settings saved",
	AdminLogout:                 "Logout",
	AdminLogin:                  "Login",
	AdminPassword:               "Password",
	AdminCertificates:           "Certificates",
	AdminCriticalThreshold:      "Critical threshold (days)",
	AdminWarningThreshold:       "Warning threshold (days)",
	AdminCORS:                   "CORS",
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
	AdminVaultRemove:            "Remove",
	AdminVaultTLSTip:            "TLS tip: Provide the CA bundle either inline as base64 (preferred) or via a PEM file path / CA directory. If \"TLS CA cert (base64)\" is set, it takes precedence and the file/path fields are ignored. Encode a PEM bundle with: cat /path/to/ca.pem | base64 | tr -d '\\n'. \"TLS server name\" overrides SNI. \"TLS insecure\" disables verification (development only).",
	AdminToggleEnable:           "Enable",
	AdminLoginHint:              "Use the bcrypt-hashed password set in VCV_ADMIN_PASSWORD.",
	AdminVaultsEmpty:            "No Vault instances configured yet. Click \"Add vault\" to get started.",
	AdminCORSOriginsHint:        "e.g. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Vault access token. This value is stored in the settings file.",
	AdminThresholdsHint:         "Certificates expiring within these thresholds are flagged in the dashboard.",
}

var frenchMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonToggleTheme:           "Changer de thème",
	ButtonClose:                 "Fermer",
	ButtonDetails:               "Détails",
	ButtonDocumentation:         "Documentation",
	ButtonDownloadPEM:           "Télécharger PEM",
	ButtonRefresh:               "Rafraîchir",
	ButtonToggleFilters:         "Afficher/masquer les filtres",
	CacheInvalidateFailed:       "Échec du vidage du cache",
	CacheInvalidated:            "Cache vidé et données actualisées",
	CertificateInformationTitle: "Informations du certificat",
	ColumnActions:               "Actions",
	ColumnCommonName:            "Nom commun",
	ColumnCreatedAt:             "Créé le",
	ColumnExpiresAt:             "Expire le",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Statut",
	DashboardCertsLabel:         "certs",
	DashboardClearFilter:        "Effacer le filtre",
	DashboardCritical:           "Expire très bientôt",
	DashboardCriticalDesc:       "Certificats valides expirant dans le seuil critique",
	DashboardExpired:            "Expirés",
	DashboardExpiredDesc:        "Certificats ayant dépassé leur date d'expiration",
	DashboardWarning:            "Expire bientôt",
	DashboardWarningDesc:        "Certificats valides expirant dans le seuil d'avertissement",
	DashboardFilterHint:         "Cliquez sur un statut pour filtrer le tableau",
	DashboardRevoked:            "Révoqués",
	DashboardRevokedDesc:        "Certificats explicitement révoqués dans Vault",
	DashboardTotal:              "Total des certificats",
	DashboardValid:              "Valides",
	DashboardValidDesc:          "Certificats n'expirant pas dans le seuil d'avertissement",
	DaysRemaining:               "{{days}} jours restants",
	DaysRemainingShort:          "{{days}}j",
	DaysRemainingSingular:       "{{days}} jour restant",
	ExpiredSince:                "Expiré depuis le {{date}}",
	ExpiredToday:                "Expiré aujourd'hui",
	ExpiredDays:                 "Expiré il y a {{days}} jours",
	ExpiredDaysSingular:         "Expiré il y a {{days}} jour",
	ExpiringToday:               "Expire aujourd'hui",
	DeselectAll:                 "Tout désélectionner",
	DownloadPEMFailed:           "Échec du téléchargement du certificat PEM ({{status}})",
	DownloadPEMNetworkError:     "Erreur réseau lors du téléchargement du certificat PEM. Veuillez réessayer.",
	DownloadPEMSuccess:          "Certificat PEM téléchargé avec succès",
	AdminDocsTitle:              "Documentation admin",
	ExpiryFilter30Days:          "≤ 30 jours",
	ExpiryFilter7Days:           "≤ 7 jours",
	ExpiryFilter90Days:          "≤ 90 jours",
	ExpiryFilterAll:             "Toutes les dates",
	FooterVaultConnected:        "Vault : connecté",
	FooterVaultDisconnected:     "Vault : déconnecté",
	FooterVaultLoading:          "Vault : …",
	FooterVaultNotConfigured:    "Vault : non configuré",
	FooterVaultSummary:          "Vaults : {{up}}/{{total}} OK",
	FooterVersion:               "VCV v{{version}}",
	LabelFingerprintSHA1:        "Empreinte SHA-1",
	LabelFingerprintSHA256:      "Empreinte SHA-256",
	LabelIssuer:                 "Émetteur",
	LabelKeyAlgorithm:           "Algorithme de clé",
	LabelLanguage:               "Langue",
	LabelLoading:                "Chargement...",
	LabelPEM:                    "Certificat PEM",
	LabelSerialNumber:           "Numéro de série",
	LabelSubject:                "Sujet",
	LabelUsage:                  "Utilisation",
	LegendExpiredText:           "Date d'expiration dépassée.",
	LegendExpiredTitle:          "Expiré",
	LegendRevokedText:           "Révoqué explicitement dans Vault.",
	LegendRevokedTitle:          "Révoqué",
	LegendValidText:             "Non expiré et non révoqué.",
	LegendValidTitle:            "Valide",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsFailed:           "Échec du chargement des détails du certificat ({{status}})",
	LoadDetailsNetworkError:     "Erreur réseau lors du chargement des détails du certificat. Veuillez réessayer.",
	LoadFailed:                  "Échec du chargement des certificats ({{status}})",
	LoadNetworkError:            "Erreur réseau lors du chargement des certificats. Veuillez réessayer.",
	LoadSuccess:                 "Certificats chargés avec succès",
	LoadUnexpectedFormat:        "Format de réponse inattendu du serveur",
	LoadingDetails:              "Chargement des détails du certificat...",
	ModalDetailsTitle:           "Détails du certificat",
	ModalVaultStatusTitle:       "Statut Vaults",
	MountSearchPlaceholder:      "Rechercher des vaults ou moteurs PKI...",
	MountSelectorTitle:          "Vaults et montages PKI",
	MountSelectorTooltip:        "Filtrer les certificats par instance Vault et montage PKI",
	MountStatsSelected:          "Sélectionnés",
	MountStatsTotal:             "Total",
	NoData:                      "Aucune donnée",
	NotificationCritical:        "{{count}} certificat(s) expirant dans {{threshold}} jours ou moins !",
	NotificationWarning:         "{{count}} certificat(s) expirant dans {{threshold}} jours ou moins",
	PaginationAll:               "Tous les résultats",
	PaginationInfo:              "Page {{current}} sur {{total}}",
	PaginationNext:              "Suivant",
	PaginationPageSizeLabel:     "Résultats par page",
	PaginationPrev:              "Précédent",
	SearchPlaceholder:           "Rechercher par Common Name ou SAN",
	SelectAll:                   "Tout sélectionner",
	StatusFilterAll:             "Tous",
	StatusFilterExpired:         "Expiré",
	StatusFilterExpiring:        "Expiration proche",
	StatusFilterRevoked:         "Révoqué",
	StatusFilterTitle:           "Filtre des statuts",
	StatusFilterValid:           "Valide",
	StatusLabelExpired:          "Expiré",
	StatusLabelRevoked:          "Révoqué",
	StatusLabelValid:            "Valide",
	SummaryAll:                  "{{total}} certificats",
	SummaryNoCertificates:       "Aucun certificat.",
	SummarySome:                 "{{visible}} sur {{total}} certificats affichés",
	TechnicalDetailsTitle:       "Détails techniques",
	VaultConnectionLost:         "Connexion à Vault perdue",
	VaultConnectionRestored:     "Connexion à Vault rétablie",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Retour à VCV",
	AdminSettings:               "Paramètres",
	AdminSettingsSaved:          "Paramètres enregistrés",
	AdminLogout:                 "Déconnexion",
	AdminLogin:                  "Connexion",
	AdminPassword:               "Mot de passe",
	AdminCertificates:           "Certificats",
	AdminCriticalThreshold:      "Seuil critique (jours)",
	AdminWarningThreshold:       "Seuil d'avertissement (jours)",
	AdminCORS:                   "CORS",
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
	AdminVaultRemove:            "Supprimer",
	AdminVaultTLSTip:            "Astuce TLS : Fournissez le bundle CA soit en ligne en base64 (préféré) soit via un chemin de fichier PEM / répertoire CA. Si \"Certificat CA TLS (base64)\" est défini, il a la priorité et les champs fichier/chemin sont ignorés. Encodez un bundle PEM avec : cat /chemin/vers/ca.pem | base64 | tr -d '\\n'. \"Nom du serveur TLS\" remplace SNI. \"TLS non sécurisé\" désactive la vérification (développement uniquement).",
	AdminToggleEnable:           "Activer",
	AdminLoginHint:              "Utilisez le mot de passe haché en bcrypt défini dans VCV_ADMIN_PASSWORD.",
	AdminVaultsEmpty:            "Aucune instance Vault configurée. Cliquez sur « Ajouter un vault » pour commencer.",
	AdminCORSOriginsHint:        "ex. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Jeton d'accès Vault. Cette valeur est stockée dans le fichier de paramètres.",
	AdminThresholdsHint:         "Les certificats expirant dans ces seuils sont signalés dans le tableau de bord.",
}

var spanishMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonToggleTheme:           "Cambiar tema",
	ButtonClose:                 "Cerrar",
	ButtonDetails:               "Detalles",
	ButtonDocumentation:         "Documentación",
	ButtonDownloadPEM:           "Descargar PEM",
	ButtonRefresh:               "Actualizar",
	ButtonToggleFilters:         "Mostrar/ocultar filtros",
	CacheInvalidateFailed:       "Error al borrar el caché",
	CacheInvalidated:            "Caché borrado y datos actualizados",
	CertificateInformationTitle: "Información del certificado",
	ColumnActions:               "Acciones",
	ColumnCommonName:            "Nombre común",
	ColumnCreatedAt:             "Creado el",
	ColumnExpiresAt:             "Caduca el",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Estado",
	DashboardCertsLabel:         "certs",
	DashboardClearFilter:        "Borrar filtro",
	DashboardCritical:           "Caduca muy pronto",
	DashboardCriticalDesc:       "Certificados válidos que caducan dentro del umbral crítico",
	DashboardExpired:            "Caducados",
	DashboardExpiredDesc:        "Certificados que han superado su fecha de caducidad",
	DashboardWarning:            "Caduca pronto",
	DashboardWarningDesc:        "Certificados válidos que caducan dentro del umbral de advertencia",
	DashboardFilterHint:         "Haga clic en un estado para filtrar la tabla",
	DashboardRevoked:            "Revocados",
	DashboardRevokedDesc:        "Certificados explícitamente revocados en Vault",
	DashboardTotal:              "Total de certificados",
	DashboardValid:              "Válidos",
	DashboardValidDesc:          "Certificados que no caducan dentro del umbral de advertencia",
	DaysRemaining:               "{{days}} días restantes",
	DaysRemainingShort:          "{{days}}d",
	DaysRemainingSingular:       "{{days}} día restante",
	ExpiredSince:                "Vencido desde el {{date}}",
	ExpiredToday:                "Vencido hoy",
	ExpiredDays:                 "Vencido hace {{days}} días",
	ExpiredDaysSingular:         "Vencido hace {{days}} día",
	ExpiringToday:               "Vence hoy",
	DeselectAll:                 "Deseleccionar todo",
	DownloadPEMFailed:           "Error al descargar el certificado PEM ({{status}})",
	DownloadPEMNetworkError:     "Error de red al descargar el certificado PEM. Por favor intente nuevamente.",
	DownloadPEMSuccess:          "Certificado PEM descargado exitosamente",
	AdminDocsTitle:              "Documentación admin",
	ExpiryFilter30Days:          "≤ 30 días",
	ExpiryFilter7Days:           "≤ 7 días",
	ExpiryFilter90Days:          "≤ 90 días",
	ExpiryFilterAll:             "Todas las fechas",
	FooterVaultConnected:        "Vault: conectado",
	FooterVaultDisconnected:     "Vault: desconectado",
	FooterVaultLoading:          "Vault: …",
	FooterVaultNotConfigured:    "Vault: no configurado",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} OK",
	FooterVersion:               "VCV v{{version}}",
	LabelFingerprintSHA1:        "Huella SHA-1",
	LabelFingerprintSHA256:      "Huella SHA-256",
	LabelIssuer:                 "Emisor",
	LabelKeyAlgorithm:           "Algoritmo de clave",
	LabelLanguage:               "Idioma",
	LabelLoading:                "Cargando...",
	LabelPEM:                    "Certificado PEM",
	LabelSerialNumber:           "Número de serie",
	LabelSubject:                "Sujeto",
	LabelUsage:                  "Uso",
	LegendExpiredText:           "Fecha de vencimiento superada.",
	LegendExpiredTitle:          "Caducado",
	LegendRevokedText:           "Revocado explícitamente en Vault.",
	LegendRevokedTitle:          "Revocado",
	LegendValidText:             "No caducado y no revocado.",
	LegendValidTitle:            "Válido",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsFailed:           "Error al cargar los detalles del certificado ({{status}})",
	LoadDetailsNetworkError:     "Error de red al cargar los detalles del certificado. Por favor intente nuevamente.",
	LoadFailed:                  "Error al cargar los certificados ({{status}})",
	LoadNetworkError:            "Error de red al cargar los certificados. Por favor intente nuevamente.",
	LoadSuccess:                 "Certificados cargados exitosamente",
	LoadUnexpectedFormat:        "Formato de respuesta inesperado del servidor",
	LoadingDetails:              "Cargando detalles del certificado...",
	ModalDetailsTitle:           "Detalles del certificado",
	ModalVaultStatusTitle:       "Estado Vault",
	MountSearchPlaceholder:      "Buscar vaults o motores PKI...",
	MountSelectorTitle:          "Vaults y montajes PKI",
	MountSelectorTooltip:        "Filtrar certificados por instancia de Vault y montaje PKI",
	MountStatsSelected:          "Seleccionados",
	MountStatsTotal:             "Total",
	NoData:                      "Sin datos",
	NotificationCritical:        "{{count}} certificado(s) caducando en {{threshold}} días o menos!",
	NotificationWarning:         "{{count}} certificado(s) caducando en {{threshold}} días o menos",
	PaginationAll:               "Todos los resultados",
	PaginationInfo:              "Página {{current}} de {{total}}",
	PaginationNext:              "Siguiente",
	PaginationPageSizeLabel:     "Resultados por página",
	PaginationPrev:              "Anterior",
	SearchPlaceholder:           "Buscar por Common Name o SAN",
	SelectAll:                   "Seleccionar todo",
	StatusFilterAll:             "Todos",
	StatusFilterExpired:         "Caducado",
	StatusFilterExpiring:        "Próximo a caducar",
	StatusFilterRevoked:         "Revocado",
	StatusFilterTitle:           "Filtro de estado",
	StatusFilterValid:           "Válido",
	StatusLabelExpired:          "Caducado",
	StatusLabelRevoked:          "Revocado",
	StatusLabelValid:            "Válido",
	SummaryAll:                  "{{total}} certificados",
	SummaryNoCertificates:       "Ningún certificado.",
	SummarySome:                 "{{visible}} de {{total}} certificados mostrados",
	TechnicalDetailsTitle:       "Detalles técnicos",
	VaultConnectionLost:         "Conexión a Vault perdida",
	VaultConnectionRestored:     "Conexión a Vault restablecida",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Volver a VCV",
	AdminSettings:               "Configuración",
	AdminSettingsSaved:          "Configuración guardada",
	AdminLogout:                 "Cerrar sesión",
	AdminLogin:                  "Iniciar sesión",
	AdminPassword:               "Contraseña",
	AdminCertificates:           "Certificados",
	AdminCriticalThreshold:      "Umbral crítico (días)",
	AdminWarningThreshold:       "Umbral de advertencia (días)",
	AdminCORS:                   "CORS",
	AdminCORSOrigins:            "Orígenes permitidos (separados por comas)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Administrar instancias de Vault configuradas.",
	AdminAddVault:               "Agregar vault",
	AdminSaveSettings:           "Guardar settings.json",
	AdminRestartNote:            "Los cambios se guardan en el archivo de configuración. Es posible que se requiera reiniciar el servidor para que todos los cambios surtan efecto.",
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
	AdminToggleEnable:           "Habilitar",
	AdminLoginHint:              "Use la contraseña hasheada con bcrypt configurada en VCV_ADMIN_PASSWORD.",
	AdminVaultsEmpty:            "No hay instancias de Vault configuradas. Haga clic en \"Agregar vault\" para comenzar.",
	AdminCORSOriginsHint:        "ej. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Token de acceso de Vault. Este valor se almacena en el archivo de configuración.",
	AdminThresholdsHint:         "Los certificados que expiran dentro de estos umbrales se señalan en el panel.",
}

var germanMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonToggleTheme:           "Design umschalten",
	ButtonClose:                 "Schließen",
	ButtonDetails:               "Details",
	ButtonDocumentation:         "Dokumentation",
	ButtonDownloadPEM:           "PEM herunterladen",
	ButtonRefresh:               "Aktualisieren",
	ButtonToggleFilters:         "Filter ein-/ausblenden",
	CacheInvalidateFailed:       "Cache konnte nicht geleert werden",
	CacheInvalidated:            "Cache geleert und Daten aktualisiert",
	CertificateInformationTitle: "Zertifikatsinformationen",
	ColumnActions:               "Aktionen",
	ColumnCommonName:            "Allgemeiner Name",
	ColumnCreatedAt:             "Erstellt am",
	ColumnExpiresAt:             "Gültig bis",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Status",
	DashboardCertsLabel:         "Zert.",
	DashboardClearFilter:        "Filter löschen",
	DashboardCritical:           "Läuft sehr bald ab",
	DashboardCriticalDesc:       "Gültige Zertifikate, die innerhalb des kritischen Schwellenwerts ablaufen",
	DashboardExpired:            "Abgelaufen",
	DashboardExpiredDesc:        "Zertifikate, deren Ablaufdatum überschritten ist",
	DashboardWarning:            "Läuft bald ab",
	DashboardWarningDesc:        "Gültige Zertifikate, die innerhalb des Warnschwellenwerts ablaufen",
	DashboardFilterHint:         "Klicken Sie auf einen Status, um die Tabelle zu filtern",
	DashboardRevoked:            "Widerrufen",
	DashboardRevokedDesc:        "Zertifikate, die explizit in Vault widerrufen wurden",
	DashboardTotal:              "Zertifikate gesamt",
	DashboardValid:              "Gültig",
	DashboardValidDesc:          "Zertifikate, die nicht innerhalb des Warnschwellenwerts ablaufen",
	DaysRemaining:               "{{days}} verbleibende Tage",
	DaysRemainingShort:          "{{days}}T",
	DaysRemainingSingular:       "{{days}} verbleibender Tag",
	ExpiredSince:                "Abgelaufen seit dem {{date}}",
	ExpiredToday:                "Heute abgelaufen",
	ExpiredDays:                 "Vor {{days}} Tagen abgelaufen",
	ExpiredDaysSingular:         "Vor {{days}} Tag abgelaufen",
	ExpiringToday:               "Läuft heute ab",
	DeselectAll:                 "Alle abwählen",
	DownloadPEMFailed:           "Zertifikat-PEM konnte nicht heruntergeladen werden ({{status}})",
	DownloadPEMNetworkError:     "Netzwerkfehler beim Herunterladen des Zertifikat-PEM. Bitte versuchen Sie es erneut.",
	DownloadPEMSuccess:          "Zertifikat-PEM erfolgreich heruntergeladen",
	AdminDocsTitle:              "Admin-dokumentation",
	ExpiryFilter30Days:          "≤ 30 Tage",
	ExpiryFilter7Days:           "≤ 7 Tage",
	ExpiryFilter90Days:          "≤ 90 Tage",
	ExpiryFilterAll:             "Alle Daten",
	FooterVaultConnected:        "Vault: verbunden",
	FooterVaultDisconnected:     "Vault: getrennt",
	FooterVaultLoading:          "Vault: …",
	FooterVaultNotConfigured:    "Vault: nicht konfiguriert",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} OK",
	FooterVersion:               "VCV v{{version}}",
	LabelFingerprintSHA1:        "SHA-1-Fingerabdruck",
	LabelFingerprintSHA256:      "SHA-256-Fingerabdruck",
	LabelIssuer:                 "Aussteller",
	LabelKeyAlgorithm:           "Schlüsselalgorithmus",
	LabelLanguage:               "Sprache",
	LabelLoading:                "Wird geladen...",
	LabelPEM:                    "PEM-Zertifikat",
	LabelSerialNumber:           "Seriennummer",
	LabelSubject:                "Betreff",
	LabelUsage:                  "Verwendung",
	LegendExpiredText:           "Ablaufdatum überschritten.",
	LegendExpiredTitle:          "Abgelaufen",
	LegendRevokedText:           "Explizit in Vault widerrufen.",
	LegendRevokedTitle:          "Widerrufen",
	LegendValidText:             "Nicht abgelaufen und nicht widerrufen.",
	LegendValidTitle:            "Gültig",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsFailed:           "Zertifikatsdetails konnten nicht geladen werden ({{status}})",
	LoadDetailsNetworkError:     "Netzwerkfehler beim Laden der Zertifikatsdetails. Bitte versuchen Sie es erneut.",
	LoadFailed:                  "Zertifikate konnten nicht geladen werden ({{status}})",
	LoadNetworkError:            "Netzwerkfehler beim Laden der Zertifikate. Bitte versuchen Sie es erneut.",
	LoadSuccess:                 "Zertifikate erfolgreich geladen",
	LoadUnexpectedFormat:        "Unerwartetes Antwortformat vom Server",
	LoadingDetails:              "Zertifikatsdetails werden geladen...",
	ModalDetailsTitle:           "Zertifikatsdetails",
	ModalVaultStatusTitle:       "Vault-Status",
	MountSearchPlaceholder:      "Vaults oder PKI-Motoren suchen...",
	MountSelectorTitle:          "Vaults & PKI-Mounts",
	MountSelectorTooltip:        "Zertifikate nach Vault-Instanz und PKI-Mount filtern",
	MountStatsSelected:          "Ausgewählt",
	MountStatsTotal:             "Gesamt",
	NoData:                      "Keine Daten",
	NotificationCritical:        "{{count}} Zertifikat(e) laufen in {{threshold}} Tagen oder weniger ab!",
	NotificationWarning:         "{{count}} Zertifikat(e) laufen in {{threshold}} Tagen oder weniger ab",
	PaginationAll:               "Alle Ergebnisse",
	PaginationInfo:              "Seite {{current}} von {{total}}",
	PaginationNext:              "Weiter",
	PaginationPageSizeLabel:     "Ergebnisse pro Seite",
	PaginationPrev:              "Zurück",
	SearchPlaceholder:           "Suche nach Common Name oder SAN",
	SelectAll:                   "Alle auswählen",
	StatusFilterAll:             "Alle",
	StatusFilterExpired:         "Abgelaufen",
	StatusFilterExpiring:        "Bald ablaufend",
	StatusFilterRevoked:         "Widerrufen",
	StatusFilterTitle:           "Statusfilter",
	StatusFilterValid:           "Gültig",
	StatusLabelExpired:          "Abgelaufen",
	StatusLabelRevoked:          "Widerrufen",
	StatusLabelValid:            "Gültig",
	SummaryAll:                  "{{total}} Zertifikate",
	SummaryNoCertificates:       "Keine Zertifikate.",
	SummarySome:                 "{{visible}} von {{total}} Zertifikaten angezeigt",
	TechnicalDetailsTitle:       "Technische Details",
	VaultConnectionLost:         "Verbindung zu Vault unterbrochen",
	VaultConnectionRestored:     "Verbindung zu Vault wiederhergestellt",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Zurück zu VCV",
	AdminSettings:               "Einstellungen",
	AdminSettingsSaved:          "Einstellungen gespeichert",
	AdminLogout:                 "Abmelden",
	AdminLogin:                  "Anmelden",
	AdminPassword:               "Passwort",
	AdminCertificates:           "Zertifikate",
	AdminCriticalThreshold:      "Kritischer Schwellenwert (Tage)",
	AdminWarningThreshold:       "Warnschwellenwert (Tage)",
	AdminCORS:                   "CORS",
	AdminCORSOrigins:            "Erlaubte Ursprünge (durch Kommas getrennt)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Konfigurierte Vault-Instanzen verwalten.",
	AdminAddVault:               "Vault hinzufügen",
	AdminSaveSettings:           "settings.json speichern",
	AdminRestartNote:            "Änderungen werden in der Einstellungsdatei gespeichert. Ein Neustart des Servers kann erforderlich sein, damit alle Änderungen wirksam werden.",
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
	AdminToggleEnable:           "Aktivieren",
	AdminLoginHint:              "Verwenden Sie das bcrypt-gehashte Passwort aus VCV_ADMIN_PASSWORD.",
	AdminVaultsEmpty:            "Noch keine Vault-Instanzen konfiguriert. Klicken Sie auf \"Vault hinzufügen\", um zu beginnen.",
	AdminCORSOriginsHint:        "z.B. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Vault-Zugriffstoken. Dieser Wert wird in der Einstellungsdatei gespeichert.",
	AdminThresholdsHint:         "Zertifikate, die innerhalb dieser Schwellenwerte ablaufen, werden im Dashboard markiert.",
}

var italianMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonToggleTheme:           "Cambia tema",
	ButtonClose:                 "Chiudi",
	ButtonDetails:               "Dettagli",
	ButtonDocumentation:         "Documentazione",
	ButtonDownloadPEM:           "Scarica PEM",
	ButtonRefresh:               "Aggiorna",
	ButtonToggleFilters:         "Mostra/nascondi filtri",
	CacheInvalidateFailed:       "Impossibile cancellare la cache",
	CacheInvalidated:            "Cache cancellata e dati aggiornati",
	CertificateInformationTitle: "Informazioni sul certificato",
	ColumnActions:               "Azioni",
	ColumnCommonName:            "Nome comune",
	ColumnCreatedAt:             "Creato il",
	ColumnExpiresAt:             "Scade il",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Stato",
	DashboardCertsLabel:         "cert.",
	DashboardClearFilter:        "Cancella filtro",
	DashboardCritical:           "Scade molto presto",
	DashboardCriticalDesc:       "Certificati validi in scadenza entro la soglia critica",
	DashboardExpired:            "Scaduti",
	DashboardExpiredDesc:        "Certificati che hanno superato la data di scadenza",
	DashboardWarning:            "In scadenza",
	DashboardWarningDesc:        "Certificati validi in scadenza entro la soglia di avviso",
	DashboardFilterHint:         "Clicca su uno stato per filtrare la tabella",
	DashboardRevoked:            "Revocati",
	DashboardRevokedDesc:        "Certificati esplicitamente revocati in Vault",
	DashboardTotal:              "Certificati totali",
	DashboardValid:              "Validi",
	DashboardValidDesc:          "Certificati che non scadono entro la soglia di avviso",
	DaysRemaining:               "{{days}} giorni rimanenti",
	DaysRemainingShort:          "{{days}}g",
	DaysRemainingSingular:       "{{days}} giorno rimanente",
	ExpiredSince:                "Scaduto dal {{date}}",
	ExpiredToday:                "Scaduto oggi",
	ExpiredDays:                 "Scaduto {{days}} giorni fa",
	ExpiredDaysSingular:         "Scaduto {{days}} giorno fa",
	ExpiringToday:               "Scade oggi",
	DeselectAll:                 "Deseleziona tutto",
	DownloadPEMFailed:           "Impossibile scaricare il certificato PEM ({{status}})",
	DownloadPEMNetworkError:     "Errore di rete durante il download del certificato PEM. Riprova.",
	DownloadPEMSuccess:          "Certificato PEM scaricato con successo",
	AdminDocsTitle:              "Documentazione admin",
	ExpiryFilter30Days:          "≤ 30 giorni",
	ExpiryFilter7Days:           "≤ 7 giorni",
	ExpiryFilter90Days:          "≤ 90 giorni",
	ExpiryFilterAll:             "Tutte le date",
	FooterVaultConnected:        "Vault: connesso",
	FooterVaultDisconnected:     "Vault: disconnesso",
	FooterVaultLoading:          "Vault: …",
	FooterVaultNotConfigured:    "Vault: non configurato",
	FooterVaultSummary:          "Vaults: {{up}}/{{total}} OK",
	FooterVersion:               "VCV v{{version}}",
	LabelFingerprintSHA1:        "Impronta SHA-1",
	LabelFingerprintSHA256:      "Impronta SHA-256",
	LabelIssuer:                 "Emittente",
	LabelKeyAlgorithm:           "Algoritmo della chiave",
	LabelLanguage:               "Lingua",
	LabelLoading:                "Caricamento...",
	LabelPEM:                    "Certificato PEM",
	LabelSerialNumber:           "Numero di serie",
	LabelSubject:                "Soggetto",
	LabelUsage:                  "Utilizzo",
	LegendExpiredText:           "Data di scadenza superata.",
	LegendExpiredTitle:          "Scaduto",
	LegendRevokedText:           "Revocato esplicitamente in Vault.",
	LegendRevokedTitle:          "Revocato",
	LegendValidText:             "Non scaduto e non revocato.",
	LegendValidTitle:            "Valido",
	LabelVault:                  "Vault",
	LabelPKI:                    "PKI",
	LoadDetailsFailed:           "Impossibile caricare i dettagli del certificato ({{status}})",
	LoadDetailsNetworkError:     "Errore di rete durante il caricamento dei dettagli del certificato. Riprova.",
	LoadFailed:                  "Impossibile caricare i certificati ({{status}})",
	LoadNetworkError:            "Errore di rete durante il caricamento dei certificati. Riprova.",
	LoadSuccess:                 "Certificati caricati correttamente",
	LoadUnexpectedFormat:        "Formato di risposta inatteso dal server",
	ModalDetailsTitle:           "Dettagli del certificato",
	ModalVaultStatusTitle:       "Stato Vault",
	MountSearchPlaceholder:      "Cerca vaults o motori PKI...",
	MountSelectorTitle:          "Vaults e mount PKI",
	MountSelectorTooltip:        "Filtra i certificati per istanza Vault e mount PKI",
	MountStatsSelected:          "Selezionati",
	MountStatsTotal:             "Totale",
	NoData:                      "Nessun dato",
	NotificationCritical:        "{{count}} certificato/i in scadenza entro {{threshold}} giorni o meno!",
	NotificationWarning:         "{{count}} certificato/i in scadenza entro {{threshold}} giorni o meno",
	PaginationAll:               "Tutti i risultati",
	PaginationInfo:              "Pagina {{current}} di {{total}}",
	PaginationNext:              "Successivo",
	PaginationPageSizeLabel:     "Risultati per pagina",
	PaginationPrev:              "Precedente",
	SearchPlaceholder:           "Cerca per Common Name o SAN",
	SelectAll:                   "Seleziona tutto",
	StatusFilterAll:             "Tutti",
	StatusFilterExpired:         "Scaduto",
	StatusFilterExpiring:        "In scadenza",
	StatusFilterRevoked:         "Revocato",
	StatusFilterTitle:           "Filtro di stato",
	StatusFilterValid:           "Valido",
	StatusLabelExpired:          "Scaduto",
	StatusLabelRevoked:          "Revocato",
	StatusLabelValid:            "Valido",
	SummaryAll:                  "{{total}} certificati",
	SummaryNoCertificates:       "Nessun certificato.",
	SummarySome:                 "Mostrati {{visible}} di {{total}} certificati",
	TechnicalDetailsTitle:       "Dettagli tecnici",
	VaultConnectionLost:         "Connessione al Vault interrotta",
	VaultConnectionRestored:     "Connessione al Vault ripristinata",
	AdminTitle:                  "VaultCertsViewer Admin",
	AdminBackToVCV:              "Torna a VCV",
	AdminSettings:               "Impostazioni",
	AdminSettingsSaved:          "Impostazioni salvate",
	AdminLogout:                 "Disconnetti",
	AdminLogin:                  "Accedi",
	AdminPassword:               "Password",
	AdminCertificates:           "Certificati",
	AdminCriticalThreshold:      "Soglia critica (giorni)",
	AdminWarningThreshold:       "Soglia di avviso (giorni)",
	AdminCORS:                   "CORS",
	AdminCORSOrigins:            "Origini consentite (separate da virgole)",
	AdminVaults:                 "Vaults",
	AdminVaultsHint:             "Gestisci le istanze Vault configurate.",
	AdminAddVault:               "Aggiungi vault",
	AdminSaveSettings:           "Salva settings.json",
	AdminRestartNote:            "Le modifiche vengono salvate nel file delle impostazioni. Potrebbe essere necessario riavviare il server affinché tutte le modifiche abbiano effetto.",
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
	AdminToggleEnable:           "Abilita",
	AdminLoginHint:              "Utilizzare la password hashata con bcrypt impostata in VCV_ADMIN_PASSWORD.",
	AdminVaultsEmpty:            "Nessuna istanza Vault configurata. Fai clic su \"Aggiungi vault\" per iniziare.",
	AdminCORSOriginsHint:        "es. https://example.com, https://other.example.com",
	AdminVaultTokenHint:         "Token di accesso Vault. Questo valore viene memorizzato nel file delle impostazioni.",
	AdminThresholdsHint:         "I certificati in scadenza entro queste soglie vengono segnalati nella dashboard.",
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
