// Package index keeps a persistent SQLite catalogue of everything under the
// data root: one stable id per file or directory, a change journal and the
// bookkeeping the background scanner needs.
//
// It plays the role the "file cache" plays in Nextcloud/OneDrive style
// services: clients address items by id (so identities survive renames,
// reinstalls and moves), and ask "what changed after sequence N" instead of
// walking the tree. The file system stays the source of truth; the index is
// reconciled from it by the scanner, by every write that goes through the API
// and on demand whenever a directory is listed.
package index

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

// Item is a catalogued file or directory. Paths are gateway paths ("/share/dir/file").
type Item struct {
	ID       string
	ParentID string
	Path     string
	Name     string
	IsDir    bool
	Size     int64
	MTime    time.Time
	Seq      int64
}

// Change is one journal entry.
type Change struct {
	Seq     int64
	Kind    string // "upsert" | "delete" | "move"
	ID      string
	Path    string // current path (for delete: the path that vanished)
	OldPath string // move only
	Item    *Item  // upsert/move: the item after the change
}

// RootID is the id of the data root (parent of the shares).
const RootID = "root"

// Index is safe for concurrent use.
type Index struct {
	db *sql.DB
	// bootstrapping suppresses journal rows while an empty index is filled for the
	// first time: nobody can replay that far back anyway (since=0 answers reset),
	// and it keeps the database a fraction of the size.
	bootstrapping bool
}

// SetBootstrapping toggles journal suppression (see Index.bootstrapping).
func (ix *Index) SetBootstrapping(on bool) { ix.bootstrapping = on }

// Open opens or creates the database at dbPath. Use ":memory:" for tests.
func Open(dbPath string) (*Index, error) {
	dsn := dbPath
	if dbPath != ":memory:" {
		dsn = "file:" + dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // sqlite: one writer, keeps things simple and safe
	ix := &Index{db: db}
	if err := ix.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return ix, nil
}

func (ix *Index) Close() error { return ix.db.Close() }

func (ix *Index) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY, parent TEXT NOT NULL, path TEXT NOT NULL UNIQUE, name TEXT NOT NULL,
			is_dir INTEGER NOT NULL, size INTEGER NOT NULL, mtime INTEGER NOT NULL,
			ino INTEGER NOT NULL DEFAULT 0, dev INTEGER NOT NULL DEFAULT 0, seq INTEGER NOT NULL DEFAULT 0)`,
		`CREATE INDEX IF NOT EXISTS items_parent ON items(parent)`,
		`CREATE INDEX IF NOT EXISTS items_ino ON items(dev, ino)`,
		`CREATE TABLE IF NOT EXISTS changes (
			seq INTEGER PRIMARY KEY AUTOINCREMENT, kind TEXT NOT NULL, id TEXT NOT NULL,
			path TEXT NOT NULL, old_path TEXT NOT NULL DEFAULT '', ts INTEGER NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS changes_path ON changes(path)`,
		`CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
	}
	for _, s := range stmts {
		if _, err := ix.db.Exec(s); err != nil {
			return fmt.Errorf("index schema: %w", err)
		}
	}
	return nil
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// clean normalises a gateway path.
func clean(p string) string {
	p = path.Clean("/" + strings.ReplaceAll(p, "\\", "/"))
	return p
}

func parentOf(p string) string {
	if p == "/" {
		return ""
	}
	return path.Dir(p)
}

// ---- reads -----------------------------------------------------------------

func scanItem(row interface{ Scan(...any) error }) (*Item, error) {
	var it Item
	var isDir int
	var mt int64
	if err := row.Scan(&it.ID, &it.ParentID, &it.Path, &it.Name, &isDir, &it.Size, &mt, &it.Seq); err != nil {
		return nil, err
	}
	it.IsDir = isDir == 1
	it.MTime = time.Unix(0, mt).UTC()
	return &it, nil
}

const itemCols = "id, parent, path, name, is_dir, size, mtime, seq"

// ByID returns the item with the given id, or nil.
func (ix *Index) ByID(id string) (*Item, error) {
	if id == RootID {
		return &Item{ID: RootID, Path: "/", Name: "", IsDir: true}, nil
	}
	it, err := scanItem(ix.db.QueryRow(`SELECT `+itemCols+` FROM items WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return it, err
}

// ByPath returns the item at path, or nil.
func (ix *Index) ByPath(p string) (*Item, error) {
	p = clean(p)
	if p == "/" {
		return ix.ByID(RootID)
	}
	it, err := scanItem(ix.db.QueryRow(`SELECT `+itemCols+` FROM items WHERE path = ?`, p))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return it, err
}

// Children lists the direct children of a directory path as known to the index.
func (ix *Index) Children(dir string) ([]*Item, error) {
	parentID := RootID
	if d := clean(dir); d != "/" {
		p, err := ix.ByPath(d)
		if err != nil || p == nil {
			return nil, err
		}
		parentID = p.ID
	}
	rows, err := ix.db.Query(`SELECT `+itemCols+` FROM items WHERE parent = ? ORDER BY name`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// LatestSeq is the sequence number of the newest journal entry (0 when empty).
func (ix *Index) LatestSeq() (int64, error) {
	var n sql.NullInt64
	if err := ix.db.QueryRow(`SELECT MAX(seq) FROM changes`).Scan(&n); err != nil {
		return 0, err
	}
	return n.Int64, nil
}

// journalStart is the oldest sequence still in the journal (0 when empty).
func (ix *Index) journalStart() (int64, error) {
	var n sql.NullInt64
	if err := ix.db.QueryRow(`SELECT MIN(seq) FROM changes`).Scan(&n); err != nil {
		return 0, err
	}
	return n.Int64, nil
}

// Changes returns up to limit journal entries after `since`, in order.
// reset is true when the journal no longer covers `since` (the client must
// re-enumerate from scratch and continue from latest).
func (ix *Index) Changes(since int64, limit int) (changes []Change, latest int64, truncated, reset bool, err error) {
	latest, err = ix.LatestSeq()
	if err != nil {
		return
	}
	start, err := ix.journalStart()
	if err != nil {
		return
	}
	// since=0 means "never synced": the client enumerates and starts from latest.
	// A since older than the retained journal cannot be replayed either.
	if since == 0 || (start > 0 && since < start-1) {
		return nil, latest, false, true, nil
	}
	rows, err := ix.db.Query(`SELECT seq, kind, id, path, old_path FROM changes WHERE seq > ? ORDER BY seq LIMIT ?`, since, limit+1)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c Change
		if err = rows.Scan(&c.Seq, &c.Kind, &c.ID, &c.Path, &c.OldPath); err != nil {
			return
		}
		changes = append(changes, c)
	}
	if err = rows.Err(); err != nil {
		return
	}
	if len(changes) > limit {
		changes = changes[:limit]
		truncated = true
		latest = changes[len(changes)-1].Seq
	}
	// Attach the current item for upserts/moves (nil if it vanished since).
	for i := range changes {
		if changes[i].Kind != "delete" {
			changes[i].Item, _ = ix.ByID(changes[i].ID)
		}
	}
	return
}

// ---- writes ----------------------------------------------------------------

// Observed is what the file system reports about one directory entry.
type Observed struct {
	Name  string
	IsDir bool
	Size  int64
	MTime time.Time
	Ino   uint64
	Dev   uint64
}

// FromFileInfo builds an Observed from a stat result.
func FromFileInfo(fi fs.FileInfo) Observed {
	o := Observed{Name: fi.Name(), IsDir: fi.IsDir(), Size: fi.Size(), MTime: fi.ModTime()}
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		o.Ino = uint64(st.Ino)
		o.Dev = uint64(st.Dev)
	}
	return o
}

// Reconcile makes the index agree with the observed contents of dir: new
// entries are inserted (a vanished entry with the same inode is recognised as
// a rename and keeps its id), changed entries are updated, missing entries are
// deleted together with their subtree. It returns the id of every current child
// by name. dir must already be indexed (or be "/").
func (ix *Index) Reconcile(dir string, observed []Observed) (map[string]string, error) {
	dir = clean(dir)
	parentID := RootID
	if dir != "/" {
		p, err := ix.ByPath(dir)
		if err != nil {
			return nil, err
		}
		if p == nil {
			// Index the ancestors first so the parent exists.
			if err := ix.ensurePath(dir); err != nil {
				return nil, err
			}
			p, err = ix.ByPath(dir)
			if err != nil || p == nil {
				return nil, fmt.Errorf("index: parent %s not indexed", dir)
			}
		}
		parentID = p.ID
	}
	tx, err := ix.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	now := time.Now().UnixNano()

	// Current children by name.
	type row struct {
		id    string
		isDir bool
		size  int64
		mtime int64
		ino   uint64
		dev   uint64
	}
	known := map[string]row{}
	rows, err := tx.Query(`SELECT id, name, is_dir, size, mtime, ino, dev FROM items WHERE parent = ?`, parentID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r row
		var name string
		var isDir int
		if err := rows.Scan(&r.id, &name, &isDir, &r.size, &r.mtime, &r.ino, &r.dev); err != nil {
			rows.Close()
			return nil, err
		}
		r.isDir = isDir == 1
		known[name] = r
	}
	rows.Close()

	ids := make(map[string]string, len(observed))
	seen := map[string]bool{}
	var missing []string
	for name := range known {
		missing = append(missing, name)
	}
	for _, o := range observed {
		seen[o.Name] = true
		p := path.Join(dir, o.Name)
		if k, ok := known[o.Name]; ok {
			ids[o.Name] = k.id
			if k.isDir != o.IsDir || k.size != o.Size || k.mtime != o.MTime.UnixNano() {
				if _, err := tx.Exec(`UPDATE items SET is_dir=?, size=?, mtime=?, ino=?, dev=?, seq=(SELECT COALESCE(MAX(seq),0)+1 FROM changes) WHERE id=?`,
					b2i(o.IsDir), o.Size, o.MTime.UnixNano(), o.Ino, o.Dev, k.id); err != nil {
					return nil, err
				}
				if err := ix.journal(tx, "upsert", k.id, p, "", now); err != nil {
					return nil, err
				}
			}
			continue
		}
		// New name in this directory: a rename when an entry that vanished from the
		// same directory had the same inode (the common "rename in place" case);
		// otherwise a creation. Moves across directories made through the API are
		// recorded by Move(); ones made behind the gateway's back become
		// delete + create.
		var id string
		if o.Ino != 0 {
			for _, name := range missing {
				k := known[name]
				if !seen[name] && k.ino == o.Ino && k.dev == o.Dev && k.isDir == o.IsDir {
					seen[name] = true // consumed: not a deletion
					if err := movePath(tx, k.id, path.Join(dir, name), p, parentID, now); err != nil {
						return nil, err
					}
					if _, err := tx.Exec(`UPDATE items SET size=?, mtime=? WHERE id=?`, o.Size, o.MTime.UnixNano(), k.id); err != nil {
						return nil, err
					}
					id = k.id
					break
				}
			}
			if id != "" {
				ids[o.Name] = id
				continue
			}
		}
		id = newID()
		if _, err := tx.Exec(`INSERT INTO items(id,parent,path,name,is_dir,size,mtime,ino,dev,seq) VALUES(?,?,?,?,?,?,?,?,?,(SELECT COALESCE(MAX(seq),0)+1 FROM changes))`,
			id, parentID, p, o.Name, b2i(o.IsDir), o.Size, o.MTime.UnixNano(), o.Ino, o.Dev); err != nil {
			return nil, err
		}
		if err := ix.journal(tx, "upsert", id, p, "", now); err != nil {
			return nil, err
		}
		ids[o.Name] = id
	}
	for _, name := range missing {
		if seen[name] {
			continue
		}
		if err := deleteSubtree(tx, known[name].id, path.Join(dir, name), now); err != nil {
			return nil, err
		}
	}
	return ids, tx.Commit()
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ensurePath indexes dir and its ancestors from the file system-independent
// information we have (names only); sizes/mtimes get corrected by the next
// reconcile of each level.
func (ix *Index) ensurePath(dir string) error {
	dir = clean(dir)
	if dir == "/" {
		return nil
	}
	parent := parentOf(dir)
	if err := ix.ensurePath(parent); err != nil {
		return err
	}
	if it, err := ix.ByPath(dir); err != nil || it != nil {
		return err
	}
	parentID := RootID
	if parent != "/" {
		p, err := ix.ByPath(parent)
		if err != nil {
			return err
		}
		parentID = p.ID
	}
	_, err := ix.db.Exec(`INSERT INTO items(id,parent,path,name,is_dir,size,mtime,seq) VALUES(?,?,?,?,1,0,0,(SELECT COALESCE(MAX(seq),0)+1 FROM changes))`,
		newID(), parentID, dir, path.Base(dir))
	return err
}

// journal records one change unless the index is bootstrapping.
func (ix *Index) journal(tx interface {
	Exec(string, ...any) (sql.Result, error)
}, kind, id, p, oldPath string, now int64) error {
	if ix.bootstrapping {
		return nil
	}
	_, err := tx.Exec(`INSERT INTO changes(kind,id,path,old_path,ts) VALUES(?,?,?,?,?)`, kind, id, p, oldPath, now)
	return err
}

func movePath(tx *sql.Tx, id, oldPath, newPath, newParent string, now int64) error {
	if _, err := tx.Exec(`UPDATE items SET path=?, name=?, parent=?, seq=(SELECT COALESCE(MAX(seq),0)+1 FROM changes) WHERE id=?`, newPath, path.Base(newPath), newParent, id); err != nil {
		return err
	}
	// Re-path the subtree.
	if _, err := tx.Exec(`UPDATE items SET path = ? || substr(path, ?) WHERE path LIKE ? ESCAPE '\'`, newPath, len(oldPath)+1, likePrefix(oldPath)+"/%"); err != nil {
		return err
	}
	_, err := tx.Exec(`INSERT INTO changes(kind,id,path,old_path,ts) VALUES('move',?,?,?,?)`, id, newPath, oldPath, now)
	return err
}

func likePrefix(p string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(p)
}

func deleteSubtree(tx *sql.Tx, id, p string, now int64) error {
	rows, err := tx.Query(`SELECT id, path FROM items WHERE path LIKE ? ESCAPE '\' ORDER BY length(path) DESC`, likePrefix(p)+"/%")
	if err != nil {
		return err
	}
	type kv struct{ id, path string }
	var subs []kv
	for rows.Next() {
		var k kv
		if err := rows.Scan(&k.id, &k.path); err != nil {
			rows.Close()
			return err
		}
		subs = append(subs, k)
	}
	rows.Close()
	for _, s := range subs {
		if _, err := tx.Exec(`DELETE FROM items WHERE id=?`, s.id); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO changes(kind,id,path,ts) VALUES('delete',?,?,?)`, s.id, s.path, now); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM items WHERE id=?`, id); err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO changes(kind,id,path,ts) VALUES('delete',?,?,?)`, id, p, now)
	return err
}

// Upsert records one path as observed (write-through after PUT/mkdir/copy).
// The parent must exist in the index (it is created on demand).
func (ix *Index) Upsert(p string, o Observed) (string, error) {
	p = clean(p)
	parent := parentOf(p)
	if err := ix.ensurePath(parent); err != nil {
		return "", err
	}
	existing, err := ix.ByPath(p)
	if err != nil {
		return "", err
	}
	now := time.Now().UnixNano()
	if existing != nil {
		_, err = ix.db.Exec(`UPDATE items SET is_dir=?, size=?, mtime=?, ino=?, dev=?, seq=(SELECT COALESCE(MAX(seq),0)+1 FROM changes) WHERE id=?`,
			b2i(o.IsDir), o.Size, o.MTime.UnixNano(), o.Ino, o.Dev, existing.ID)
		if err != nil {
			return "", err
		}
		_, err = ix.db.Exec(`INSERT INTO changes(kind,id,path,ts) VALUES('upsert',?,?,?)`, existing.ID, p, now)
		return existing.ID, err
	}
	parentID := RootID
	if parent != "/" {
		pp, err := ix.ByPath(parent)
		if err != nil || pp == nil {
			return "", fmt.Errorf("index: parent of %s missing", p)
		}
		parentID = pp.ID
	}
	id := newID()
	if _, err := ix.db.Exec(`INSERT INTO items(id,parent,path,name,is_dir,size,mtime,ino,dev,seq) VALUES(?,?,?,?,?,?,?,?,?,(SELECT COALESCE(MAX(seq),0)+1 FROM changes))`,
		id, parentID, p, path.Base(p), b2i(o.IsDir), o.Size, o.MTime.UnixNano(), o.Ino, o.Dev); err != nil {
		return "", err
	}
	_, err = ix.db.Exec(`INSERT INTO changes(kind,id,path,ts) VALUES('upsert',?,?,?)`, id, p, now)
	return id, err
}

// Delete records that p (and everything below it) is gone.
func (ix *Index) Delete(p string) error {
	p = clean(p)
	it, err := ix.ByPath(p)
	if err != nil || it == nil {
		return err
	}
	tx, err := ix.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := deleteSubtree(tx, it.ID, p, time.Now().UnixNano()); err != nil {
		return err
	}
	return tx.Commit()
}

// Move records a rename/move keeping the id (and the ids of the subtree).
func (ix *Index) Move(from, to string) error {
	from, to = clean(from), clean(to)
	it, err := ix.ByPath(from)
	if err != nil {
		return err
	}
	if it == nil {
		return nil // never indexed: the next reconcile of the target dir will add it
	}
	if err := ix.ensurePath(parentOf(to)); err != nil {
		return err
	}
	newParent := RootID
	if parentOf(to) != "/" {
		pp, err := ix.ByPath(parentOf(to))
		if err != nil || pp == nil {
			return fmt.Errorf("index: parent of %s missing", to)
		}
		newParent = pp.ID
	}
	tx, err := ix.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// An item already at the destination (overwrite) is gone. (Queried through the
	// transaction: the pool has a single connection.)
	var oldID string
	if err := tx.QueryRow(`SELECT id FROM items WHERE path = ?`, to).Scan(&oldID); err == nil && oldID != it.ID {
		if err := deleteSubtree(tx, oldID, to, time.Now().UnixNano()); err != nil {
			return err
		}
	}
	if err := movePath(tx, it.ID, from, to, newParent, time.Now().UnixNano()); err != nil {
		return err
	}
	return tx.Commit()
}

// MarkBootstrapped writes a single journal row once the initial catalogue is complete,
// so sequence numbers start from 1 and clients can anchor to it. The row refers to the
// root and is skipped by the API.
func (ix *Index) MarkBootstrapped() error {
	_, err := ix.db.Exec(`INSERT INTO changes(kind,id,path,ts) VALUES('upsert',?,?,?)`, RootID, "/", time.Now().UnixNano())
	return err
}

// Prune drops journal entries older than keep.
func (ix *Index) Prune(keep time.Duration) error {
	_, err := ix.db.Exec(`DELETE FROM changes WHERE ts < ?`, time.Now().Add(-keep).UnixNano())
	return err
}

// Meta get/set.
func (ix *Index) GetMeta(key string) (string, error) {
	var v string
	err := ix.db.QueryRow(`SELECT value FROM meta WHERE key=?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}
func (ix *Index) SetMeta(key, value string) error {
	_, err := ix.db.Exec(`INSERT INTO meta(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// Count returns the number of indexed items.
func (ix *Index) Count() (int64, error) {
	var n int64
	err := ix.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&n)
	return n, err
}

// Ping checks the database is usable.
func (ix *Index) Ping(ctx context.Context) error { return ix.db.PingContext(ctx) }
