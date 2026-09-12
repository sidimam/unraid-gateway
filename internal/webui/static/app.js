(() => {
  const $ = (id) => document.getElementById(id);
  const API = '/api/v1';

  // ---- i18n -----------------------------------------------------------------
  // The same seven languages as the Unraid Drive app. "system" follows the browser.
  const I18N = {
    en: {
      connecting: 'connecting…', online: 'gateway online', unreachable: 'gateway unreachable',
      system: 'System', theme_system: 'Theme: system', theme_light: 'Theme: light', theme_dark: 'Theme: dark',
      forget_key: 'Forget stored key', forget_key_title: 'Delete the key stored on the gateway', disconnect: 'Disconnect',
      login_title: 'Connect', login_intro: 'Use an Unraid API key (Settings → Management Access → API Keys). VIEWER is enough for files; ADMIN also manages the server and its keys. The key is validated by Unraid and exchanged for a session token.',
      stored_connect: 'Connect with the key stored on this gateway', stored_env: 'from the WEBUI_API_KEY variable', stored_file: 'remembered on this gateway',
      pm_account: 'Account name (for your password manager)', api_key: 'Unraid API key', api_key_ph: 'paste your Unraid API key',
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
      col_key: 'Key', col_registered: 'Registered', col_logins: 'Logins', no_devices: 'No device registered yet (older app builds do not register; update to 1.3 build 29 or later).',
      this_browser: '(this browser)', old: 'old', last_seen_prefix: 'last seen', replaced_title: 'Replaced by a new installation on {date}. Safe to remove.', registration_off: 'registration is off (DEVICE_REGISTRATION=off)', devices_unavailable: 'devices unavailable: {err}',
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
    },
    it: {
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
      col_key: 'Chiave', col_registered: 'Registrato', col_logins: 'Accessi', no_devices: 'Nessun dispositivo registrato (le build vecchie dell\'app non si registrano: aggiorna alla 1.3 build 29 o successiva).',
      this_browser: '(questo browser)', old: 'vecchio', last_seen_prefix: 'ultimo accesso', replaced_title: 'Sostituito da una nuova installazione il {date}. Si può rimuovere.', registration_off: 'registrazione disattivata (DEVICE_REGISTRATION=off)', devices_unavailable: 'dispositivi non disponibili: {err}',
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
    },
    es: {
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
    loadStatus(); if (token && !$('app').hidden) { list(cwd); loadActivity(); loadDevices(); loadNotifyChannels(); }
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

  // ---- tabs -----------------------------------------------------------------
  function showTab(name) {
    document.querySelectorAll('#tabs button').forEach((b) => b.classList.toggle('active', b.dataset.tab === name));
    document.querySelectorAll('.tab').forEach((c) => { c.hidden = c.id !== 'tab-' + name; });
    localStorage.setItem('ugw.tab', name);
    if (name === 'devices') loadDevices();
    if (name === 'keys' && !$('keys').hidden) loadKeys();
  }
  $('tabs').addEventListener('click', (e) => { const b = e.target.closest('button'); if (b) showTab(b.dataset.tab); });

  // ---- status ---------------------------------------------------------------
  async function loadStatus() {
    try {
      const s = await (await fetch('/healthz')).json();
      $('version').textContent = s.version || '';
      $('status').textContent = t('online');
      $('st-gateway').textContent = (s.version || '') + ' · ' + location.host;
    } catch { $('status').textContent = t('unreachable'); }
    $('origin').textContent = location.origin;
    $('tls-note').textContent = location.protocol === 'https:' ? t('tls_ok') : t('tls_no');
  }

  // ---- auth -----------------------------------------------------------------
  function showApp(identity, user, shares) {
    $('login').hidden = true; $('app').hidden = false; $('who').hidden = false;
    const key = identity ? `${identity.name || t('key_label')} · ${(identity.roles || []).join(', ')}` : '';
    $('who-name').textContent = user ? `${user} · ${key}` : key;
    window.ugwShares = shares || null; // share → 'rw' | 'ro' when a user is logged in
    showTab(localStorage.getItem('ugw.tab') || 'files');
    list(cwd);
    startActivity();
    loadDevices();
    loadNotifyChannels();
    $('forget').hidden = !(storedInfo.available && storedInfo.canForget);
  }
  function logout() {
    if (token) fetch(API + '/auth/logout', { method: 'POST', headers: { Authorization: 'Bearer ' + token } }).catch(() => {});
    token = ''; sessionStorage.removeItem('ugw.token'); clearInterval(activityTimer);
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
        const act = document.createElement('td'); act.className = 'actions';
        const btn = (label, fn) => { const b = document.createElement('button'); b.className = 'ghost'; b.textContent = label; b.onclick = fn; act.appendChild(b); };
        if (cwd !== '/') { btn(t('rename'), () => rename(e)); btn(t('delete'), () => remove(e)); }
        tr.append(name, size, mt, act); tb.appendChild(tr);
      }
    } catch (err) { alert(err.message); }
  }

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
    } catch (err) { $('activity-scope').textContent = t('activity_unavailable', { err: err.message }); }
  }
  function startActivity() {
    clearInterval(activityTimer); loadActivity();
    activityTimer = setInterval(() => { if ($('activity-live').checked && !document.hidden) loadActivity(); }, 3000);
  }
  $('activity-live').addEventListener('change', () => { if ($('activity-live').checked) loadActivity(); });

  // ---- devices --------------------------------------------------------------
  async function loadDevices() {
    try {
      const d = await api('/devices');
      const now = new Date();
      $('devices-scope').textContent = d.registration === 'off' ? t('registration_off') : '';
      $('st-devices').textContent = String(d.devices.filter((x) => !x.supersededBy).length);
      const tb = $('devices').querySelector('tbody'); tb.innerHTML = '';
      $('devices-empty').hidden = d.devices.length > 0;
      $('devices-remove-all').hidden = d.devices.length === 0;
      for (const x of d.devices) {
        const tr = document.createElement('tr');
        const mine = x.id === d.thisDevice;
        const old = !!x.supersededBy;
        if (old) tr.className = 'old';
        const label = x.name + (mine ? ' ' + t('this_browser') : '') + (old ? ' — ' + t('old') : '');
        [label, x.user || '—', x.key || '—', ago(x.firstSeen, now), (old ? t('last_seen_prefix') + ' ' : '') + ago(x.lastSeen, now), x.lastIp || '', String(x.logins)].forEach((v, i) => {
          const td = document.createElement('td'); td.textContent = v; if (i === 0) td.className = 'wrap'; if (i === 6) td.className = 'num';
          if (i === 0 && old) td.title = t('replaced_title', { date: new Date(x.supersededAt).toLocaleString(locale()) });
          tr.appendChild(td);
        });
        const act = document.createElement('td'); act.className = 'actions';
        const b = document.createElement('button'); b.className = 'ghost'; b.textContent = t('remove');
        b.onclick = async () => {
          if (!confirm(t('confirm_remove_device', { name: x.name }) + (mine ? t('confirm_remove_this') : ''))) return;
          try { await api('/devices/' + encodeURIComponent(x.id), { method: 'DELETE' }); if (mine) logout(); else loadDevices(); } catch (err) { alert(err.message); }
        };
        act.appendChild(b); tr.appendChild(act); tb.appendChild(tr);
      }
    } catch (err) { $('devices-scope').textContent = t('devices_unavailable', { err: err.message }); }
  }
  $('devices-refresh').addEventListener('click', loadDevices);
  $('devices-remove-all').addEventListener('click', async () => {
    if (!confirm(t('confirm_remove_all'))) return;
    try { await api('/devices', { method: 'DELETE' }); logout(); } catch (err) { alert(err.message); }
  });

  // ---- notifications --------------------------------------------------------
  async function loadNotifyChannels() {
    try { const c = await api('/notify/channels'); $('notify-channels').textContent = c.channels.length ? t('channels', { list: c.channels.join(', ') }) : t('no_channel'); } catch {}
  }
  $('notify-test').addEventListener('click', async () => {
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
