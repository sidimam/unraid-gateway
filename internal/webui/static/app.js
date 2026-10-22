(() => {
  const $ = (id) => document.getElementById(id);
  const API = '/api/v1';

  // ---- i18n -----------------------------------------------------------------
  // The same seven languages as the Unraid Drive app. "system" follows the browser.
  const I18N = {
    en: {
      tab_status: 'Status', tab_settings: 'Settings', health_title: 'Health', health_checks: 'All checks', health_ok: 'All fine', health_warning: 'Warnings', health_error: 'Problems', health_unknown: 'Checking…', health_unavailable: 'health unavailable: {err}', checked_at: 'checked {when}', sys_title: 'System information', sys_unavailable: 'system information unavailable: {err}', sys_host: 'Host', sys_os: 'Unraid', sys_kernel: 'Kernel', sys_uptime: 'Up since', sys_cpu: 'Processor', sys_cores: 'Cores / threads', sys_clock: 'Clock', sys_memory: 'Memory', sys_memory_use: 'In use', sys_board: 'Mainboard', sys_versions: 'Unraid API', sys_gpu: 'Graphics', sys_network: 'Network', appearance: 'Appearance', language: 'Language', theme: 'Theme',
      connecting: 'connecting…', online: 'gateway online', unreachable: 'gateway unreachable',
      system: 'System', theme_system: 'Theme: system', theme_light: 'Theme: light', theme_dark: 'Theme: dark',
      forget_key: 'Forget stored key', forget_key_title: 'Delete the key stored on the gateway', disconnect: 'Disconnect',
      login_title: 'Connect', login_intro: 'Use an Unraid API key (Settings → Management Access → API Keys). VIEWER is enough for files; ADMIN also manages the server and its keys. The key is validated by Unraid and exchanged for a session token.',
      stored_connect: 'Connect with the key stored on this gateway', stored_env: 'from the WEBUI_API_KEY variable', stored_file: 'remembered on this gateway',
      pm_account: 'Account name (for your password manager — edit it freely)', api_key: 'Unraid API key', api_key_ph: 'paste your Unraid API key',
      remember_key: 'Remember this key on the gateway (/config/webui.key): anyone who can open this page could then use it — keep the gateway on the LAN or behind Cloudflare Access',
      connect: 'Connect', as_user: 'Sign in as an Unraid user…', user_intro: 'With an Unraid user and password the gateway applies that user\'s share permissions (public / secure / private, read or write), exactly as over SMB. The API key above is still required.',
      unraid_user: 'Unraid user', unraid_password: 'Unraid password', connect_as_user: 'Connect as this user', need_key: 'Paste the API key first (or use the key stored on the gateway).',
      stat_connected: 'Connected now', stat_streams: 'Streams / transfers', stat_devices: 'Registered devices', stat_gateway: 'Gateway',
      tab_files: 'Files', tab_activity: 'Activity', tab_devices: 'Devices', tab_notify: 'Notifications', tab_keys: 'API keys', tab_advanced: 'Advanced',
      refresh: 'Refresh', new_folder: 'New folder', upload: 'Upload', shares: 'shares', read_only: 'read-only',
      root_hint: 'These are the shares mounted into the container. Open one to upload files or create folders. File access is decided by the volume mounts (read/write or read-only) and, with an Unraid user, by that user\'s share permissions.',
      col_name: 'Name', col_size: 'Size', col_modified: 'Modified', empty_folder: 'Empty folder.', rename: 'Rename', delete: 'Delete',
      new_name: 'New name', folder_name: 'Folder name', open_share_first: 'The root only lists mounted shares. Open a share first.',
      confirm_delete_file: 'Delete file "{name}"?', confirm_delete_folder: 'Delete folder "{name}" and everything inside it?', upload_failed: 'upload failed: {name}',
      live: 'live', connected_devices: 'Connected devices', col_device: 'Device', col_user: 'User', col_from: 'From', col_since: 'Since', col_last_seen: 'Last seen', col_requests: 'Requests', col_last_file: 'Last file',
      nobody_connected: 'Nobody connected right now (sessions disappear 30 minutes after their last request).', streams_title: 'Streams and transfers in progress',
      col_what: 'What', col_file: 'File', col_who: 'Who', col_progress: 'Progress', col_speed: 'Speed', col_elapsed: 'Elapsed', nothing_active: 'Nothing playing or transferring.',
      recent_transfers: 'Recent transfers', col_bytes: 'Bytes', col_ended: 'Ended', scope_user: 'your devices only', scope_all: 'all users', activity_unavailable: 'activity unavailable: {err}',
      k_download: '⬇︎ download', k_stream: '▶︎ stream', k_upload: '⬆︎ upload', now: 'now', s_ago: '{n} s ago', min_ago: '{n} min ago', h_ago: '{n} h ago',
      dev_older: 'Unraid Drive (older build)', dev_browser: 'Browser', dev_mpv: 'mpv player', unknown: 'unknown',
      remove_all: 'Remove all', devices_intro: 'Every Unraid Drive installation registers itself when you add the server. Removing a device closes its sessions and the app asks to sign in again (which registers it anew and sends a notification). A reinstalled app supersedes its previous entry, shown greyed as "old".',
      col_key: 'Key', col_registered: 'Registered', col_logins: 'Sign-ins', no_devices: 'No device registered yet (older app builds do not register; update to 1.3 build 29 or later).',
      this_browser: '(this browser)', old: 'old', last_seen_prefix: 'last seen', replaced_title: 'Replaced by a new installation on {date}. Safe to remove.', registration_off: 'registration is off (DEVICE_REGISTRATION=off)', devices_unavailable: 'devices unavailable: {err}',
      users_title: 'Users', users_intro: 'One row per Unraid user (or API key) with registered devices: how many devices and syncs, when last seen, what moved today. Open a user for its devices, the folder syncs and backups each device runs (pause or resume them), the files they moved and the user\'s activity over time.', col_user: 'User', col_devices: 'Devices', col_syncs: 'Syncs', col_today: 'Today', no_users: 'No user yet: users appear when their devices register.', users_unavailable: 'users unavailable: {err}', open: 'Open', n_paused: '{n} paused', n_warnings: '{n} warning(s)', all_fine: 'All fine', user_devices: 'Devices', user_syncs: 'Folder syncs and backups', user_syncs_intro: 'Defined by the apps on each device. Pausing a sync here stops it on the device at its next run (the app shows it as paused) until you resume it.', no_syncs: 'No sync defined on this user\'s devices.', no_syncs_device: 'No sync defined on this device.', user_log_intro: 'every device of this user, newest first', kind_folder: 'Folder sync', kind_photos: 'Photos backup', kind_downloads: 'Downloads', mode_up: 'device → Unraid', mode_down: 'Unraid → device', mode_both: 'two-way', sync_paused: 'paused', sync_paused_by: 'paused by {who}', sync_never: 'never run yet', sync_failed: 'last run failed', sync_partial: 'last run with {c} conflicts, {e} errors', sync_ok: 'last run fine', sync_last: 'last run {when}: {n} files, {bytes}', sync_files_btn: 'Files ({n})', sync_no_files: 'No file moved yet.', col_file: 'File', col_direction: 'Direction', col_outcome: 'Outcome', dir_up: 'uploaded', dir_down: 'downloaded', dir_delete: 'deleted', dir_conflict: 'conflict', out_ok: 'ok', out_conflict: 'conflict', out_failed: 'failed', out_skipped: 'skipped', syncs_unavailable: 'syncs unavailable: {err}', ev_sync: 'Sync',
      active_today: 'active today', stale: 'not seen for {n} days', last_sync: 'Last sync', prune_label: 'Remove the devices not seen for', prune_days: 'days', prune_run: 'Remove stale devices', confirm_prune: 'Remove every device not seen for {n} days? Their apps will ask to sign in again if they come back.', pruned: '{n} device(s) removed:\n{list}', pruned_none: 'No device was that old.', back: 'Back', charts_title: 'Data moved', range_24h: 'Last 24 hours', range_7d: 'Last 7 days', range_30d: 'Last 30 days', range_90d: 'Last 90 days', chart_empty: 'Nothing moved in this period.', log_title: 'Log', log_intro: 'sign-ins, registrations, removals and interrupted transfers', col_when: 'When', col_event: 'Event', col_detail: 'Detail', log_empty: 'No events recorded yet.', state: 'State', tot_today: 'Today', tot_week: 'Last 7 days', tot_month: 'Last 30 days', tot_detail: '{req} requests · {dl} downloads/streams · {ul} uploads', metrics_unavailable: 'metrics unavailable: {err}', legend_down: 'downloaded / streamed', legend_up: 'uploaded', ev_login: 'Sign-in', ev_register: 'Registration', ev_revoked: 'Revoked', ev_transfer: 'Transfer', ev_admin: 'Administration', log_unavailable: 'log unavailable: {err}',
      remove: 'Remove', confirm_remove_device: 'Remove "{name}"?\nIts sessions are closed and the app will ask to sign in again.', confirm_remove_this: '\n\nThis is the browser you are using: you will be signed out.', confirm_remove_all: 'Remove ALL devices? Every app (and this browser) will have to sign in again.',
      send_test: 'Send a test', notify_intro: 'A new or removed device raises a notification on every configured channel: Unraid\'s notifications (NOTIFY_UNRAID — set NOTIFY_UNRAID_API_KEY to an ADMIN key if yours is VIEWER), e-mail (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_FROM, SMTP_TO, SMTP_TLS) and Telegram (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID). Set them in the container settings.',
      channels: 'channels: {list}', no_channel: 'no channel configured', sending: 'sending…', no_channel_long: 'No channel configured: set NOTIFY_UNRAID / SMTP_* / TELEGRAM_* in the container.', errors: 'Errors: {list}', sent_to: 'Sent to {list}.', error: 'error: {err}',
      needs_admin: '(needs an ADMIN key)', keys_intro: 'List, create, delete and rotate the API keys of your Unraid server. A new key is shown once: copy it into the apps (Edit server or credentials). Rotate = create a replacement, optionally remember it for this web UI and delete the old one.',
      list_keys: 'List keys', key_name_ph: 'new key name (letters, digits, spaces)', create: 'Create', col_roles: 'Roles', col_created: 'Created', rotate: 'Rotate',
      rotate_name: 'Name of the replacement key', rotate_remember: 'Also remember the new key for this web UI (one-click login)?', rotate_delete: 'Delete the old key "{name}" right away?\nEvery app still using it stops working until you update it. Choose Cancel to delete it later.',
      confirm_delete_key: 'Delete the API key "{name}"? Every app using it stops working.', give_name: 'Give the key a name',
      new_key: 'New key "{name}" ({roles}):', key_not_returned: '(Unraid did not return the key)', copy_now: 'Copy it now: it is not shown again.', remembered: 'Remembered for this web UI.', old_deleted: 'Old key deleted.', old_not_deleted: 'Old key NOT deleted: {err}', not_remembered: 'Not remembered: {err}',
      for_the_app: 'For the apps', server_url: 'Server URL to enter in Unraid Drive:', tls_ok: 'Served over HTTPS: the Files app extension will accept this URL.', tls_no: 'You are on plain HTTP. iOS extensions require a valid HTTPS certificate: put Nginx Proxy Manager, SWAG or Cloudflare Tunnel in front before using the app remotely.',
      graphql: 'GraphQL proxy', run: 'Run', docs: 'Documentation',
      session_expired: 'session expired', device_revoked: 'this browser was removed from the gateway: sign in again to register it', remember_failed: 'Could not remember the key: {err}', confirm_forget: 'Delete the API key stored on the gateway? You will paste it again next time.',
      key_label: 'api key', user_scope_note: 'per-user permissions',
      health_open: 'Open the Health page', notify_unraid: 'Unraid notifications', notify_unraid_on: "Send to Unraid's notification system", notify_unraid_key: 'ADMIN API key (only if the web UI key is VIEWER; leave empty when /tmp/notifications is mounted)', notify_mail: 'E-mail (SMTP)', notify_smtp_host: 'Server', notify_smtp_port: 'Port', notify_smtp_tls: 'Encryption', none: 'none', notify_smtp_user: 'User', notify_smtp_password: 'Password', notify_smtp_from: 'From', notify_smtp_to: 'To (comma separated)', notify_tg_token: 'Bot token', notify_tg_chat: 'Chat id', save: 'Save', notify_reset: 'Back to the container variables', notify_source_file: 'Channels set from this page (/config/notify.json); the container variables are ignored.', notify_source_env: 'Channels from the container variables; saving here overrides them (kept in /config/notify.json).', notify_saved: 'Saved.', notify_admin_only: 'Only an ADMIN key or an administrator can change the channels.', secret_set: 'set — leave empty to keep', confirm_notify_reset: 'Forget the channels saved here and go back to the container variables?', tab_health: 'Health', donut_checks: 'Checks', donut_devices: 'Devices', donut_memory: 'Memory', donut_shares: 'Shares', load_title: 'Unraid load', load_nokey: 'no key remembered on the gateway: sampling is off (WEBUI_API_KEY or Remember this key)', load_empty: 'No samples yet: the gateway reads the load every minute while it runs.', traffic_title: 'Traffic through the gateway', legend_cpu: 'CPU %', legend_mem: 'memory %', legend_requests: 'requests', read_write: 'read/write', mounted: 'mounted', n_devices_seen: '{n} device(s)', users_devices_title: 'Devices of {name}', logins_note: 'since gateway 0.16 one sign-in per device every 10 minutes is counted (the File Provider used to sign in at every request)', pm_account_ph: 'e.g. unraid-gateway at home', filter_ph: 'filter…', search_ph: 'search in the log…', col_unraid_path: 'Unraid path', unraid_path_title: 'Where the file lives on the Unraid server (user share path)', all_devices: 'All devices', all_events: 'All events', show_more: 'Show more ({n} left)', n_events: '{n} events', no_user_group: 'no user (API key only)',
      conflict_newest: 'conflicts: newest wins', conflict_keepBoth: 'conflicts: keep both', conflict_keepExisting: 'conflicts: keep existing', conflict_overwrite: 'conflicts: device wins', conflict_ask: 'conflicts: ask', hash_check: 'identical files skipped',
    },
    it: {
      tab_status: 'Stato', tab_settings: 'Impostazioni', health_title: 'Stato di salute', health_checks: 'Tutti i controlli', health_ok: 'Tutto regolare', health_warning: 'Avvisi', health_error: 'Problemi', health_unknown: 'Controllo…', health_unavailable: 'stato non disponibile: {err}', checked_at: 'controllato {when}', sys_title: 'Informazioni di sistema', sys_unavailable: 'informazioni di sistema non disponibili: {err}', sys_host: 'Host', sys_os: 'Unraid', sys_kernel: 'Kernel', sys_uptime: 'Acceso dal', sys_cpu: 'Processore', sys_cores: 'Core / thread', sys_clock: 'Frequenza', sys_memory: 'Memoria', sys_memory_use: 'In uso', sys_board: 'Scheda madre', sys_versions: 'API Unraid', sys_gpu: 'Grafica', sys_network: 'Rete', appearance: 'Aspetto', language: 'Lingua', theme: 'Tema',
      connecting: 'connessione…', online: 'gateway online', unreachable: 'gateway non raggiungibile',
      system: 'Sistema', theme_system: 'Tema: sistema', theme_light: 'Tema: chiaro', theme_dark: 'Tema: scuro',
      forget_key: 'Dimentica la chiave salvata', forget_key_title: 'Elimina la chiave salvata sul gateway', disconnect: 'Disconnetti',
      login_title: 'Collegati', login_intro: 'Usa una chiave API di Unraid (Settings → Management Access → API Keys). VIEWER basta per i file; ADMIN gestisce anche il server e le sue chiavi. La chiave è verificata da Unraid e scambiata con un token di sessione.',
      stored_connect: 'Collegati con la chiave salvata su questo gateway', stored_env: 'dalla variabile WEBUI_API_KEY', stored_file: 'ricordata su questo gateway',
      pm_account: 'Nome account (per il gestore password)', api_key: 'Chiave API di Unraid', api_key_ph: 'incolla la chiave API di Unraid',
      remember_key: 'Ricorda questa chiave sul gateway (/config/webui.key): chiunque possa aprire questa pagina potrebbe usarla — tieni il gateway in LAN o dietro Cloudflare Access',
      connect: 'Collegati', as_user: 'Accedi come utente Unraid…', user_intro: 'Con utente e password Unraid il gateway applica i permessi delle share di quell\'utente (pubblica / sicura / privata, lettura o scrittura), come via SMB. La chiave API qui sopra resta necessaria.',
      unraid_user: 'Utente Unraid', unraid_password: 'Password Unraid', connect_as_user: 'Collegati come questo utente', need_key: 'Incolla prima la chiave API (o usa quella salvata sul gateway).',
      stat_connected: 'Collegati ora', stat_streams: 'Stream / trasferimenti', stat_devices: 'Dispositivi registrati', stat_gateway: 'Gateway',
      tab_files: 'File', tab_activity: 'Attività', tab_devices: 'Dispositivi', tab_notify: 'Notifiche', tab_keys: 'Chiavi API', tab_advanced: 'Avanzate',
      refresh: 'Aggiorna', new_folder: 'Nuova cartella', upload: 'Carica', shares: 'share', read_only: 'sola lettura',
      root_hint: 'Queste sono le share montate nel container. Aprine una per caricare file o creare cartelle. L\'accesso ai file è deciso dai volumi (lettura/scrittura o sola lettura) e, con un utente Unraid, dai permessi delle sue share.',
      col_name: 'Nome', col_size: 'Dimensione', col_modified: 'Modificato', empty_folder: 'Cartella vuota.', rename: 'Rinomina', delete: 'Elimina',
      new_name: 'Nuovo nome', folder_name: 'Nome cartella', open_share_first: 'La radice elenca solo le share montate. Apri prima una share.',
      confirm_delete_file: 'Eliminare il file "{name}"?', confirm_delete_folder: 'Eliminare la cartella "{name}" e tutto il suo contenuto?', upload_failed: 'caricamento fallito: {name}',
      live: 'in diretta', connected_devices: 'Dispositivi collegati', col_device: 'Dispositivo', col_user: 'Utente', col_from: 'Da', col_since: 'Dal', col_last_seen: 'Ultimo accesso', col_requests: 'Richieste', col_last_file: 'Ultimo file',
      nobody_connected: 'Nessuno collegato adesso (le sessioni spariscono 30 minuti dopo l\'ultima richiesta).', streams_title: 'Stream e trasferimenti in corso',
      col_what: 'Cosa', col_file: 'File', col_who: 'Chi', col_progress: 'Avanzamento', col_speed: 'Velocità', col_elapsed: 'Trascorso', nothing_active: 'Niente in riproduzione o in trasferimento.',
      recent_transfers: 'Trasferimenti recenti', col_bytes: 'Byte', col_ended: 'Terminato', scope_user: 'solo i tuoi dispositivi', scope_all: 'tutti gli utenti', activity_unavailable: 'attività non disponibile: {err}',
      k_download: '⬇︎ download', k_stream: '▶︎ stream', k_upload: '⬆︎ upload', now: 'adesso', s_ago: '{n} s fa', min_ago: '{n} min fa', h_ago: '{n} h fa',
      dev_older: 'Unraid Drive (build precedente)', dev_browser: 'Browser', dev_mpv: 'lettore mpv', unknown: 'sconosciuto',
      remove_all: 'Rimuovi tutti', devices_intro: 'Ogni installazione di Unraid Drive si registra quando aggiungi il server. Rimuovere un dispositivo chiude le sue sessioni e l\'app chiede di accedere di nuovo (si registra da capo e parte una notifica). Un\'app reinstallata sostituisce la voce precedente, mostrata in grigio come "old".',
      col_key: 'Chiave', col_registered: 'Registrato', col_logins: 'Accessi (uno ogni 10 min)', no_devices: 'Nessun dispositivo registrato (le build vecchie dell\'app non si registrano: aggiorna alla 1.3 build 29 o successiva).',
      this_browser: '(questo browser)', old: 'vecchio', last_seen_prefix: 'ultimo accesso', replaced_title: 'Sostituito da una nuova installazione il {date}. Si può rimuovere.', registration_off: 'registrazione disattivata (DEVICE_REGISTRATION=off)', devices_unavailable: 'dispositivi non disponibili: {err}',
      users_title: 'Utenti', users_intro: 'Una riga per utente Unraid (o chiave API) con dispositivi registrati: quanti dispositivi e sincronizzazioni, ultimo accesso, dati movimentati oggi. Apri un utente per i suoi dispositivi, le sincronizzazioni di cartelle e i backup di ogni dispositivo (da mettere in pausa o riprendere), i file movimentati e l\'attività nel tempo.', col_user: 'Utente', col_devices: 'Dispositivi', col_syncs: 'Sync', col_today: 'Oggi', no_users: 'Nessun utente: compaiono quando i loro dispositivi si registrano.', users_unavailable: 'utenti non disponibili: {err}', open: 'Apri', n_paused: '{n} in pausa', n_warnings: '{n} avvisi', all_fine: 'Tutto regolare', user_devices: 'Dispositivi', user_syncs: 'Sincronizzazioni di cartelle e backup', user_syncs_intro: 'Definite dalle app su ogni dispositivo. Mettere in pausa una sincronizzazione qui la ferma sul dispositivo alla prossima esecuzione (l\'app la mostra in pausa) finché non la riprendi.', no_syncs: 'Nessuna sincronizzazione sui dispositivi di questo utente.', no_syncs_device: 'Nessuna sincronizzazione su questo dispositivo.', user_log_intro: 'tutti i dispositivi di questo utente, dal più recente', kind_folder: 'Sync cartella', kind_photos: 'Backup Foto', kind_downloads: 'Download', mode_up: 'dispositivo → Unraid', mode_down: 'Unraid → dispositivo', mode_both: 'bidirezionale', sync_paused: 'in pausa', sync_paused_by: 'in pausa da {who}', sync_never: 'mai eseguita', sync_failed: 'ultima esecuzione fallita', sync_partial: 'ultima esecuzione con {c} conflitti, {e} errori', sync_ok: 'ultima esecuzione regolare', sync_last: 'ultima esecuzione {when}: {n} file, {bytes}', sync_files_btn: 'File ({n})', sync_no_files: 'Nessun file movimentato.', col_file: 'File', col_direction: 'Direzione', col_outcome: 'Esito', dir_up: 'caricato', dir_down: 'scaricato', dir_delete: 'eliminato', dir_conflict: 'conflitto', out_ok: 'ok', out_conflict: 'conflitto', out_failed: 'fallito', out_skipped: 'saltato', syncs_unavailable: 'sincronizzazioni non disponibili: {err}', ev_sync: 'Sincronizzazione',
      active_today: 'attivo oggi', stale: 'non visto da {n} giorni', last_sync: 'Ultima sincronizzazione', prune_label: 'Rimuovi i dispositivi non visti da', prune_days: 'giorni', prune_run: 'Rimuovi i dispositivi inattivi', confirm_prune: 'Rimuovere tutti i dispositivi non visti da {n} giorni? Le loro app chiederanno di accedere di nuovo se tornano.', pruned: '{n} dispositivo/i rimosso/i:\n{list}', pruned_none: 'Nessun dispositivo così vecchio.', back: 'Indietro', charts_title: 'Dati movimentati', range_24h: 'Ultime 24 ore', range_7d: 'Ultimi 7 giorni', range_30d: 'Ultimi 30 giorni', range_90d: 'Ultimi 90 giorni', chart_empty: 'Nessun dato movimentato nel periodo.', log_title: 'Registro', log_intro: 'accessi, registrazioni, rimozioni e trasferimenti interrotti', col_when: 'Quando', col_event: 'Evento', col_detail: 'Dettaglio', log_empty: 'Nessun evento registrato.', state: 'Stato', tot_today: 'Oggi', tot_week: 'Ultimi 7 giorni', tot_month: 'Ultimi 30 giorni', tot_detail: '{req} richieste · {dl} download/stream · {ul} upload', metrics_unavailable: 'metriche non disponibili: {err}', legend_down: 'scaricato / in streaming', legend_up: 'caricato', ev_login: 'Accesso', ev_register: 'Registrazione', ev_revoked: 'Revocato', ev_transfer: 'Trasferimento', ev_admin: 'Amministrazione', log_unavailable: 'registro non disponibile: {err}',
      remove: 'Rimuovi', confirm_remove_device: 'Rimuovere "{name}"?\nLe sue sessioni vengono chiuse e l\'app chiederà di accedere di nuovo.', confirm_remove_this: '\n\nÈ il browser che stai usando: verrai disconnesso.', confirm_remove_all: 'Rimuovere TUTTI i dispositivi? Ogni app (e questo browser) dovrà accedere di nuovo.',
      send_test: 'Invia una prova', notify_intro: 'Un dispositivo nuovo o rimosso genera una notifica su ogni canale configurato: notifiche di Unraid (NOTIFY_UNRAID — imposta NOTIFY_UNRAID_API_KEY a una chiave ADMIN se la tua è VIEWER), e-mail (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_FROM, SMTP_TO, SMTP_TLS) e Telegram (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID). Si impostano nelle impostazioni del container.',
      channels: 'canali: {list}', no_channel: 'nessun canale configurato', sending: 'invio…', no_channel_long: 'Nessun canale configurato: imposta NOTIFY_UNRAID / SMTP_* / TELEGRAM_* nel container.', errors: 'Errori: {list}', sent_to: 'Inviata a {list}.', error: 'errore: {err}',
      needs_admin: '(serve una chiave ADMIN)', keys_intro: 'Elenca, crea, elimina e ruota le chiavi API del tuo server Unraid. Una chiave nuova viene mostrata una sola volta: copiala nelle app (Modifica server o credenziali). Ruota = crea una sostituta, eventualmente la ricorda per questa web UI ed elimina la vecchia.',
      list_keys: 'Elenca chiavi', key_name_ph: 'nome della nuova chiave (lettere, cifre, spazi)', create: 'Crea', col_roles: 'Ruoli', col_created: 'Creata', rotate: 'Ruota',
      rotate_name: 'Nome della chiave sostitutiva', rotate_remember: 'Ricordare la nuova chiave anche per questa web UI (accesso con un clic)?', rotate_delete: 'Eliminare subito la vecchia chiave "{name}"?\nOgni app che la usa ancora smette di funzionare finché non la aggiorni. Scegli Annulla per eliminarla dopo.',
      confirm_delete_key: 'Eliminare la chiave API "{name}"? Ogni app che la usa smette di funzionare.', give_name: 'Dai un nome alla chiave',
      new_key: 'Nuova chiave "{name}" ({roles}):', key_not_returned: '(Unraid non ha restituito la chiave)', copy_now: 'Copiala adesso: non verrà mostrata di nuovo.', remembered: 'Ricordata per questa web UI.', old_deleted: 'Vecchia chiave eliminata.', old_not_deleted: 'Vecchia chiave NON eliminata: {err}', not_remembered: 'Non ricordata: {err}',
      for_the_app: 'Per le app', server_url: 'URL del server da inserire in Unraid Drive:', tls_ok: 'Servito in HTTPS: l\'estensione dell\'app File accetterà questo URL.', tls_no: 'Sei in HTTP semplice. Le estensioni iOS richiedono un certificato HTTPS valido: metti davanti Nginx Proxy Manager, SWAG o Cloudflare Tunnel prima di usare l\'app da fuori.',
      graphql: 'Proxy GraphQL', run: 'Esegui', docs: 'Documentazione',
      session_expired: 'sessione scaduta', device_revoked: 'questo browser è stato rimosso dal gateway: accedi di nuovo per registrarlo', remember_failed: 'Impossibile ricordare la chiave: {err}', confirm_forget: 'Eliminare la chiave API salvata sul gateway? La incollerai di nuovo la prossima volta.',
      key_label: 'chiave api', user_scope_note: 'permessi per utente',
      health_open: 'Apri la pagina Stato di salute', notify_unraid: 'Notifiche di Unraid', notify_unraid_on: 'Invia al sistema di notifiche di Unraid', notify_unraid_key: 'Chiave API ADMIN (solo se la chiave della web UI è VIEWER; vuota se /tmp/notifications è montato)', notify_mail: 'E-mail (SMTP)', notify_smtp_host: 'Server', notify_smtp_port: 'Porta', notify_smtp_tls: 'Cifratura', none: 'nessuna', notify_smtp_user: 'Utente', notify_smtp_password: 'Password', notify_smtp_from: 'Mittente', notify_smtp_to: 'Destinatari (separati da virgola)', notify_tg_token: 'Token del bot', notify_tg_chat: 'Chat id', save: 'Salva', notify_reset: 'Torna alle variabili del container', notify_source_file: 'Canali impostati da questa pagina (/config/notify.json); le variabili del container sono ignorate.', notify_source_env: 'Canali dalle variabili del container; salvando qui le sostituisci (restano in /config/notify.json).', notify_saved: 'Salvato.', notify_admin_only: 'Solo una chiave ADMIN o un amministratore può cambiare i canali.', secret_set: 'impostata — lascia vuoto per mantenerla', confirm_notify_reset: 'Dimenticare i canali salvati qui e tornare alle variabili del container?', tab_health: 'Salute', donut_checks: 'Controlli', donut_devices: 'Dispositivi', donut_memory: 'Memoria', donut_shares: 'Share', load_title: 'Carico di Unraid', load_nokey: 'nessuna chiave ricordata sul gateway: campionamento spento (WEBUI_API_KEY o «Ricorda questa chiave»)', load_empty: 'Ancora nessun campione: il gateway legge il carico ogni minuto mentre è in esecuzione.', traffic_title: 'Traffico attraverso il gateway', legend_cpu: 'CPU %', legend_mem: 'memoria %', legend_requests: 'richieste', read_write: 'lettura/scrittura', mounted: 'montate', n_devices_seen: '{n} dispositivi', users_devices_title: 'Dispositivi di {name}', logins_note: 'dal gateway 0.16 si conta un accesso per dispositivo ogni 10 minuti (il File Provider accedeva a ogni richiesta)', pm_account_ph: 'es. unraid-gateway di casa', filter_ph: 'filtra…', search_ph: 'cerca nel registro…', col_unraid_path: 'Percorso Unraid', unraid_path_title: 'Dove sta il file sul server Unraid (percorso della share utente)', all_devices: 'Tutti i dispositivi', all_events: 'Tutti gli eventi', show_more: 'Mostra altri ({n} rimasti)', n_events: '{n} eventi', no_user_group: 'senza utente (solo chiave API)',
      conflict_newest: 'conflitti: vince il più recente', conflict_keepBoth: 'conflitti: tieni entrambi', conflict_keepExisting: 'conflitti: tieni l\'esistente', conflict_overwrite: 'conflitti: vince il dispositivo', conflict_ask: 'conflitti: chiedi', hash_check: 'file identici saltati',
      pm_account: 'Nome account (per il gestore password — modificabile)',
    },
    es: {
      tab_status: 'Estado', tab_settings: 'Ajustes', health_title: 'Salud', health_checks: 'Todas las comprobaciones', health_ok: 'Todo correcto', health_warning: 'Avisos', health_error: 'Problemas', health_unknown: 'Comprobando…', health_unavailable: 'estado no disponible: {err}', checked_at: 'comprobado {when}', sys_title: 'Información del sistema', sys_unavailable: 'información del sistema no disponible: {err}', sys_host: 'Host', sys_os: 'Unraid', sys_kernel: 'Kernel', sys_uptime: 'Encendido desde', sys_cpu: 'Procesador', sys_cores: 'Núcleos / hilos', sys_clock: 'Frecuencia', sys_memory: 'Memoria', sys_memory_use: 'En uso', sys_board: 'Placa base', sys_versions: 'API de Unraid', sys_gpu: 'Gráficos', sys_network: 'Red', appearance: 'Apariencia', language: 'Idioma', theme: 'Tema',
      connecting: 'conectando…', online: 'gateway en línea', unreachable: 'gateway inalcanzable',
      system: 'Sistema', theme_system: 'Tema: sistema', theme_light: 'Tema: claro', theme_dark: 'Tema: oscuro',
      forget_key: 'Olvidar la clave guardada', forget_key_title: 'Eliminar la clave guardada en el gateway', disconnect: 'Desconectar',
      login_title: 'Conectar', login_intro: 'Usa una clave API de Unraid (Settings → Management Access → API Keys). VIEWER basta para los archivos; ADMIN también gestiona el servidor y sus claves. Unraid valida la clave y la cambia por un token de sesión.',
      stored_connect: 'Conectar con la clave guardada en este gateway', stored_env: 'de la variable WEBUI_API_KEY', stored_file: 'recordada en este gateway',
      pm_account: 'Nombre de cuenta (para tu gestor de contraseñas)', api_key: 'Clave API de Unraid', api_key_ph: 'pega tu clave API de Unraid',
      remember_key: 'Recordar esta clave en el gateway (/config/webui.key): cualquiera que abra esta página podría usarla — mantén el gateway en la LAN o detrás de Cloudflare Access',
      connect: 'Conectar', as_user: 'Iniciar sesión como usuario de Unraid…', user_intro: 'Con usuario y contraseña de Unraid el gateway aplica los permisos de recursos compartidos de ese usuario (público / seguro / privado, lectura o escritura), igual que por SMB. La clave API sigue siendo necesaria.',
      unraid_user: 'Usuario de Unraid', unraid_password: 'Contraseña de Unraid', connect_as_user: 'Conectar como este usuario', need_key: 'Pega primero la clave API (o usa la guardada en el gateway).',
      stat_connected: 'Conectados ahora', stat_streams: 'Streams / transferencias', stat_devices: 'Dispositivos registrados', stat_gateway: 'Gateway',
      tab_files: 'Archivos', tab_activity: 'Actividad', tab_devices: 'Dispositivos', tab_notify: 'Notificaciones', tab_keys: 'Claves API', tab_advanced: 'Avanzado',
      refresh: 'Actualizar', new_folder: 'Nueva carpeta', upload: 'Subir', shares: 'recursos', read_only: 'solo lectura',
      root_hint: 'Estos son los recursos compartidos montados en el contenedor. Abre uno para subir archivos o crear carpetas. El acceso lo deciden los volúmenes (lectura/escritura o solo lectura) y, con un usuario de Unraid, sus permisos.',
      col_name: 'Nombre', col_size: 'Tamaño', col_modified: 'Modificado', empty_folder: 'Carpeta vacía.', rename: 'Renombrar', delete: 'Eliminar',
      new_name: 'Nuevo nombre', folder_name: 'Nombre de la carpeta', open_share_first: 'La raíz solo lista los recursos montados. Abre uno primero.',
      confirm_delete_file: '¿Eliminar el archivo "{name}"?', confirm_delete_folder: '¿Eliminar la carpeta "{name}" y todo su contenido?', upload_failed: 'subida fallida: {name}',
      live: 'en vivo', connected_devices: 'Dispositivos conectados', col_device: 'Dispositivo', col_user: 'Usuario', col_from: 'Desde', col_since: 'Desde', col_last_seen: 'Última vez', col_requests: 'Peticiones', col_last_file: 'Último archivo',
      nobody_connected: 'Nadie conectado ahora (las sesiones desaparecen 30 minutos después de la última petición).', streams_title: 'Streams y transferencias en curso',
      col_what: 'Qué', col_file: 'Archivo', col_who: 'Quién', col_progress: 'Progreso', col_speed: 'Velocidad', col_elapsed: 'Transcurrido', nothing_active: 'Nada reproduciéndose ni transfiriéndose.',
      recent_transfers: 'Transferencias recientes', col_bytes: 'Bytes', col_ended: 'Finalizado', scope_user: 'solo tus dispositivos', scope_all: 'todos los usuarios', activity_unavailable: 'actividad no disponible: {err}',
      k_download: '⬇︎ descarga', k_stream: '▶︎ stream', k_upload: '⬆︎ subida', now: 'ahora', s_ago: 'hace {n} s', min_ago: 'hace {n} min', h_ago: 'hace {n} h',
      dev_older: 'Unraid Drive (build anterior)', dev_browser: 'Navegador', dev_mpv: 'reproductor mpv', unknown: 'desconocido',
      remove_all: 'Eliminar todos', devices_intro: 'Cada instalación de Unraid Drive se registra al añadir el servidor. Eliminar un dispositivo cierra sus sesiones y la app pide iniciar sesión de nuevo (se registra otra vez y envía una notificación). Una app reinstalada sustituye su entrada anterior, mostrada en gris como "old".',
      col_key: 'Clave', col_registered: 'Registrado', col_logins: 'Accesos', no_devices: 'Ningún dispositivo registrado aún (las builds antiguas no se registran; actualiza a 1.3 build 29 o posterior).',
      this_browser: '(este navegador)', old: 'antiguo', last_seen_prefix: 'última vez', replaced_title: 'Sustituido por una nueva instalación el {date}. Se puede eliminar.', registration_off: 'registro desactivado (DEVICE_REGISTRATION=off)', devices_unavailable: 'dispositivos no disponibles: {err}',
      users_title: 'Usuarios', users_intro: 'Una fila por usuario de Unraid (o clave API) con dispositivos registrados: cuántos dispositivos y sincronizaciones, última vez visto, datos movidos hoy. Abre un usuario para ver sus dispositivos, las sincronizaciones de carpetas y copias de cada dispositivo (pausarlas o reanudarlas), los archivos movidos y su actividad en el tiempo.', col_user: 'Usuario', col_devices: 'Dispositivos', col_syncs: 'Sincr.', col_today: 'Hoy', no_users: 'Ningún usuario aún: aparecen cuando sus dispositivos se registran.', users_unavailable: 'usuarios no disponibles: {err}', open: 'Abrir', n_paused: '{n} en pausa', n_warnings: '{n} aviso(s)', all_fine: 'Todo correcto', user_devices: 'Dispositivos', user_syncs: 'Sincronizaciones de carpetas y copias', user_syncs_intro: 'Definidas por las apps en cada dispositivo. Pausar una aquí la detiene en el dispositivo en su siguiente ejecución (la app la muestra en pausa) hasta que la reanudes.', no_syncs: 'Ninguna sincronización en los dispositivos de este usuario.', no_syncs_device: 'Ninguna sincronización en este dispositivo.', user_log_intro: 'todos los dispositivos de este usuario, más recientes primero', kind_folder: 'Sincronización de carpeta', kind_photos: 'Copia de Fotos', kind_downloads: 'Descargas', mode_up: 'dispositivo → Unraid', mode_down: 'Unraid → dispositivo', mode_both: 'bidireccional', sync_paused: 'en pausa', sync_paused_by: 'en pausa por {who}', sync_never: 'nunca ejecutada', sync_failed: 'última ejecución fallida', sync_partial: 'última ejecución con {c} conflictos, {e} errores', sync_ok: 'última ejecución correcta', sync_last: 'última ejecución {when}: {n} archivos, {bytes}', sync_files_btn: 'Archivos ({n})', sync_no_files: 'Ningún archivo movido aún.', col_file: 'Archivo', col_direction: 'Dirección', col_outcome: 'Resultado', dir_up: 'subido', dir_down: 'descargado', dir_delete: 'eliminado', dir_conflict: 'conflicto', out_ok: 'ok', out_conflict: 'conflicto', out_failed: 'fallido', out_skipped: 'omitido', syncs_unavailable: 'sincronizaciones no disponibles: {err}', ev_sync: 'Sincronización',
      active_today: 'activo hoy', stale: 'sin actividad desde hace {n} días', last_sync: 'Última sincronización', prune_label: 'Eliminar los dispositivos sin actividad desde hace', prune_days: 'días', prune_run: 'Eliminar dispositivos inactivos', confirm_prune: '¿Eliminar todos los dispositivos sin actividad desde hace {n} días? Sus apps pedirán iniciar sesión de nuevo si vuelven.', pruned: '{n} dispositivo(s) eliminado(s):\n{list}', pruned_none: 'Ningún dispositivo era tan antiguo.', back: 'Atrás', charts_title: 'Datos movidos', range_24h: 'Últimas 24 horas', range_7d: 'Últimos 7 días', range_30d: 'Últimos 30 días', range_90d: 'Últimos 90 días', chart_empty: 'Nada movido en este periodo.', log_title: 'Registro', log_intro: 'accesos, registros, eliminaciones y transferencias interrumpidas', col_when: 'Cuándo', col_event: 'Evento', col_detail: 'Detalle', log_empty: 'Ningún evento registrado aún.', state: 'Estado', tot_today: 'Hoy', tot_week: 'Últimos 7 días', tot_month: 'Últimos 30 días', tot_detail: '{req} peticiones · {dl} descargas/streams · {ul} subidas', metrics_unavailable: 'métricas no disponibles: {err}', legend_down: 'descargado / en streaming', legend_up: 'subido', ev_login: 'Acceso', ev_register: 'Registro', ev_revoked: 'Revocado', ev_transfer: 'Transferencia', ev_admin: 'Administración', log_unavailable: 'registro no disponible: {err}',
      remove: 'Eliminar', confirm_remove_device: '¿Eliminar "{name}"?\nSus sesiones se cierran y la app pedirá iniciar sesión de nuevo.', confirm_remove_this: '\n\nEs el navegador que estás usando: se cerrará tu sesión.', confirm_remove_all: '¿Eliminar TODOS los dispositivos? Cada app (y este navegador) tendrá que iniciar sesión de nuevo.',
      send_test: 'Enviar una prueba', notify_intro: 'Un dispositivo nuevo o eliminado genera una notificación en cada canal configurado: notificaciones de Unraid (NOTIFY_UNRAID — pon NOTIFY_UNRAID_API_KEY con una clave ADMIN si la tuya es VIEWER), correo (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_FROM, SMTP_TO, SMTP_TLS) y Telegram (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID). Se configuran en los ajustes del contenedor.',
      channels: 'canales: {list}', no_channel: 'ningún canal configurado', sending: 'enviando…', no_channel_long: 'Ningún canal configurado: define NOTIFY_UNRAID / SMTP_* / TELEGRAM_* en el contenedor.', errors: 'Errores: {list}', sent_to: 'Enviada a {list}.', error: 'error: {err}',
      needs_admin: '(requiere una clave ADMIN)', keys_intro: 'Lista, crea, elimina y rota las claves API de tu servidor Unraid. Una clave nueva se muestra una sola vez: cópiala en las apps (Editar servidor o credenciales). Rotar = crear una sustituta, opcionalmente recordarla para esta web UI y eliminar la antigua.',
      list_keys: 'Listar claves', key_name_ph: 'nombre de la nueva clave (letras, dígitos, espacios)', create: 'Crear', col_roles: 'Roles', col_created: 'Creada', rotate: 'Rotar',
      rotate_name: 'Nombre de la clave sustituta', rotate_remember: '¿Recordar también la nueva clave para esta web UI (acceso con un clic)?', rotate_delete: '¿Eliminar ya la clave antigua "{name}"?\nToda app que aún la use dejará de funcionar hasta que la actualices. Elige Cancelar para eliminarla más tarde.',
      confirm_delete_key: '¿Eliminar la clave API "{name}"? Toda app que la use dejará de funcionar.', give_name: 'Dale un nombre a la clave',
      new_key: 'Nueva clave "{name}" ({roles}):', key_not_returned: '(Unraid no devolvió la clave)', copy_now: 'Cópiala ahora: no se mostrará de nuevo.', remembered: 'Recordada para esta web UI.', old_deleted: 'Clave antigua eliminada.', old_not_deleted: 'Clave antigua NO eliminada: {err}', not_remembered: 'No recordada: {err}',
      for_the_app: 'Para las apps', server_url: 'URL del servidor para Unraid Drive:', tls_ok: 'Servido por HTTPS: la extensión de Archivos aceptará esta URL.', tls_no: 'Estás en HTTP plano. Las extensiones de iOS requieren un certificado HTTPS válido: pon Nginx Proxy Manager, SWAG o Cloudflare Tunnel delante antes de usar la app desde fuera.',
      graphql: 'Proxy GraphQL', run: 'Ejecutar', docs: 'Documentación',
      session_expired: 'sesión caducada', device_revoked: 'este navegador fue eliminado del gateway: inicia sesión de nuevo para registrarlo', remember_failed: 'No se pudo recordar la clave: {err}', confirm_forget: '¿Eliminar la clave API guardada en el gateway? La pegarás de nuevo la próxima vez.',
      key_label: 'clave api', user_scope_note: 'permisos por usuario',
    },
    fr: {
      tab_status: 'État', tab_settings: 'Réglages', health_title: 'État de santé', health_checks: 'Toutes les vérifications', health_ok: 'Tout va bien', health_warning: 'Avertissements', health_error: 'Problèmes', health_unknown: 'Vérification…', health_unavailable: 'état indisponible : {err}', checked_at: 'vérifié {when}', sys_title: 'Informations système', sys_unavailable: 'informations système indisponibles : {err}', sys_host: 'Hôte', sys_os: 'Unraid', sys_kernel: 'Noyau', sys_uptime: 'Allumé depuis', sys_cpu: 'Processeur', sys_cores: 'Cœurs / threads', sys_clock: 'Fréquence', sys_memory: 'Mémoire', sys_memory_use: 'Utilisée', sys_board: 'Carte mère', sys_versions: 'API Unraid', sys_gpu: 'Carte graphique', sys_network: 'Réseau', appearance: 'Apparence', language: 'Langue', theme: 'Thème',
      connecting: 'connexion…', online: 'passerelle en ligne', unreachable: 'passerelle injoignable',
      system: 'Système', theme_system: 'Thème : système', theme_light: 'Thème : clair', theme_dark: 'Thème : sombre',
      forget_key: 'Oublier la clé enregistrée', forget_key_title: 'Supprimer la clé enregistrée sur la passerelle', disconnect: 'Déconnecter',
      login_title: 'Connexion', login_intro: 'Utilisez une clé API Unraid (Settings → Management Access → API Keys). VIEWER suffit pour les fichiers ; ADMIN gère aussi le serveur et ses clés. La clé est validée par Unraid et échangée contre un jeton de session.',
      stored_connect: 'Se connecter avec la clé enregistrée sur cette passerelle', stored_env: 'depuis la variable WEBUI_API_KEY', stored_file: 'mémorisée sur cette passerelle',
      pm_account: 'Nom du compte (pour votre gestionnaire de mots de passe)', api_key: 'Clé API Unraid', api_key_ph: 'collez votre clé API Unraid',
      remember_key: 'Mémoriser cette clé sur la passerelle (/config/webui.key) : quiconque ouvre cette page pourrait l\'utiliser — gardez la passerelle sur le LAN ou derrière Cloudflare Access',
      connect: 'Se connecter', as_user: 'Se connecter comme utilisateur Unraid…', user_intro: 'Avec un utilisateur et un mot de passe Unraid, la passerelle applique les permissions de partage de cet utilisateur (public / sécurisé / privé, lecture ou écriture), comme via SMB. La clé API ci-dessus reste nécessaire.',
      unraid_user: 'Utilisateur Unraid', unraid_password: 'Mot de passe Unraid', connect_as_user: 'Se connecter comme cet utilisateur', need_key: 'Collez d\'abord la clé API (ou utilisez celle enregistrée sur la passerelle).',
      stat_connected: 'Connectés maintenant', stat_streams: 'Flux / transferts', stat_devices: 'Appareils enregistrés', stat_gateway: 'Passerelle',
      tab_files: 'Fichiers', tab_activity: 'Activité', tab_devices: 'Appareils', tab_notify: 'Notifications', tab_keys: 'Clés API', tab_advanced: 'Avancé',
      refresh: 'Actualiser', new_folder: 'Nouveau dossier', upload: 'Envoyer', shares: 'partages', read_only: 'lecture seule',
      root_hint: 'Voici les partages montés dans le conteneur. Ouvrez-en un pour envoyer des fichiers ou créer des dossiers. L\'accès dépend des volumes (lecture/écriture ou lecture seule) et, avec un utilisateur Unraid, de ses permissions.',
      col_name: 'Nom', col_size: 'Taille', col_modified: 'Modifié', empty_folder: 'Dossier vide.', rename: 'Renommer', delete: 'Supprimer',
      new_name: 'Nouveau nom', folder_name: 'Nom du dossier', open_share_first: 'La racine ne liste que les partages montés. Ouvrez d\'abord un partage.',
      confirm_delete_file: 'Supprimer le fichier « {name} » ?', confirm_delete_folder: 'Supprimer le dossier « {name} » et tout son contenu ?', upload_failed: 'envoi échoué : {name}',
      live: 'direct', connected_devices: 'Appareils connectés', col_device: 'Appareil', col_user: 'Utilisateur', col_from: 'Depuis', col_since: 'Depuis', col_last_seen: 'Vu', col_requests: 'Requêtes', col_last_file: 'Dernier fichier',
      nobody_connected: 'Personne n\'est connecté (les sessions disparaissent 30 minutes après la dernière requête).', streams_title: 'Flux et transferts en cours',
      col_what: 'Quoi', col_file: 'Fichier', col_who: 'Qui', col_progress: 'Progression', col_speed: 'Vitesse', col_elapsed: 'Écoulé', nothing_active: 'Rien en lecture ni en transfert.',
      recent_transfers: 'Transferts récents', col_bytes: 'Octets', col_ended: 'Terminé', scope_user: 'vos appareils seulement', scope_all: 'tous les utilisateurs', activity_unavailable: 'activité indisponible : {err}',
      k_download: '⬇︎ téléchargement', k_stream: '▶︎ flux', k_upload: '⬆︎ envoi', now: 'maintenant', s_ago: 'il y a {n} s', min_ago: 'il y a {n} min', h_ago: 'il y a {n} h',
      dev_older: 'Unraid Drive (build ancienne)', dev_browser: 'Navigateur', dev_mpv: 'lecteur mpv', unknown: 'inconnu',
      remove_all: 'Tout supprimer', devices_intro: 'Chaque installation d\'Unraid Drive s\'enregistre quand vous ajoutez le serveur. Supprimer un appareil ferme ses sessions et l\'app demande de se reconnecter (elle se réenregistre et une notification part). Une app réinstallée remplace son ancienne entrée, grisée « old ».',
      col_key: 'Clé', col_registered: 'Enregistré', col_logins: 'Connexions', no_devices: 'Aucun appareil enregistré (les anciennes builds ne s\'enregistrent pas ; mettez à jour vers 1.3 build 29 ou plus).',
      this_browser: '(ce navigateur)', old: 'ancien', last_seen_prefix: 'vu', replaced_title: 'Remplacé par une nouvelle installation le {date}. Peut être supprimé.', registration_off: 'enregistrement désactivé (DEVICE_REGISTRATION=off)', devices_unavailable: 'appareils indisponibles : {err}',
      users_title: 'Utilisateurs', users_intro: 'Une ligne par utilisateur Unraid (ou clé API) ayant des appareils enregistrés : nombre d’appareils et de synchronisations, dernière activité, données transférées aujourd’hui. Ouvrez un utilisateur pour ses appareils, les synchronisations de dossiers et sauvegardes de chaque appareil (à suspendre ou reprendre), les fichiers transférés et son activité dans le temps.', col_user: 'Utilisateur', col_devices: 'Appareils', col_syncs: 'Synchros', col_today: 'Aujourd’hui', no_users: 'Aucun utilisateur : ils apparaissent quand leurs appareils s’enregistrent.', users_unavailable: 'utilisateurs indisponibles : {err}', open: 'Ouvrir', n_paused: '{n} en pause', n_warnings: '{n} avertissement(s)', all_fine: 'Tout va bien', user_devices: 'Appareils', user_syncs: 'Synchronisations de dossiers et sauvegardes', user_syncs_intro: 'Définies par les apps sur chaque appareil. Suspendre une synchronisation ici l’arrête sur l’appareil à sa prochaine exécution (l’app l’affiche en pause) jusqu’à sa reprise.', no_syncs: 'Aucune synchronisation sur les appareils de cet utilisateur.', no_syncs_device: 'Aucune synchronisation sur cet appareil.', user_log_intro: 'tous les appareils de cet utilisateur, du plus récent', kind_folder: 'Synchro de dossier', kind_photos: 'Sauvegarde Photos', kind_downloads: 'Téléchargements', mode_up: 'appareil → Unraid', mode_down: 'Unraid → appareil', mode_both: 'bidirectionnelle', sync_paused: 'en pause', sync_paused_by: 'en pause par {who}', sync_never: 'jamais exécutée', sync_failed: 'dernière exécution échouée', sync_partial: 'dernière exécution avec {c} conflits, {e} erreurs', sync_ok: 'dernière exécution réussie', sync_last: 'dernière exécution {when} : {n} fichiers, {bytes}', sync_files_btn: 'Fichiers ({n})', sync_no_files: 'Aucun fichier transféré.', col_file: 'Fichier', col_direction: 'Sens', col_outcome: 'Résultat', dir_up: 'envoyé', dir_down: 'téléchargé', dir_delete: 'supprimé', dir_conflict: 'conflit', out_ok: 'ok', out_conflict: 'conflit', out_failed: 'échec', out_skipped: 'ignoré', syncs_unavailable: 'synchronisations indisponibles : {err}', ev_sync: 'Synchronisation',
      active_today: 'actif aujourd’hui', stale: 'pas vu depuis {n} jours', last_sync: 'Dernière synchronisation', prune_label: 'Supprimer les appareils pas vus depuis', prune_days: 'jours', prune_run: 'Supprimer les appareils inactifs', confirm_prune: 'Supprimer tous les appareils pas vus depuis {n} jours ? Leurs apps redemanderont de se connecter si elles reviennent.', pruned: '{n} appareil(s) supprimé(s) :\n{list}', pruned_none: 'Aucun appareil aussi ancien.', back: 'Retour', charts_title: 'Données transférées', range_24h: 'Dernières 24 heures', range_7d: '7 derniers jours', range_30d: '30 derniers jours', range_90d: '90 derniers jours', chart_empty: 'Rien transféré sur cette période.', log_title: 'Journal', log_intro: 'connexions, enregistrements, suppressions et transferts interrompus', col_when: 'Quand', col_event: 'Événement', col_detail: 'Détail', log_empty: 'Aucun événement enregistré.', state: 'État', tot_today: 'Aujourd’hui', tot_week: '7 derniers jours', tot_month: '30 derniers jours', tot_detail: '{req} requêtes · {dl} téléchargements/flux · {ul} envois', metrics_unavailable: 'métriques indisponibles : {err}', legend_down: 'téléchargé / diffusé', legend_up: 'envoyé', ev_login: 'Connexion', ev_register: 'Enregistrement', ev_revoked: 'Révoqué', ev_transfer: 'Transfert', ev_admin: 'Administration', log_unavailable: 'journal indisponible : {err}',
      remove: 'Supprimer', confirm_remove_device: 'Supprimer « {name} » ?\nSes sessions sont fermées et l\'app demandera de se reconnecter.', confirm_remove_this: '\n\nC\'est le navigateur que vous utilisez : vous serez déconnecté.', confirm_remove_all: 'Supprimer TOUS les appareils ? Chaque app (et ce navigateur) devra se reconnecter.',
      send_test: 'Envoyer un test', notify_intro: 'Un appareil nouveau ou supprimé déclenche une notification sur chaque canal configuré : notifications Unraid (NOTIFY_UNRAID — mettez NOTIFY_UNRAID_API_KEY à une clé ADMIN si la vôtre est VIEWER), e-mail (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_FROM, SMTP_TO, SMTP_TLS) et Telegram (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID). À régler dans les paramètres du conteneur.',
      channels: 'canaux : {list}', no_channel: 'aucun canal configuré', sending: 'envoi…', no_channel_long: 'Aucun canal configuré : définissez NOTIFY_UNRAID / SMTP_* / TELEGRAM_* dans le conteneur.', errors: 'Erreurs : {list}', sent_to: 'Envoyé à {list}.', error: 'erreur : {err}',
      needs_admin: '(nécessite une clé ADMIN)', keys_intro: 'Listez, créez, supprimez et renouvelez les clés API de votre serveur Unraid. Une nouvelle clé n\'est affichée qu\'une fois : copiez-la dans les apps (Modifier le serveur ou les identifiants). Renouveler = créer une remplaçante, éventuellement la mémoriser pour cette interface et supprimer l\'ancienne.',
      list_keys: 'Lister les clés', key_name_ph: 'nom de la nouvelle clé (lettres, chiffres, espaces)', create: 'Créer', col_roles: 'Rôles', col_created: 'Créée', rotate: 'Renouveler',
      rotate_name: 'Nom de la clé de remplacement', rotate_remember: 'Mémoriser aussi la nouvelle clé pour cette interface (connexion en un clic) ?', rotate_delete: 'Supprimer tout de suite l\'ancienne clé « {name} » ?\nToute app qui l\'utilise encore cessera de fonctionner jusqu\'à sa mise à jour. Choisissez Annuler pour la supprimer plus tard.',
      confirm_delete_key: 'Supprimer la clé API « {name} » ? Toute app qui l\'utilise cessera de fonctionner.', give_name: 'Donnez un nom à la clé',
      new_key: 'Nouvelle clé « {name} » ({roles}) :', key_not_returned: '(Unraid n\'a pas renvoyé la clé)', copy_now: 'Copiez-la maintenant : elle ne sera plus affichée.', remembered: 'Mémorisée pour cette interface.', old_deleted: 'Ancienne clé supprimée.', old_not_deleted: 'Ancienne clé NON supprimée : {err}', not_remembered: 'Non mémorisée : {err}',
      for_the_app: 'Pour les apps', server_url: 'URL du serveur à saisir dans Unraid Drive :', tls_ok: 'Servi en HTTPS : l\'extension Fichiers acceptera cette URL.', tls_no: 'Vous êtes en HTTP simple. Les extensions iOS exigent un certificat HTTPS valide : placez Nginx Proxy Manager, SWAG ou Cloudflare Tunnel devant avant d\'utiliser l\'app à distance.',
      graphql: 'Proxy GraphQL', run: 'Exécuter', docs: 'Documentation',
      session_expired: 'session expirée', device_revoked: 'ce navigateur a été supprimé de la passerelle : reconnectez-vous pour l\'enregistrer', remember_failed: 'Impossible de mémoriser la clé : {err}', confirm_forget: 'Supprimer la clé API enregistrée sur la passerelle ? Vous la collerez de nouveau la prochaine fois.',
      key_label: 'clé api', user_scope_note: 'permissions par utilisateur',
    },
    de: {
      tab_status: 'Status', tab_settings: 'Einstellungen', health_title: 'Zustand', health_checks: 'Alle Prüfungen', health_ok: 'Alles in Ordnung', health_warning: 'Warnungen', health_error: 'Probleme', health_unknown: 'Prüfe …', health_unavailable: 'Zustand nicht verfügbar: {err}', checked_at: 'geprüft {when}', sys_title: 'Systeminformationen', sys_unavailable: 'Systeminformationen nicht verfügbar: {err}', sys_host: 'Host', sys_os: 'Unraid', sys_kernel: 'Kernel', sys_uptime: 'Läuft seit', sys_cpu: 'Prozessor', sys_cores: 'Kerne / Threads', sys_clock: 'Takt', sys_memory: 'Arbeitsspeicher', sys_memory_use: 'In Benutzung', sys_board: 'Mainboard', sys_versions: 'Unraid-API', sys_gpu: 'Grafik', sys_network: 'Netzwerk', appearance: 'Darstellung', language: 'Sprache', theme: 'Design',
      connecting: 'verbinde…', online: 'Gateway online', unreachable: 'Gateway nicht erreichbar',
      system: 'System', theme_system: 'Design: System', theme_light: 'Design: hell', theme_dark: 'Design: dunkel',
      forget_key: 'Gespeicherten Schlüssel vergessen', forget_key_title: 'Den auf dem Gateway gespeicherten Schlüssel löschen', disconnect: 'Trennen',
      login_title: 'Verbinden', login_intro: 'Verwende einen Unraid-API-Schlüssel (Settings → Management Access → API Keys). VIEWER reicht für Dateien; ADMIN verwaltet auch den Server und seine Schlüssel. Unraid prüft den Schlüssel und tauscht ihn gegen ein Sitzungstoken.',
      stored_connect: 'Mit dem auf diesem Gateway gespeicherten Schlüssel verbinden', stored_env: 'aus der Variable WEBUI_API_KEY', stored_file: 'auf diesem Gateway gemerkt',
      pm_account: 'Kontoname (für deinen Passwortmanager)', api_key: 'Unraid-API-Schlüssel', api_key_ph: 'Unraid-API-Schlüssel einfügen',
      remember_key: 'Diesen Schlüssel auf dem Gateway merken (/config/webui.key): wer diese Seite öffnen kann, könnte ihn nutzen — Gateway im LAN lassen oder hinter Cloudflare Access',
      connect: 'Verbinden', as_user: 'Als Unraid-Benutzer anmelden…', user_intro: 'Mit Unraid-Benutzer und Passwort wendet das Gateway die Freigabe-Berechtigungen dieses Benutzers an (öffentlich / sicher / privat, Lesen oder Schreiben), wie über SMB. Der API-Schlüssel oben bleibt nötig.',
      unraid_user: 'Unraid-Benutzer', unraid_password: 'Unraid-Passwort', connect_as_user: 'Als dieser Benutzer verbinden', need_key: 'Zuerst den API-Schlüssel einfügen (oder den auf dem Gateway gespeicherten verwenden).',
      stat_connected: 'Jetzt verbunden', stat_streams: 'Streams / Übertragungen', stat_devices: 'Registrierte Geräte', stat_gateway: 'Gateway',
      tab_files: 'Dateien', tab_activity: 'Aktivität', tab_devices: 'Geräte', tab_notify: 'Mitteilungen', tab_keys: 'API-Schlüssel', tab_advanced: 'Erweitert',
      refresh: 'Aktualisieren', new_folder: 'Neuer Ordner', upload: 'Hochladen', shares: 'Freigaben', read_only: 'nur lesen',
      root_hint: 'Das sind die im Container eingehängten Freigaben. Öffne eine, um Dateien hochzuladen oder Ordner anzulegen. Der Zugriff hängt von den Volumes ab (Lesen/Schreiben oder nur Lesen) und, mit einem Unraid-Benutzer, von dessen Berechtigungen.',
      col_name: 'Name', col_size: 'Größe', col_modified: 'Geändert', empty_folder: 'Leerer Ordner.', rename: 'Umbenennen', delete: 'Löschen',
      new_name: 'Neuer Name', folder_name: 'Ordnername', open_share_first: 'Die Wurzel listet nur eingehängte Freigaben. Öffne zuerst eine Freigabe.',
      confirm_delete_file: 'Datei „{name}“ löschen?', confirm_delete_folder: 'Ordner „{name}“ mit allem Inhalt löschen?', upload_failed: 'Hochladen fehlgeschlagen: {name}',
      live: 'live', connected_devices: 'Verbundene Geräte', col_device: 'Gerät', col_user: 'Benutzer', col_from: 'Von', col_since: 'Seit', col_last_seen: 'Zuletzt', col_requests: 'Anfragen', col_last_file: 'Letzte Datei',
      nobody_connected: 'Gerade niemand verbunden (Sitzungen verschwinden 30 Minuten nach der letzten Anfrage).', streams_title: 'Laufende Streams und Übertragungen',
      col_what: 'Was', col_file: 'Datei', col_who: 'Wer', col_progress: 'Fortschritt', col_speed: 'Tempo', col_elapsed: 'Dauer', nothing_active: 'Nichts läuft oder wird übertragen.',
      recent_transfers: 'Letzte Übertragungen', col_bytes: 'Bytes', col_ended: 'Beendet', scope_user: 'nur deine Geräte', scope_all: 'alle Benutzer', activity_unavailable: 'Aktivität nicht verfügbar: {err}',
      k_download: '⬇︎ Download', k_stream: '▶︎ Stream', k_upload: '⬆︎ Upload', now: 'jetzt', s_ago: 'vor {n} s', min_ago: 'vor {n} min', h_ago: 'vor {n} h',
      dev_older: 'Unraid Drive (ältere Build)', dev_browser: 'Browser', dev_mpv: 'mpv-Player', unknown: 'unbekannt',
      remove_all: 'Alle entfernen', devices_intro: 'Jede Unraid-Drive-Installation registriert sich beim Hinzufügen des Servers. Entfernen schließt die Sitzungen des Geräts und die App fragt erneut nach der Anmeldung (registriert sich neu und löst eine Mitteilung aus). Eine neu installierte App ersetzt ihren alten Eintrag, grau als „old“ gezeigt.',
      col_key: 'Schlüssel', col_registered: 'Registriert', col_logins: 'Anmeldungen', no_devices: 'Noch kein Gerät registriert (ältere App-Builds registrieren sich nicht; auf 1.3 Build 29 oder neuer aktualisieren).',
      this_browser: '(dieser Browser)', old: 'alt', last_seen_prefix: 'zuletzt', replaced_title: 'Am {date} durch eine neue Installation ersetzt. Kann entfernt werden.', registration_off: 'Registrierung aus (DEVICE_REGISTRATION=off)', devices_unavailable: 'Geräte nicht verfügbar: {err}',
      users_title: 'Benutzer', users_intro: 'Eine Zeile pro Unraid-Benutzer (oder API-Schlüssel) mit registrierten Geräten: Anzahl Geräte und Synchronisierungen, zuletzt gesehen, heute übertragene Daten. Öffne einen Benutzer für seine Geräte, die Ordner-Synchronisierungen und Backups jedes Geräts (pausieren oder fortsetzen), die übertragenen Dateien und seine Aktivität im Zeitverlauf.', col_user: 'Benutzer', col_devices: 'Geräte', col_syncs: 'Syncs', col_today: 'Heute', no_users: 'Noch kein Benutzer: sie erscheinen, sobald ihre Geräte sich registrieren.', users_unavailable: 'Benutzer nicht verfügbar: {err}', open: 'Öffnen', n_paused: '{n} pausiert', n_warnings: '{n} Warnung(en)', all_fine: 'Alles in Ordnung', user_devices: 'Geräte', user_syncs: 'Ordner-Synchronisierungen und Backups', user_syncs_intro: 'Von den Apps auf jedem Gerät definiert. Hier pausiert, stoppt eine Synchronisierung beim nächsten Lauf auf dem Gerät (die App zeigt sie pausiert), bis du sie fortsetzt.', no_syncs: 'Keine Synchronisierung auf den Geräten dieses Benutzers.', no_syncs_device: 'Keine Synchronisierung auf diesem Gerät.', user_log_intro: 'alle Geräte dieses Benutzers, neueste zuerst', kind_folder: 'Ordner-Sync', kind_photos: 'Fotos-Backup', kind_downloads: 'Downloads', mode_up: 'Gerät → Unraid', mode_down: 'Unraid → Gerät', mode_both: 'beidseitig', sync_paused: 'pausiert', sync_paused_by: 'pausiert von {who}', sync_never: 'noch nie gelaufen', sync_failed: 'letzter Lauf fehlgeschlagen', sync_partial: 'letzter Lauf mit {c} Konflikten, {e} Fehlern', sync_ok: 'letzter Lauf in Ordnung', sync_last: 'letzter Lauf {when}: {n} Dateien, {bytes}', sync_files_btn: 'Dateien ({n})', sync_no_files: 'Noch keine Datei übertragen.', col_file: 'Datei', col_direction: 'Richtung', col_outcome: 'Ergebnis', dir_up: 'hochgeladen', dir_down: 'heruntergeladen', dir_delete: 'gelöscht', dir_conflict: 'Konflikt', out_ok: 'ok', out_conflict: 'Konflikt', out_failed: 'fehlgeschlagen', out_skipped: 'übersprungen', syncs_unavailable: 'Synchronisierungen nicht verfügbar: {err}', ev_sync: 'Synchronisierung',
      active_today: 'heute aktiv', stale: 'seit {n} Tagen nicht gesehen', last_sync: 'Letzte Synchronisierung', prune_label: 'Geräte entfernen, die nicht gesehen wurden seit', prune_days: 'Tagen', prune_run: 'Inaktive Geräte entfernen', confirm_prune: 'Alle Geräte entfernen, die seit {n} Tagen nicht gesehen wurden? Ihre Apps fragen bei der Rückkehr erneut nach der Anmeldung.', pruned: '{n} Gerät(e) entfernt:\n{list}', pruned_none: 'Kein Gerät war so alt.', back: 'Zurück', charts_title: 'Übertragene Daten', range_24h: 'Letzte 24 Stunden', range_7d: 'Letzte 7 Tage', range_30d: 'Letzte 30 Tage', range_90d: 'Letzte 90 Tage', chart_empty: 'In diesem Zeitraum nichts übertragen.', log_title: 'Protokoll', log_intro: 'Anmeldungen, Registrierungen, Entfernungen und abgebrochene Übertragungen', col_when: 'Wann', col_event: 'Ereignis', col_detail: 'Details', log_empty: 'Noch keine Ereignisse.', state: 'Zustand', tot_today: 'Heute', tot_week: 'Letzte 7 Tage', tot_month: 'Letzte 30 Tage', tot_detail: '{req} Anfragen · {dl} Downloads/Streams · {ul} Uploads', metrics_unavailable: 'Messwerte nicht verfügbar: {err}', legend_down: 'heruntergeladen / gestreamt', legend_up: 'hochgeladen', ev_login: 'Anmeldung', ev_register: 'Registrierung', ev_revoked: 'Widerrufen', ev_transfer: 'Übertragung', ev_admin: 'Verwaltung', log_unavailable: 'Protokoll nicht verfügbar: {err}',
      remove: 'Entfernen', confirm_remove_device: '„{name}“ entfernen?\nSeine Sitzungen werden geschlossen und die App fragt erneut nach der Anmeldung.', confirm_remove_this: '\n\nDas ist der Browser, den du gerade benutzt: du wirst abgemeldet.', confirm_remove_all: 'ALLE Geräte entfernen? Jede App (und dieser Browser) muss sich neu anmelden.',
      send_test: 'Test senden', notify_intro: 'Ein neues oder entferntes Gerät löst auf jedem konfigurierten Kanal eine Mitteilung aus: Unraid-Benachrichtigungen (NOTIFY_UNRAID — NOTIFY_UNRAID_API_KEY auf einen ADMIN-Schlüssel setzen, wenn deiner VIEWER ist), E-Mail (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_FROM, SMTP_TO, SMTP_TLS) und Telegram (TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID). In den Container-Einstellungen setzen.',
      channels: 'Kanäle: {list}', no_channel: 'kein Kanal konfiguriert', sending: 'sende…', no_channel_long: 'Kein Kanal konfiguriert: NOTIFY_UNRAID / SMTP_* / TELEGRAM_* im Container setzen.', errors: 'Fehler: {list}', sent_to: 'Gesendet an {list}.', error: 'Fehler: {err}',
      needs_admin: '(braucht einen ADMIN-Schlüssel)', keys_intro: 'API-Schlüssel deines Unraid-Servers auflisten, anlegen, löschen und rotieren. Ein neuer Schlüssel wird nur einmal gezeigt: in die Apps kopieren (Server oder Zugangsdaten bearbeiten). Rotieren = Ersatz anlegen, optional für diese Web-UI merken und den alten löschen.',
      list_keys: 'Schlüssel auflisten', key_name_ph: 'Name des neuen Schlüssels (Buchstaben, Ziffern, Leerzeichen)', create: 'Anlegen', col_roles: 'Rollen', col_created: 'Erstellt', rotate: 'Rotieren',
      rotate_name: 'Name des Ersatzschlüssels', rotate_remember: 'Den neuen Schlüssel auch für diese Web-UI merken (Ein-Klick-Anmeldung)?', rotate_delete: 'Den alten Schlüssel „{name}“ sofort löschen?\nJede App, die ihn noch nutzt, funktioniert nicht mehr, bis du sie aktualisierst. Abbrechen wählen, um ihn später zu löschen.',
      confirm_delete_key: 'API-Schlüssel „{name}“ löschen? Jede App, die ihn nutzt, funktioniert nicht mehr.', give_name: 'Gib dem Schlüssel einen Namen',
      new_key: 'Neuer Schlüssel „{name}“ ({roles}):', key_not_returned: '(Unraid hat den Schlüssel nicht zurückgegeben)', copy_now: 'Jetzt kopieren: er wird nicht noch einmal gezeigt.', remembered: 'Für diese Web-UI gemerkt.', old_deleted: 'Alter Schlüssel gelöscht.', old_not_deleted: 'Alter Schlüssel NICHT gelöscht: {err}', not_remembered: 'Nicht gemerkt: {err}',
      for_the_app: 'Für die Apps', server_url: 'Server-URL für Unraid Drive:', tls_ok: 'Über HTTPS ausgeliefert: die Dateien-Erweiterung akzeptiert diese URL.', tls_no: 'Du bist auf reinem HTTP. iOS-Erweiterungen brauchen ein gültiges HTTPS-Zertifikat: Nginx Proxy Manager, SWAG oder Cloudflare Tunnel davorschalten, bevor du die App von unterwegs nutzt.',
      graphql: 'GraphQL-Proxy', run: 'Ausführen', docs: 'Dokumentation',
      session_expired: 'Sitzung abgelaufen', device_revoked: 'dieser Browser wurde vom Gateway entfernt: erneut anmelden, um ihn zu registrieren', remember_failed: 'Schlüssel konnte nicht gemerkt werden: {err}', confirm_forget: 'Den auf dem Gateway gespeicherten API-Schlüssel löschen? Beim nächsten Mal fügst du ihn wieder ein.',
      key_label: 'API-Schlüssel', user_scope_note: 'Berechtigungen pro Benutzer',
    },
    'zh-Hans': {
      tab_status: '状态', tab_settings: '设置', health_title: '健康状态', health_checks: '全部检查', health_ok: '一切正常', health_warning: '警告', health_error: '问题', health_unknown: '检查中…', health_unavailable: '状态不可用：{err}', checked_at: '检查于 {when}', sys_title: '系统信息', sys_unavailable: '系统信息不可用：{err}', sys_host: '主机', sys_os: 'Unraid', sys_kernel: '内核', sys_uptime: '开机时间', sys_cpu: '处理器', sys_cores: '核心 / 线程', sys_clock: '频率', sys_memory: '内存', sys_memory_use: '使用中', sys_board: '主板', sys_versions: 'Unraid API', sys_gpu: '显卡', sys_network: '网络', appearance: '外观', language: '语言', theme: '主题',
      connecting: '正在连接…', online: '网关在线', unreachable: '无法连接网关',
      system: '系统', theme_system: '主题：跟随系统', theme_light: '主题：浅色', theme_dark: '主题：深色',
      forget_key: '忘记已保存的密钥', forget_key_title: '删除保存在网关上的密钥', disconnect: '断开',
      login_title: '连接', login_intro: '使用 Unraid API 密钥（Settings → Management Access → API Keys）。VIEWER 足以访问文件；ADMIN 还可管理服务器及其密钥。密钥由 Unraid 验证并换取会话令牌。',
      stored_connect: '使用保存在此网关上的密钥连接', stored_env: '来自 WEBUI_API_KEY 变量', stored_file: '已保存在此网关',
      pm_account: '账户名（供密码管理器使用）', api_key: 'Unraid API 密钥', api_key_ph: '粘贴你的 Unraid API 密钥',
      remember_key: '在网关上记住此密钥（/config/webui.key）：任何能打开此页面的人都可能使用它——请将网关保留在局域网内或置于 Cloudflare Access 之后',
      connect: '连接', as_user: '以 Unraid 用户登录…', user_intro: '提供 Unraid 用户名和密码后，网关将应用该用户的共享权限（公开 / 安全 / 私有，读或写），与 SMB 一致。上面的 API 密钥仍然必需。',
      unraid_user: 'Unraid 用户', unraid_password: 'Unraid 密码', connect_as_user: '以此用户连接', need_key: '请先粘贴 API 密钥（或使用网关上保存的密钥）。',
      stat_connected: '当前连接', stat_streams: '串流 / 传输', stat_devices: '已注册设备', stat_gateway: '网关',
      tab_files: '文件', tab_activity: '活动', tab_devices: '设备', tab_notify: '通知', tab_keys: 'API 密钥', tab_advanced: '高级',
      refresh: '刷新', new_folder: '新建文件夹', upload: '上传', shares: '共享', read_only: '只读',
      root_hint: '这些是挂载到容器中的共享。打开一个以上传文件或创建文件夹。文件访问由卷挂载（读写或只读）决定；使用 Unraid 用户时还受其共享权限约束。',
      col_name: '名称', col_size: '大小', col_modified: '修改时间', empty_folder: '文件夹为空。', rename: '重命名', delete: '删除',
      new_name: '新名称', folder_name: '文件夹名称', open_share_first: '根目录只列出已挂载的共享。请先打开一个共享。',
      confirm_delete_file: '删除文件“{name}”？', confirm_delete_folder: '删除文件夹“{name}”及其全部内容？', upload_failed: '上传失败：{name}',
      live: '实时', connected_devices: '已连接的设备', col_device: '设备', col_user: '用户', col_from: '来源', col_since: '起始', col_last_seen: '最近活动', col_requests: '请求数', col_last_file: '最近文件',
      nobody_connected: '当前无人连接（会话在最后一次请求 30 分钟后消失）。', streams_title: '进行中的串流与传输',
      col_what: '类型', col_file: '文件', col_who: '谁', col_progress: '进度', col_speed: '速度', col_elapsed: '已用时', nothing_active: '没有正在播放或传输的内容。',
      recent_transfers: '最近的传输', col_bytes: '字节', col_ended: '结束', scope_user: '仅你的设备', scope_all: '所有用户', activity_unavailable: '活动不可用：{err}',
      k_download: '⬇︎ 下载', k_stream: '▶︎ 串流', k_upload: '⬆︎ 上传', now: '刚刚', s_ago: '{n} 秒前', min_ago: '{n} 分钟前', h_ago: '{n} 小时前',
      dev_older: 'Unraid Drive（旧版本）', dev_browser: '浏览器', dev_mpv: 'mpv 播放器', unknown: '未知',
      remove_all: '全部移除', devices_intro: '每个 Unraid Drive 安装在添加服务器时都会自行注册。移除设备会关闭其会话，应用会要求重新登录（重新注册并发送通知）。重装的应用会取代旧条目，旧条目以灰色“old”显示。',
      col_key: '密钥', col_registered: '注册时间', col_logins: '登录次数', no_devices: '尚无已注册设备（旧版应用不会注册；请更新到 1.3 build 29 或更高）。',
      this_browser: '（此浏览器）', old: '旧', last_seen_prefix: '最近活动', replaced_title: '已于 {date} 被新安装取代。可以安全移除。', registration_off: '注册已关闭（DEVICE_REGISTRATION=off）', devices_unavailable: '设备不可用：{err}',
      users_title: '用户', users_intro: '每个拥有已注册设备的 Unraid 用户（或 API 密钥）一行：设备和同步数量、最近出现时间、今天传输的数据。打开用户可查看其设备、每台设备运行的文件夹同步和备份（可暂停或恢复）、传输的文件以及该用户随时间的活动。', col_user: '用户', col_devices: '设备', col_syncs: '同步', col_today: '今天', no_users: '尚无用户：设备注册后会出现。', users_unavailable: '用户不可用：{err}', open: '打开', n_paused: '{n} 个已暂停', n_warnings: '{n} 条警告', all_fine: '一切正常', user_devices: '设备', user_syncs: '文件夹同步与备份', user_syncs_intro: '由每台设备上的 App 定义。在此暂停同步后，设备会在下次运行时停止（App 显示为已暂停），直到你恢复。', no_syncs: '该用户的设备上没有定义同步。', no_syncs_device: '此设备上没有定义同步。', user_log_intro: '该用户的所有设备，最新在前', kind_folder: '文件夹同步', kind_photos: '照片备份', kind_downloads: '下载', mode_up: '设备 → Unraid', mode_down: 'Unraid → 设备', mode_both: '双向', sync_paused: '已暂停', sync_paused_by: '由 {who} 暂停', sync_never: '尚未运行', sync_failed: '上次运行失败', sync_partial: '上次运行有 {c} 个冲突、{e} 个错误', sync_ok: '上次运行正常', sync_last: '上次运行 {when}：{n} 个文件，{bytes}', sync_files_btn: '文件（{n}）', sync_no_files: '尚未传输文件。', col_file: '文件', col_direction: '方向', col_outcome: '结果', dir_up: '已上传', dir_down: '已下载', dir_delete: '已删除', dir_conflict: '冲突', out_ok: '成功', out_conflict: '冲突', out_failed: '失败', out_skipped: '已跳过', syncs_unavailable: '同步不可用：{err}', ev_sync: '同步',
      active_today: '今天活跃', stale: '已 {n} 天未出现', last_sync: '上次同步', prune_label: '移除以下天数内未出现的设备：', prune_days: '天', prune_run: '移除不活跃设备', confirm_prune: '移除所有 {n} 天未出现的设备？它们的应用回来时需要重新登录。', pruned: '已移除 {n} 台设备：\n{list}', pruned_none: '没有这么久未出现的设备。', back: '返回', charts_title: '传输的数据', range_24h: '最近 24 小时', range_7d: '最近 7 天', range_30d: '最近 30 天', range_90d: '最近 90 天', chart_empty: '此期间没有传输。', log_title: '日志', log_intro: '登录、注册、移除和中断的传输', col_when: '时间', col_event: '事件', col_detail: '详情', log_empty: '尚无事件。', state: '状态', tot_today: '今天', tot_week: '最近 7 天', tot_month: '最近 30 天', tot_detail: '{req} 次请求 · {dl} 次下载/流 · {ul} 次上传', metrics_unavailable: '指标不可用：{err}', legend_down: '下载 / 流', legend_up: '上传', ev_login: '登录', ev_register: '注册', ev_revoked: '已撤销', ev_transfer: '传输', ev_admin: '管理', log_unavailable: '日志不可用：{err}',
      remove: '移除', confirm_remove_device: '移除“{name}”？\n其会话将被关闭，应用会要求重新登录。', confirm_remove_this: '\n\n这是你正在使用的浏览器：你将被登出。', confirm_remove_all: '移除所有设备？每个应用（以及此浏览器）都必须重新登录。',
      send_test: '发送测试', notify_intro: '新设备或被移除的设备会在每个已配置的渠道上发出通知：Unraid 通知（NOTIFY_UNRAID——若你的密钥是 VIEWER，请将 NOTIFY_UNRAID_API_KEY 设为 ADMIN 密钥）、电子邮件（SMTP_HOST、SMTP_PORT、SMTP_USER、SMTP_PASSWORD、SMTP_FROM、SMTP_TO、SMTP_TLS）和 Telegram（TELEGRAM_BOT_TOKEN、TELEGRAM_CHAT_ID）。在容器设置中配置。',
      channels: '渠道：{list}', no_channel: '未配置渠道', sending: '发送中…', no_channel_long: '未配置渠道：请在容器中设置 NOTIFY_UNRAID / SMTP_* / TELEGRAM_*。', errors: '错误：{list}', sent_to: '已发送至 {list}。', error: '错误：{err}',
      needs_admin: '（需要 ADMIN 密钥）', keys_intro: '列出、创建、删除和轮换你的 Unraid 服务器的 API 密钥。新密钥只显示一次：将其复制到应用中（编辑服务器或凭证）。轮换 = 创建替代密钥，可选地为此网页记住它，并删除旧密钥。',
      list_keys: '列出密钥', key_name_ph: '新密钥名称（字母、数字、空格）', create: '创建', col_roles: '角色', col_created: '创建时间', rotate: '轮换',
      rotate_name: '替代密钥的名称', rotate_remember: '是否也为此网页记住新密钥（一键登录）？', rotate_delete: '立即删除旧密钥“{name}”？\n仍在使用它的应用将停止工作，直到你更新它。选择“取消”以稍后删除。',
      confirm_delete_key: '删除 API 密钥“{name}”？使用它的所有应用都将停止工作。', give_name: '请为密钥命名',
      new_key: '新密钥“{name}”（{roles}）：', key_not_returned: '（Unraid 未返回密钥）', copy_now: '请立即复制：不会再次显示。', remembered: '已为此网页记住。', old_deleted: '旧密钥已删除。', old_not_deleted: '旧密钥未删除：{err}', not_remembered: '未记住：{err}',
      for_the_app: '供应用使用', server_url: '在 Unraid Drive 中输入的服务器地址：', tls_ok: '通过 HTTPS 提供：“文件”扩展将接受此地址。', tls_no: '你正在使用纯 HTTP。iOS 扩展需要有效的 HTTPS 证书：在远程使用应用前，请在前面部署 Nginx Proxy Manager、SWAG 或 Cloudflare Tunnel。',
      graphql: 'GraphQL 代理', run: '运行', docs: '文档',
      session_expired: '会话已过期', device_revoked: '此浏览器已从网关移除：请重新登录以注册', remember_failed: '无法记住密钥：{err}', confirm_forget: '删除保存在网关上的 API 密钥？下次需要重新粘贴。',
      key_label: 'API 密钥', user_scope_note: '按用户的权限',
    },
    ar: {
      tab_status: 'الحالة', tab_settings: 'الإعدادات', health_title: 'الحالة الصحية', health_checks: 'كل الفحوصات', health_ok: 'كل شيء على ما يرام', health_warning: 'تحذيرات', health_error: 'مشاكل', health_unknown: 'جارٍ الفحص…', health_unavailable: 'الحالة غير متاحة: {err}', checked_at: 'فُحص {when}', sys_title: 'معلومات النظام', sys_unavailable: 'معلومات النظام غير متاحة: {err}', sys_host: 'المضيف', sys_os: 'Unraid', sys_kernel: 'النواة', sys_uptime: 'يعمل منذ', sys_cpu: 'المعالج', sys_cores: 'الأنوية / الخيوط', sys_clock: 'التردد', sys_memory: 'الذاكرة', sys_memory_use: 'قيد الاستخدام', sys_board: 'اللوحة الأم', sys_versions: 'واجهة Unraid API', sys_gpu: 'الرسوميات', sys_network: 'الشبكة', appearance: 'المظهر', language: 'اللغة', theme: 'السمة',
      connecting: 'جارٍ الاتصال…', online: 'البوابة متصلة', unreachable: 'تعذّر الوصول إلى البوابة',
      system: 'النظام', theme_system: 'المظهر: النظام', theme_light: 'المظهر: فاتح', theme_dark: 'المظهر: داكن',
      forget_key: 'نسيان المفتاح المحفوظ', forget_key_title: 'حذف المفتاح المحفوظ على البوابة', disconnect: 'قطع الاتصال',
      login_title: 'الاتصال', login_intro: 'استخدم مفتاح API من Unraid (Settings → Management Access → API Keys). يكفي VIEWER للملفات؛ أما ADMIN فيدير الخادم ومفاتيحه أيضًا. يتحقق Unraid من المفتاح ويستبدله برمز جلسة.',
      stored_connect: 'الاتصال بالمفتاح المحفوظ على هذه البوابة', stored_env: 'من المتغير WEBUI_API_KEY', stored_file: 'محفوظ على هذه البوابة',
      pm_account: 'اسم الحساب (لمدير كلمات المرور)', api_key: 'مفتاح API لـ Unraid', api_key_ph: 'الصق مفتاح API الخاص بـ Unraid',
      remember_key: 'تذكّر هذا المفتاح على البوابة (/config/webui.key): يمكن لأي شخص يفتح هذه الصفحة استخدامه — أبقِ البوابة داخل الشبكة المحلية أو خلف Cloudflare Access',
      connect: 'اتصال', as_user: 'تسجيل الدخول كمستخدم Unraid…', user_intro: 'باستخدام مستخدم وكلمة مرور Unraid تطبّق البوابة صلاحيات المشاركات لذلك المستخدم (عامة / آمنة / خاصة، قراءة أو كتابة) تمامًا كما عبر SMB. ويبقى مفتاح API أعلاه مطلوبًا.',
      unraid_user: 'مستخدم Unraid', unraid_password: 'كلمة مرور Unraid', connect_as_user: 'الاتصال بهذا المستخدم', need_key: 'الصق مفتاح API أولًا (أو استخدم المفتاح المحفوظ على البوابة).',
      stat_connected: 'المتصلون الآن', stat_streams: 'البث / النقل', stat_devices: 'الأجهزة المسجّلة', stat_gateway: 'البوابة',
      tab_files: 'الملفات', tab_activity: 'النشاط', tab_devices: 'الأجهزة', tab_notify: 'الإشعارات', tab_keys: 'مفاتيح API', tab_advanced: 'متقدم',
      refresh: 'تحديث', new_folder: 'مجلد جديد', upload: 'رفع', shares: 'المشاركات', read_only: 'للقراءة فقط',
      root_hint: 'هذه هي المشاركات المركّبة داخل الحاوية. افتح إحداها لرفع الملفات أو إنشاء المجلدات. يحدد تركيب وحدات التخزين إمكانية الوصول (قراءة/كتابة أو قراءة فقط)، ومع مستخدم Unraid تُطبَّق صلاحياته أيضًا.',
      col_name: 'الاسم', col_size: 'الحجم', col_modified: 'آخر تعديل', empty_folder: 'المجلد فارغ.', rename: 'إعادة تسمية', delete: 'حذف',
      new_name: 'الاسم الجديد', folder_name: 'اسم المجلد', open_share_first: 'الجذر يعرض المشاركات المركّبة فقط. افتح مشاركة أولًا.',
      confirm_delete_file: 'حذف الملف "{name}"؟', confirm_delete_folder: 'حذف المجلد "{name}" وكل ما بداخله؟', upload_failed: 'فشل الرفع: {name}',
      live: 'مباشر', connected_devices: 'الأجهزة المتصلة', col_device: 'الجهاز', col_user: 'المستخدم', col_from: 'من', col_since: 'منذ', col_last_seen: 'آخر ظهور', col_requests: 'الطلبات', col_last_file: 'آخر ملف',
      nobody_connected: 'لا أحد متصل الآن (تختفي الجلسات بعد 30 دقيقة من آخر طلب).', streams_title: 'البث وعمليات النقل الجارية',
      col_what: 'ماذا', col_file: 'الملف', col_who: 'من', col_progress: 'التقدم', col_speed: 'السرعة', col_elapsed: 'المنقضي', nothing_active: 'لا شيء قيد التشغيل أو النقل.',
      recent_transfers: 'عمليات النقل الأخيرة', col_bytes: 'بايت', col_ended: 'انتهى', scope_user: 'أجهزتك فقط', scope_all: 'كل المستخدمين', activity_unavailable: 'النشاط غير متاح: {err}',
      k_download: '⬇︎ تنزيل', k_stream: '▶︎ بث', k_upload: '⬆︎ رفع', now: 'الآن', s_ago: 'قبل {n} ث', min_ago: 'قبل {n} د', h_ago: 'قبل {n} س',
      dev_older: 'Unraid Drive (إصدار أقدم)', dev_browser: 'متصفح', dev_mpv: 'مشغّل mpv', unknown: 'غير معروف',
      remove_all: 'إزالة الكل', devices_intro: 'يسجّل كل تثبيت لـ Unraid Drive نفسه عند إضافة الخادم. إزالة جهاز تُغلق جلساته ويطلب التطبيق تسجيل الدخول مجددًا (فيسجّل من جديد ويُرسل إشعار). التطبيق المعاد تثبيته يحلّ محل إدخاله السابق الذي يظهر رماديًا بعلامة "old".',
      col_key: 'المفتاح', col_registered: 'تاريخ التسجيل', col_logins: 'مرات الدخول', no_devices: 'لا يوجد جهاز مسجّل بعد (إصدارات التطبيق القديمة لا تسجّل؛ حدّث إلى 1.3 build 29 أو أحدث).',
      this_browser: '(هذا المتصفح)', old: 'قديم', last_seen_prefix: 'آخر ظهور', replaced_title: 'استُبدل بتثبيت جديد في {date}. يمكن إزالته بأمان.', registration_off: 'التسجيل متوقف (DEVICE_REGISTRATION=off)', devices_unavailable: 'الأجهزة غير متاحة: {err}',
      users_title: 'المستخدمون', users_intro: 'صف لكل مستخدم Unraid (أو مفتاح API) لديه أجهزة مسجلة: عدد الأجهزة والمزامنات، آخر ظهور، البيانات المنقولة اليوم. افتح مستخدمًا لعرض أجهزته ومزامنات المجلدات والنسخ الاحتياطية لكل جهاز (إيقافها مؤقتًا أو استئنافها) والملفات المنقولة ونشاطه عبر الزمن.', col_user: 'المستخدم', col_devices: 'الأجهزة', col_syncs: 'المزامنات', col_today: 'اليوم', no_users: 'لا مستخدمين بعد: يظهرون عند تسجيل أجهزتهم.', users_unavailable: 'المستخدمون غير متاحين: {err}', open: 'فتح', n_paused: '{n} متوقفة مؤقتًا', n_warnings: '{n} تحذير', all_fine: 'كل شيء على ما يرام', user_devices: 'الأجهزة', user_syncs: 'مزامنات المجلدات والنسخ الاحتياطية', user_syncs_intro: 'تحددها التطبيقات على كل جهاز. إيقاف مزامنة مؤقتًا هنا يوقفها على الجهاز عند التشغيل التالي (يعرضها التطبيق متوقفة) حتى تستأنفها.', no_syncs: 'لا مزامنات على أجهزة هذا المستخدم.', no_syncs_device: 'لا مزامنات على هذا الجهاز.', user_log_intro: 'كل أجهزة هذا المستخدم، الأحدث أولًا', kind_folder: 'مزامنة مجلد', kind_photos: 'نسخ الصور احتياطيًا', kind_downloads: 'التنزيلات', mode_up: 'الجهاز ← Unraid', mode_down: 'Unraid ← الجهاز', mode_both: 'ثنائية الاتجاه', sync_paused: 'متوقفة مؤقتًا', sync_paused_by: 'أوقفها {who} مؤقتًا', sync_never: 'لم تُشغّل بعد', sync_failed: 'فشل آخر تشغيل', sync_partial: 'آخر تشغيل مع {c} تعارضات و{e} أخطاء', sync_ok: 'آخر تشغيل سليم', sync_last: 'آخر تشغيل {when}: {n} ملفات، {bytes}', sync_files_btn: 'الملفات ({n})', sync_no_files: 'لم يُنقل أي ملف بعد.', col_file: 'الملف', col_direction: 'الاتجاه', col_outcome: 'النتيجة', dir_up: 'مرفوع', dir_down: 'منزّل', dir_delete: 'محذوف', dir_conflict: 'تعارض', out_ok: 'تم', out_conflict: 'تعارض', out_failed: 'فشل', out_skipped: 'تم تخطيه', syncs_unavailable: 'المزامنات غير متاحة: {err}', ev_sync: 'مزامنة',
      active_today: 'نشط اليوم', stale: 'لم يُر منذ {n} يومًا', last_sync: 'آخر مزامنة', prune_label: 'إزالة الأجهزة التي لم تُر منذ', prune_days: 'يومًا', prune_run: 'إزالة الأجهزة الخاملة', confirm_prune: 'إزالة كل الأجهزة التي لم تُر منذ {n} يومًا؟ ستطلب تطبيقاتها تسجيل الدخول مجددًا عند عودتها.', pruned: 'تمت إزالة {n} جهاز:\n{list}', pruned_none: 'لا يوجد جهاز بهذا القِدم.', back: 'رجوع', charts_title: 'البيانات المنقولة', range_24h: 'آخر 24 ساعة', range_7d: 'آخر 7 أيام', range_30d: 'آخر 30 يومًا', range_90d: 'آخر 90 يومًا', chart_empty: 'لم يُنقل شيء في هذه الفترة.', log_title: 'السجل', log_intro: 'تسجيلات الدخول والتسجيلات والإزالات والنقل المتوقف', col_when: 'متى', col_event: 'الحدث', col_detail: 'التفاصيل', log_empty: 'لا أحداث بعد.', state: 'الحالة', tot_today: 'اليوم', tot_week: 'آخر 7 أيام', tot_month: 'آخر 30 يومًا', tot_detail: '{req} طلبات · {dl} تنزيلات/بث · {ul} رفع', metrics_unavailable: 'المقاييس غير متاحة: {err}', legend_down: 'تم تنزيله / بثه', legend_up: 'تم رفعه', ev_login: 'تسجيل دخول', ev_register: 'تسجيل', ev_revoked: 'مُلغى', ev_transfer: 'نقل', ev_admin: 'إدارة', log_unavailable: 'السجل غير متاح: {err}',
      remove: 'إزالة', confirm_remove_device: 'إزالة "{name}"؟\nستُغلق جلساته وسيطلب التطبيق تسجيل الدخول مجددًا.', confirm_remove_this: '\n\nهذا هو المتصفح الذي تستخدمه: سيتم تسجيل خروجك.', confirm_remove_all: 'إزالة كل الأجهزة؟ سيتعين على كل تطبيق (وهذا المتصفح) تسجيل الدخول مجددًا.',
      send_test: 'إرسال اختبار', notify_intro: 'يُطلق الجهاز الجديد أو المُزال إشعارًا على كل قناة مضبوطة: إشعارات Unraid (NOTIFY_UNRAID — اضبط NOTIFY_UNRAID_API_KEY على مفتاح ADMIN إذا كان مفتاحك VIEWER)، البريد الإلكتروني (SMTP_HOST وSMTP_PORT وSMTP_USER وSMTP_PASSWORD وSMTP_FROM وSMTP_TO وSMTP_TLS) وTelegram (TELEGRAM_BOT_TOKEN وTELEGRAM_CHAT_ID). تُضبط في إعدادات الحاوية.',
      channels: 'القنوات: {list}', no_channel: 'لا توجد قناة مضبوطة', sending: 'جارٍ الإرسال…', no_channel_long: 'لا توجد قناة مضبوطة: اضبط NOTIFY_UNRAID / SMTP_* / TELEGRAM_* في الحاوية.', errors: 'أخطاء: {list}', sent_to: 'أُرسل إلى {list}.', error: 'خطأ: {err}',
      needs_admin: '(يتطلب مفتاح ADMIN)', keys_intro: 'اعرض مفاتيح API لخادم Unraid وأنشئها واحذفها ودوّرها. يُعرض المفتاح الجديد مرة واحدة: انسخه إلى التطبيقات (تعديل الخادم أو بيانات الاعتماد). التدوير = إنشاء بديل، وتذكّره اختياريًا لهذه الواجهة، وحذف القديم.',
      list_keys: 'عرض المفاتيح', key_name_ph: 'اسم المفتاح الجديد (حروف وأرقام ومسافات)', create: 'إنشاء', col_roles: 'الأدوار', col_created: 'أُنشئ', rotate: 'تدوير',
      rotate_name: 'اسم المفتاح البديل', rotate_remember: 'هل تريد أيضًا تذكّر المفتاح الجديد لهذه الواجهة (دخول بنقرة واحدة)؟', rotate_delete: 'حذف المفتاح القديم "{name}" فورًا؟\nسيتوقف كل تطبيق ما زال يستخدمه حتى تحدّثه. اختر إلغاء لحذفه لاحقًا.',
      confirm_delete_key: 'حذف مفتاح API "{name}"؟ سيتوقف كل تطبيق يستخدمه.', give_name: 'أعطِ المفتاح اسمًا',
      new_key: 'مفتاح جديد "{name}" ({roles}):', key_not_returned: '(لم يُعد Unraid المفتاح)', copy_now: 'انسخه الآن: لن يُعرض مرة أخرى.', remembered: 'تم تذكّره لهذه الواجهة.', old_deleted: 'حُذف المفتاح القديم.', old_not_deleted: 'لم يُحذف المفتاح القديم: {err}', not_remembered: 'لم يُتذكّر: {err}',
      for_the_app: 'للتطبيقات', server_url: 'عنوان الخادم لإدخاله في Unraid Drive:', tls_ok: 'يُقدَّم عبر HTTPS: ستقبل إضافة الملفات هذا العنوان.', tls_no: 'أنت على HTTP عادي. تتطلب إضافات iOS شهادة HTTPS صالحة: ضع Nginx Proxy Manager أو SWAG أو Cloudflare Tunnel أمامها قبل استخدام التطبيق عن بُعد.',
      graphql: 'وكيل GraphQL', run: 'تشغيل', docs: 'التوثيق',
      session_expired: 'انتهت الجلسة', device_revoked: 'أُزيل هذا المتصفح من البوابة: سجّل الدخول مجددًا لتسجيله', remember_failed: 'تعذّر تذكّر المفتاح: {err}', confirm_forget: 'حذف مفتاح API المحفوظ على البوابة؟ ستلصقه مجددًا في المرة القادمة.',
      key_label: 'مفتاح api', user_scope_note: 'صلاحيات لكل مستخدم',
    },
  };
  const LANGS = ['en', 'it', 'es', 'fr', 'de', 'zh-Hans', 'ar'];
  const pickLang = () => {
    const pref = localStorage.getItem('ugw.lang') || 'system';
    if (pref !== 'system' && I18N[pref]) return pref;
    for (const l of navigator.languages || [navigator.language || 'en']) {
      const low = l.toLowerCase();
      if (low.startsWith('zh')) return 'zh-Hans';
      const two = low.slice(0, 2);
      if (LANGS.includes(two)) return two;
    }
    return 'en';
  };
  let lang = pickLang();
  const t = (key, vars) => {
    let s = (I18N[lang] && I18N[lang][key]) || I18N.en[key] || key;
    if (vars) for (const k of Object.keys(vars)) s = s.split('{' + k + '}').join(String(vars[k]));
    return s;
  };
  const locale = () => (lang === 'zh-Hans' ? 'zh-Hans' : lang);
  function applyLang() {
    document.documentElement.lang = locale();
    document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr';
    document.querySelectorAll('[data-i18n]').forEach((el) => { el.textContent = t(el.dataset.i18n); });
    document.querySelectorAll('[data-i18n-title]').forEach((el) => { el.title = t(el.dataset.i18nTitle); });
    document.querySelectorAll('[data-i18n-placeholder]').forEach((el) => { el.placeholder = t(el.dataset.i18nPlaceholder); });
    $('lang').value = localStorage.getItem('ugw.lang') || 'system';
  }
  $('lang').addEventListener('change', () => {
    localStorage.setItem('ugw.lang', $('lang').value); lang = pickLang(); applyLang();
    loadStatus(); if (token && !$('app').hidden) { list(cwd); loadActivity(); loadDevices(); loadUsers(); loadNotifyChannels(); loadHealth(); loadSystem(); }
  });

  // ---- theme ----------------------------------------------------------------
  function applyTheme() {
    const pref = localStorage.getItem('ugw.theme') || 'system';
    if (pref === 'system') document.documentElement.removeAttribute('data-theme'); else document.documentElement.dataset.theme = pref;
    $('theme').value = pref;
  }
  $('theme').addEventListener('change', () => { localStorage.setItem('ugw.theme', $('theme').value); applyTheme(); });
  applyTheme();
  applyLang();

  // ---- state ----------------------------------------------------------------
  let token = sessionStorage.getItem('ugw.token') || '';
  // This browser is a "device" like an app installation: a stable id, registered on manual login.
  let deviceId = localStorage.getItem('ugw.device') || '';
  if (!deviceId) { deviceId = (crypto.randomUUID ? crypto.randomUUID() : String(Date.now()) + Math.random().toString(16).slice(2)); localStorage.setItem('ugw.device', deviceId); }
  const deviceName = () => `Web UI · ${(navigator.userAgent.match(/(Firefox|Edg|Chrome|Safari)\/[\d.]+/) || ['browser'])[0]} on ${navigator.platform || 'unknown'}`;
  let cwd = '/';

  const fmtSize = (n) => {
    if (n < 1024) return n + ' B';
    const u = ['KB', 'MB', 'GB', 'TB']; let i = -1;
    do { n /= 1024; i++; } while (n >= 1024 && i < u.length - 1);
    return n.toFixed(n < 10 ? 1 : 0) + ' ' + u[i];
  };
  const fmtDate = (s) => new Date(s).toLocaleString(locale());
  const encPath = (p) => encodeURIComponent(p);

  async function api(path, opts = {}) {
    opts.headers = Object.assign({}, opts.headers, token ? { Authorization: 'Bearer ' + token } : {});
    const res = await fetch(API + path, opts);
    if (res.status === 401 && token) {
      let msg = t('session_expired');
      try { const b = await res.clone().json(); if (b && b.code === 'device_revoked') msg = t('device_revoked'); } catch {}
      logout(); $('login-error').textContent = msg; $('login-error').hidden = false; throw new Error(msg);
    }
    if (res.status === 204) return null;
    const ct = res.headers.get('content-type') || '';
    const body = ct.includes('json') ? await res.json() : await res.text();
    if (!res.ok) throw new Error((body && body.error) || res.statusText);
    return body;
  }

  // ---- 0.15 helpers: text filters on tables, resizable columns, controls inside <summary> ------
  const norm = (v) => String(v || '').toLowerCase();
  // Hide the rows of the given tables whose text does not contain the filter.
  function filterRows(tableIds, text) {
    const q = norm(text).trim();
    for (const id of tableIds) for (const tr of $(id).querySelectorAll('tbody tr')) tr.classList.toggle('hidden-row', q !== '' && !norm(tr.textContent).includes(q));
  }
  const bindFilter = (inputId, tableIds) => $(inputId).addEventListener('input', () => filterRows(tableIds, $(inputId).value));
  // Column resizing (0.17): pointer events (mouse, touch, pencil), wide visible handles, widths remembered
  // per table in this browser, double-tap on a divider resets that column. The table switches to
  // table-layout: fixed only once a width is set, so the default layout is never clipped.
  function makeResizable(table) {
    const ths = [...table.querySelectorAll('thead th')];
    const key = 'ugw.cols.' + table.id;
    const saved = (() => { try { return JSON.parse(localStorage.getItem(key) || 'null'); } catch { return null; } })();
    const persist = () => { try { localStorage.setItem(key, JSON.stringify(ths.map((x) => x.style.width || ''))); } catch {} };
    const freeze = () => { if (!table.dataset.sized) { ths.forEach((x) => { if (!x.style.width) x.style.width = x.offsetWidth + 'px'; }); table.dataset.sized = '1'; table.classList.add('resizable'); } };
    if (saved && saved.some(Boolean) && !table.dataset.sized) { ths.forEach((x, i) => { if (saved[i]) x.style.width = saved[i]; }); freeze(); }
    ths.forEach((th, i) => {
      if (i === ths.length - 1 || th.querySelector('.rz')) return;
      const h = document.createElement('span'); h.className = 'rz'; th.appendChild(h);
      let lastTap = 0;
      h.addEventListener('pointerdown', (e) => {
        e.preventDefault(); e.stopPropagation();
        const now = Date.now();
        if (now - lastTap < 350) { th.style.width = ''; persist(); lastTap = 0; return; }   // double tap/click: auto width
        lastTap = now;
        freeze();
        h.setPointerCapture(e.pointerId); h.classList.add('active');
        const x0 = e.clientX, w0 = th.offsetWidth, rtl = document.documentElement.dir === 'rtl';
        const move = (ev) => { th.style.width = Math.max(60, w0 + (ev.clientX - x0) * (rtl ? -1 : 1)) + 'px'; };
        const up = () => { h.removeEventListener('pointermove', move); h.removeEventListener('pointerup', up); h.removeEventListener('pointercancel', up); h.classList.remove('active'); persist(); };
        h.addEventListener('pointermove', move); h.addEventListener('pointerup', up); h.addEventListener('pointercancel', up);
      });
      h.addEventListener('click', (e) => e.stopPropagation());
    });
  }
  // ⓘ popovers (0.17): the long explanations live behind an info button, shown on click / hover.
  const tipEl = $('tip');
  function showTip(anchor, text) {
    tipEl.textContent = text; tipEl.hidden = false;
    const r = anchor.getBoundingClientRect(), tw = tipEl.offsetWidth, th = tipEl.offsetHeight;
    let x = Math.min(window.innerWidth - tw - 12, Math.max(12, r.left - 8)), y = r.bottom + 8;
    if (y + th > window.innerHeight - 8) y = Math.max(8, r.top - th - 8);
    tipEl.style.left = x + 'px'; tipEl.style.top = y + 'px'; tipEl.dataset.for = anchor.id || anchor.dataset.tip || '';
  }
  function hideTip() { tipEl.hidden = true; tipEl.dataset.for = ''; }
  document.addEventListener('click', (e) => {
    const b = e.target.closest('.info');
    if (!b) { if (!e.target.closest('#tip')) hideTip(); return; }
    e.preventDefault(); e.stopPropagation();
    const text = b.id === 'tls-info' ? $('tls-note').textContent : t(b.dataset.tip);
    if (!tipEl.hidden && tipEl.dataset.for === (b.id || b.dataset.tip)) hideTip(); else showTip(b, text);
  }, true);
  document.querySelectorAll('.info').forEach((b) => { b.addEventListener('mouseenter', () => { if (tipEl.hidden) showTip(b, b.id === 'tls-info' ? $('tls-note').textContent : t(b.dataset.tip)); }); b.addEventListener('mouseleave', () => { if (tipEl.dataset.for === (b.id || b.dataset.tip)) hideTip(); }); });
  window.addEventListener('scroll', hideTip, { passive: true });
  // Donut chart: segments [{v, color, label}], text in the middle.
  function drawDonut(canvas, segs, center, legendEl) {
    const dpr = window.devicePixelRatio || 1, S = 140;
    canvas.width = S * dpr; canvas.height = S * dpr;
    const ctx = canvas.getContext('2d'); ctx.scale(dpr, dpr); ctx.clearRect(0, 0, S, S);
    const css = getComputedStyle(document.documentElement), line = css.getPropertyValue('--line').trim(), fg = css.getPropertyValue('--fg').trim() || '#222';
    const total = segs.reduce((a, x) => a + x.v, 0), cx = S / 2, cy = S / 2, R = 60, r = 42;
    let a0 = -Math.PI / 2;
    if (total <= 0) { ctx.beginPath(); ctx.arc(cx, cy, R, 0, 2 * Math.PI); ctx.arc(cx, cy, r, 2 * Math.PI, 0, true); ctx.fillStyle = line; ctx.fill(); }
    for (const x of segs) {
      if (x.v <= 0) continue;
      const a1 = a0 + (2 * Math.PI * x.v) / total;
      ctx.beginPath(); ctx.arc(cx, cy, R, a0, a1); ctx.arc(cx, cy, r, a1, a0, true); ctx.closePath(); ctx.fillStyle = x.color; ctx.fill();
      a0 = a1;
    }
    ctx.fillStyle = fg; ctx.textAlign = 'center'; ctx.textBaseline = 'middle'; ctx.font = '600 15px -apple-system, sans-serif'; ctx.fillText(center, cx, cy);
    if (legendEl) { legendEl.innerHTML = ''; for (const x of segs) { if (x.v <= 0) continue; const d = document.createElement('div'); const i = document.createElement('i'); i.style.background = x.color; d.appendChild(i); d.appendChild(document.createTextNode(`${x.label}: ${x.text ?? x.v}`)); legendEl.appendChild(d); } }
  }
  // Line chart over time: samples [{t, ...}], lines [{key, color, label}], ymax fixed or auto.
  function drawLines(canvas, samples, lines, opts, emptyEl, emptyText) {
    const empty = !samples.length;
    if (emptyEl) { emptyEl.hidden = !empty; if (empty) emptyEl.textContent = emptyText || t('chart_empty'); }
    canvas.hidden = empty; if (empty) return;
    const dpr = window.devicePixelRatio || 1, W = canvas.clientWidth || 600, H = Number(canvas.getAttribute('height')) || 200;
    canvas.width = W * dpr; canvas.height = H * dpr;
    const ctx = canvas.getContext('2d'); ctx.scale(dpr, dpr); ctx.clearRect(0, 0, W, H);
    const css = getComputedStyle(document.documentElement), fg = css.getPropertyValue('--muted').trim(), grid = css.getPropertyValue('--line').trim();
    const padL = 48, padR = 8, padT = 8, padB = 40, iw = W - padL - padR, ih = H - padT - padB;
    const max = opts.max || Math.max(1, ...samples.flatMap((p) => lines.map((l) => p[l.key] || 0)));
    const t0 = new Date(samples[0].t).getTime(), t1 = Math.max(t0 + 1, new Date(samples[samples.length - 1].t).getTime());
    ctx.strokeStyle = grid; ctx.fillStyle = fg; ctx.font = '11px -apple-system, sans-serif'; ctx.textAlign = 'right'; ctx.textBaseline = 'middle';
    for (let i = 0; i <= 4; i++) { const y = padT + ih - (ih * i) / 4; ctx.beginPath(); ctx.moveTo(padL, y); ctx.lineTo(W - padR, y); ctx.stroke(); ctx.fillText(opts.fmt ? opts.fmt((max * i) / 4) : String(Math.round((max * i) / 4)), padL - 6, y); }
    for (const l of lines) {
      ctx.beginPath(); ctx.strokeStyle = l.color; ctx.lineWidth = 2;
      samples.forEach((p, i) => { const x = padL + (iw * (new Date(p.t).getTime() - t0)) / (t1 - t0), y = padT + ih - (ih * Math.min(max, p[l.key] || 0)) / max; i ? ctx.lineTo(x, y) : ctx.moveTo(x, y); });
      ctx.stroke();
    }
    ctx.lineWidth = 1; ctx.fillStyle = fg; ctx.textAlign = 'center'; ctx.textBaseline = 'top';
    const n = Math.max(2, Math.floor(iw / 90)), span = t1 - t0;
    for (let i = 0; i <= n; i++) { const tt = new Date(t0 + (span * i) / n); ctx.fillText(span > 36e5 * 36 ? tt.toLocaleDateString(locale(), { day: 'numeric', month: 'short' }) : tt.toLocaleTimeString(locale(), { hour: '2-digit', minute: '2-digit' }), padL + (iw * i) / n, padT + ih + 6); }
    ctx.textAlign = 'left'; let lx = padL;
    for (const l of lines) { ctx.fillStyle = l.color; ctx.fillRect(lx, H - 12, 10, 10); ctx.fillStyle = fg; ctx.fillText(l.label, lx + 14, H - 12); lx += 14 + ctx.measureText(l.label).width + 18; }
  }

  // Inputs, selects, buttons and labels inside a card header must not toggle the card.
  document.querySelectorAll('details.card > summary').forEach((sm) => sm.querySelectorAll('input, select, button, label, code').forEach((el) => el.addEventListener('click', (e) => e.stopPropagation())));

  // ---- tabs -----------------------------------------------------------------
  function showTab(name) {
    document.querySelectorAll('#tabs button').forEach((b) => b.classList.toggle('active', b.dataset.tab === name));
    document.querySelectorAll('.tab').forEach((c) => { c.hidden = c.id !== 'tab-' + name; });
    localStorage.setItem('ugw.tab', name);
    if (name === 'health') { loadHealth(); loadSystem(); loadDevices(); loadLoad(); loadTraffic(); }
    if (name === 'settings') { loadDevices(); loadUsers(); loadNotifyChannels(); loadNotifySettings(); if (!$('keys').hidden) loadKeys(); }
  }
  $('tabs').addEventListener('click', (e) => { const b = e.target.closest('button'); if (b) showTab(b.dataset.tab); });
  // The header dot opens Settings › Health (0.15).
  $('health-pill').addEventListener('click', () => { if ($('app').hidden) return; showTab('health'); window.scrollTo({ top: 0, behavior: 'smooth' }); });

  // ---- status ---------------------------------------------------------------
  async function loadStatus() {
    try {
      const s = await (await fetch('/healthz')).json();
      $('version').textContent = s.version || '';
      if (!token) { $('status').textContent = t('online'); $('health-dot').className = 'dot'; }
      $('st-gateway').textContent = (s.version || '') + ' · ' + location.host;
    } catch { $('status').textContent = t('unreachable'); $('health-dot').className = 'dot error'; }
    $('origin').textContent = location.origin; $('origin2').textContent = location.origin;
    $('tls-note').textContent = location.protocol === 'https:' ? t('tls_ok') : t('tls_no');
  }

  // ---- auth -----------------------------------------------------------------
  function showApp(identity, user, shares) {
    $('login').hidden = true; $('app').hidden = false; $('who').hidden = false;
    const key = identity ? `${identity.name || t('key_label')} · ${(identity.roles || []).join(', ')}` : '';
    $('who-name').textContent = user ? `${user} · ${key}` : key;
    window.ugwShares = shares || null; // share → 'rw' | 'ro' when a user is logged in
    showTab(['status', 'files', 'health', 'settings'].includes(localStorage.getItem('ugw.tab')) ? localStorage.getItem('ugw.tab') : 'status');
    list(cwd);
    startActivity();
    loadDevices();
    loadNotifyChannels();
    startHealth();
    $('forget').hidden = !(storedInfo.available && storedInfo.canForget);
  }
  function logout() {
    if (token) fetch(API + '/auth/logout', { method: 'POST', headers: { Authorization: 'Bearer ' + token } }).catch(() => {});
    token = ''; sessionStorage.removeItem('ugw.token'); clearInterval(activityTimer); clearInterval(healthTimer);
    setHealth(null);
    $('login').hidden = false; $('app').hidden = true; $('who').hidden = true;
  }
  // A key kept on the gateway (WEBUI_API_KEY or /config/webui.key) logs the web UI in with one click.
  let storedInfo = { available: false, canRemember: false, canForget: false };
  async function loadStoredKey() {
    try { storedInfo = await (await fetch(API + '/auth/stored-key')).json(); } catch { storedInfo = { available: false }; }
    $('stored').hidden = !storedInfo.available;
    $('stored-note').textContent = storedInfo.available ? (storedInfo.source === 'env' ? t('stored_env') : t('stored_file')) : '';
    $('remember-label').hidden = !storedInfo.canRemember;
    $('apikey').required = !storedInfo.available;
  }
  // Offer the credentials to the browser's password manager (Chromium); Safari and Firefox
  // pick them up from the submitted <form> (username + current-password fields) on their own.
  function offerToPasswordManager(id, password) {
    try {
      if (window.PasswordCredential && navigator.credentials && id && password) {
        navigator.credentials.store(new PasswordCredential({ id, password })).catch(() => {});
      }
    } catch {}
  }
  async function login(body, withUser) {
    $('login-error').hidden = true;
    try {
      if (withUser && $('username').value.trim()) { body.username = $('username').value.trim(); body.password = $('password').value; }
      body.deviceId = deviceId; body.deviceName = deviceName(); body.registerDevice = true;
      const r = await api('/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      token = r.token; sessionStorage.setItem('ugw.token', token);
      if (body.apiKey) offerToPasswordManager($('account').value.trim() || 'unraid-gateway', body.apiKey);
      if (body.username) offerToPasswordManager(body.username, body.password);
      if (body.apiKey && $('remember').checked) {
        try { await api('/auth/remember', { method: 'POST' }); } catch (err) { alert(t('remember_failed', { err: err.message })); }
        $('remember').checked = false;
      }
      $('apikey').value = ''; $('password').value = '';
      showApp(r.identity, r.user, r.shares);
      loadStoredKey();
    } catch (err) { $('login-error').textContent = err.message; $('login-error').hidden = false; }
  }
  const currentKey = () => $('apikey').value.trim();
  $('login-form').addEventListener('submit', (e) => {
    e.preventDefault();
    const key = currentKey();
    if (!key && storedInfo.available) { login({ useStoredKey: true }, false); return; }
    login({ apiKey: key }, false);
  });
  $('user-toggle').addEventListener('click', () => { $('user-form').hidden = !$('user-form').hidden; if (!$('user-form').hidden) $('username').focus(); });
  $('user-form').addEventListener('submit', (e) => {
    e.preventDefault();
    const key = currentKey();
    if (!key && !storedInfo.available) { $('login-error').textContent = t('need_key'); $('login-error').hidden = false; return; }
    login(key ? { apiKey: key } : { useStoredKey: true }, true);
  });
  $('stored-connect').addEventListener('click', () => login({ useStoredKey: true }, false));
  $('forget').addEventListener('click', async () => {
    if (!confirm(t('confirm_forget'))) return;
    try { await api('/auth/remember', { method: 'DELETE' }); await loadStoredKey(); $('forget').hidden = true; } catch (err) { alert(err.message); }
  });
  $('logout').addEventListener('click', logout);

  // ---- browser --------------------------------------------------------------
  function crumbs(path) {
    const parts = path.split('/').filter(Boolean);
    const nav = $('crumbs'); nav.innerHTML = '';
    const mk = (label, p) => { const a = document.createElement('a'); a.href = '#'; a.textContent = label; a.onclick = (e) => { e.preventDefault(); list(p); }; return a; };
    nav.appendChild(mk(t('shares'), '/'));
    let acc = '';
    parts.forEach((p) => { acc += '/' + p; nav.appendChild(Object.assign(document.createElement('span'), { textContent: '/' })); nav.appendChild(mk(p, acc)); });
  }

  async function list(path) {
    try {
      const r = await api('/fs/list?path=' + encPath(path));
      cwd = r.path; crumbs(cwd);
      const atRoot = cwd === '/';
      $('mkdir').hidden = atRoot; $('upload-label').hidden = atRoot; $('root-hint').hidden = !atRoot;
      const tb = $('files').querySelector('tbody'); tb.innerHTML = '';
      $('empty').hidden = r.entries.length > 0;
      for (const e of r.entries) {
        const tr = document.createElement('tr');
        const name = document.createElement('td'); name.className = 'name';
        const a = document.createElement('a'); a.href = '#'; a.textContent = (e.type === 'dir' ? '📁 ' : '📄 ') + e.name;
        a.onclick = (ev) => { ev.preventDefault(); e.type === 'dir' ? list(e.path) : download(e); };
        name.appendChild(a);
        const size = document.createElement('td'); size.className = 'num';
        const ro = cwd === '/' && window.ugwShares && window.ugwShares[e.name] === 'ro';
        size.textContent = e.type === 'dir' ? (ro ? t('read_only') : '') : fmtSize(e.size);
        const mt = document.createElement('td'); mt.textContent = fmtDate(e.mtime);
        // The container sees /data/<share>/…; on Unraid that is the user share /mnt/user/<share>/… (0.15).
        const up = document.createElement('td'); up.className = 'wrap small muted'; up.textContent = '/mnt/user' + e.path;
        const act = document.createElement('td'); act.className = 'actions';
        const btn = (label, fn) => { const b = document.createElement('button'); b.className = 'ghost'; b.textContent = label; b.onclick = fn; act.appendChild(b); };
        if (cwd !== '/') { btn(t('rename'), () => rename(e)); btn(t('delete'), () => remove(e)); }
        tr.append(name, size, mt, up, act); tb.appendChild(tr);
      }
      filterRows(['files'], $('files-filter').value); makeResizable($('files'));
    } catch (err) { alert(err.message); }
  }
  bindFilter('files-filter', ['files']);

  async function download(e) {
    try {
      const res = await fetch(API + '/fs/content?path=' + encPath(e.path), { headers: { Authorization: 'Bearer ' + token } });
      if (!res.ok) throw new Error((await res.json()).error);
      const url = URL.createObjectURL(await res.blob());
      const a = document.createElement('a'); a.href = url; a.download = e.name; document.body.appendChild(a); a.click(); a.remove();
      setTimeout(() => URL.revokeObjectURL(url), 10000);
    } catch (err) { alert(err.message); }
  }
  async function rename(e) {
    const to = prompt(t('new_name'), e.name); if (!to || to === e.name) return;
    const dir = e.path.slice(0, e.path.lastIndexOf('/')) || '';
    try { await api('/fs/move', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ from: e.path, to: dir + '/' + to }) }); list(cwd); }
    catch (err) { alert(err.message); }
  }
  async function remove(e) {
    if (!confirm(t(e.type === 'dir' ? 'confirm_delete_folder' : 'confirm_delete_file', { name: e.name }))) return;
    try { await api('/fs/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path: e.path, recursive: true }) }); list(cwd); }
    catch (err) { alert(err.message); }
  }
  $('mkdir').addEventListener('click', async () => {
    if (cwd === '/') { alert(t('open_share_first')); return; }
    const name = prompt(t('folder_name')); if (!name) return;
    try { await api('/fs/mkdir', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path: (cwd === '/' ? '' : cwd) + '/' + name }) }); list(cwd); }
    catch (err) { alert(err.message); }
  });
  $('refresh').addEventListener('click', () => list(cwd));

  $('upload').addEventListener('change', async (ev) => {
    const files = [...ev.target.files]; ev.target.value = '';
    if (cwd === '/') { alert(t('open_share_first')); return; }
    const bar = $('progress'); bar.hidden = false;
    for (let i = 0; i < files.length; i++) {
      const f = files[i];
      await new Promise((resolve) => {
        const xhr = new XMLHttpRequest();
        xhr.open('PUT', API + '/fs/content?path=' + encPath(cwd + '/' + f.name));
        xhr.setRequestHeader('Authorization', 'Bearer ' + token);
        xhr.setRequestHeader('X-Mtime', new Date(f.lastModified).toISOString());
        xhr.upload.onprogress = (p) => { if (p.lengthComputable) { const pct = Math.round(p.loaded / p.total * 100); bar.firstElementChild.style.width = pct + '%'; bar.lastElementChild.textContent = `${f.name} · ${pct}% (${i + 1}/${files.length})`; } };
        xhr.onload = () => { if (xhr.status >= 300) { try { alert(JSON.parse(xhr.responseText).error); } catch { alert(xhr.statusText); } } resolve(); };
        xhr.onerror = () => { alert(t('upload_failed', { name: f.name })); resolve(); };
        xhr.send(f);
      });
    }
    bar.hidden = true; bar.firstElementChild.style.width = '0';
    list(cwd);
  });

  // ---- graphql --------------------------------------------------------------
  $('gql-run').addEventListener('click', async () => {
    $('gql-out').textContent = '…';
    try {
      const r = await api('/graphql', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ query: $('gql').value }) });
      $('gql-out').textContent = typeof r === 'string' ? r : JSON.stringify(r, null, 2);
    } catch (err) { $('gql-out').textContent = t('error', { err: err.message }); }
  });

  // ---- health (0.12) ----------------------------------------------------------
  // One dot for gateway and server: green, yellow for warnings, red for problems (GET /api/v1/health).
  let healthTimer = null;
  function setHealth(h) {
    const lvl = h ? h.level : '';
    const cls = 'dot' + (lvl ? ' ' + lvl : '');
    $('health-dot').className = cls; $('health-dot-big').className = cls;
    const label = h ? t(lvl === 'ok' ? 'health_ok' : lvl === 'warning' ? 'health_warning' : 'health_error') : t('health_unknown');
    $('status').textContent = label; $('health-level').textContent = label;
    $('health-pill').title = (h && h.reasons.length ? h.reasons.join(' · ') + ' — ' : '') + t('health_open');
    const ul = $('health-reasons'); ul.innerHTML = '';
    (h ? h.reasons : []).forEach((r) => { const li = document.createElement('li'); li.textContent = r; ul.appendChild(li); });
    const checks = h ? h.checks : [];
    const cnt = (lv) => checks.filter((c) => c.level === lv).length;
    drawDonut($('donut-checks'), [{ v: cnt('ok'), color: '#34c759', label: t('health_ok') }, { v: cnt('warning'), color: '#ff9f0a', label: t('health_warning') }, { v: cnt('error'), color: '#ff3b30', label: t('health_error') }], String(checks.length), $('donut-checks-legend'));
    const tb = $('health-checks').querySelector('tbody'); tb.innerHTML = '';
    (h ? h.checks : []).forEach((c) => {
      const tr = document.createElement('tr');
      const d = document.createElement('td'); d.className = 'num'; const dot = document.createElement('span'); dot.className = 'dot ' + c.level; d.appendChild(dot);
      const n = document.createElement('td'); n.textContent = c.name;
      const v = document.createElement('td'); v.className = 'wrap'; v.textContent = c.detail || '';
      tr.append(d, n, v); tb.appendChild(tr);
    });
  }
  async function loadHealth() {
    if (!token || $('app').hidden) return;
    try { const h = await api('/health'); setHealth(h); $('health-when').textContent = t('checked_at', { when: new Date().toLocaleTimeString(locale()) }); }
    catch (err) { $('health-when').textContent = t('health_unavailable', { err: err.message }); }
  }
  function startHealth() { clearInterval(healthTimer); loadHealth(); healthTimer = setInterval(() => { if (!document.hidden) loadHealth(); }, 30000); }
  $('health-refresh').addEventListener('click', loadHealth);

  // ---- system information (0.12) ----------------------------------------------
  async function loadSystem() {
    if (!token || $('app').hidden) return;
    const kv = $('sys-kv'); $('sys-note').textContent = '';
    try {
      const d = await api('/system'); const i = d.info || {}; kv.innerHTML = '';
      const put = (k, v) => { if (!v) return; const a = document.createElement('div'); a.className = 'k'; a.textContent = t(k); const b = document.createElement('div'); b.textContent = v; kv.append(a, b); };
      const os = i.os || {}, cpu = i.cpu || {}, mem = (i.memory && i.memory.layout) || [], mt = (d.metrics && d.metrics.memory) || {};
      put('sys_host', os.fqdn || os.hostname);
      put('sys_os', [os.distro, os.release].filter(Boolean).join(' ') + (os.uefi ? ' · UEFI' : ''));
      put('sys_kernel', os.kernel);
      if (os.uptime) { const b = new Date(os.uptime); if (!isNaN(b)) put('sys_uptime', b.toLocaleString(locale())); }
      put('sys_cpu', [cpu.manufacturer, cpu.brand].filter(Boolean).join(' ') + (cpu.socket ? ' · ' + cpu.socket : ''));
      if (cpu.cores) put('sys_cores', `${cpu.cores} / ${cpu.threads || '?'}`);
      if (cpu.speed) put('sys_clock', cpu.speedmax ? `${cpu.speed.toFixed(2)} – ${cpu.speedmax.toFixed(2)} GHz` : `${cpu.speed.toFixed(2)} GHz`);
      // unraid-api reports module sizes in KiB (33554432 = 32 GiB).
      const installed = mem.reduce((s, m) => s + (m.size || 0), 0) * 1024;
      if (installed) put('sys_memory', fmtSize(installed) + (mem.length ? ' · ' + mem.map((m) => [m.size ? fmtSize(m.size * 1024) : null, m.type, m.clockSpeed ? m.clockSpeed + ' MHz' : null].filter(Boolean).join(' ')).join(', ') : ''));
      if (mt.used && mt.total) { put('sys_memory_use', fmtSize(mt.used) + ' / ' + fmtSize(mt.total)); drawDonut($('donut-memory'), [{ v: mt.used, color: '#3b8bd9', label: t('sys_memory_use'), text: fmtSize(mt.used) }, { v: Math.max(0, mt.total - mt.used), color: getComputedStyle(document.documentElement).getPropertyValue('--line').trim(), label: t('sys_memory'), text: fmtSize(mt.total) }], Math.round(mt.used * 100 / mt.total) + '%', $('donut-memory-legend')); }
      if (i.baseboard) put('sys_board', [i.baseboard.manufacturer, i.baseboard.model, i.baseboard.version].filter(Boolean).join(' '));
      if (i.versions && i.versions.core) put('sys_versions', i.versions.core.api);
      const gpus = (i.devices && i.devices.gpu) || []; if (gpus.length) put('sys_gpu', gpus.map((g) => [g.vendorname, g.productid].filter(Boolean).join(' ')).join(', '));
      const nics = (i.networkInterfaces || []).filter((n) => n.ipAddress && n.name !== 'lo' && !/^(veth|docker|virbr)/.test(n.name));
      if (nics.length) put('sys_network', nics.map((n) => `${n.name} ${n.ipAddress}${n.speed ? ' · ' + n.speed + ' Mb/s' : ''}`).join(', '));
    } catch (err) { $('sys-note').textContent = t('sys_unavailable', { err: err.message }); }
    loadSharesDonut();
  }
  async function loadSharesDonut() {
    try {
      const r = await api('/fs/list?path=/'); const shares = (r.entries || []).filter((e) => e.type === 'dir');
      const perms = window.ugwShares;
      const segs = perms ? [{ v: shares.filter((e) => perms[e.name] !== 'ro').length, color: '#34c759', label: t('read_write') }, { v: shares.filter((e) => perms[e.name] === 'ro').length, color: '#ff9f0a', label: t('read_only') }] : [{ v: shares.length, color: '#34c759', label: t('mounted') }];
      drawDonut($('donut-shares'), segs, String(shares.length), $('donut-shares-legend'));
    } catch {}
  }
  // Unraid load over time (0.17): CPU and memory sampled by the gateway every minute.
  async function loadLoad() {
    if (!token || $('app').hidden) return;
    try {
      const r = await api('/system/history?range=' + $('load-range').value);
      $('load-note').textContent = r.sampling ? '' : t('load_nokey');
      drawLines($('load-chart'), r.samples || [], [{ key: 'cpu', color: '#ff9f0a', label: t('legend_cpu') }, { key: 'mem', color: '#3b8bd9', label: t('legend_mem') }], { max: 100, fmt: (v) => Math.round(v) + '%' }, $('load-empty'), t('load_empty'));
    } catch (err) { $('load-note').textContent = t('metrics_unavailable', { err: err.message }); }
  }
  $('load-range').addEventListener('change', loadLoad);
  // Traffic through the gateway (0.17): every device summed.
  async function loadTraffic() {
    if (!token || $('app').hidden) return;
    try {
      const m = await api('/metrics?range=' + $('traffic-range').value);
      $('traffic-note').textContent = t('n_devices_seen', { n: m.devices });
      const tot = $('traffic-totals'); tot.innerHTML = '';
      for (const [k, v] of [['today', m.today], ['week', m.week], ['month', m.month]]) {
        const d = document.createElement('div'); d.className = 'stat';
        d.innerHTML = `<span class="k"></span><span class="v"></span><span class="muted small"></span>`;
        d.querySelector('.k').textContent = t('tot_' + k);
        d.querySelector('.v').textContent = '↓ ' + fmtSize(v.down) + '  ↑ ' + fmtSize(v.up);
        d.querySelector('.muted').textContent = t('tot_detail', { req: v.requests, dl: v.downloads + v.streams, ul: v.uploads });
        tot.appendChild(d);
      }
      drawChart($('traffic-chart'), m.series || [], m.step, $('traffic-empty'));
      drawLines($('requests-chart'), m.series || [], [{ key: 'requests', color: '#8e8e93', label: t('legend_requests') }], {}, null);
    } catch (err) { $('traffic-note').textContent = t('metrics_unavailable', { err: err.message }); }
  }
  $('traffic-range').addEventListener('change', loadTraffic);

  // ---- activity -------------------------------------------------------------
  const ago = (iso, now) => {
    const d = Math.max(0, (now - new Date(iso)) / 1000);
    if (d < 5) return t('now'); if (d < 60) return t('s_ago', { n: Math.round(d) });
    if (d < 3600) return t('min_ago', { n: Math.round(d / 60) }); if (d < 86400) return t('h_ago', { n: Math.round(d / 3600) });
    return fmtDate(iso);
  };
  const dur = (sec) => { sec = Math.max(0, Math.round(sec)); const h = Math.floor(sec / 3600), m = Math.floor(sec % 3600 / 60), s = sec % 60; return (h ? h + ':' : '') + String(m).padStart(h ? 2 : 1, '0') + ':' + String(s).padStart(2, '0'); };
  const device = (c) => {
    if (!c.agent) return t('unknown');
    if (c.agent.startsWith('Unraid Drive')) return c.agent;
    if (/CFNetwork/.test(c.agent)) return t('dev_older') + ' · ' + (c.agent.match(/Darwin\/[\d.]+/) || [''])[0];
    if (/Mozilla/.test(c.agent)) return t('dev_browser') + ' · ' + ((c.agent.match(/(Firefox|Edg|Chrome|Safari)\/[\d.]+/) || ['browser'])[0]);
    if (/mpv|libmpv/i.test(c.agent)) return t('dev_mpv');
    return c.agent.slice(0, 60);
  };
  const who = (x) => (x.user ? x.user : (x.key ? t('key_label') + ': ' + x.key : '—'));
  const kindLabel = (k) => t({ download: 'k_download', stream: 'k_stream', upload: 'k_upload' }[k] || k);
  let activityTimer = null;
  const speeds = {};
  async function loadActivity() {
    if (!token || $('app').hidden) return;
    try {
      const a = await api('/activity');
      const now = new Date(a.now);
      $('activity-scope').textContent = a.scope === 'user' ? t('scope_user') : t('scope_all');
      $('st-clients').textContent = String(a.clients.length);
      $('st-active').textContent = String(a.active.length);
      const tc = $('act-clients').querySelector('tbody'); tc.innerHTML = '';
      $('act-clients-empty').hidden = a.clients.length > 0;
      for (const c of a.clients) {
        const tr = document.createElement('tr');
        const cells = [device(c), who(c), c.ip, ago(c.firstSeen, now), ago(c.lastSeen, now), String(c.requests), c.lastPath ? `${c.lastPath} (${ago(c.lastPathAt, now)})` : ''];
        cells.forEach((v, i) => { const td = document.createElement('td'); td.textContent = v; if (i === 5) td.className = 'num'; if (i === 6 || i === 0) td.className = 'wrap'; tr.appendChild(td); });
        if (c.active) tr.classList.add('live');
        tc.appendChild(tr);
      }
      const ta = $('act-active').querySelector('tbody'); ta.innerHTML = '';
      $('act-active-empty').hidden = a.active.length > 0;
      for (const x of a.active) {
        const tr = document.createElement('tr');
        const prev = speeds[x.id]; speeds[x.id] = [x.bytes, now.getTime()];
        const speed = prev && now.getTime() > prev[1] ? (x.bytes - prev[0]) / ((now.getTime() - prev[1]) / 1000) : 0;
        const pct = x.size > 0 ? Math.min(100, Math.round(x.bytes / x.size * 100)) : null;
        const kind = document.createElement('td'); kind.textContent = kindLabel(x.kind);
        const file = document.createElement('td'); file.className = 'wrap'; file.textContent = x.path;
        const w = document.createElement('td'); w.className = 'wrap'; w.textContent = `${who(x)} · ${device(x)} · ${x.ip}`;
        const prog = document.createElement('td'); prog.className = 'wrap';
        prog.innerHTML = `<div class="bar"><div style="width:${pct ?? 0}%"></div></div><span class="small muted">${fmtSize(x.bytes)}${x.size > 0 ? ' / ' + fmtSize(x.size) + ' · ' + pct + '%' : ''}${x.range ? ' · ' + x.range : ''}</span>`;
        const sp = document.createElement('td'); sp.className = 'num'; sp.textContent = speed > 0 ? fmtSize(speed) + '/s' : '';
        const el = document.createElement('td'); el.className = 'num'; el.textContent = dur((now - new Date(x.started)) / 1000);
        tr.append(kind, file, w, prog, sp, el); ta.appendChild(tr);
      }
      for (const id of Object.keys(speeds)) if (!a.active.some((x) => String(x.id) === id)) delete speeds[id];
      const trc = $('act-recent').querySelector('tbody'); trc.innerHTML = '';
      for (const x of a.recent.slice(0, 30)) {
        const tr = document.createElement('tr');
        [kindLabel(x.kind), x.path, `${who(x)} · ${device(x)}`, fmtSize(x.bytes), ago(x.ended, now)].forEach((v, i) => { const td = document.createElement('td'); td.textContent = v; if (i === 3) td.className = 'num'; if (i === 1 || i === 2) td.className = 'wrap'; tr.appendChild(td); });
        trc.appendChild(tr);
      }
      filterRows(['act-clients', 'act-active', 'act-recent'], $('activity-filter').value);
    } catch (err) { $('activity-scope').textContent = t('activity_unavailable', { err: err.message }); }
  }
  bindFilter('activity-filter', ['act-clients', 'act-active', 'act-recent']);
  function startActivity() {
    clearInterval(activityTimer); loadActivity();
    activityTimer = setInterval(() => { if ($('activity-live').checked && !document.hidden) loadActivity(); }, 3000);
  }
  $('activity-live').addEventListener('change', () => { if ($('activity-live').checked) loadActivity(); });

  // ---- devices --------------------------------------------------------------
  // Icon by device kind, read from the registration name ("Unraid Drive 1.4 (44) · iPhone · iOS 27.1 · App").
  const DEV_ICONS = {
    phone: '<svg viewBox="0 0 24 24"><rect x="7" y="2" width="10" height="20" rx="2"/><circle cx="12" cy="18.5" r="1" fill="currentColor" stroke="none"/></svg>',
    tablet: '<svg viewBox="0 0 24 24"><rect x="4" y="2" width="16" height="20" rx="2"/><circle cx="12" cy="18.5" r="1" fill="currentColor" stroke="none"/></svg>',
    mac: '<svg viewBox="0 0 24 24"><rect x="2" y="4" width="20" height="13" rx="2"/><path d="M8 21h8M12 17v4"/></svg>',
    tv: '<svg viewBox="0 0 24 24"><rect x="2" y="5" width="20" height="13" rx="2"/><path d="M7 22h10"/></svg>',
    vision: '<svg viewBox="0 0 24 24"><path d="M3 12c0-3 2-5 5-5h8c3 0 5 2 5 5s-2 5-5 5c-2 0-3-1-4-2-1 1-2 2-4 2-3 0-5-2-5-5z"/></svg>',
    web: '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3c3 3 3 15 0 18M12 3c-3 3-3 15 0 18"/></svg>',
    other: '<svg viewBox="0 0 24 24"><rect x="3" y="3" width="18" height="18" rx="3"/></svg>',
  };
  const deviceKind = (name) => {
    const n = (name || '').toLowerCase();
    if (n.startsWith('web ui')) return 'web';
    if (n.includes('ipad')) return 'tablet';
    if (n.includes('iphone')) return 'phone';
    if (n.includes('macos') || n.includes('mac ')) return 'mac';
    if (n.includes('tvos') || n.includes('apple tv')) return 'tv';
    if (n.includes('visionos') || n.includes('vision pro')) return 'vision';
    return 'other';
  };
  // "Unraid Drive 1.4 (44) · Simone's iPhone · iOS 27.1 · App" → ["Simone's iPhone", "iOS 27.1 · App", "Unraid Drive 1.4 (44)"]
  const splitName = (name) => {
    const parts = (name || '').split(' · ');
    if (parts.length >= 3 && parts[0].startsWith('Unraid Drive')) return [parts[1], parts.slice(2).join(' · '), parts[0]];
    return [name || '—', '', ''];
  };
  const deviceState = (x, now) => {
    if (x.supersededBy) return ['', t('old')];
    const days = (now - new Date(x.lastSeen)) / 86400000;
    if (days < 1) return ['ok', t('active_today')];
    if (days > 30) return ['warning', t('stale', { n: Math.floor(days) })];
    return ['', t('last_seen_prefix') + ' ' + ago(x.lastSeen, now)];
  };
  let devicesCache = { devices: [], thisDevice: '' };
  async function loadDevices() {
    try {
      const d = await api('/devices'); d.devices = d.devices || [];
      devicesCache = d;
      const now = new Date();
      $('devices-scope').textContent = d.registration === 'off' ? t('registration_off') : '';
      $('st-devices').textContent = String(d.devices.filter((x) => !x.supersededBy).length);
      const wrap = $('devices-cards'); wrap.innerHTML = '';
      $('devices-empty').hidden = d.devices.length > 0;
      $('devices-prune').hidden = d.devices.length === 0;
      $('devices-remove-all').hidden = d.devices.length === 0;
      // Grouped per Unraid user (0.15); devices registered with a bare API key go last.
      const groups = new Map();
      const groupOf = (x) => x.user || (x.key ? t('key_label') + ' ' + x.key : '');
      for (const x of d.devices) { const g = groupOf(x); if (!groups.has(g)) groups.set(g, []); groups.get(g).push(x); }
      const order = [...groups.keys()].sort((a, b) => (a === '') - (b === '') || a.localeCompare(b));
      for (const g of order) {
        const head = document.createElement('div'); head.className = 'group-title';
        head.innerHTML = '<span class="gname"></span><span class="pill"></span>';
        head.querySelector('.gname').textContent = g || t('no_user_group'); head.querySelector('.pill').textContent = String(groups.get(g).filter((x) => !x.supersededBy).length);
        const grid = document.createElement('div'); grid.className = 'devices-cards'; grid.dataset.group = g;
        wrap.append(head, grid);
      }
      for (const x of d.devices) {
        const grid = wrap.querySelector(`.devices-cards[data-group="${CSS.escape(groupOf(x))}"]`);
        const mine = x.id === d.thisDevice;
        const [title, sub, app] = splitName(x.name);
        const [cls, state] = deviceState(x, now);
        const card = document.createElement('button'); card.className = 'dev-card' + (x.supersededBy ? ' old' : ''); card.type = 'button';
        card.innerHTML = `<span class="dev-icon">${DEV_ICONS[deviceKind(x.name)]}</span>
          <span class="dev-main"><span class="dev-title"></span><span class="dev-sub muted small"></span><span class="dev-meta muted small"></span></span>
          <span class="dot ${cls}"></span>`;
        card.querySelector('.dev-title').textContent = title + (mine ? ' ' + t('this_browser') : '');
        card.querySelector('.dev-sub').textContent = [sub, app].filter(Boolean).join(' · ');
        card.querySelector('.dev-meta').textContent = [x.user ? t('col_user') + ': ' + x.user : null, t('last_sync') + ': ' + ago(x.lastSeen, now), state].filter(Boolean).join(' · ');
        if (x.supersededBy) card.title = t('replaced_title', { date: new Date(x.supersededAt).toLocaleString(locale()) });
        card.onclick = () => openDevice(x.id);
        card.dataset.text = norm([x.name, x.user, x.key, x.lastIp, state].join(' '));
        grid.appendChild(card);
      }
      applyDevicesFilter();
      const states = { ok: 0, '': 0, warning: 0, old: 0 };
      for (const x of d.devices) { if (x.supersededBy) states.old++; else states[deviceState(x, now)[0]]++; }
      drawDonut($('donut-devices'), [{ v: states.ok, color: '#34c759', label: t('active_today') }, { v: states[''], color: '#8e8e93', label: t('last_seen_prefix') + ' > 1 d' }, { v: states.warning, color: '#ff9f0a', label: t('stale', { n: 30 }) }, { v: states.old, color: getComputedStyle(document.documentElement).getPropertyValue('--line').trim(), label: t('old') }], String(d.devices.filter((x) => !x.supersededBy).length), $('donut-devices-legend'));
      if (!$('device-detail').hidden && !d.devices.some((x) => x.id === currentDevice)) closeDevice();
    } catch (err) { $('devices-scope').textContent = t('devices_unavailable', { err: err.message }); }
  }
  function applyDevicesFilter() {
    const q = norm($('devices-filter').value).trim();
    for (const card of $('devices-cards').querySelectorAll('.dev-card')) card.classList.toggle('hidden-row', q !== '' && !card.dataset.text.includes(q));
    for (const grid of $('devices-cards').querySelectorAll('.devices-cards')) { const any = [...grid.children].some((c) => !c.classList.contains('hidden-row')); grid.classList.toggle('hidden-row', !any); grid.previousElementSibling.classList.toggle('hidden-row', !any); }
  }
  $('devices-filter').addEventListener('input', applyDevicesFilter);
  $('devices-refresh').addEventListener('click', loadDevices);
  $('devices-remove-all').addEventListener('click', async () => {
    if (!confirm(t('confirm_remove_all'))) return;
    try { await api('/devices', { method: 'DELETE' }); logout(); } catch (err) { alert(err.message); }
  });
  $('prune-run').addEventListener('click', async () => {
    const days = Number($('prune-days').value);
    if (!confirm(t('confirm_prune', { n: days }))) return;
    try {
      const r = await api('/devices/prune', { method: 'POST', body: JSON.stringify({ days }) });
      alert(r.removed.length ? t('pruned', { n: r.removed.length, list: r.removed.join('\n') }) : t('pruned_none'));
      loadDevices();
    } catch (err) { alert(err.message); }
  });

  // ---- device detail: facts, charts, log ------------------------------------
  let currentDevice = '';
  const showGlobals = (on) => document.querySelectorAll('#tab-settings .settings-global').forEach((c) => { c.hidden = !on; });
  function closeDevice() { currentDevice = ''; $('device-detail').hidden = true; if ($('user-detail').hidden) showGlobals(true); }
  $('device-back').addEventListener('click', closeDevice);
  $('device-range').addEventListener('change', () => { if (currentDevice) loadDeviceMetrics(currentDevice); });
  $('device-remove').addEventListener('click', async () => {
    const x = devicesCache.devices.find((y) => y.id === currentDevice); if (!x) return;
    const mine = x.id === devicesCache.thisDevice;
    if (!confirm(t('confirm_remove_device', { name: x.name }) + (mine ? t('confirm_remove_this') : ''))) return;
    try { await api('/devices/' + encodeURIComponent(x.id), { method: 'DELETE' }); if (mine) logout(); else { closeDevice(); loadDevices(); } } catch (err) { alert(err.message); }
  });
  async function openDevice(id) {
    const x = devicesCache.devices.find((y) => y.id === id); if (!x) return;
    currentDevice = id;
    const now = new Date();
    const [title, sub, app] = splitName(x.name);
    const [cls, state] = deviceState(x, now);
    $('device-icon').innerHTML = DEV_ICONS[deviceKind(x.name)];
    $('device-title').textContent = title;
    $('device-dot').className = 'dot ' + cls; $('device-dot').title = state;
    const facts = [[t('col_device'), [sub, app].filter(Boolean).join(' · ')], [t('col_user'), x.user || '—'], [t('col_key'), x.key || '—'],
      [t('col_registered'), fmtDate(x.firstSeen)], [t('last_sync'), ago(x.lastSeen, now)], [t('col_from'), x.lastIp || '—'], [t('col_logins'), String(x.logins) + ' ⓘ'], [t('state'), state], ['ID', x.id]];
    const dl = $('device-facts'); dl.innerHTML = '';
    for (const [k, v] of facts) { const dt = document.createElement('dt'); dt.textContent = k; const dd = document.createElement('dd'); dd.textContent = v; if (k === t('col_logins')) { dd.title = t('logins_note'); dd.style.cursor = 'help'; } dl.append(dt, dd); }
    showGlobals(false); $('user-detail').hidden = true; $('device-detail').hidden = false;
    await Promise.all([loadDeviceMetrics(id), loadDeviceLog(id), loadSyncs('device-syncs', 'device-syncs-empty', { device: id })]);
  }
  async function loadDeviceMetrics(id) {
    try {
      const m = await api('/devices/' + encodeURIComponent(id) + '/metrics?range=' + $('device-range').value);
      const tot = $('device-totals'); tot.innerHTML = '';
      for (const [k, v] of [['today', m.today], ['week', m.week], ['month', m.month]]) {
        const s = document.createElement('div'); s.className = 'stat';
        s.innerHTML = `<span class="k"></span><span class="v"></span><span class="muted small"></span>`;
        s.querySelector('.k').textContent = t('tot_' + k);
        s.querySelector('.v').textContent = '↓ ' + fmtSize(v.down) + '  ↑ ' + fmtSize(v.up);
        s.querySelector('.muted').textContent = t('tot_detail', { req: v.requests, dl: v.downloads + v.streams, ul: v.uploads });
        tot.appendChild(s);
      }
      drawChart($('device-chart'), m.series || [], m.step);
    } catch (err) { $('device-chart-empty').hidden = false; $('device-chart-empty').textContent = t('metrics_unavailable', { err: err.message }); }
  }
  function drawChart(canvas, series, step, emptyEl) {
    emptyEl = emptyEl || $('device-chart-empty');
    const empty = !series.some((p) => p.up || p.down);
    emptyEl.hidden = !empty; if (!empty) emptyEl.textContent = t('chart_empty');
    canvas.hidden = empty; if (empty) return;
    const dpr = window.devicePixelRatio || 1, W = canvas.clientWidth || 600, H = 200;
    canvas.width = W * dpr; canvas.height = H * dpr;
    const ctx = canvas.getContext('2d'); ctx.scale(dpr, dpr); ctx.clearRect(0, 0, W, H);
    const css = getComputedStyle(document.documentElement);
    const fg = css.getPropertyValue('--muted').trim(), line = css.getPropertyValue('--line').trim(), down = css.getPropertyValue('--accent').trim(), up = '#3b8bd9';
    const padL = 56, padR = 8, padT = 8, padB = 46, iw = W - padL - padR, ih = H - padT - padB;
    const max = Math.max(1, ...series.map((p) => p.up + p.down));
    ctx.strokeStyle = line; ctx.fillStyle = fg; ctx.font = '11px -apple-system, sans-serif'; ctx.textAlign = 'right'; ctx.textBaseline = 'middle';
    for (let i = 0; i <= 4; i++) { const y = padT + ih - (ih * i) / 4; ctx.beginPath(); ctx.moveTo(padL, y); ctx.lineTo(W - padR, y); ctx.stroke(); ctx.fillText(fmtSize((max * i) / 4), padL - 6, y); }
    const n = series.length, bw = Math.max(1, iw / n - 2);
    series.forEach((p, i) => {
      const x = padL + (iw * i) / n + 1;
      const hd = (ih * p.down) / max, hu = (ih * p.up) / max;
      ctx.fillStyle = down; ctx.fillRect(x, padT + ih - hd, bw, hd);
      ctx.fillStyle = up; ctx.fillRect(x, padT + ih - hd - hu, bw, hu);
    });
    ctx.fillStyle = fg; ctx.textAlign = 'center'; ctx.textBaseline = 'top';
    const every = Math.max(1, Math.ceil(n / Math.floor(iw / 64)));
    series.forEach((p, i) => {
      if (i % every) return;
      const d = new Date(p.t);
      const label = step === '1h0m0s' ? d.toLocaleTimeString(locale(), { hour: '2-digit', minute: '2-digit' }) : d.toLocaleDateString(locale(), { day: 'numeric', month: 'short' });
      ctx.fillText(label, padL + (iw * (i + 0.5)) / n, padT + ih + 6);
    });
    ctx.textAlign = 'left'; ctx.fillStyle = down; ctx.fillRect(padL, H - 12, 10, 10); ctx.fillStyle = fg; ctx.fillText(t('legend_down'), padL + 14, H - 12);
    ctx.fillStyle = up; ctx.fillRect(padL + 110, H - 12, 10, 10); ctx.fillStyle = fg; ctx.fillText(t('legend_up'), padL + 124, H - 12);
  }
  async function loadDeviceLog(id) {
    try {
      const r = await api('/devices/' + encodeURIComponent(id) + '/log?limit=200');
      const tb = $('device-log').querySelector('tbody'); tb.innerHTML = '';
      const ev = r.events || [];
      $('device-log-empty').hidden = ev.length > 0; $('device-log').hidden = ev.length === 0;
      for (const e of ev) {
        const tr = document.createElement('tr'); if (e.level !== 'info') tr.className = e.level;
        [fmtDate(e.time), t('ev_' + e.kind) || e.kind, e.detail].forEach((v, i) => { const td = document.createElement('td'); td.textContent = v; if (i === 2) td.className = 'wrap'; tr.appendChild(td); });
        tb.appendChild(tr);
      }
    } catch (err) { $('device-log-empty').hidden = false; $('device-log-empty').textContent = t('log_unavailable', { err: err.message }); }
  }

  // ---- syncs (0.14): the folder syncs / backups the apps run ------------------
  const SYNC_ICONS = { folder: '⇄', photos: '▣', downloads: '⤓' };
  const syncKind = (k) => t('kind_' + k) || k;
  const syncMode = (m) => t('mode_' + m) || m;
  const syncState = (x) => {
    if (!x.enabled) return ['warning', x.disabledBy ? t('sync_paused_by', { who: x.disabledBy }) : t('sync_paused')];
    if (!x.lastRun) return ['', t('sync_never')];
    if (x.lastRun.status === 'failed') return ['error', t('sync_failed')];
    if (x.lastRun.errors || x.lastRun.conflicts || x.lastRun.status === 'partial') return ['warning', t('sync_partial', { c: x.lastRun.conflicts, e: x.lastRun.errors })];
    return ['ok', t('sync_ok')];
  };
  async function loadSyncs(containerId, emptyId, filter) {
    const wrap = $(containerId); wrap.innerHTML = '';
    try {
      const q = Object.entries(filter).map(([k, v]) => k + '=' + encodeURIComponent(v)).join('&');
      const r = await api('/syncs?' + q); const list = r.syncs || [];
      $(emptyId).hidden = list.length > 0;
      const now = new Date();
      for (const x of list) {
        const [cls, state] = syncState(x);
        const row = document.createElement('div'); row.className = 'sync-row' + (x.enabled ? '' : ' paused');
        row.innerHTML = `<span class="sync-icon">${SYNC_ICONS[x.kind] || '⇄'}</span>
          <span class="sync-main"><span class="sync-title"></span><span class="sync-sub muted small"></span><span class="sync-last muted small"></span></span>
          <span class="dot ${cls}"></span>
          <label class="switch"><input type="checkbox" ${x.enabled ? 'checked' : ''}><span></span></label>
          <button class="ghost small-btn sync-files"></button>`;
        row.querySelector('.sync-title').textContent = `${x.localLabel} ${x.mode === 'down' ? '←' : x.mode === 'both' ? '⇄' : '→'} ${x.remotePath}`;
        const dev = devicesCache.devices.find((d) => d.id === x.deviceId);
        row.querySelector('.sync-sub').textContent = [syncKind(x.kind), syncMode(x.mode), dev ? splitName(dev.name)[0] : x.deviceId, x.user, x.conflicts ? t('conflict_' + x.conflicts) : null, x.hashCheck ? t('hash_check') : null].filter(Boolean).join(' · ');
        row.querySelector('.sync-last').textContent = x.lastRun ? t('sync_last', { when: ago(x.lastRun.at, now), n: x.lastRun.files, bytes: fmtSize(x.lastRun.bytes) }) + ' · ' + state : state;
        row.querySelector('.dot').title = state;
        const sw = row.querySelector('input');
        sw.onchange = async () => {
          try { await api('/syncs/' + encodeURIComponent(x.id), { method: 'PATCH', body: JSON.stringify({ enabled: sw.checked }) }); loadSyncs(containerId, emptyId, filter); }
          catch (err) { alert(err.message); sw.checked = !sw.checked; }
        };
        const fb = row.querySelector('.sync-files'); fb.textContent = t('sync_files_btn', { n: (x.files || []).length });
        const files = document.createElement('div'); files.className = 'sync-files-list'; files.hidden = true;
        if ((x.files || []).length) {
          const tb = document.createElement('table'); tb.innerHTML = `<thead><tr><th>${t('col_when')}</th><th>${t('col_file')}</th><th>${t('col_direction')}</th><th class="num">${t('col_bytes')}</th><th>${t('col_outcome')}</th></tr></thead><tbody></tbody>`;
          for (const f of x.files) {
            const tr = document.createElement('tr'); if (f.outcome !== 'ok') tr.className = f.outcome === 'failed' ? 'error' : 'warning';
            [fmtDate(f.time), f.path, t('dir_' + f.direction) || f.direction, fmtSize(f.bytes || 0), (t('out_' + f.outcome) || f.outcome) + (f.detail ? ' · ' + f.detail : '')].forEach((v, i) => { const td = document.createElement('td'); td.textContent = v; if (i === 1) td.className = 'wrap'; if (i === 3) td.className = 'num'; tr.appendChild(td); });
            tb.querySelector('tbody').appendChild(tr);
          }
          const tw = document.createElement('div'); tw.className = 'table-wrap'; tw.appendChild(tb); files.appendChild(tw);
        } else { files.innerHTML = `<p class="muted small">${t('sync_no_files')}</p>`; }
        fb.onclick = () => { files.hidden = !files.hidden; };
        wrap.appendChild(row); wrap.appendChild(files);
      }
    } catch (err) { $(emptyId).hidden = false; $(emptyId).textContent = t('syncs_unavailable', { err: err.message }); }
  }

  // ---- users (0.14): per-user view ---------------------------------------------
  let currentUser = '';
  async function loadUsers() {
    try {
      const r = await api('/users'); const users = r.users || [];
      const tb = $('users').querySelector('tbody'); tb.innerHTML = '';
      $('users-empty').hidden = users.length > 0; $('users').hidden = users.length === 0;
      const now = new Date();
      for (const u of users) {
        const tr = document.createElement('tr');
        [u.user, String(u.devices), u.syncs ? `${u.syncs}${u.pausedSyncs ? ' (' + t('n_paused', { n: u.pausedSyncs }) + ')' : ''}` : '0', u.lastSeen ? ago(u.lastSeen, now) : '—', '↓ ' + fmtSize(u.today.down) + '  ↑ ' + fmtSize(u.today.up)].forEach((v, i) => {
          const td = document.createElement('td');
          if (i === 0) { const b = document.createElement('button'); b.className = 'user-link'; b.textContent = v; b.title = t('users_devices_title', { name: v }); b.onclick = () => showUserDialog(v); td.appendChild(b); } else td.textContent = v;
          if (i === 1 || i === 2) td.className = 'num'; tr.appendChild(td);
        });
        const act = document.createElement('td'); act.className = 'actions';
        const dot = document.createElement('span'); dot.className = 'dot ' + (u.warnings ? 'warning' : 'ok'); dot.title = u.warnings ? t('n_warnings', { n: u.warnings }) : t('all_fine'); dot.style.marginRight = '8px';
        const b = document.createElement('button'); b.className = 'ghost'; b.textContent = t('open'); b.onclick = () => openUser(u.user);
        act.appendChild(dot); act.appendChild(b); tr.appendChild(act); tb.appendChild(tr);
      }
      filterRows(['users'], $('users-filter').value);
    } catch (err) { $('users-scope').textContent = t('users_unavailable', { err: err.message }); }
  }
  bindFilter('users-filter', ['users']);
  $('users-refresh').addEventListener('click', loadUsers);
  // Popup with a user's devices (0.17): the same cards as the Devices group, without leaving the page.
  async function showUserDialog(name) {
    const dlg = $('user-dialog'); $('user-dialog-title').textContent = t('users_devices_title', { name }); const wrap = $('user-dialog-devices'); wrap.innerHTML = '';
    $('user-dialog-open').onclick = () => { dlg.close(); openUser(name); };
    dlg.showModal();
    try {
      const u = await api('/users/' + encodeURIComponent(name)); const now = new Date();
      for (const x of u.devices || []) {
        const [title, sub, app] = splitName(x.name); const [cls, state] = deviceState(x, now);
        const card = document.createElement('button'); card.type = 'button'; card.className = 'dev-card' + (x.supersededBy ? ' old' : '');
        card.innerHTML = `<span class="dev-icon">${DEV_ICONS[deviceKind(x.name)]}</span><span class="dev-main"><span class="dev-title"></span><span class="dev-sub muted small"></span><span class="dev-meta muted small"></span></span><span class="dot ${cls}"></span>`;
        card.querySelector('.dev-title').textContent = title; card.querySelector('.dev-sub').textContent = [sub, app].filter(Boolean).join(' · ');
        card.querySelector('.dev-meta').textContent = t('last_sync') + ': ' + ago(x.lastSeen, now) + ' · ' + state;
        card.onclick = () => { dlg.close(); if (!devicesCache.devices.some((d) => d.id === x.id)) devicesCache.devices.push(x); openDevice(x.id); };
        wrap.appendChild(card);
      }
      if (!(u.devices || []).length) wrap.innerHTML = `<p class="muted small">${t('no_devices')}</p>`;
    } catch (err) { wrap.innerHTML = `<p class="error">${err.message}</p>`; }
  }
  $('user-dialog-close').addEventListener('click', () => $('user-dialog').close());
  $('user-dialog').addEventListener('click', (e) => { if (e.target === $('user-dialog')) $('user-dialog').close(); });
  function closeUser() { currentUser = ''; $('user-detail').hidden = true; if ($('device-detail').hidden) showGlobals(true); }
  $('user-back').addEventListener('click', closeUser);
  $('user-range').addEventListener('change', () => { if (currentUser) loadUserMetrics(currentUser); });
  async function openUser(name) {
    currentUser = name;
    $('user-title').textContent = name;
    showGlobals(false); $('device-detail').hidden = true; $('user-detail').hidden = false;
    try {
      const u = await api('/users/' + encodeURIComponent(name));
      const now = new Date();
      const wrap = $('user-devices'); wrap.innerHTML = '';
      let warn = 0;
      for (const x of u.devices || []) {
        const [title, sub, app] = splitName(x.name);
        const [cls, state] = deviceState(x, now); if (cls === 'warning') warn++;
        const card = document.createElement('div'); card.className = 'dev-card' + (x.supersededBy ? ' old' : '');
        card.innerHTML = `<span class="dev-icon">${DEV_ICONS[deviceKind(x.name)]}</span>
          <span class="dev-main"><span class="dev-title"></span><span class="dev-sub muted small"></span><span class="dev-meta muted small"></span></span>
          <span class="dot ${cls}"></span><button class="ghost small-btn"></button>`;
        card.querySelector('.dev-title').textContent = title;
        card.querySelector('.dev-sub').textContent = [sub, app].filter(Boolean).join(' · ');
        card.querySelector('.dev-meta').textContent = t('last_sync') + ': ' + ago(x.lastSeen, now) + ' · ' + state;
        const rb = card.querySelector('button'); rb.textContent = t('remove');
        rb.onclick = async () => {
          if (!confirm(t('confirm_remove_device', { name: x.name }) + (x.id === devicesCache.thisDevice ? t('confirm_remove_this') : ''))) return;
          try { await api('/devices/' + encodeURIComponent(x.id), { method: 'DELETE' }); if (x.id === devicesCache.thisDevice) logout(); else { await loadDevices(); openUser(name); } } catch (err) { alert(err.message); }
        };
        card.querySelector('.dev-main').onclick = () => { closeUser(); openDevice(x.id); };
        wrap.appendChild(card);
      }
      const bad = (u.syncs || []).filter((x) => !x.enabled || (x.lastRun && (x.lastRun.status === 'failed' || x.lastRun.errors || x.lastRun.conflicts))).length;
      $('user-dot').className = 'dot ' + (warn + bad ? 'warning' : 'ok'); $('user-dot').title = warn + bad ? t('n_warnings', { n: warn + bad }) : t('all_fine');
      const tot = $('user-totals'); tot.innerHTML = '';
      for (const [k, v] of [['today', u.today], ['week', u.week], ['month', u.month]]) {
        const s = document.createElement('div'); s.className = 'stat';
        s.innerHTML = `<span class="k"></span><span class="v"></span><span class="muted small"></span>`;
        s.querySelector('.k').textContent = t('tot_' + k);
        s.querySelector('.v').textContent = '↓ ' + fmtSize(v.down) + '  ↑ ' + fmtSize(v.up);
        s.querySelector('.muted').textContent = t('tot_detail', { req: v.requests, dl: v.downloads + v.streams, ul: v.uploads });
        tot.appendChild(s);
      }
      await Promise.all([loadSyncs('user-syncs', 'user-syncs-empty', { user: name, all: 1 }), loadUserMetrics(name), loadUserLog(name)]);
    } catch (err) { alert(err.message); closeUser(); }
  }
  async function loadUserMetrics(name) {
    try {
      const m = await api('/users/' + encodeURIComponent(name) + '/metrics?range=' + $('user-range').value);
      drawChart($('user-chart'), m.series || [], m.step, $('user-chart-empty'));
    } catch (err) { $('user-chart-empty').hidden = false; $('user-chart-empty').textContent = t('metrics_unavailable', { err: err.message }); }
  }
  // The user log is fetched once (up to 500 events) and filtered/paginated in the browser (0.15).
  let userLog = { events: [], shown: 0 };
  const ULOG_PAGE = 50;
  function renderUserLog(more) {
    const dev = $('ulog-device').value, kind = $('ulog-event').value, q = norm($('ulog-search').value).trim();
    const rows = userLog.events.filter((e) => (!dev || e.deviceName === dev) && (!kind || e.kind === kind) && (!q || norm([fmtDate(e.time), splitName(e.deviceName)[0], t('ev_' + e.kind), e.detail].join(' ')).includes(q)));
    userLog.shown = more ? userLog.shown + ULOG_PAGE : ULOG_PAGE;
    const tb = $('user-log').querySelector('tbody'); tb.innerHTML = '';
    for (const e of rows.slice(0, userLog.shown)) {
      const tr = document.createElement('tr'); if (e.level !== 'info') tr.className = e.level;
      [fmtDate(e.time), splitName(e.deviceName)[0], t('ev_' + e.kind) || e.kind, e.detail].forEach((v, i) => { const td = document.createElement('td'); td.textContent = v; if (i === 3) td.className = 'wrap'; tr.appendChild(td); });
      tb.appendChild(tr);
    }
    $('user-log-empty').hidden = rows.length > 0; $('user-log').hidden = rows.length === 0;
    $('user-log-count').textContent = t('n_events', { n: rows.length });
    const left = rows.length - userLog.shown;
    $('ulog-more').hidden = left <= 0; $('ulog-more').textContent = t('show_more', { n: Math.max(0, left) });
  }
  async function loadUserLog(name) {
    try {
      const r = await api('/users/' + encodeURIComponent(name) + '/log?limit=500');
      userLog.events = r.events || [];
      const fill = (sel, values, label) => { const cur = sel.value; sel.length = 1; for (const v of values) { const o = document.createElement('option'); o.value = v; o.textContent = label(v); sel.appendChild(o); } sel.value = values.includes(cur) ? cur : ''; };
      fill($('ulog-device'), [...new Set(userLog.events.map((e) => e.deviceName).filter(Boolean))].sort(), (v) => splitName(v)[0]);
      fill($('ulog-event'), [...new Set(userLog.events.map((e) => e.kind))].sort(), (v) => t('ev_' + v) || v);
      renderUserLog(false);
    } catch (err) { $('user-log-empty').hidden = false; $('user-log-empty').textContent = t('log_unavailable', { err: err.message }); }
  }
  ['ulog-device', 'ulog-event'].forEach((id) => $(id).addEventListener('change', () => renderUserLog(false)));
  $('ulog-search').addEventListener('input', () => renderUserLog(false));
  $('ulog-more').addEventListener('click', () => renderUserLog(true));

  // ---- notifications --------------------------------------------------------
  async function loadNotifyChannels() {
    try { const c = await api('/notify/channels'); $('notify-channels').textContent = c.channels.length ? t('channels', { list: c.channels.join(', ') }) : t('no_channel'); } catch {}
  }

  // Notification channels editable from the web UI (0.18): secrets never come back, only "set".
  async function loadNotifySettings() {
    try {
      const x = await api('/notify/settings');
      $('notify-form').hidden = false; $('notify-source').textContent = t(x.source === 'file' ? 'notify_source_file' : 'notify_source_env');
      $('nf-unraid').checked = !!x.unraidEnabled; $('nf-unraid-key').value = ''; $('nf-unraid-key').placeholder = x.hasUnraidApiKey ? t('secret_set') : '';
      $('nf-smtp-host').value = x.smtpHost || ''; $('nf-smtp-port').value = x.smtpPort || ''; $('nf-smtp-tls').value = x.smtpTls || 'starttls'; $('nf-smtp-user').value = x.smtpUser || '';
      $('nf-smtp-password').value = ''; $('nf-smtp-password').placeholder = x.hasSmtpPassword ? t('secret_set') : ''; $('nf-smtp-from').value = x.smtpFrom || ''; $('nf-smtp-to').value = x.smtpTo || '';
      $('nf-tg-token').value = ''; $('nf-tg-token').placeholder = x.hasTelegramToken ? t('secret_set') : ''; $('nf-tg-chat').value = x.telegramChatId || '';
      $('notify-channels').textContent = x.channels.length ? t('channels', { list: x.channels.join(', ') }) : t('no_channel');
    } catch (err) { $('notify-form').hidden = true; $('notify-source').textContent = err.message.includes('admin') ? t('notify_admin_only') : t('error', { err: err.message }); }
  }
  $('notify-form').addEventListener('submit', async (e) => {
    e.preventDefault(); $('notify-saved').textContent = t('sending');
    const body = { unraidEnabled: $('nf-unraid').checked, unraidApiKey: $('nf-unraid-key').value, smtpHost: $('nf-smtp-host').value, smtpPort: $('nf-smtp-port').value, smtpTls: $('nf-smtp-tls').value, smtpUser: $('nf-smtp-user').value, smtpPassword: $('nf-smtp-password').value, smtpFrom: $('nf-smtp-from').value, smtpTo: $('nf-smtp-to').value, telegramToken: $('nf-tg-token').value, telegramChatId: $('nf-tg-chat').value };
    try { await api('/notify/settings', { method: 'PUT', body: JSON.stringify(body) }); $('notify-saved').textContent = t('notify_saved'); await loadNotifySettings(); } catch (err) { $('notify-saved').textContent = t('error', { err: err.message }); }
  });
  $('notify-reset').addEventListener('click', async () => {
    if (!confirm(t('confirm_notify_reset'))) return;
    try { await api('/notify/settings', { method: 'DELETE' }); $('notify-saved').textContent = t('notify_saved'); await loadNotifySettings(); } catch (err) { $('notify-saved').textContent = t('error', { err: err.message }); }
  });
  $('notify-test').addEventListener('click', async (e) => {
    e.preventDefault(); e.stopPropagation(); $('notify-card').open = true;
    $('notify-result').hidden = false; $('notify-result').textContent = t('sending');
    try {
      const r = await api('/notify/test', { method: 'POST' });
      $('notify-result').textContent = r.channels.length === 0 ? t('no_channel_long') : (r.errors.length ? t('errors', { list: r.errors.join(' · ') }) : t('sent_to', { list: r.channels.join(', ') }));
    } catch (err) { $('notify-result').textContent = t('error', { err: err.message }); }
  });

  // ---- api keys -------------------------------------------------------------
  async function loadKeys() {
    $('keys-error').hidden = true;
    try {
      const r = await api('/keys');
      const items = (r && r.apiKeys) || [];
      const tb = $('keys').querySelector('tbody'); tb.innerHTML = ''; $('keys').hidden = false;
      for (const k of items) {
        const tr = document.createElement('tr');
        [k.name, (k.roles || []).join(', '), k.createdAt ? fmtDate(k.createdAt) : ''].forEach((v) => { const td = document.createElement('td'); td.textContent = v; tr.appendChild(td); });
        const act = document.createElement('td'); act.className = 'actions';
        const rot = document.createElement('button'); rot.className = 'ghost'; rot.textContent = t('rotate');
        rot.onclick = async () => {
          const name = prompt(t('rotate_name'), k.name + ' new'); if (!name) return;
          const remember = confirm(t('rotate_remember'));
          const del = confirm(t('rotate_delete', { name: k.name }));
          try {
            const res = await api('/keys/rotate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, roles: k.roles, remember, deleteOldId: del ? k.id : '' }) });
            showNewKey(res); loadKeys();
          } catch (err) { $('keys-error').textContent = err.message; $('keys-error').hidden = false; }
        };
        const delb = document.createElement('button'); delb.className = 'ghost'; delb.textContent = t('delete');
        delb.onclick = async () => {
          if (!confirm(t('confirm_delete_key', { name: k.name }))) return;
          try { await api('/keys/' + encodeURIComponent(k.id), { method: 'DELETE' }); loadKeys(); } catch (err) { $('keys-error').textContent = err.message; $('keys-error').hidden = false; }
        };
        act.append(rot, delb); tr.appendChild(act); tb.appendChild(tr);
      }
    } catch (err) { $('keys-error').textContent = err.message; $('keys-error').hidden = false; }
  }
  function showNewKey(res) {
    const box = $('key-new'); box.hidden = false;
    box.textContent = t('new_key', { name: res.name, roles: (res.roles || []).join(', ') }) + `\n\n${res.key || t('key_not_returned')}\n\n` + t('copy_now')
      + (res.remembered ? '\n' + t('remembered') : '') + (res.deletedOld ? '\n' + t('old_deleted') : '') + (res.deleteError ? '\n' + t('old_not_deleted', { err: res.deleteError }) : '') + (res.rememberError ? '\n' + t('not_remembered', { err: res.rememberError }) : '');
  }
  $('keys-refresh').addEventListener('click', loadKeys);
  $('key-create').addEventListener('click', async () => {
    const name = $('key-name').value.trim(); if (!name) { alert(t('give_name')); return; }
    $('keys-error').hidden = true;
    try {
      const r = await api('/keys', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, roles: [$('key-role').value] }) });
      const c = (r && r.apiKey && r.apiKey.create) || {};
      showNewKey({ name: c.name || name, roles: c.roles || [$('key-role').value], key: c.key });
      $('key-name').value = ''; loadKeys();
    } catch (err) { $('keys-error').textContent = err.message; $('keys-error').hidden = false; }
  });

  // ---- boot -----------------------------------------------------------------
  loadStatus();
  loadStoredKey();
  if (token) api('/auth/session').then((s) => showApp(s.identity, s.user, s.shares)).catch(() => logout());
})();
