(() => {
  const $ = (id) => document.getElementById(id);
  const API = '/api/v1';
  let token = sessionStorage.getItem('ugw.token') || '';
  let cwd = '/';

  const fmtSize = (n) => {
    if (n < 1024) return n + ' B';
    const u = ['KB', 'MB', 'GB', 'TB']; let i = -1;
    do { n /= 1024; i++; } while (n >= 1024 && i < u.length - 1);
    return n.toFixed(n < 10 ? 1 : 0) + ' ' + u[i];
  };
  const fmtDate = (s) => new Date(s).toLocaleString();
  const encPath = (p) => encodeURIComponent(p);

  async function api(path, opts = {}) {
    opts.headers = Object.assign({}, opts.headers, token ? { Authorization: 'Bearer ' + token } : {});
    const res = await fetch(API + path, opts);
    if (res.status === 401 && token) { logout(); throw new Error('session expired'); }
    if (res.status === 204) return null;
    const ct = res.headers.get('content-type') || '';
    const body = ct.includes('json') ? await res.json() : await res.text();
    if (!res.ok) throw new Error((body && body.error) || res.statusText);
    return body;
  }

  // ---- status ---------------------------------------------------------------
  async function loadStatus() {
    try {
      const s = await (await fetch('/healthz')).json();
      $('version').textContent = s.version || '';
      $('status').textContent = 'gateway online';
    } catch { $('status').textContent = 'gateway unreachable'; }
    $('origin').textContent = location.origin;
    $('tls-note').textContent = location.protocol === 'https:'
      ? 'Served over HTTPS: the Files app extension will accept this URL.'
      : 'You are on plain HTTP. iOS extensions require a valid HTTPS certificate: put Nginx Proxy Manager, SWAG or Cloudflare Tunnel in front before using the app remotely.';
  }

  // ---- auth -----------------------------------------------------------------
  function showApp(identity, user, shares) {
    $('login').hidden = true; $('app').hidden = false; $('who').hidden = false;
    const key = identity ? `${identity.name || 'api key'} · ${(identity.roles || []).join(', ')}` : '';
    $('who-name').textContent = user ? `${user} · ${key}` : key;
    window.ugwShares = shares || null; // share → 'rw' | 'ro' when a user is logged in
    list(cwd);
    startActivity();
  }
  function logout() {
    if (token) fetch(API + '/auth/logout', { method: 'POST', headers: { Authorization: 'Bearer ' + token } }).catch(() => {});
    token = ''; sessionStorage.removeItem('ugw.token'); clearInterval(activityTimer);
    $('login').hidden = false; $('app').hidden = true; $('who').hidden = true;
  }
  $('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    $('login-error').hidden = true;
    try {
      const body = { apiKey: $('apikey').value.trim() };
      if ($('username').value.trim()) { body.username = $('username').value.trim(); body.password = $('password').value; }
      const r = await api('/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      token = r.token; sessionStorage.setItem('ugw.token', token);
      $('apikey').value = ''; $('password').value = '';
      showApp(r.identity, r.user, r.shares);
    } catch (err) { $('login-error').textContent = err.message; $('login-error').hidden = false; }
  });
  $('logout').addEventListener('click', logout);

  // ---- browser --------------------------------------------------------------
  function crumbs(path) {
    const parts = path.split('/').filter(Boolean);
    const nav = $('crumbs'); nav.innerHTML = '';
    const mk = (label, p) => { const a = document.createElement('a'); a.href = '#'; a.textContent = label; a.onclick = (e) => { e.preventDefault(); list(p); }; return a; };
    nav.appendChild(mk('shares', '/'));
    let acc = '';
    parts.forEach((p) => { acc += '/' + p; nav.appendChild(Object.assign(document.createElement('span'), { textContent: '/' })); nav.appendChild(mk(p, acc)); });
  }

  async function list(path) {
    try {
      const r = await api('/fs/list?path=' + encPath(path));
      cwd = r.path; crumbs(cwd);
      // The root only lists mounted shares: writing happens inside a share.
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
        size.textContent = e.type === 'dir' ? (ro ? 'read-only' : '') : fmtSize(e.size);
        const mt = document.createElement('td'); mt.textContent = fmtDate(e.mtime);
        const act = document.createElement('td'); act.className = 'actions';
        const btn = (label, fn) => { const b = document.createElement('button'); b.className = 'ghost'; b.textContent = label; b.onclick = fn; act.appendChild(b); };
        if (cwd !== '/') { // share roots are mount points: no rename/delete
          btn('Rename', () => rename(e));
          btn('Delete', () => remove(e));
        }
        tr.append(name, size, mt, act); tb.appendChild(tr);
      }
    } catch (err) { alert(err.message); }
  }

  async function download(e) {
    // Files are fetched with the bearer header and handed to the browser as a blob.
    try {
      const res = await fetch(API + '/fs/content?path=' + encPath(e.path), { headers: { Authorization: 'Bearer ' + token } });
      if (!res.ok) throw new Error((await res.json()).error);
      const url = URL.createObjectURL(await res.blob());
      const a = document.createElement('a'); a.href = url; a.download = e.name; document.body.appendChild(a); a.click(); a.remove();
      setTimeout(() => URL.revokeObjectURL(url), 10000);
    } catch (err) { alert(err.message); }
  }

  async function rename(e) {
    const to = prompt('New name', e.name); if (!to || to === e.name) return;
    const dir = e.path.slice(0, e.path.lastIndexOf('/')) || '';
    try { await api('/fs/move', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ from: e.path, to: dir + '/' + to }) }); list(cwd); }
    catch (err) { alert(err.message); }
  }
  async function remove(e) {
    if (!confirm(`Delete ${e.type === 'dir' ? 'folder' : 'file'} "${e.name}"${e.type === 'dir' ? ' and everything inside it' : ''}?`)) return;
    try { await api('/fs/delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path: e.path, recursive: true }) }); list(cwd); }
    catch (err) { alert(err.message); }
  }
  $('mkdir').addEventListener('click', async () => {
    if (cwd === '/') { alert('The root only lists mounted shares. Open a share first.'); return; }
    const name = prompt('Folder name'); if (!name) return;
    try { await api('/fs/mkdir', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ path: (cwd === '/' ? '' : cwd) + '/' + name }) }); list(cwd); }
    catch (err) { alert(err.message); }
  });
  $('refresh').addEventListener('click', () => list(cwd));

  // Upload uses XHR for a progress bar; whole-file PUT is atomic on the server.
  $('upload').addEventListener('change', async (ev) => {
    const files = [...ev.target.files]; ev.target.value = '';
    if (cwd === '/') { alert('Pick a share first: the root only lists mounted shares.'); return; }
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
        xhr.onerror = () => { alert('upload failed: ' + f.name); resolve(); };
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
    } catch (err) { $('gql-out').textContent = 'error: ' + err.message; }
  });

  // ---- activity -------------------------------------------------------------
  // Who is connected, from which device, what is streaming. Polled every 3 s while "live" is on.
  const ago = (iso, now) => {
    const d = Math.max(0, (now - new Date(iso)) / 1000);
    if (d < 5) return 'now'; if (d < 60) return Math.round(d) + ' s ago';
    if (d < 3600) return Math.round(d / 60) + ' min ago'; if (d < 86400) return Math.round(d / 3600) + ' h ago';
    return fmtDate(iso);
  };
  const dur = (sec) => { sec = Math.max(0, Math.round(sec)); const h = Math.floor(sec / 3600), m = Math.floor(sec % 3600 / 60), s = sec % 60; return (h ? h + ':' : '') + String(m).padStart(h ? 2 : 1, '0') + ':' + String(s).padStart(2, '0'); };
  const device = (c) => {
    // "Unraid Drive 1.3 (28) · iPhone 17 Pro · iOS 26.1 · App" from the apps; anything else is a browser/other UA.
    if (!c.agent) return 'unknown';
    if (c.agent.startsWith('Unraid Drive')) return c.agent;
    if (/CFNetwork/.test(c.agent)) return 'Unraid Drive (older build) · ' + (c.agent.match(/Darwin\/[\d.]+/) || [''])[0];
    if (/Mozilla/.test(c.agent)) return 'Browser · ' + ((c.agent.match(/(Firefox|Edg|Chrome|Safari)\/[\d.]+/) || ['browser'])[0]);
    if (/mpv|libmpv/i.test(c.agent)) return 'mpv player';
    return c.agent.slice(0, 60);
  };
  const who = (t) => (t.user ? t.user : (t.key ? 'key: ' + t.key : '—'));
  const kindLabel = { download: '⬇︎ download', stream: '▶︎ stream', upload: '⬆︎ upload' };
  let activityTimer = null;
  const speeds = {}; // id → [bytes, time]
  async function loadActivity() {
    if (!token || $('app').hidden) return;
    try {
      const a = await api('/activity');
      const now = new Date(a.now);
      $('activity-scope').textContent = a.scope === 'user' ? 'your devices only' : 'all users';
      const tc = $('act-clients').querySelector('tbody'); tc.innerHTML = '';
      $('act-clients-empty').hidden = a.clients.length > 0;
      for (const c of a.clients) {
        const tr = document.createElement('tr');
        const cells = [device(c), c.user || (c.key ? 'key: ' + c.key : '—'), c.ip, ago(c.firstSeen, now), ago(c.lastSeen, now), String(c.requests), c.lastPath ? `${c.lastPath} (${ago(c.lastPathAt, now)})` : ''];
        cells.forEach((v, i) => { const td = document.createElement('td'); td.textContent = v; if (i === 5) td.className = 'num'; if (i === 6 || i === 0) td.className = 'wrap'; tr.appendChild(td); });
        if (c.active) tr.classList.add('live');
        tc.appendChild(tr);
      }
      const ta = $('act-active').querySelector('tbody'); ta.innerHTML = '';
      $('act-active-empty').hidden = a.active.length > 0;
      for (const t of a.active) {
        const tr = document.createElement('tr');
        const prev = speeds[t.id]; speeds[t.id] = [t.bytes, now.getTime()];
        const speed = prev && now.getTime() > prev[1] ? (t.bytes - prev[0]) / ((now.getTime() - prev[1]) / 1000) : 0;
        const pct = t.size > 0 ? Math.min(100, Math.round(t.bytes / t.size * 100)) : null;
        const kind = document.createElement('td'); kind.textContent = kindLabel[t.kind] || t.kind;
        const file = document.createElement('td'); file.className = 'wrap'; file.textContent = t.path;
        const w = document.createElement('td'); w.className = 'wrap'; w.textContent = `${who(t)} · ${device(t)} · ${t.ip}`;
        const prog = document.createElement('td'); prog.className = 'wrap';
        prog.innerHTML = `<div class="bar"><div style="width:${pct ?? 0}%"></div></div><span class="small">${fmtSize(t.bytes)}${t.size > 0 ? ' / ' + fmtSize(t.size) + ' · ' + pct + '%' : ''}${t.range ? ' · ' + t.range : ''}</span>`;
        const sp = document.createElement('td'); sp.className = 'num'; sp.textContent = speed > 0 ? fmtSize(speed) + '/s' : '';
        const el = document.createElement('td'); el.className = 'num'; el.textContent = dur((now - new Date(t.started)) / 1000);
        tr.append(kind, file, w, prog, sp, el); ta.appendChild(tr);
      }
      for (const id of Object.keys(speeds)) if (!a.active.some((t) => String(t.id) === id)) delete speeds[id];
      const trc = $('act-recent').querySelector('tbody'); trc.innerHTML = '';
      for (const t of a.recent.slice(0, 30)) {
        const tr = document.createElement('tr');
        [kindLabel[t.kind] || t.kind, t.path, `${who(t)} · ${device(t)}`, fmtSize(t.bytes), ago(t.ended, now)].forEach((v, i) => { const td = document.createElement('td'); td.textContent = v; if (i === 3) td.className = 'num'; if (i === 1 || i === 2) td.className = 'wrap'; tr.appendChild(td); });
        trc.appendChild(tr);
      }
    } catch (err) { $('activity-scope').textContent = 'activity unavailable: ' + err.message; }
  }
  function startActivity() {
    clearInterval(activityTimer); loadActivity();
    activityTimer = setInterval(() => { if ($('activity-live').checked && !document.hidden) loadActivity(); }, 3000);
  }
  $('activity-live').addEventListener('change', () => { if ($('activity-live').checked) loadActivity(); });

  // ---- boot -----------------------------------------------------------------
  loadStatus();
  if (token) api('/auth/session').then((s) => showApp(s.identity, s.user, s.shares)).catch(() => logout());
})();
