# Documentazione utente - VaultCertsViewer (VCV)

## Cos'è VCV?

VaultCertsViewer (VCV) è un'interfaccia web leggera progettata per visualizzare e monitorare i certificati gestiti dai motori PKI di HashiCorp Vault. Fornisce una dashboard centralizzata per tracciare le date di scadenza, lo stato (valido, scaduto, revocato) e i dettagli tecnici dei certificati su più istanze Vault e punti di montaggio PKI.

## Funzionalità

- **Supporto multi-vault**: Connettiti a una o più istanze Vault.
- **Rilevamento motori PKI**: Rileva automaticamente i punti di montaggio PKI a cui hai accesso.
- **Dashboard**: Statistiche in tempo reale sulla distribuzione dello stato dei certificati e sulla cronologia delle scadenze.
- **Ricerca e filtraggio**: Cerca per Common Name (CN) o Subject Alternative Names (SAN). Filtra per Vault, punto di montaggio PKI, stato o soglia di scadenza.
- **Vista dettagliata**: Accedi ai metadati completi del certificato, inclusi emittente, impronte digitali e contenuto PEM.
- **Esportazione**: Scarica i file PEM dei certificati direttamente dall'interfaccia utente.
- **I18n**: Supporto completo per inglese, francese, spagnolo, tedesco e italiano.
- **Dark mode**: Interfaccia utente moderna con commutatore modalità scura/chiara.

## Configurazione

VCV è configurato principalmente tramite variabili d'ambiente o un file `settings.json`.

### Principali variabili d'ambiente

- `VAULT_ADDRS`: Elenco di indirizzi Vault separati da virgole.
- `VCV_EXPIRE_WARNING`: Soglia in giorni per le notifiche di avviso (predefinito: 30).
- `VCV_EXPIRE_CRITICAL`: Soglia in giorni per le notifiche critiche (predefinito: 7).
- `LOG_LEVEL`: Dettaglio del registro (info, debug, error).

## Limiti e cosa NON fa

- **Sola lettura**: VCV è attualmente uno strumento di visualizzazione. **Non** consente l'emissione, il rinnovo o la revoca dei certificati.
- **Autenticazione**: VCV presuppone che tu abbia fornito token validi o configurato l'autenticazione per le istanze Vault a cui si connette.
- **Gestione Vault**: Non gestisce le policy di Vault né la configurazione PKI; legge solo i dati esistenti.

## Supporto

Per problemi o richieste di funzionalità, fare riferimento al repository del progetto.
