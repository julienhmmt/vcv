# Guida utente - VaultCertsViewer (VCV)

## Cos'è VCV?

VaultCertsViewer (VCV) è un'interfaccia web leggera progettata per visualizzare e monitorare i certificati gestiti dai motori PKI di HashiCorp Vault (o OpenBao). Fornisce una dashboard centralizzata per tracciare le date di scadenza, lo stato (valido, scaduto, revocato) e i dettagli tecnici dei certificati su più istanze Vault e punti di montaggio PKI.

## Funzionalità

- **Supporto multi-vault**: Connessione simultanea a una o più istanze Vault.
- **Selettore motori PKI**: Filtra i certificati per istanza Vault e punto di montaggio PKI tramite un modale interattivo con ricerca, selezione/deselezione per vault o globalmente.
- **Dashboard**: Grafico ad anello con statistiche in tempo reale sulla distribuzione dello stato dei certificati (valido, in scadenza, scaduto, revocato). Clicca su un segmento o una scheda di stato per filtrare la tabella istantaneamente.
- **Ricerca e filtraggio**: Cerca per Common Name (CN) o Subject Alternative Names (SAN). Filtra per stato tramite le schede della dashboard.
- **Ordinamento**: Ordina la tabella dei certificati per Common Name, data di creazione, data di scadenza, nome del Vault o punto di montaggio PKI. Clicca su un'intestazione di colonna per alternare tra ordine crescente/decrescente.
- **Paginazione**: Paginazione lato server con dimensioni di pagina configurabili (25, 50, 100 o Tutti).
- **Vista dettagliata**: Accedi ai metadati completi del certificato in un modale organizzato: identità (soggetto, emittente, numero di serie, SANs), date di validità con stato di scadenza, e dettagli tecnici (algoritmo di chiave, utilizzo della chiave, impronte digitali SHA-1/SHA-256).
- **Download PEM**: Scarica i file PEM direttamente dalla tabella.
- **Stato Vault**: Un indicatore nell'intestazione (icona scudo con punto di stato) mostra lo stato di connessione in tempo reale delle istanze Vault. Clicca per aprire un modale dettagliato con informazioni di salute per vault e un pulsante di aggiornamento.
- **Notifiche di scadenza**: Un banner in cima alla pagina avvisa dei certificati in scadenza entro le soglie configurate (critico / avviso).
- **Notifiche toast**: Messaggi toast in tempo reale per cambiamenti di connessione Vault, errori e feedback utente.
- **Cache e aggiornamento**: I dati dei certificati sono memorizzati nella cache lato server (TTL di 15 min). Usa il pulsante di aggiornamento (↻) nell'intestazione per invalidare la cache e ottenere dati freschi.
- **Documentazione integrata**: Accedi a questa guida utente e al riferimento di configurazione direttamente dall'interfaccia tramite il pulsante documentazione (📖).
- **Sincronizzazione URL**: Filtri, ricerca, ordinamento, paginazione e selezione dei montaggi si riflettono nell'URL per condivisione e segnalibri.
- **I18n**: Supporto completo per inglese, francese, spagnolo, tedesco e italiano. Cambia lingua con il menu a tendina nell'intestazione.
- **Dark mode**: Interfaccia moderna con commutatore persistente modalità scura/chiara.
- **Pannello di amministrazione**: Gestisci il file `settings.json` visualmente (aggiungi/rimuovi istanze Vault, configura soglie, registrazione, CORS). Richiede la variabile d'ambiente `VCV_ADMIN_PASSWORD`.
- **Metriche Prometheus**: Esponi metriche di certificati e connessione su `/metrics` per monitoraggio e allerte.

## Utilizzo dell'interfaccia

### Dashboard

La dashboard mostra un grafico ad anello e quattro schede di stato (Valido, In scadenza, Scaduto, Revocato). Clicca su qualsiasi scheda o segmento del grafico per filtrare la tabella dei certificati per quello stato. Appare un pulsante «Cancella filtro» per ripristinare il filtro.

### Selettore motori PKI

Clicca sul pulsante «PKI Engines» nella barra dei filtri per aprire il modale di selezione dei montaggi. I montaggi sono raggruppati per istanza Vault. Puoi:

- Cercare montaggi per nome.
- Selezionare/deselezionare singoli montaggi.
- Selezionare/deselezionare tutti i montaggi di una specifica istanza Vault.
- Selezionare/deselezionare tutti i montaggi globalmente.

La tabella dei certificati si aggiorna automaticamente quando alterni i montaggi.

### Dettagli del certificato

Clicca sul pulsante «Dettagli» su qualsiasi riga per aprire un modale con i metadati completi del certificato, organizzati in tre sezioni: identità (soggetto, emittente, numero di serie, SANs), validità (date di creazione/scadenza con conto alla rovescia), e dettagli tecnici (algoritmo di chiave, utilizzo della chiave, impronte digitali SHA-1/SHA-256).

### Stato Vault

L'icona scudo nell'intestazione mostra lo stato complessivo di connessione Vault (verde = tutti connessi, rosso = almeno uno disconnesso). Clicca per vedere lo stato per vault. Puoi forzare un controllo di salute dal modale.

## Configurazione

VCV è configurato principalmente tramite un file `settings.json`. Il pannello di amministrazione consente di modificare questo file visualmente. Consulta la documentazione di configurazione per tutti i dettagli.

Tutti i parametri dell'applicazione (istanze Vault, soglie di scadenza, registrazione, CORS, ecc.) sono definiti in `settings.json`. Solo due variabili d'ambiente sono ancora necessarie:

- `VCV_ADMIN_PASSWORD`: Hash bcrypt per attivare il pannello di amministrazione (mantenuto come variabile d'ambiente per sicurezza — non deve essere memorizzato in un file modificabile dall'interfaccia).
- `SETTINGS_PATH`: Percorso verso un file `settings.json` personalizzato (necessario solo se il file non si trova in una posizione predefinita).

> **Nota:** Le variabili d'ambiente (`VAULT_ADDRS`, `LOG_LEVEL`, ecc.) sono ancora supportate come fallback legacy quando non viene trovato alcun `settings.json`, ma l'uso di `settings.json` è l'approccio consigliato.

## Limiti e cosa NON fa

- **Sola lettura**: VCV è uno strumento di visualizzazione. **Non** consente l'emissione, il rinnovo o la revoca dei certificati.
- **Autenticazione**: VCV presuppone che tu abbia fornito token validi per le istanze Vault a cui si connette.
- **Gestione Vault**: Non gestisce le policy di Vault né la configurazione PKI; legge solo i dati esistenti.

## Supporto

Per problemi o richieste di funzionalità, fare riferimento al repository del progetto.
