# Benutzerhandbuch - VaultCertsViewer (VCV)

## Was ist VCV?

VaultCertsViewer (VCV) ist eine leichtgewichtige Web-Schnittstelle zur Visualisierung und Überwachung von Zertifikaten, die von HashiCorp Vault (oder OpenBao) PKI-Engines verwaltet werden. Es bietet ein zentralisiertes Dashboard zur Verfolgung von Ablaufdaten, Status (gültig, abgelaufen, widerrufen) und technischen Details Ihrer Zertifikate über mehrere Vault-Instanzen und PKI-Mounts hinweg.

## Funktionen

- **Multi-Vault-Unterstützung**: Gleichzeitige Verbindung zu einer oder mehreren Vault-Instanzen.
- **PKI-Engine-Auswahl**: Filtern Sie Zertifikate nach Vault-Instanz und PKI-Mount über ein interaktives Modal mit Suche, Auswahl/Abwahl pro Vault oder global.
- **Dashboard**: Ringdiagramm mit Echtzeit-Statistiken zur Zertifikatsstatusverteilung (gültig, ablaufend, abgelaufen, widerrufen). Klicken Sie auf ein Segment oder eine Statuskarte, um die Tabelle sofort zu filtern.
- **Suchen & Filtern**: Suche nach Common Name (CN) oder Subject Alternative Names (SAN). Filtern nach Status über die Dashboard-Karten.
- **Sortierung**: Sortieren Sie die Zertifikatstabelle nach Common Name, Erstellungsdatum, Ablaufdatum, Vault-Name oder PKI-Mount. Klicken Sie auf eine Spaltenüberschrift, um zwischen auf-/absteigender Sortierung umzuschalten.
- **Paginierung**: Serverseitige Paginierung mit konfigurierbaren Seitengrößen (25, 50, 100 oder Alle).
- **Detailansicht**: Zugriff auf vollständige Zertifikatsmetadaten in einem übersichtlichen Modal: Identität (Betreff, Aussteller, Seriennummer, SANs), Gültigkeitsdaten mit Ablaufstatus und technische Details (Schlüsselalgorithmus, Schlüsselverwendung, Fingerabdrücke SHA-1/SHA-256).
- **PEM-Download**: Laden Sie Zertifikats-PEM-Dateien direkt aus der Tabelle herunter.
- **Vault-Status**: Ein Indikator im Header (Schild-Symbol mit Statuspunkt) zeigt den Live-Verbindungsstatus Ihrer Vault-Instanzen. Klicken Sie darauf, um ein detailliertes Status-Modal mit Gesundheitsinformationen pro Vault und einer Aktualisierungsschaltfläche zu öffnen.
- **Ablaufbenachrichtigungen**: Ein Banner oben auf der Seite warnt vor Zertifikaten, die innerhalb der konfigurierten Schwellenwerte ablaufen (kritisch / Warnung).
- **Toast-Benachrichtigungen**: Echtzeit-Toast-Nachrichten bei Vault-Verbindungsänderungen, Fehlern und Benutzer-Feedback.
- **Cache & Aktualisierung**: Zertifikatsdaten werden serverseitig gecacht (15 Min. TTL). Verwenden Sie die Aktualisierungsschaltfläche (↻) im Header, um den Cache zu invalidieren und frische Daten abzurufen.
- **Integrierte Dokumentation**: Greifen Sie auf dieses Benutzerhandbuch und die Konfigurationsreferenz direkt über die Dokumentationsschaltfläche (📖) in der Oberfläche zu.
- **URL-Synchronisation**: Filter, Suche, Sortierreihenfolge, Paginierung und Mount-Auswahl werden in der URL abgebildet, um Lesezeichen und Teilen zu ermöglichen.
- **I18n**: Volle Unterstützung für Englisch, Französisch, Spanisch, Deutsch und Italienisch. Wechseln Sie die Sprache über das Dropdown-Menü im Header.
- **Dunkelmodus**: Moderne Benutzeroberfläche mit persistentem Dunkel-/Hellmodus-Umschalter.
- **Admin-Panel**: Verwalten Sie die `settings.json` visuell (Vault-Instanzen hinzufügen/entfernen, Schwellenwerte konfigurieren, Protokollierung, CORS). Erfordert die Umgebungsvariable `VCV_ADMIN_PASSWORD`.
- **Prometheus-Metriken**: Zertifikats- und Verbindungsmetriken unter `/metrics` für Überwachung und Alarmierung.

## Bedienung der Oberfläche

### Dashboard

Das Dashboard zeigt ein Ringdiagramm und vier Statuskarten (Gültig, Ablaufend, Abgelaufen, Widerrufen). Klicken Sie auf eine Karte oder ein Diagrammsegment, um die Zertifikatstabelle nach diesem Status zu filtern. Eine Schaltfläche „Filter löschen" erscheint, um den Filter zurückzusetzen.

### PKI-Engine-Auswahl

Klicken Sie auf die Schaltfläche „PKI Engines" in der Filterleiste, um das Mount-Auswahl-Modal zu öffnen. Mounts sind nach Vault-Instanz gruppiert. Sie können:

- Mounts nach Name suchen.
- Einzelne Mounts auswählen/abwählen.
- Alle Mounts einer bestimmten Vault-Instanz auswählen/abwählen.
- Alle Mounts global auswählen/abwählen.

Die Zertifikatstabelle wird automatisch aktualisiert, wenn Sie Mounts umschalten.

### Zertifikatsdetails

Klicken Sie auf die Schaltfläche „Details" in einer Zeile, um ein Modal mit den vollständigen Zertifikatsmetadaten zu öffnen, organisiert in drei Abschnitten: Identität (Betreff, Aussteller, Seriennummer, SANs), Gültigkeit (Erstellungs-/Ablaufdaten mit Countdown) und technische Details (Schlüsselalgorithmus, Schlüsselverwendung, SHA-1/SHA-256-Fingerabdrücke).

### Vault-Status

Das Schild-Symbol im Header zeigt den gesamten Vault-Verbindungsstatus (grün = alle verbunden, rot = mindestens einer getrennt). Klicken Sie darauf, um den Status pro Vault zu sehen. Sie können eine Gesundheitsprüfung über das Modal erzwingen.

## Konfiguration

VCV wird hauptsächlich über eine `settings.json`-Datei konfiguriert. Das Admin-Panel ermöglicht die visuelle Bearbeitung dieser Datei. Siehe die Konfigurationsdokumentation für alle Details.

Alle Anwendungseinstellungen (Vault-Instanzen, Ablaufschwellenwerte, Protokollierung, CORS usw.) werden in `settings.json` definiert. Nur zwei Umgebungsvariablen sind noch erforderlich:

- `VCV_ADMIN_PASSWORD`: Bcrypt-Hash zur Aktivierung des Admin-Panels (aus Sicherheitsgründen als Umgebungsvariable belassen — sollte nicht in einer über die Oberfläche bearbeitbaren Datei gespeichert werden).
- `SETTINGS_PATH`: Pfad zu einer benutzerdefinierten `settings.json`-Datei (nur erforderlich, wenn die Datei nicht an einem Standardspeicherort liegt).

> **Hinweis:** Umgebungsvariablen (`VAULT_ADDRS`, `LOG_LEVEL` usw.) werden weiterhin als Legacy-Fallback unterstützt, wenn keine `settings.json` gefunden wird, aber die Verwendung von `settings.json` ist der empfohlene Ansatz.

## Einschränkungen & was es NICHT tut

- **Nur Lesezugriff**: VCV ist ein Visualisierungswerkzeug. Es erlaubt **kein** Ausstellen, Erneuern oder Widerrufen von Zertifikaten.
- **Authentifizierung**: VCV setzt voraus, dass Sie gültige Token für die Vault-Instanzen bereitgestellt haben.
- **Vault-Verwaltung**: Es verwaltet keine Vault-Richtlinien oder PKI-Konfigurationen; es liest nur vorhandene Daten.

## Support

Bei Problemen oder Funktionsanfragen wenden Sie sich bitte an das Projekt-Repository.
