package i18n

import "strings"

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
	ButtonClose                 string `json:"buttonClose"`
	ButtonDetails               string `json:"buttonDetails"`
	ButtonDownloadPEM           string `json:"buttonDownloadPEM"`
	ButtonRefresh               string `json:"buttonRefresh"`
	CacheInvalidateFailed       string `json:"cacheInvalidateFailed"`
	CacheInvalidated            string `json:"cacheInvalidated"`
	CertificateInformationTitle string `json:"certificateInformationTitle"`
	ChartExpiryTimeline         string `json:"chartExpiryTimeline"`
	ChartLegendExpired          string `json:"chartLegendExpired"`
	ChartLegendRevoked          string `json:"chartLegendRevoked"`
	ChartLegendValid            string `json:"chartLegendValid"`
	ChartStatusDistribution     string `json:"chartStatusDistribution"`
	ColumnActions               string `json:"columnActions"`
	ColumnCommonName            string `json:"columnCommonName"`
	ColumnCreatedAt             string `json:"columnCreatedAt"`
	ColumnExpiresAt             string `json:"columnExpiresAt"`
	ColumnSAN                   string `json:"columnSan"`
	ColumnStatus                string `json:"columnStatus"`
	DashboardExpired            string `json:"dashboardExpired"`
	DashboardExpiring           string `json:"dashboardExpiring"`
	DashboardRevoked            string `json:"dashboardRevoked"`
	DashboardTotal              string `json:"dashboardTotal"`
	DashboardValid              string `json:"dashboardValid"`
	DaysRemaining               string `json:"daysRemaining"`
	DaysRemainingShort          string `json:"daysRemainingShort"`
	DaysRemainingSingular       string `json:"daysRemainingSingular"`
	ExpiredSince                string `json:"expiredSince"`
	DeselectAll                 string `json:"deselectAll"`
	DownloadPEMFailed           string `json:"downloadPEMFailed"`
	DownloadPEMNetworkError     string `json:"downloadPEMNetworkError"`
	DownloadPEMSuccess          string `json:"downloadPEMSuccess"`
	DualStatusNote              string `json:"dualStatusNote"`
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
	LoadDetailsFailed           string `json:"loadDetailsFailed"`
	LoadDetailsNetworkError     string `json:"loadDetailsNetworkError"`
	LoadFailed                  string `json:"loadFailed"`
	LoadNetworkError            string `json:"loadNetworkError"`
	LoadSuccess                 string `json:"loadSuccess"`
	LoadUnexpectedFormat        string `json:"loadUnexpectedFormat"`
	LoadingDetails              string `json:"loadingDetails"`
	ModalDetailsTitle           string `json:"modalDetailsTitle"`
	MountSelectorTitle          string `json:"mountSelectorTitle"`
	NoCertsExpiringSoon         string `json:"noCertsExpiringSoon"`
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
}

// Response is the payload returned by the /api/i18n endpoint.
type Response struct {
	Language Language `json:"language"`
	Messages Messages `json:"messages"`
}

var englishMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonClose:                 "Close",
	ButtonDetails:               "Details",
	ButtonDownloadPEM:           "Download PEM",
	ButtonRefresh:               "Refresh",
	CacheInvalidateFailed:       "Failed to clear cache",
	CacheInvalidated:            "Cache cleared and data refreshed",
	CertificateInformationTitle: "Certificate information",
	ChartExpiryTimeline:         "Expiration timeline",
	ChartLegendExpired:          "Expired",
	ChartLegendRevoked:          "Revoked",
	ChartLegendValid:            "Valid",
	ChartStatusDistribution:     "Status distribution",
	ColumnActions:               "Actions",
	ColumnCommonName:            "Common name",
	ColumnCreatedAt:             "Created at",
	ColumnExpiresAt:             "Expires at",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Status",
	DashboardExpired:            "Expired",
	DashboardExpiring:           "Expiring soon",
	DashboardRevoked:            "Revoked",
	DashboardTotal:              "Total certificates",
	DashboardValid:              "Valid",
	DaysRemaining:               "{{days}} days remaining",
	DaysRemainingShort:          "{{days}}d",
	DaysRemainingSingular:       "{{days}} day remaining",
	ExpiredSince:                "Expired since the {{date}}",
	DeselectAll:                 "Deselect all",
	DownloadPEMFailed:           "Failed to download certificate PEM ({{status}})",
	DownloadPEMNetworkError:     "Network error downloading certificate PEM. Please try again.",
	DownloadPEMSuccess:          "Certificate PEM downloaded successfully",
	DualStatusNote:              "{{count}} certificate(s) are both expired and revoked",
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
	LoadDetailsFailed:           "Failed to load certificate details ({{status}})",
	LoadDetailsNetworkError:     "Network error loading certificate details. Please try again.",
	LoadFailed:                  "Failed to load certificates ({{status}})",
	LoadNetworkError:            "Network error loading certificates. Please try again.",
	LoadSuccess:                 "Certificates loaded successfully",
	LoadUnexpectedFormat:        "Unexpected response format from server",
	LoadingDetails:              "Loading certificate details...",
	ModalDetailsTitle:           "Certificate details",
	MountSelectorTitle:          "PKI engines",
	NoCertsExpiringSoon:         "No certificates expiring soon",
	NoData:                      "No data",
	NotificationCritical:        "{{count}} certificate(s) expiring within {{threshold}} days or less!",
	NotificationWarning:         "{{count}} certificate(s) expiring within {{threshold}} days or less",
	PaginationAll:               "All results",
	PaginationInfo:              "Page {{current}} of {{total}}",
	PaginationNext:              "Next",
	PaginationPageSizeLabel:     "Results per page",
	PaginationPrev:              "Previous",
	SearchPlaceholder:           "CN or SAN",
	SelectAll:                   "Select all",
	StatusFilterAll:             "All",
	StatusFilterExpired:         "Expired",
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
}

var frenchMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonClose:                 "Fermer",
	ButtonDetails:               "Détails",
	ButtonDownloadPEM:           "Télécharger PEM",
	ButtonRefresh:               "Rafraîchir",
	CacheInvalidateFailed:       "Échec du vidage du cache",
	CacheInvalidated:            "Cache vidé et données actualisées",
	CertificateInformationTitle: "Informations du certificat",
	ChartExpiryTimeline:         "Chronologie des expirations",
	ChartLegendExpired:          "Expiré",
	ChartLegendRevoked:          "Révoqué",
	ChartLegendValid:            "Valide",
	ChartStatusDistribution:     "Répartition par statut",
	ColumnActions:               "Actions",
	ColumnCommonName:            "Nom commun",
	ColumnCreatedAt:             "Créé le",
	ColumnExpiresAt:             "Expire le",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Statut",
	DashboardExpired:            "Expirés",
	DashboardExpiring:           "Expirant bientôt",
	DashboardRevoked:            "Révoqués",
	DashboardTotal:              "Total des certificats",
	DashboardValid:              "Valides",
	DaysRemaining:               "{{days}} jours restants",
	DaysRemainingShort:          "{{days}}j",
	DaysRemainingSingular:       "{{days}} jour restant",
	ExpiredSince:                "Expiré depuis le {{date}}",
	DeselectAll:                 "Tout désélectionner",
	DownloadPEMFailed:           "Échec du téléchargement du certificat PEM ({{status}})",
	DownloadPEMNetworkError:     "Erreur réseau lors du téléchargement du certificat PEM. Veuillez réessayer.",
	DownloadPEMSuccess:          "Certificat PEM téléchargé avec succès",
	DualStatusNote:              "{{count}} certificat(s) sont à la fois expirés et révoqués",
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
	LabelPEM:                    "Certificat PEM",
	LabelSerialNumber:           "Numéro de série",
	LabelSubject:                "Sujet",
	LabelUsage:                  "Usage",
	LegendExpiredText:           "Date d'expiration dépassée.",
	LegendExpiredTitle:          "Expiré",
	LegendRevokedText:           "Révoqué explicitement dans Vault.",
	LegendRevokedTitle:          "Révoqué",
	LegendValidText:             "Non expiré et non révoqué.",
	LegendValidTitle:            "Valide",
	LoadDetailsFailed:           "Échec du chargement des détails du certificat ({{status}})",
	LoadDetailsNetworkError:     "Erreur réseau lors du chargement des détails du certificat. Veuillez réessayer.",
	LoadFailed:                  "Échec du chargement des certificats ({{status}})",
	LoadNetworkError:            "Erreur réseau lors du chargement des certificats. Veuillez réessayer.",
	LoadSuccess:                 "Certificats chargés avec succès",
	LoadUnexpectedFormat:        "Format de réponse inattendu du serveur",
	LoadingDetails:              "Chargement des détails du certificat...",
	ModalDetailsTitle:           "Détails du certificat",
	MountSelectorTitle:          "Moteurs PKI",
	NoCertsExpiringSoon:         "Aucun certificat expirant bientôt",
	NoData:                      "Aucune donnée",
	NotificationCritical:        "{{count}} certificat(s) expirant dans {{threshold}} jours ou moins !",
	NotificationWarning:         "{{count}} certificat(s) expirant dans {{threshold}} jours ou moins",
	PaginationAll:               "Tous les résultats",
	PaginationInfo:              "Page {{current}} sur {{total}}",
	PaginationNext:              "Suivant",
	PaginationPageSizeLabel:     "Résultats par page",
	PaginationPrev:              "Précédent",
	SearchPlaceholder:           "CN ou SAN",
	SelectAll:                   "Tout sélectionner",
	StatusFilterAll:             "Tous",
	StatusFilterExpired:         "Expiré",
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
}

var spanishMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonClose:                 "Cerrar",
	ButtonDetails:               "Detalles",
	ButtonDownloadPEM:           "Descargar PEM",
	ButtonRefresh:               "Actualizar",
	CacheInvalidateFailed:       "Error al borrar el caché",
	CacheInvalidated:            "Caché borrado y datos actualizados",
	CertificateInformationTitle: "Información del certificado",
	ChartExpiryTimeline:         "Cronología de vencimientos",
	ChartLegendExpired:          "Vencido",
	ChartLegendRevoked:          "Revocado",
	ChartLegendValid:            "Válido",
	ChartStatusDistribution:     "Distribución por estado",
	ColumnActions:               "Acciones",
	ColumnCommonName:            "Nombre común",
	ColumnCreatedAt:             "Creado el",
	ColumnExpiresAt:             "Caduca el",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Estado",
	DashboardExpired:            "Caducados",
	DashboardExpiring:           "Caducando pronto",
	DashboardRevoked:            "Revocados",
	DashboardTotal:              "Total de certificados",
	DashboardValid:              "Válidos",
	DaysRemaining:               "{{days}} días restantes",
	DaysRemainingShort:          "{{days}}d",
	DaysRemainingSingular:       "{{days}} día restante",
	ExpiredSince:                "Vencido desde el {{date}}",
	DeselectAll:                 "Deseleccionar todo",
	DownloadPEMFailed:           "Error al descargar el certificado PEM ({{status}})",
	DownloadPEMNetworkError:     "Error de red al descargar el certificado PEM. Por favor intente nuevamente.",
	DownloadPEMSuccess:          "Certificado PEM descargado exitosamente",
	DualStatusNote:              "{{count}} certificado(s) están tanto caducados como revocados",
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
	LoadDetailsFailed:           "Error al cargar los detalles del certificado ({{status}})",
	LoadDetailsNetworkError:     "Error de red al cargar los detalles del certificado. Por favor intente nuevamente.",
	LoadFailed:                  "Error al cargar los certificados ({{status}})",
	LoadNetworkError:            "Error de red al cargar los certificados. Por favor intente nuevamente.",
	LoadSuccess:                 "Certificados cargados exitosamente",
	LoadUnexpectedFormat:        "Formato de respuesta inesperado del servidor",
	LoadingDetails:              "Cargando detalles del certificado...",
	ModalDetailsTitle:           "Detalles del certificado",
	MountSelectorTitle:          "Motores PKI",
	NoCertsExpiringSoon:         "Ningún certificado caducando pronto",
	NoData:                      "Sin datos",
	NotificationCritical:        "{{count}} certificado(s) caducando en {{threshold}} días o menos!",
	NotificationWarning:         "{{count}} certificado(s) caducando en {{threshold}} días o menos",
	PaginationAll:               "Todos los resultados",
	PaginationInfo:              "Página {{current}} de {{total}}",
	PaginationNext:              "Siguiente",
	PaginationPageSizeLabel:     "Resultados por página",
	PaginationPrev:              "Anterior",
	SearchPlaceholder:           "CN o SAN",
	SelectAll:                   "Seleccionar todo",
	StatusFilterAll:             "Todos",
	StatusFilterExpired:         "Caducado",
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
}

var germanMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonClose:                 "Schließen",
	ButtonDetails:               "Details",
	ButtonDownloadPEM:           "PEM herunterladen",
	ButtonRefresh:               "Aktualisieren",
	CacheInvalidateFailed:       "Cache konnte nicht geleert werden",
	CacheInvalidated:            "Cache geleert und Daten aktualisiert",
	CertificateInformationTitle: "Zertifikatsinformationen",
	ChartExpiryTimeline:         "Ablauf-Zeitachse",
	ChartLegendExpired:          "Abgelaufen",
	ChartLegendRevoked:          "Widerrufen",
	ChartLegendValid:            "Gültig",
	ChartStatusDistribution:     "Statusverteilung",
	ColumnActions:               "Aktionen",
	ColumnCommonName:            "Allgemeiner Name",
	ColumnCreatedAt:             "Erstellt am",
	ColumnExpiresAt:             "Gültig bis",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Status",
	DashboardExpired:            "Abgelaufen",
	DashboardExpiring:           "Laufen bald ab",
	DashboardRevoked:            "Widerrufen",
	DashboardTotal:              "Zertifikate gesamt",
	DashboardValid:              "Gültig",
	DaysRemaining:               "{{days}} verbleibende Tage",
	DaysRemainingShort:          "{{days}}T",
	DaysRemainingSingular:       "{{days}} verbleibender Tag",
	ExpiredSince:                "Abgelaufen seit dem {{date}}",
	DeselectAll:                 "Alle abwählen",
	DownloadPEMFailed:           "Zertifikat-PEM konnte nicht heruntergeladen werden ({{status}})",
	DownloadPEMNetworkError:     "Netzwerkfehler beim Herunterladen des Zertifikat-PEM. Bitte versuchen Sie es erneut.",
	DownloadPEMSuccess:          "Zertifikat-PEM erfolgreich heruntergeladen",
	DualStatusNote:              "{{count}} Zertifikat(e) sind sowohl abgelaufen als auch widerrufen",
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
	LoadDetailsFailed:           "Zertifikatsdetails konnten nicht geladen werden ({{status}})",
	LoadDetailsNetworkError:     "Netzwerkfehler beim Laden der Zertifikatsdetails. Bitte versuchen Sie es erneut.",
	LoadFailed:                  "Zertifikate konnten nicht geladen werden ({{status}})",
	LoadNetworkError:            "Netzwerkfehler beim Laden der Zertifikate. Bitte versuchen Sie es erneut.",
	LoadSuccess:                 "Zertifikate erfolgreich geladen",
	LoadUnexpectedFormat:        "Unerwartetes Antwortformat vom Server",
	LoadingDetails:              "Zertifikatsdetails werden geladen...",
	ModalDetailsTitle:           "Zertifikatsdetails",
	MountSelectorTitle:          "PKI-Motoren",
	NoCertsExpiringSoon:         "Keine Zertifikate, die bald ablaufen",
	NoData:                      "Keine Daten",
	NotificationCritical:        "{{count}} Zertifikat(e) laufen in {{threshold}} Tagen oder weniger ab!",
	NotificationWarning:         "{{count}} Zertifikat(e) laufen in {{threshold}} Tagen oder weniger ab",
	PaginationAll:               "Alle Ergebnisse",
	PaginationInfo:              "Seite {{current}} von {{total}}",
	PaginationNext:              "Weiter",
	PaginationPageSizeLabel:     "Ergebnisse pro Seite",
	PaginationPrev:              "Zurück",
	SearchPlaceholder:           "CN oder SAN",
	SelectAll:                   "Alle auswählen",
	StatusFilterAll:             "Alle",
	StatusFilterExpired:         "Abgelaufen",
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
}

var italianMessages = Messages{
	AppTitle:                    "VaultCertsViewer",
	ButtonClose:                 "Chiudi",
	ButtonDetails:               "Dettagli",
	ButtonDownloadPEM:           "Scarica PEM",
	ButtonRefresh:               "Aggiorna",
	CacheInvalidateFailed:       "Impossibile cancellare la cache",
	CacheInvalidated:            "Cache cancellata e dati aggiornati",
	CertificateInformationTitle: "Informazioni sul certificato",
	ChartExpiryTimeline:         "Cronologia scadenze",
	ChartLegendExpired:          "Scaduto",
	ChartLegendRevoked:          "Revocato",
	ChartLegendValid:            "Valido",
	ChartStatusDistribution:     "Distribuzione stato",
	ColumnActions:               "Azioni",
	ColumnCommonName:            "Nome comune",
	ColumnCreatedAt:             "Creato il",
	ColumnExpiresAt:             "Scade il",
	ColumnSAN:                   "SAN",
	ColumnStatus:                "Stato",
	DashboardExpired:            "Scaduti",
	DashboardExpiring:           "In scadenza",
	DashboardRevoked:            "Revocati",
	DashboardTotal:              "Certificati totali",
	DashboardValid:              "Validi",
	DaysRemaining:               "{{days}} giorni rimanenti",
	DaysRemainingShort:          "{{days}}g",
	DaysRemainingSingular:       "{{days}} giorno rimanente",
	ExpiredSince:                "Scaduto dal {{date}}",
	DeselectAll:                 "Deseleziona tutto",
	DownloadPEMFailed:           "Impossibile scaricare il certificato PEM ({{status}})",
	DownloadPEMNetworkError:     "Errore di rete durante il download del certificato PEM. Riprova.",
	DownloadPEMSuccess:          "Certificato PEM scaricato con successo",
	DualStatusNote:              "{{count}} certificato(i) sono sia scaduti che revocati",
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
	LoadDetailsFailed:           "Impossibile caricare i dettagli del certificato ({{status}})",
	LoadDetailsNetworkError:     "Errore di rete durante il caricamento dei dettagli del certificato. Riprova.",
	LoadFailed:                  "Impossibile caricare i certificati ({{status}})",
	LoadNetworkError:            "Errore di rete durante il caricamento dei certificati. Riprova.",
	LoadSuccess:                 "Certificati caricati correttamente",
	LoadUnexpectedFormat:        "Formato di risposta inatteso dal server",
	LoadingDetails:              "Caricamento dei dettagli del certificato...",
	ModalDetailsTitle:           "Dettagli del certificato",
	MountSelectorTitle:          "Motori PKI",
	NoCertsExpiringSoon:         "Nessun certificato in scadenza a breve",
	NoData:                      "Nessun dato",
	NotificationCritical:        "{{count}} certificato/i in scadenza entro {{threshold}} giorni o meno!",
	NotificationWarning:         "{{count}} certificato/i in scadenza entro {{threshold}} giorni o meno",
	PaginationAll:               "Tutti i risultati",
	PaginationInfo:              "Pagina {{current}} di {{total}}",
	PaginationNext:              "Successivo",
	PaginationPageSizeLabel:     "Risultati per pagina",
	PaginationPrev:              "Precedente",
	SearchPlaceholder:           "CN o SAN",
	SelectAll:                   "Seleziona tutto",
	StatusFilterAll:             "Tutti",
	StatusFilterExpired:         "Scaduto",
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
	code := strings.ToLower(strings.TrimSpace(value))
	if code == string(LanguageEnglish) {
		return LanguageEnglish, true
	}
	if code == string(LanguageFrench) {
		return LanguageFrench, true
	}
	if code == string(LanguageSpanish) {
		return LanguageSpanish, true
	}
	if code == string(LanguageGerman) {
		return LanguageGerman, true
	}
	if code == string(LanguageItalian) {
		return LanguageItalian, true
	}
	return "", false
}

// FromAcceptLanguage inspects an Accept-Language header and picks a best-effort language.
func FromAcceptLanguage(headerValue string) Language {
	lowered := strings.ToLower(headerValue)
	if lowered == "" {
		return LanguageEnglish
	}
	if strings.Contains(lowered, "fr") {
		return LanguageFrench
	}
	if strings.Contains(lowered, "es") {
		return LanguageSpanish
	}
	if strings.Contains(lowered, "de") {
		return LanguageGerman
	}
	if strings.Contains(lowered, "it") {
		return LanguageItalian
	}
	return LanguageEnglish
}
