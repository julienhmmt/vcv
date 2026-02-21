# Konfigurationsreferenz

## 📋 Übersicht

VaultCertsViewer (VCV) wird hauptsächlich über eine `settings.json`-Datei konfiguriert. Das Admin-Panel ermöglicht es Ihnen, diese Datei direkt über die Web-Oberfläche zu verwalten. Umgebungsvariablen werden als Legacy-Fallback unterstützt, wenn keine `settings.json` gefunden wird.

VCV verwendet eine serverseitig gerenderte Architektur basierend auf [HTMX](https://htmx.org/). Alle Filter-, Sortier- und Paginierungsvorgänge werden serverseitig für optimale Leistung verarbeitet.

> **⚠️ Wichtig:** Nach dem Speichern von Änderungen kann ein Neustart des Servers erforderlich sein, damit alle Einstellungen wirksam werden.

## 🔐 Zugang zum Admin-Panel

### Admin-Passwort in der Konfigurationsdatei

Um das Admin-Panel zu aktivieren, fügen Sie einen `admin`-Abschnitt zu Ihrer `settings.json`-Datei mit einem bcrypt-Passwort hinzu:

```json
{
  "admin": {
    "password": "$2b$12$Ihr-bcrypt-hash-hier"
  }
}
```

Das Passwort muss ein **bcrypt-Hash** sein (Präfix `$2a$`, `$2b$`, oder `$2y$`).

```bash
# Einen bcrypt-Hash generieren (Beispiel mit htpasswd)
htpasswd -nbBC 10 admin IhrSicheresPasswort | cut -d: -f2

# Oder mit Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'IhrPasswort', bcrypt.gensalt()).decode())"
```

Sie können auch den 'bcrypt'-Dienst von <https://tools.hommet.net/bcrypt> verwenden, um einen bcrypt-Hash zu generieren (es werden keine Daten gespeichert).

**Standard-Benutzername:** `admin` (nicht änderbar, Standardwert)
**Sitzungsdauer:** 12 Stunden (nicht änderbar, Standardwert)
**Login-Ratenbegrenzung:** 10 Versuche pro 5 Minuten (nicht änderbar, Standardwert)

## 📁 Anwendungseinstellungen

### Umgebung (app.env)

Definiert die Anwendungsumgebung. Beeinflusst Sicherheitsfunktionen und Protokollierungsverhalten.

- `dev` - Entwicklungsmodus (ausführliche Protokollierung, keine Ratenbegrenzung)
- `prod` - Produktionsmodus (sichere Cookies, Ratenbegrenzung aktiviert)

**Standard:** `prod`

### Port (app.port)

HTTP-Server-Listening-Port.

**Standard:** `52000`

### Pfad der Konfigurationsdatei

VCV sucht die Konfigurationsdatei in dieser Reihenfolge:

1. `settings.dev.json`
2. `settings.prod.json`
3. `settings.json`
4. `./settings.json`
5. `/app/settings.json`

### Protokollierung (app.logging)

Konfigurieren Sie das Protokollierungsverhalten der Anwendung:

- **level**: `debug`, `info`, `warn`, `error`
- **format**: `json` oder `text`
- **output**: `stdout`, `file` oder `both`
- **file_path**: Pfad der Protokolldatei wenn output `file` oder `both` ist

**Standardwerte:**

- level: `info`
- format: `json`
- output: `stdout`
- file_path: `/var/log/app/vcv.log`

## 📜 Zertifikatseinstellungen

### Ablaufschwellenwerte (certificates.expiration_thresholds)

Konfigurieren Sie, wann Zertifikate als bald ablaufend gekennzeichnet werden:

- **critical**: Tage vor Ablauf für kritische Warnung
- **warning**: Tage vor Ablauf für Warnung

Diese Schwellenwerte steuern:

- Benachrichtigungsbanner oben auf der Seite
- Farbcodierung in der Zertifikatstabelle (rot für kritisch, gelb für Warnung)
- Zeitachsen-Visualisierung auf dem Dashboard
- Prometheus-Metriken (`vcv_certificates_expiring_critical`, `vcv_certificates_expiring_warning`)

**Standardwerte:**

- critical: `7`
- warning: `30`

## 🌐 CORS-Einstellungen (cors)

### Erlaubte Ursprünge (cors.allowed_origins)

Array erlaubter CORS-Ursprünge. Verwenden Sie `["*"]` um alle Ursprünge zuzulassen (in Produktion nicht empfohlen).

**Beispiel:**

```json
"allowed_origins": ["https://example.com", "https://app.example.com"]
```

### Anmeldedaten erlauben (cors.allow_credentials)

Boolean um Anmeldedaten in CORS-Anfragen zu erlauben.

**Standard:** `false`

**Hinweis:** CORS ist hauptsächlich nützlich, wenn Sie VCV in eine andere Webanwendung einbetten oder von einer anderen Domain darauf zugreifen.

## 🔐 Vault-Konfiguration

### Mehrere Vault-Instanzen

VaultCertsViewer unterstützt die gleichzeitige Überwachung mehrerer Vault-Instanzen. Jede Vault-Instanz erfordert:

- **ID**: Eindeutiger Bezeichner für diese Vault-Instanz (erforderlich)
- **Display name**: Lesbarer Name, der in der Oberfläche angezeigt wird (optional)
- **Address**: Vault-Server-URL (z.B. `https://vault.example.com:8200`)
- **Token**: Nur-Lese-Vault-Token mit PKI-Zugriff (erforderlich)
- **PKI mounts**: Array von PKI-Mount-Pfaden (z.B. `["pki", "pki2", "pki-prod"]`)
- **Enabled**: Ob diese Vault-Instanz aktiv ist

### TLS-Konfiguration

Für Vaults mit benutzerdefinierten CA-Zertifikaten oder selbstsignierten Zertifikaten:

- **TLS CA cert (Base64)**: Base64-kodiertes PEM-CA-Bundle (bevorzugte Methode)
- **TLS CA cert path**: Dateipfad zum PEM-CA-Bundle
- **TLS CA path**: Verzeichnis mit CA-Zertifikaten
- **TLS server name**: SNI-Servernamen-Überschreibung
- **TLS insecure**: TLS-Überprüfung überspringen (⚠️ nur Entwicklung, nicht empfohlen)

```bash
# Ein Zertifikat in Base64 kodieren
cat pfad-zum-zertifikat.pem | base64 | tr -d '\n'
```

**Vorrang:** Wenn `tls_ca_cert_base64` gesetzt ist, hat es Vorrang vor Dateipfaden.

### Vault-Token-Berechtigungen

Der Vault-Token muss Nur-Lese-Zugriff auf PKI-Mounts haben. Beispiel-Policy:

```hcl
# Explizite Mounts (empfohlen für Produktion)
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

# Wildcard-Muster (für dynamische Umgebungen)
vault policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"    { capabilities = ["read"] }
EOF

# Token erstellen
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Sie müssen 'pki' und 'pki2' durch die PKI-Mount-Pfade Ihres Vaults ersetzen. Fügen Sie so viele PKI-Mount-Pfade hinzu, wie Sie in Ihrem Vault haben.

## ⚡ Leistungsoptimierungen

### Caching

VaultCertsViewer implementiert Caching zur Verbesserung der Leistung:

- **Zertifikats-Cache-TTL:** 15 Minuten (reduziert Vault-API-Aufrufe)
- **Gesundheitsprüfungs-Cache:** 30 Sekunden (für den Statusindikator im Header)
- **Parallele Abfragen:** Mehrere Vaults werden gleichzeitig abgefragt
- **Cache-Invalidierung:** Verwenden Sie die Aktualisierungsschaltfläche (↻) im Header oder `POST /api/cache/invalidate` um den Zertifikats-Cache zu leeren

Mit mehreren Vaults bietet die parallele Abfrage **3-10× schnellere** Ladezeiten im Vergleich zu sequentiellen Abfragen.

## 📊 Überwachung & Metriken

### Prometheus-Metriken

Verfügbar am `/metrics`-Endpoint:

- `vcv_certificates_total` - Gesamtzahl der Zertifikate
- `vcv_certificates_valid` - Anzahl gültiger Zertifikate
- `vcv_certificates_expired` - Anzahl abgelaufener Zertifikate
- `vcv_certificates_revoked` - Anzahl widerrufener Zertifikate
- `vcv_certificates_expiring_critical` - Zertifikate, die innerhalb des kritischen Schwellenwerts ablaufen
- `vcv_certificates_expiring_warning` - Zertifikate, die innerhalb des Warnschwellenwerts ablaufen
- `vcv_vault_connected` - Vault-Verbindungsstatus (1=verbunden, 0=getrennt)
- `vcv_cache_size` - Anzahl der Cache-Einträge
- `vcv_last_fetch_timestamp` - Unix-Zeitstempel der letzten Zertifikatsabfrage

Alle Metriken enthalten Labels: `vault_id`, `vault_name`, `pki_mount`

### Gesundheits- & API-Endpoints

- `/api/health` - Einfache Gesundheitsprüfung (gibt immer 200 OK zurück)
- `/api/ready` - Bereitschaftsprüfung (prüft den Anwendungszustand)
- `/api/status` - Detaillierter Status einschließlich aller Vault-Verbindungen
- `/api/version` - Anwendungsversionsinformationen
- `/api/config` - Anwendungskonfiguration (Ablaufschwellenwerte, Vault-Liste)
- `/api/i18n` - Übersetzungen für die aktuelle Sprache
- `/api/certs` - Zertifikatsliste (JSON)
- `/api/certs/{id}/details` - Zertifikatsdetails (JSON)
- `/api/certs/{id}/pem` - PEM-Inhalt des Zertifikats (JSON)
- `/api/certs/{id}/pem/download` - PEM-Datei des Zertifikats herunterladen
- `POST /api/cache/invalidate` - Zertifikats-Cache invalidieren

### Ratenbegrenzung

Im `prod`-Modus ist die API-Ratenbegrenzung mit **300 Anfragen pro Minute** pro Client aktiviert. Folgende Pfade sind ausgenommen:

- `/api/health`, `/api/ready`, `/metrics`
- `/assets/*` (statische Dateien)

## 🔒 Sicherheits-Best-Practices

- Verwenden Sie in der Produktion immer die `prod`-Umgebung
- Schützen Sie die `settings.json`-Datei (enthält sensible Tokens)
- Verwenden Sie Nur-Lese-Vault-Tokens mit minimalen Berechtigungen
- Ratenbegrenzung ist im `prod`-Modus automatisch aktiv (300 Anfr./Min.)
- CSRF-Schutz ist bei allen zustandsändernden Anfragen aktiviert
- Sicherheits-Header (X-Content-Type-Options, X-Frame-Options, usw.) werden automatisch gesetzt
- Container mit `--read-only` und `--cap-drop=ALL` ausführen

## 📝 Beispiel settings.json

```json
{
  "app": {
    "env": "prod",
    "port": 52000,
    "logging": {
      "level": "info",
      "format": "json",
      "output": "stdout",
      "file_path": "/var/log/app/vcv.log"
    }
  },
  "certificates": {
    "expiration_thresholds": {
      "critical": 7,
      "warning": 30
    }
  },
  "cors": {
    "allowed_origins": ["https://example.com"],
    "allow_credentials": false
  },
  "vaults": [
    {
      "id": "vault-prod",
      "display_name": "Produktions-Vault",
      "address": "https://vault.example.com:8200",
      "token": "hvs.xxx",
      "pki_mounts": ["pki", "pki-intermediate"],
      "enabled": true,
      "tls_insecure": false,
      "tls_ca_cert_base64": "LS0tLS1CRUdJTi...",
      "tls_server_name": "vault.example.com"
    },
    {
      "id": "vault-dev",
      "display_name": "Entwicklungs-Vault",
      "address": "https://vault-dev.example.com:8200",
      "token": "hvs.yyy",
      "pki_mounts": ["pki_dev"],
      "enabled": true,
      "tls_insecure": true
    }
  ]
}
```

> **💡 Tipp:** Verwenden Sie das Admin-Panel, um diese Einstellungen visuell zu bearbeiten. Änderungen werden in der `settings.json`-Datei gespeichert.
