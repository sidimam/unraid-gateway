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
  }
  function logout() {
    if (token) fetch(API + '/auth/logout', { method: 'POST', headers: { Authorization: 'Bearer ' + token } }).catch(() => {});
    token = ''; sessionStorage.removeItem('ugw.token');
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

  // ---- boot -----------------------------------------------------------------
  loadStatus();
  if (token) api('/auth/session').then((s) => showApp(s.identity, s.user, s.shares)).catch(() => logout());
})();
