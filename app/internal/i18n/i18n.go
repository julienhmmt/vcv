package i18n

import "strings"

// Language represents a supported UI language.
type Language string

const (
	LanguageGerman  Language = "de"
	LanguageEnglish Language = "en"
	LanguageSpanish Language = "es"
	LanguageFrench  Language = "fr"
	LanguageItalian Language = "it"
)

// Messages contains all translatable UI strings used by the web interface.
type Messages struct {
	AppSubtitle             string `json:"appSubtitle"`
	AppTitle                string `json:"appTitle"`
	ButtonClose             string `json:"buttonClose"`
	ButtonDetails           string `json:"buttonDetails"`
	ButtonDownloadCRL       string `json:"buttonDownloadCRL"`
	ButtonDownloadPEM       string `json:"buttonDownloadPEM"`
	ButtonRefresh           string `json:"buttonRefresh"`
	ButtonRotateCRL         string `json:"buttonRotateCRL"`
	CacheInvalidateFailed   string `json:"cacheInvalidateFailed"`
	CacheInvalidated        string `json:"cacheInvalidated"`
	ChartExpiryTimeline     string `json:"chartExpiryTimeline"`
	ChartLegendExpired      string `json:"chartLegendExpired"`
	ChartLegendRevoked      string `json:"chartLegendRevoked"`
	ChartLegendValid        string `json:"chartLegendValid"`
	ChartStatusDistribution string `json:"chartStatusDistribution"`
	ColumnActions           string `json:"columnActions"`
	ColumnCommonName        string `json:"columnCommonName"`
	ColumnCreatedAt         string `json:"columnCreatedAt"`
	ColumnExpiresAt         string `json:"columnExpiresAt"`
	ColumnSAN               string `json:"columnSan"`
	ColumnStatus            string `json:"columnStatus"`
	DashboardExpired        string `json:"dashboardExpired"`
	DashboardExpiring       string `json:"dashboardExpiring"`
	DashboardRevoked        string `json:"dashboardRevoked"`
	DashboardTotal          string `json:"dashboardTotal"`
	DashboardValid          string `json:"dashboardValid"`
	DaysRemaining           string `json:"daysRemaining"`
	DaysRemainingShort      string `json:"daysRemainingShort"`
	DaysRemainingSingular   string `json:"daysRemainingSingular"`
	DownloadCRLFailed       string `json:"downloadCRLFailed"`
	DownloadCRLNetworkError string `json:"downloadCRLNetworkError"`
	DownloadPEMFailed       string `json:"downloadPEMFailed"`
	DownloadPEMNetworkError string `json:"downloadPEMNetworkError"`
	DownloadPEMSuccess      string `json:"downloadPEMSuccess"`
	DualStatusNote          string `json:"dualStatusNote"`
	ExpiryFilter30Days      string `json:"expiryFilter30Days"`
	ExpiryFilter7Days       string `json:"expiryFilter7Days"`
	ExpiryFilter90Days      string `json:"expiryFilter90Days"`
	ExpiryFilterAll         string `json:"expiryFilterAll"`
	FooterVaultConnected    string `json:"footerVaultConnected"`
	FooterVaultDisconnected string `json:"footerVaultDisconnected"`
	FooterVaultLoading      string `json:"footerVaultLoading"`
	FooterVersion           string `json:"footerVersion"`
	LabelFingerprintSHA1    string `json:"labelFingerprintSHA1"`
	LabelFingerprintSHA256  string `json:"labelFingerprintSHA256"`
	LabelIssuer             string `json:"labelIssuer"`
	LabelKeyAlgorithm       string `json:"labelKeyAlgorithm"`
	LabelPEM                string `json:"labelPem"`
	LabelSerialNumber       string `json:"labelSerialNumber"`
	LabelSubject            string `json:"labelSubject"`
	LabelUsage              string `json:"labelUsage"`
	LegendExpiredText       string `json:"legendExpiredText"`
	LegendExpiredTitle      string `json:"legendExpiredTitle"`
	LegendRevokedText       string `json:"legendRevokedText"`
	LegendRevokedTitle      string `json:"legendRevokedTitle"`
	LegendValidText         string `json:"legendValidText"`
	LegendValidTitle        string `json:"legendValidTitle"`
	LoadDetailsFailed       string `json:"loadDetailsFailed"`
	LoadDetailsNetworkError string `json:"loadDetailsNetworkError"`
	LoadFailed              string `json:"loadFailed"`
	LoadNetworkError        string `json:"loadNetworkError"`
	LoadSuccess             string `json:"loadSuccess"`
	LoadUnexpectedFormat    string `json:"loadUnexpectedFormat"`
	LoadingDetails          string `json:"loadingDetails"`
	ModalDetailsTitle       string `json:"modalDetailsTitle"`
	NoCertsExpiringSoon     string `json:"noCertsExpiringSoon"`
	NoData                  string `json:"noData"`
	NotificationCritical    string `json:"notificationCritical"`
	NotificationWarning     string `json:"notificationWarning"`
	PaginationAll           string `json:"paginationAll"`
	PaginationInfo          string `json:"paginationInfo"`
	PaginationNext          string `json:"paginationNext"`
	PaginationPageSizeLabel string `json:"paginationPageSizeLabel"`
	PaginationPrev          string `json:"paginationPrev"`
	RotateCRLFailed         string `json:"rotateCRLFailed"`
	RotateCRLNetworkError   string `json:"rotateCRLNetworkError"`
	RotateCRLSuccess        string `json:"rotateCRLSuccess"`
	SearchPlaceholder       string `json:"searchPlaceholder"`
	StatusFilterAll         string `json:"statusFilterAll"`
	StatusFilterExpired     string `json:"statusFilterExpired"`
	StatusFilterRevoked     string `json:"statusFilterRevoked"`
	StatusFilterValid       string `json:"statusFilterValid"`
	StatusLabelExpired      string `json:"statusLabelExpired"`
	StatusLabelRevoked      string `json:"statusLabelRevoked"`
	StatusLabelValid        string `json:"statusLabelValid"`
	SummaryAll              string `json:"summaryAll"`
	SummaryNoCertificates   string `json:"summaryNoCertificates"`
	SummarySome             string `json:"summarySome"`
}

// Response is the payload returned by the /api/i18n endpoint.
type Response struct {
	Language Language `json:"language"`
	Messages Messages `json:"messages"`
}

var englishMessages = Messages{
	AppSubtitle:             "Certificates from the configured Vault PKI mount",
	AppTitle:                "VaultCertsViewer",
	ButtonClose:             "Close",
	ButtonDetails:           "Details",
	ButtonDownloadCRL:       "Download CRL",
	ButtonDownloadPEM:       "Download PEM",
	ButtonRefresh:           "Refresh",
	ButtonRotateCRL:         "Rotate CRL",
	CacheInvalidateFailed:   "Failed to clear cache",
	CacheInvalidated:        "Cache cleared and data refreshed",
	ChartExpiryTimeline:     "Expiration Timeline",
	ChartLegendExpired:      "Expired",
	ChartLegendRevoked:      "Revoked",
	ChartLegendValid:        "Valid",
	ChartStatusDistribution: "Status Distribution",
	ColumnActions:           "Actions",
	ColumnCommonName:        "Common name",
	ColumnCreatedAt:         "Created at",
	ColumnExpiresAt:         "Expires at",
	ColumnSAN:               "SAN",
	ColumnStatus:            "Status",
	DashboardExpired:        "Expired",
	DashboardExpiring:       "Expiring Soon",
	DashboardRevoked:        "Revoked",
	DashboardTotal:          "Total Certificates",
	DashboardValid:          "Valid",
	DaysRemaining:           "{{days}} days remaining",
	DaysRemainingShort:      "{{days}}d",
	DaysRemainingSingular:   "{{days}} day remaining",
	DownloadCRLFailed:       "Failed to download CRL ({{status}})",
	DownloadCRLNetworkError: "Network error downloading CRL. Please try again.",
	DownloadPEMFailed:       "Failed to download certificate PEM ({{status}})",
	DownloadPEMNetworkError: "Network error downloading certificate PEM. Please try again.",
	DownloadPEMSuccess:      "Certificate PEM downloaded successfully",
	DualStatusNote:          "{{count}} certificate(s) are both expired and revoked",
	ExpiryFilter30Days:      "≤ 30 days",
	ExpiryFilter7Days:       "≤ 7 days",
	ExpiryFilter90Days:      "≤ 90 days",
	ExpiryFilterAll:         "All dates",
	FooterVaultConnected:    "Vault: connected",
	FooterVaultDisconnected: "Vault: disconnected",
	FooterVaultLoading:      "Vault: …",
	FooterVersion:           "VCV v{{version}}",
	LabelFingerprintSHA1:    "SHA-1 Fingerprint",
	LabelFingerprintSHA256:  "SHA-256 Fingerprint",
	LabelIssuer:             "Issuer",
	LabelKeyAlgorithm:       "Key Algorithm",
	LabelPEM:                "PEM Certificate",
	LabelSerialNumber:       "Serial Number",
	LabelSubject:            "Subject",
	LabelUsage:              "Usage",
	LegendExpiredText:       "Past the expiration date.",
	LegendExpiredTitle:      "Expired",
	LegendRevokedText:       "Explicitly revoked in Vault.",
	LegendRevokedTitle:      "Revoked",
	LegendValidText:         "Not expired and not revoked.",
	LegendValidTitle:        "Valid",
	LoadDetailsFailed:       "Failed to load certificate details ({{status}})",
	LoadDetailsNetworkError: "Network error loading certificate details. Please try again.",
	LoadFailed:              "Failed to load certificates ({{status}})",
	LoadNetworkError:        "Network error loading certificates. Please try again.",
	LoadSuccess:             "Certificates loaded successfully",
	LoadUnexpectedFormat:    "Unexpected response format from server",
	LoadingDetails:          "Loading certificate details...",
	ModalDetailsTitle:       "Certificate Details",
	NoCertsExpiringSoon:     "No certificates expiring soon",
	NoData:                  "No data",
	NotificationCritical:    "{{count}} certificate(s) expiring within 7 days or less!",
	NotificationWarning:     "{{count}} certificate(s) expiring within 30 days or less",
	PaginationAll:           "All results",
	PaginationInfo:          "Page {{current}} of {{total}}",
	PaginationNext:          "Next",
	PaginationPageSizeLabel: "Results per page",
	PaginationPrev:          "Previous",
	RotateCRLFailed:         "Failed to rotate CRL ({{status}})",
	RotateCRLNetworkError:   "Network error rotating CRL. Please try again.",
	RotateCRLSuccess:        "CRL rotated successfully",
	SearchPlaceholder:       "CN or SAN",
	StatusFilterAll:         "All",
	StatusFilterExpired:     "Expired",
	StatusFilterRevoked:     "Revoked",
	StatusFilterValid:       "Valid",
	StatusLabelExpired:      "Expired",
	StatusLabelRevoked:      "Revoked",
	StatusLabelValid:        "Valid",
	SummaryAll:              "{{total}} certificates",
	SummaryNoCertificates:   "No certificates.",
	SummarySome:             "{{visible}} of {{total}} certificates shown",
}

var frenchMessages = Messages{
	AppSubtitle:             "Certificats du PKI Vault configuré",
	AppTitle:                "VaultCertsViewer",
	ButtonClose:             "Fermer",
	ButtonDetails:           "Détails",
	ButtonDownloadCRL:       "Télécharger la CRL",
	ButtonDownloadPEM:       "Télécharger PEM",
	ButtonRefresh:           "Actualiser",
	ButtonRotateCRL:         "Générer la CRL",
	CacheInvalidateFailed:   "Échec du vidage du cache",
	CacheInvalidated:        "Cache vidé et données actualisées",
	ChartExpiryTimeline:     "Chronologie des expirations",
	ChartLegendExpired:      "Expiré",
	ChartLegendRevoked:      "Révoqué",
	ChartLegendValid:        "Valide",
	ChartStatusDistribution: "Répartition par statut",
	ColumnActions:           "Actions",
	ColumnCommonName:        "Nom commun",
	ColumnCreatedAt:         "Créé le",
	ColumnExpiresAt:         "Expire le",
	ColumnSAN:               "SAN",
	ColumnStatus:            "Statut",
	DashboardExpired:        "Expirés",
	DashboardExpiring:       "Expirant bientôt",
	DashboardRevoked:        "Révoqués",
	DashboardTotal:          "Total des certificats",
	DashboardValid:          "Valides",
	DaysRemaining:           "{{days}} jours restants",
	DaysRemainingShort:      "{{days}}j",
	DaysRemainingSingular:   "{{days}} jour restant",
	DownloadCRLFailed:       "Échec du téléchargement de la CRL ({{status}})",
	DownloadCRLNetworkError: "Erreur réseau lors du téléchargement de la CRL. Veuillez réessayer.",
	DownloadPEMFailed:       "Échec du téléchargement du certificat PEM ({{status}})",
	DownloadPEMNetworkError: "Erreur réseau lors du téléchargement du certificat PEM. Veuillez réessayer.",
	DownloadPEMSuccess:      "Certificat PEM téléchargé avec succès",
	DualStatusNote:          "{{count}} certificat(s) sont à la fois expirés et révoqués",
	ExpiryFilter30Days:      "≤ 30 jours",
	ExpiryFilter7Days:       "≤ 7 jours",
	ExpiryFilter90Days:      "≤ 90 jours",
	ExpiryFilterAll:         "Toutes les dates",
	FooterVaultConnected:    "Vault : connecté",
	FooterVaultDisconnected: "Vault : déconnecté",
	FooterVaultLoading:      "Vault : …",
	FooterVersion:           "VCV v{{version}}",
	LabelFingerprintSHA1:    "Empreinte SHA-1",
	LabelFingerprintSHA256:  "Empreinte SHA-256",
	LabelIssuer:             "Émetteur",
	LabelKeyAlgorithm:       "Algorithme de clé",
	LabelPEM:                "Certificat PEM",
	LabelSerialNumber:       "Numéro de série",
	LabelSubject:            "Sujet",
	LabelUsage:              "Usage",
	LegendExpiredText:       "Date d'expiration dépassée.",
	LegendExpiredTitle:      "Expiré",
	LegendRevokedText:       "Révoqué explicitement dans Vault.",
	LegendRevokedTitle:      "Révoqué",
	LegendValidText:         "Non expiré et non révoqué.",
	LegendValidTitle:        "Valide",
	LoadDetailsFailed:       "Échec du chargement des détails du certificat ({{status}})",
	LoadDetailsNetworkError: "Erreur réseau lors du chargement des détails du certificat. Veuillez réessayer.",
	LoadFailed:              "Échec du chargement des certificats ({{status}})",
	LoadNetworkError:        "Erreur réseau lors du chargement des certificats. Veuillez réessayer.",
	LoadSuccess:             "Certificats chargés avec succès",
	LoadUnexpectedFormat:    "Format de réponse inattendu du serveur",
	LoadingDetails:          "Chargement des détails du certificat...",
	ModalDetailsTitle:       "Détails du certificat",
	NoCertsExpiringSoon:     "Aucun certificat expirant bientôt",
	NoData:                  "Aucune donnée",
	NotificationCritical:    "{{count}} certificat(s) expirant dans 7 jours ou moins !",
	NotificationWarning:     "{{count}} certificat(s) expirant dans 30 jours ou moins",
	PaginationAll:           "Tous les résultats",
	PaginationInfo:          "Page {{current}} sur {{total}}",
	PaginationNext:          "Suivant",
	PaginationPageSizeLabel: "Résultats par page",
	PaginationPrev:          "Précédent",
	RotateCRLFailed:         "Échec de la génération de la CRL ({{status}})",
	RotateCRLNetworkError:   "Erreur réseau lors de la génération de la CRL. Veuillez réessayer.",
	RotateCRLSuccess:        "CRL générée avec succès",
	SearchPlaceholder:       "CN ou SAN",
	StatusFilterAll:         "Tous",
	StatusFilterExpired:     "Expiré",
	StatusFilterRevoked:     "Révoqué",
	StatusFilterValid:       "Valide",
	StatusLabelExpired:      "Expiré",
	StatusLabelRevoked:      "Révoqué",
	StatusLabelValid:        "Valide",
	SummaryAll:              "{{total}} certificats",
	SummaryNoCertificates:   "Aucun certificat.",
	SummarySome:             "{{visible}} sur {{total}} certificats affichés",
}

var spanishMessages = Messages{
	AppSubtitle:             "Certificados del PKI de Vault configurado",
	AppTitle:                "VaultCertsViewer",
	ButtonClose:             "Cerrar",
	ButtonDetails:           "Detalles",
	ButtonDownloadCRL:       "Descargar CRL",
	ButtonDownloadPEM:       "Descargar PEM",
	ButtonRefresh:           "Actualizar",
	ButtonRotateCRL:         "Rotar CRL",
	CacheInvalidateFailed:   "Error al borrar el caché",
	CacheInvalidated:        "Caché borrado y datos actualizados",
	ChartExpiryTimeline:     "Línea de tiempo de caducidad",
	ChartLegendExpired:      "Caducado",
	ChartLegendRevoked:      "Revocado",
	ChartLegendValid:        "Válido",
	ChartStatusDistribution: "Distribución por estado",
	ColumnActions:           "Acciones",
	ColumnCommonName:        "Nombre común",
	ColumnCreatedAt:         "Creado el",
	ColumnExpiresAt:         "Caduca el",
	ColumnSAN:               "SAN",
	ColumnStatus:            "Estado",
	DashboardExpired:        "Caducados",
	DashboardExpiring:       "Caducando pronto",
	DashboardRevoked:        "Revocados",
	DashboardTotal:          "Total de certificados",
	DashboardValid:          "Válidos",
	DaysRemaining:           "{{days}} días restantes",
	DaysRemainingShort:      "{{days}}d",
	DaysRemainingSingular:   "{{days}} día restante",
	DownloadCRLFailed:       "Error al descargar la CRL ({{status}})",
	DownloadCRLNetworkError: "Error de red al descargar la CRL. Por favor intente nuevamente.",
	DownloadPEMFailed:       "Error al descargar el certificado PEM ({{status}})",
	DownloadPEMNetworkError: "Error de red al descargar el certificado PEM. Por favor intente nuevamente.",
	DownloadPEMSuccess:      "Certificado PEM descargado exitosamente",
	DualStatusNote:          "{{count}} certificado(s) están tanto caducados como revocados",
	ExpiryFilter30Days:      "≤ 30 días",
	ExpiryFilter7Days:       "≤ 7 días",
	ExpiryFilter90Days:      "≤ 90 días",
	ExpiryFilterAll:         "Todas las fechas",
	FooterVaultConnected:    "Vault: conectado",
	FooterVaultDisconnected: "Vault: desconectado",
	FooterVaultLoading:      "Vault: …",
	FooterVersion:           "VCV v{{version}}",
	LabelFingerprintSHA1:    "Huella SHA-1",
	LabelFingerprintSHA256:  "Huella SHA-256",
	LabelIssuer:             "Emisor",
	LabelKeyAlgorithm:       "Algoritmo de clave",
	LabelPEM:                "Certificado PEM",
	LabelSerialNumber:       "Número de serie",
	LabelSubject:            "Sujeto",
	LabelUsage:              "Uso",
	LegendExpiredText:       "Fecha de caducidad superada.",
	LegendExpiredTitle:      "Caducado",
	LegendRevokedText:       "Revocado explícitamente en Vault.",
	LegendRevokedTitle:      "Revocado",
	LegendValidText:         "No caducado y no revocado.",
	LegendValidTitle:        "Válido",
	LoadDetailsFailed:       "Error al cargar los detalles del certificado ({{status}})",
	LoadDetailsNetworkError: "Error de red al cargar los detalles del certificado. Por favor intente nuevamente.",
	LoadFailed:              "Error al cargar los certificados ({{status}})",
	LoadNetworkError:        "Error de red al cargar los certificados. Por favor intente nuevamente.",
	LoadSuccess:             "Certificados cargados exitosamente",
	LoadUnexpectedFormat:    "Formato de respuesta inesperado del servidor",
	LoadingDetails:          "Cargando detalles del certificado...",
	ModalDetailsTitle:       "Detalles del certificado",
	NoCertsExpiringSoon:     "Ningún certificado caducando pronto",
	NoData:                  "Sin datos",
	NotificationCritical:    "{{count}} certificado(s) caducando en 7 días o menos!",
	NotificationWarning:     "{{count}} certificado(s) caducando en 30 días o menos",
	PaginationAll:           "Todos los resultados",
	PaginationInfo:          "Página {{current}} de {{total}}",
	PaginationNext:          "Siguiente",
	PaginationPageSizeLabel: "Resultados por página",
	PaginationPrev:          "Anterior",
	RotateCRLFailed:         "Error al rotar la CRL ({{status}})",
	RotateCRLNetworkError:   "Error de red al rotar la CRL. Por favor intente nuevamente.",
	RotateCRLSuccess:        "CRL rotada exitosamente",
	SearchPlaceholder:       "CN o SAN",
	StatusFilterAll:         "Todos",
	StatusFilterExpired:     "Caducado",
	StatusFilterRevoked:     "Revocado",
	StatusFilterValid:       "Válido",
	StatusLabelExpired:      "Caducado",
	StatusLabelRevoked:      "Revocado",
	StatusLabelValid:        "Válido",
	SummaryAll:              "{{total}} certificados",
	SummaryNoCertificates:   "Ningún certificado.",
	SummarySome:             "{{visible}} de {{total}} certificados mostrados",
}

var germanMessages = Messages{
	AppSubtitle:             "Zertifikate aus dem konfigurierten Vault-PKI-Mount",
	AppTitle:                "VaultCertsViewer",
	ButtonClose:             "Schließen",
	ButtonDetails:           "Details",
	ButtonDownloadCRL:       "CRL herunterladen",
	ButtonDownloadPEM:       "PEM herunterladen",
	ButtonRefresh:           "Aktualisieren",
	ButtonRotateCRL:         "CRL rotieren",
	CacheInvalidateFailed:   "Cache konnte nicht geleert werden",
	CacheInvalidated:        "Cache geleert und Daten aktualisiert",
	ChartExpiryTimeline:     "Ablaufzeitachse",
	ChartLegendExpired:      "Abgelaufen",
	ChartLegendRevoked:      "Widerrufen",
	ChartLegendValid:        "Gültig",
	ChartStatusDistribution: "Statusverteilung",
	ColumnActions:           "Aktionen",
	ColumnCommonName:        "Allgemeiner Name",
	ColumnCreatedAt:         "Erstellt am",
	ColumnExpiresAt:         "Gültig bis",
	ColumnSAN:               "SAN",
	ColumnStatus:            "Status",
	DashboardExpired:        "Abgelaufen",
	DashboardExpiring:       "Laufen bald ab",
	DashboardRevoked:        "Widerrufen",
	DashboardTotal:          "Zertifikate gesamt",
	DashboardValid:          "Gültig",
	DaysRemaining:           "{{days}} verbleibende Tage",
	DaysRemainingShort:      "{{days}}T",
	DaysRemainingSingular:   "{{days}} verbleibender Tag",
	DownloadCRLFailed:       "CRL konnte nicht heruntergeladen werden ({{status}})",
	DownloadCRLNetworkError: "Netzwerkfehler beim Herunterladen der CRL. Bitte versuchen Sie es erneut.",
	DownloadPEMFailed:       "Zertifikat-PEM konnte nicht heruntergeladen werden ({{status}})",
	DownloadPEMNetworkError: "Netzwerkfehler beim Herunterladen des Zertifikat-PEM. Bitte versuchen Sie es erneut.",
	DownloadPEMSuccess:      "Zertifikat-PEM erfolgreich heruntergeladen",
	DualStatusNote:          "{{count}} Zertifikat(e) sind sowohl abgelaufen als auch widerrufen",
	ExpiryFilter30Days:      "≤ 30 Tage",
	ExpiryFilter7Days:       "≤ 7 Tage",
	ExpiryFilter90Days:      "≤ 90 Tage",
	ExpiryFilterAll:         "Alle Daten",
	FooterVaultConnected:    "Vault: verbunden",
	FooterVaultDisconnected: "Vault: getrennt",
	FooterVaultLoading:      "Vault: …",
	FooterVersion:           "VCV v{{version}}",
	LabelFingerprintSHA1:    "SHA-1-Fingerabdruck",
	LabelFingerprintSHA256:  "SHA-256-Fingerabdruck",
	LabelIssuer:             "Aussteller",
	LabelKeyAlgorithm:       "Schlüsselalgorithmus",
	LabelPEM:                "PEM-Zertifikat",
	LabelSerialNumber:       "Seriennummer",
	LabelSubject:            "Betreff",
	LabelUsage:              "Verwendung",
	LegendExpiredText:       "Ablaufdatum überschritten.",
	LegendExpiredTitle:      "Abgelaufen",
	LegendRevokedText:       "Explizit in Vault widerrufen.",
	LegendRevokedTitle:      "Widerrufen",
	LegendValidText:         "Nicht abgelaufen und nicht widerrufen.",
	LegendValidTitle:        "Gültig",
	LoadDetailsFailed:       "Zertifikatsdetails konnten nicht geladen werden ({{status}})",
	LoadDetailsNetworkError: "Netzwerkfehler beim Laden der Zertifikatsdetails. Bitte versuchen Sie es erneut.",
	LoadFailed:              "Zertifikate konnten nicht geladen werden ({{status}})",
	LoadNetworkError:        "Netzwerkfehler beim Laden der Zertifikate. Bitte versuchen Sie es erneut.",
	LoadSuccess:             "Zertifikate erfolgreich geladen",
	LoadUnexpectedFormat:    "Unerwartetes Antwortformat vom Server",
	LoadingDetails:          "Zertifikatsdetails werden geladen...",
	ModalDetailsTitle:       "Zertifikatsdetails",
	NoCertsExpiringSoon:     "Keine Zertifikate, die bald ablaufen",
	NoData:                  "Keine Daten",
	NotificationCritical:    "{{count}} Zertifikat(e) laufen in 7 Tagen oder weniger ab!",
	NotificationWarning:     "{{count}} Zertifikat(e) laufen in 30 Tagen oder weniger ab",
	PaginationAll:           "Alle Ergebnisse",
	PaginationInfo:          "Seite {{current}} von {{total}}",
	PaginationNext:          "Weiter",
	PaginationPageSizeLabel: "Ergebnisse pro Seite",
	PaginationPrev:          "Zurück",
	RotateCRLFailed:         "CRL konnte nicht rotiert werden ({{status}})",
	RotateCRLNetworkError:   "Netzwerkfehler beim Rotieren der CRL. Bitte versuchen Sie es erneut.",
	RotateCRLSuccess:        "CRL erfolgreich rotiert",
	SearchPlaceholder:       "CN oder SAN",
	StatusFilterAll:         "Alle",
	StatusFilterExpired:     "Abgelaufen",
	StatusFilterRevoked:     "Widerrufen",
	StatusFilterValid:       "Gültig",
	StatusLabelExpired:      "Abgelaufen",
	StatusLabelRevoked:      "Widerrufen",
	StatusLabelValid:        "Gültig",
	SummaryAll:              "{{total}} Zertifikate",
	SummaryNoCertificates:   "Keine Zertifikate.",
	SummarySome:             "{{visible}} von {{total}} Zertifikaten angezeigt",
}

var italianMessages = Messages{
	AppSubtitle:             "Certificati dal mount PKI di Vault configurato",
	AppTitle:                "VaultCertsViewer",
	ButtonClose:             "Chiudi",
	ButtonDetails:           "Dettagli",
	ButtonDownloadCRL:       "Scarica CRL",
	ButtonDownloadPEM:       "Scarica PEM",
	ButtonRefresh:           "Aggiorna",
	ButtonRotateCRL:         "Ruota CRL",
	CacheInvalidateFailed:   "Impossibile svuotare la cache",
	CacheInvalidated:        "Cache svuotata e dati aggiornati",
	ChartExpiryTimeline:     "Cronologia delle scadenze",
	ChartLegendExpired:      "Scaduto",
	ChartLegendRevoked:      "Revocato",
	ChartLegendValid:        "Valido",
	ChartStatusDistribution: "Distribuzione degli stati",
	ColumnActions:           "Azioni",
	ColumnCommonName:        "Nome comune",
	ColumnCreatedAt:         "Creato il",
	ColumnExpiresAt:         "Scade il",
	ColumnSAN:               "SAN",
	ColumnStatus:            "Stato",
	DashboardExpired:        "Scaduti",
	DashboardExpiring:       "In scadenza",
	DashboardRevoked:        "Revocati",
	DashboardTotal:          "Certificati totali",
	DashboardValid:          "Validi",
	DaysRemaining:           "{{days}} giorni rimanenti",
	DaysRemainingShort:      "{{days}}g",
	DaysRemainingSingular:   "{{days}} giorno rimanente",
	DownloadCRLFailed:       "Impossibile scaricare la CRL ({{status}})",
	DownloadCRLNetworkError: "Errore di rete durante il download della CRL. Riprova.",
	DownloadPEMFailed:       "Impossibile scaricare il certificato PEM ({{status}})",
	DownloadPEMNetworkError: "Errore di rete durante il download del certificato PEM. Riprova.",
	DownloadPEMSuccess:      "Certificato PEM scaricato correttamente",
	DualStatusNote:          "{{count}} certificato(i) sono sia scaduti che revocati",
	ExpiryFilter30Days:      "≤ 30 giorni",
	ExpiryFilter7Days:       "≤ 7 giorni",
	ExpiryFilter90Days:      "≤ 90 giorni",
	ExpiryFilterAll:         "Tutte le date",
	FooterVaultConnected:    "Vault: connesso",
	FooterVaultDisconnected: "Vault: disconnesso",
	FooterVaultLoading:      "Vault: …",
	FooterVersion:           "VCV v{{version}}",
	LabelFingerprintSHA1:    "Impronta SHA-1",
	LabelFingerprintSHA256:  "Impronta SHA-256",
	LabelIssuer:             "Emittente",
	LabelKeyAlgorithm:       "Algoritmo della chiave",
	LabelPEM:                "Certificato PEM",
	LabelSerialNumber:       "Numero di serie",
	LabelSubject:            "Soggetto",
	LabelUsage:              "Utilizzo",
	LegendExpiredText:       "Data di scadenza superata.",
	LegendExpiredTitle:      "Scaduto",
	LegendRevokedText:       "Revocato esplicitamente in Vault.",
	LegendRevokedTitle:      "Revocato",
	LegendValidText:         "Non scaduto e non revocato.",
	LegendValidTitle:        "Valido",
	LoadDetailsFailed:       "Impossibile caricare i dettagli del certificato ({{status}})",
	LoadDetailsNetworkError: "Errore di rete durante il caricamento dei dettagli del certificato. Riprova.",
	LoadFailed:              "Impossibile caricare i certificati ({{status}})",
	LoadNetworkError:        "Errore di rete durante il caricamento dei certificati. Riprova.",
	LoadSuccess:             "Certificati caricati correttamente",
	LoadUnexpectedFormat:    "Formato di risposta inatteso dal server",
	LoadingDetails:          "Caricamento dei dettagli del certificato...",
	ModalDetailsTitle:       "Dettagli del certificato",
	NoCertsExpiringSoon:     "Nessun certificato in scadenza a breve",
	NoData:                  "Nessun dato",
	NotificationCritical:    "{{count}} certificato/i in scadenza entro 7 giorni o meno!",
	NotificationWarning:     "{{count}} certificato/i in scadenza entro 30 giorni o meno",
	RotateCRLFailed:         "Impossibile ruotare la CRL ({{status}})",
	RotateCRLNetworkError:   "Errore di rete durante la rotazione della CRL. Riprova.",
	RotateCRLSuccess:        "CRL ruotata correttamente",
	SearchPlaceholder:       "CN o SAN",
	StatusFilterAll:         "Tutti",
	StatusFilterExpired:     "Scaduto",
	StatusFilterRevoked:     "Revocato",
	StatusFilterValid:       "Valido",
	StatusLabelExpired:      "Scaduto",
	StatusLabelRevoked:      "Revocato",
	StatusLabelValid:        "Valido",
	SummaryAll:              "{{total}} certificati",
	SummaryNoCertificates:   "Nessun certificato.",
	SummarySome:             "{{visible}} di {{total}} certificati mostrati",
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
