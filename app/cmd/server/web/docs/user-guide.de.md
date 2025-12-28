# Benutzerdokumentation - VaultCertsViewer (VCV)

## Was ist VCV?

VaultCertsViewer (VCV) ist eine leichtgewichtige Web-Schnittstelle zur Visualisierung und Überwachung von Zertifikaten, die von HashiCorp Vault PKI-Engines verwaltet werden. Es bietet ein zentralisiertes Dashboard zur Verfolgung von Ablaufdaten, Status (gültig, abgelaufen, widerrufen) und technischen Details Ihrer Zertifikate über mehrere Vault-Instanzen und PKI-Mounts hinweg.

## Funktionen

- **Multi-vault-unterstützung**: Verbindung zu einer oder mehreren Vault-Instanzen.
- **PKI-engine-erkennung**: Erkennt automatisch PKI-Mounts, auf die Sie Zugriff haben.
- **Dashboard**: Echtzeit-Statistiken zur Zertifikatsstatusverteilung und Ablauf-Zeitachse.
- **Suchen & Filtern**: Suche nach Common Name (CN) oder Subject Alternative Names (SAN). Filtern nach Vault, PKI-Mount, Status oder Ablaufschwellenwert.
- **Detailansicht**: Zugriff auf vollständige Zertifikatsmetadaten einschließlich Aussteller, Fingerabdrücke und PEM-Inhalt.
- **Export**: Download von Zertifikats-PEM-Dateien direkt aus der Benutzeroberfläche.
- **I18n**: Volle Unterstützung für Englisch, Französisch, Spanisch, Deutsch und Italienisch.
- **Dunkelmodus**: Moderne Benutzeroberfläche mit Umschalter für Dunkel-/Hellmodus.

## Konfiguration

VCV wird primär über Umgebungsvariablen oder eine `settings.json`-Datei konfiguriert.

### Wichtigste Umgebungsvariablen

- `VAULT_ADDRS`: Kommagetrennte Liste von Vault-Adressen.
- `VCV_EXPIRE_WARNING`: Schwellenwert in Tagen für Warnmeldungen (Standard: 30).
- `VCV_EXPIRE_CRITICAL`: Schwellenwert in Tagen für kritische Meldungen (Standard: 7).
- `LOG_LEVEL`: Ausführlichkeit der Protokollierung (info, debug, error).

## Einschränkungen & Was es NICHT tut

- **Nur Lesezugriff**: VCV ist derzeit ein Visualisierungswerkzeug. Es erlaubt **kein** Ausstellen, Erneuern oder Widerrufen von Zertifikaten.
- **Authentifizierung**: VCV geht davon aus, dass Sie gültige Token bereitgestellt oder die Authentifizierung für die Vault-Instanzen konfiguriert haben.
- **Vault-Verwaltung**: Es verwaltet keine Vault-Richtlinien oder PKI-Konfigurationen; es liest nur vorhandene Daten.

## Support

Bei Problemen oder Funktionsanfragen wenden Sie sich bitte an das Projekt-Repository.
