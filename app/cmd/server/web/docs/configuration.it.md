# Riferimento di configurazione

## 📋 Panoramica

VaultCertsViewer (VCV) è configurato principalmente tramite un file `settings.json`. Il pannello di amministrazione consente di gestire questo file direttamente dall'interfaccia web. Le variabili d'ambiente sono supportate come fallback legacy quando non viene trovato alcun `settings.json`.

VCV utilizza un'architettura di rendering lato server basata su [HTMX](https://htmx.org/). Tutti i filtraggi, ordinamenti e paginazioni sono gestiti lato server per prestazioni ottimali.

> **⚠️ Importante:** Dopo aver salvato le modifiche, potrebbe essere necessario un riavvio del server affinché tutte le impostazioni abbiano effetto.

## 🔐 Accesso al pannello di amministrazione

### VCV_ADMIN_PASSWORD

Variabile d'ambiente necessaria per attivare il pannello di amministrazione. Deve essere un **hash bcrypt**.

```bash
# Generare un hash bcrypt (esempio con htpasswd)
htpasswd -nbBC 10 admin LaTuaPasswordSicura | cut -d: -f2

# O con Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'LaTuaPassword', bcrypt.gensalt()).decode())"

# Impostare la variabile d'ambiente
export VCV_ADMIN_PASSWORD='$2a$10$...'
```

Puoi anche utilizzare il servizio 'bcrypt' di <https://tools.hommet.net/bcrypt> per generare un hash bcrypt (nessun dato viene memorizzato).

**Nome utente predefinito:** `admin` (non modificabile, valore predefinito)
**Durata della sessione:** 12 ore (non modificabile, valore predefinito)
**Limitazione tentativi di accesso:** 10 tentativi per 5 minuti (non modificabile, valore predefinito)

## 📁 Impostazioni dell'applicazione

### Ambiente (app.env)

Definisce l'ambiente dell'applicazione. Influisce sulle funzionalità di sicurezza e sul comportamento della registrazione.

- `dev` - Modalità sviluppo (registrazione dettagliata, nessuna limitazione di velocità)
- `prod` - Modalità produzione (cookie sicuri, limitazione di velocità attivata)

**Predefinito:** `prod`

### Porta (app.port)

Porta di ascolto del server HTTP.

**Predefinito:** `52000`

### Percorso del file di configurazione

La variabile d'ambiente `SETTINGS_PATH` specifica il percorso del file `settings.json`. Se non impostata, VCV cerca i file in quest'ordine:

1. `settings.<env>.json` (es: `settings.dev.json`)
2. `settings.json`
3. `./settings.json`
4. `/etc/vcv/settings.json`

### Registrazione (app.logging)

Configura il comportamento della registrazione dell'applicazione:

- **level**: `debug`, `info`, `warn`, `error`
- **format**: `json` o `text`
- **output**: `stdout`, `file` o `both`
- **file_path**: Percorso del file di log quando output è `file` o `both`

**Predefiniti:**

- level: `info`
- format: `json`
- output: `stdout`
- file_path: `/var/log/app/vcv.log`

## 📜 Impostazioni dei certificati

### Soglie di scadenza (certificates.expiration_thresholds)

Configura quando i certificati vengono contrassegnati come in scadenza:

- **critical**: Giorni prima della scadenza per mostrare un avviso critico
- **warning**: Giorni prima della scadenza per mostrare un avviso

Queste soglie controllano:

- Banner di notifica in cima alla pagina
- Codifica colori nella tabella dei certificati (rosso per critico, giallo per avviso)
- Visualizzazione della cronologia sulla dashboard
- Metriche Prometheus (`vcv_certificates_expiring_critical`, `vcv_certificates_expiring_warning`)

**Predefiniti:**

- critical: `7`
- warning: `30`

## 🌐 Impostazioni CORS (cors)

### Origini consentite (cors.allowed_origins)

Array di origini CORS consentite. Usa `["*"]` per consentire tutte le origini (non raccomandato in produzione).

**Esempio:**

```json
"allowed_origins": ["https://example.com", "https://app.example.com"]
```

### Consentire credenziali (cors.allow_credentials)

Booleano per consentire le credenziali nelle richieste CORS.

**Predefinito:** `false`

**Nota:** CORS è principalmente utile se integri VCV in un'altra applicazione web o vi accedi da un dominio diverso.

## 🔐 Configurazione Vault

### Istanze Vault multiple

VaultCertsViewer supporta il monitoraggio simultaneo di più istanze Vault. Ogni istanza Vault richiede:

- **ID**: Identificatore univoco per questa istanza Vault (obbligatorio)
- **Display name**: Nome leggibile mostrato nell'interfaccia (opzionale)
- **Address**: URL del server Vault (es: `https://vault.example.com:8200`)
- **Token**: Token Vault di sola lettura con accesso PKI (obbligatorio)
- **PKI mounts**: Array dei percorsi di montaggio PKI (es: `["pki", "pki2", "pki-prod"]`)
- **Enabled**: Se questa istanza Vault è attiva

### Configurazione TLS

Per Vault che utilizzano certificati CA personalizzati o autofirmati:

- **TLS CA cert (Base64)**: Bundle CA PEM codificato in base64 (metodo preferito)
- **TLS CA cert path**: Percorso del file al bundle CA PEM
- **TLS CA path**: Directory contenente i certificati CA
- **TLS server name**: Sostituzione del nome server SNI
- **TLS insecure**: Ignora la verifica TLS (⚠️ solo sviluppo, non raccomandato)

```bash
# Codificare un certificato in base64
cat percorso-al-cert.pem | base64 | tr -d '\n'
```

**Precedenza:** Se `tls_ca_cert_base64` è impostato, ha la priorità sui percorsi dei file.

### Permessi del token Vault

Il token Vault deve avere accesso di sola lettura ai montaggi PKI. Esempio di policy:

```hcl
# Montaggi espliciti (raccomandato per la produzione)
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

# Pattern con wildcard (per ambienti dinamici)
vault policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"    { capabilities = ["read"] }
EOF

# Creare il token
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Devi sostituire 'pki' e 'pki2' con i percorsi di montaggio PKI del tuo Vault. Aggiungi tanti percorsi di montaggio PKI quanti ne hai nel tuo Vault.

## ⚡ Ottimizzazioni delle prestazioni

### Cache

VaultCertsViewer implementa la cache per migliorare le prestazioni:

- **TTL cache certificati:** 15 minuti (riduce le chiamate API a Vault)
- **Cache controllo di salute:** 30 secondi (per l'indicatore di stato nell'intestazione)
- **Recupero parallelo:** Più Vault vengono interrogati simultaneamente
- **Invalidazione cache:** Usa il pulsante di aggiornamento (↻) nell'intestazione o `POST /api/cache/invalidate` per svuotare la cache dei certificati

Con più Vault, il recupero parallelo offre tempi di caricamento **3-10× più veloci** rispetto alle query sequenziali.

## 📊 Monitoraggio e metriche

### Metriche Prometheus

Disponibili all'endpoint `/metrics`:

- `vcv_certificates_total` - Numero totale di certificati
- `vcv_certificates_valid` - Numero di certificati validi
- `vcv_certificates_expired` - Numero di certificati scaduti
- `vcv_certificates_revoked` - Numero di certificati revocati
- `vcv_certificates_expiring_critical` - Certificati in scadenza entro la soglia critica
- `vcv_certificates_expiring_warning` - Certificati in scadenza entro la soglia di avviso
- `vcv_vault_connected` - Stato di connessione Vault (1=connesso, 0=disconnesso)
- `vcv_cache_size` - Numero di voci nella cache
- `vcv_last_fetch_timestamp` - Timestamp Unix dell'ultimo recupero di certificati

Tutte le metriche includono etichette: `vault_id`, `vault_name`, `pki_mount`

### Endpoint salute e API

- `/api/health` - Controllo di salute base (restituisce sempre 200 OK)
- `/api/ready` - Sonda di disponibilità (verifica lo stato dell'applicazione)
- `/api/status` - Stato dettagliato incluse tutte le connessioni Vault
- `/api/version` - Informazioni sulla versione dell'applicazione
- `/api/config` - Configurazione dell'applicazione (soglie di scadenza, lista vault)
- `/api/i18n` - Traduzioni per la lingua corrente
- `/api/certs` - Lista certificati (JSON)
- `/api/certs/{id}/details` - Dettagli del certificato (JSON)
- `/api/certs/{id}/pem` - Contenuto PEM del certificato (JSON)
- `/api/certs/{id}/pem/download` - Scarica il file PEM del certificato
- `POST /api/cache/invalidate` - Invalida la cache dei certificati

### Limitazione di velocità

In modalità `prod`, la limitazione di velocità dell'API è attivata a **300 richieste al minuto** per client. I seguenti percorsi sono esenti:

- `/api/health`, `/api/ready`, `/metrics`
- `/assets/*` (file statici)

## 🔒 Best practice di sicurezza

- Utilizza sempre l'ambiente `prod` in produzione
- Proteggi il file `settings.json` (contiene token sensibili)
- Usa token Vault di sola lettura con permessi minimi
- La limitazione di velocità è automatica in modalità `prod` (300 rich./min)
- La protezione CSRF è attivata su tutte le richieste che modificano lo stato
- Gli header di sicurezza (X-Content-Type-Options, X-Frame-Options, ecc.) sono impostati automaticamente
- Esegui il container con `--read-only` e `--cap-drop=ALL`

## 📝 Esempio settings.json

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
      "display_name": "Vault Produzione",
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
      "display_name": "Vault Sviluppo",
      "address": "https://vault-dev.example.com:8200",
      "token": "hvs.yyy",
      "pki_mounts": ["pki_dev"],
      "enabled": true,
      "tls_insecure": true
    }
  ]
}
```

> **💡 Suggerimento:** Usa il pannello di amministrazione per modificare queste impostazioni visualmente. Le modifiche vengono salvate nel file `settings.json`.
