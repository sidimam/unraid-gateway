package fsapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestAPI(t *testing.T) (*API, *http.ServeMux, string) {
	t.Helper()
	dir := t.TempDir()
	api, err := New(dir, Options{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	api.Register(mux, "/fs")
	return api, mux, api.root.Path()
}

func do(mux *http.ServeMux, method, target string, body io.Reader, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, body)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestResolveRejectsTraversal(t *testing.T) {
	api, _, root := newTestAPI(t)
	// ".." segments are clamped at the root, never escape it.
	for _, p := range []string{"../x", "/a/../../x", "a/b/../../../x", "..\\..\\x"} {
		got, err := api.root.Resolve(p)
		if err != nil || got != filepath.Join(root, "x") {
			t.Errorf("%q: got %q err %v", p, got, err)
		}
	}
	got, err := api.root.Resolve("/a/./b//c")
	if err != nil || got != filepath.Join(root, "a", "b", "c") {
		t.Errorf("clean path: got %q err %v", got, err)
	}
	if got, _ := api.root.Resolve(""); got != root {
		t.Errorf("empty path should be root, got %q", got)
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	api, _, root := newTestAPI(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skip("symlinks not supported")
	}
	if _, err := api.root.Resolve("/escape/file.txt"); err == nil {
		t.Fatal("symlink escape not detected")
	}
	if err := os.Mkdir(filepath.Join(root, "inside"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "inside"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := api.root.Resolve("/link/new.txt"); err != nil {
		t.Fatalf("internal symlink should be allowed: %v", err)
	}
}

func TestPutListDownloadDelete(t *testing.T) {
	_, mux, _ := newTestAPI(t)
	rr := do(mux, "PUT", "/fs/content?path=/docs/hello.txt", strings.NewReader("hello world"), nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("put: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "GET", "/fs/list?path=/docs", nil, nil)
	var lr listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &lr); err != nil || len(lr.Entries) != 1 || lr.Entries[0].Name != "hello.txt" {
		t.Fatalf("list: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "GET", "/fs/content?path=/docs/hello.txt", nil, map[string]string{"Range": "bytes=6-10"})
	if rr.Code != http.StatusPartialContent || rr.Body.String() != "world" {
		t.Fatalf("range: %d %q", rr.Code, rr.Body.String())
	}
	rr = do(mux, "PUT", "/fs/content?path=/docs/hello.txt", strings.NewReader("x"), map[string]string{"If-Match": "\"bogus\""})
	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("if-match: %d", rr.Code)
	}
	rr = do(mux, "POST", "/fs/delete", strings.NewReader(`{"path":"/docs"}`), nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("delete non-empty dir should conflict: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "POST", "/fs/delete", strings.NewReader(`{"path":"/docs","recursive":true}`), nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete recursive: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "GET", "/fs/stat?path=/docs", nil, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("stat after delete: %d", rr.Code)
	}
}

func TestResumableUpload(t *testing.T) {
	_, mux, root := newTestAPI(t)
	rr := do(mux, "POST", "/fs/uploads", strings.NewReader(`{"path":"/big.bin","size":10}`), nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body)
	}
	var u upload
	_ = json.Unmarshal(rr.Body.Bytes(), &u)
	rr = do(mux, "PATCH", "/fs/uploads/"+u.ID, strings.NewReader("01234"), map[string]string{"Upload-Offset": "0"})
	if rr.Code != http.StatusNoContent || rr.Header().Get("Upload-Offset") != "5" {
		t.Fatalf("chunk1: %d %s", rr.Code, rr.Header().Get("Upload-Offset"))
	}
	rr = do(mux, "PATCH", "/fs/uploads/"+u.ID, strings.NewReader("zz"), map[string]string{"Upload-Offset": "3"})
	if rr.Code != http.StatusConflict {
		t.Fatalf("wrong offset should conflict: %d", rr.Code)
	}
	rr = do(mux, "POST", "/fs/uploads/"+u.ID+"/commit", nil, nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("incomplete commit should fail: %d", rr.Code)
	}
	rr = do(mux, "PATCH", "/fs/uploads/"+u.ID, strings.NewReader("56789"), map[string]string{"Upload-Offset": "5"})
	if rr.Code != http.StatusNoContent {
		t.Fatalf("chunk2: %d", rr.Code)
	}
	rr = do(mux, "POST", "/fs/uploads/"+u.ID+"/commit", nil, nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("commit: %d %s", rr.Code, rr.Body)
	}
	data, err := os.ReadFile(filepath.Join(root, "big.bin"))
	if err != nil || string(data) != "0123456789" {
		t.Fatalf("content: %q %v", data, err)
	}
	ents, _ := os.ReadDir(root)
	for _, e := range ents {
		if strings.Contains(e.Name(), partSuffix) {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestMoveCopyAndChanges(t *testing.T) {
	_, mux, _ := newTestAPI(t)
	do(mux, "PUT", "/fs/content?path=/a/one.txt", strings.NewReader("1"), nil)
	rr := do(mux, "POST", "/fs/copy", strings.NewReader(`{"from":"/a","to":"/b"}`), nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("copy: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "POST", "/fs/move", strings.NewReader(`{"from":"/b/one.txt","to":"/b/two.txt"}`), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("move: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "POST", "/fs/move", strings.NewReader(`{"from":"/a","to":"/a/sub"}`), nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("move into itself: %d", rr.Code)
	}
	rr = do(mux, "GET", "/fs/changes?path=/&since=0", nil, nil)
	var cr changesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &cr); err != nil {
		t.Fatal(err)
	}
	if len(cr.Files) != 2 || len(cr.Dirs) < 3 {
		t.Fatalf("changes: %s", rr.Body)
	}
	rr = do(mux, "GET", "/fs/changes?path=/&since="+itoa(cr.Cursor+int64(3e9)), nil, nil)
	_ = json.Unmarshal(rr.Body.Bytes(), &cr)
	if len(cr.Files) != 0 || len(cr.Dirs) != 0 {
		t.Fatalf("future cursor should report nothing: %s", rr.Body)
	}
}

func TestChangesResume(t *testing.T) {
	_, mux, _ := newTestAPI(t)
	// "a-b" sorts before "a/1.txt" lexically but after the whole "a" subtree in walk order.
	for _, f := range []string{"/a/1.txt", "/a/2.txt", "/a-b/x.txt", "/b/3.txt", "/b/c/4.txt", "/d.txt"} {
		do(mux, "PUT", "/fs/content?path="+f, strings.NewReader("x"), nil)
	}
	seen := map[string]bool{}
	after, cursor, pages := "", "", 0
	for {
		url := "/fs/changes?path=/&since=0&limit=3"
		if after != "" {
			url += "&after=" + after + "&cursor=" + cursor
		}
		rr := do(mux, "GET", url, nil, nil)
		var cr changesResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &cr); err != nil {
			t.Fatalf("page %d: %s", pages, rr.Body)
		}
		pages++
		for _, d := range cr.Dirs {
			if seen[d] {
				t.Fatalf("dir %s reported twice", d)
			}
			seen[d] = true
		}
		for _, f := range cr.Files {
			if seen[f.Path] {
				t.Fatalf("file %s reported twice", f.Path)
			}
			seen[f.Path] = true
		}
		if !cr.Truncated {
			break
		}
		if cr.Next == "" {
			t.Fatal("truncated page without next")
		}
		after, cursor = cr.Next, itoa(cr.Cursor)
		if pages > 20 {
			t.Fatal("pagination did not terminate")
		}
	}
	for _, want := range []string{"/", "/a", "/a/1.txt", "/a/2.txt", "/a-b", "/a-b/x.txt", "/b", "/b/3.txt", "/b/c", "/b/c/4.txt", "/d.txt"} {
		if !seen[want] {
			t.Errorf("%s never reported (pages=%d, seen=%v)", want, pages, seen)
		}
	}
	if pages < 2 {
		t.Fatalf("expected multiple pages, got %d", pages)
	}
}

func TestWalkCompare(t *testing.T) {
	order := []string{"/", "/a", "/a/1.txt", "/a/z", "/a/z/deep", "/a-b", "/a-b/x", "/b"}
	for i := range order {
		for j := range order {
			got := walkCompare(order[i], order[j])
			switch {
			case i < j && got >= 0, i > j && got <= 0, i == j && got != 0:
				t.Errorf("walkCompare(%q,%q)=%d", order[i], order[j], got)
			}
		}
	}
	if !isAncestor("/", "/a") || !isAncestor("/a", "/a/b/c") || isAncestor("/a", "/a-b") || isAncestor("/a", "/a") {
		t.Fatal("isAncestor")
	}
}

func TestShareRootsAreImmutable(t *testing.T) {
	_, mux, _ := newTestAPI(t)
	do(mux, "PUT", "/fs/content?path=/media/a.txt", strings.NewReader("x"), nil)
	rr := do(mux, "POST", "/fs/delete", strings.NewReader(`{"path":"/media","recursive":true}`), nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("delete share root: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "POST", "/fs/move", strings.NewReader(`{"from":"/media","to":"/media2"}`), nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("rename share root: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "POST", "/fs/move", strings.NewReader(`{"from":"/media/a.txt","to":"/newshare","overwrite":true}`), nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("move onto root level: %d %s", rr.Code, rr.Body)
	}
	rr = do(mux, "POST", "/fs/mkdir", strings.NewReader(`{"path":"/newshare"}`), nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("mkdir at root: %d %s", rr.Code, rr.Body)
	}
	if rr := do(mux, "GET", "/fs/stat?path=/media/a.txt", nil, nil); rr.Code != http.StatusOK {
		t.Fatalf("file should still exist: %d", rr.Code)
	}
	// Inside a share everything still works.
	if rr := do(mux, "POST", "/fs/mkdir", strings.NewReader(`{"path":"/media/sub"}`), nil); rr.Code != http.StatusCreated {
		t.Fatalf("mkdir inside share: %d %s", rr.Code, rr.Body)
	}
}

func TestReadOnly(t *testing.T) {
	dir := t.TempDir()
	api, _ := New(dir, Options{ReadOnly: true}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	api.Register(mux, "/fs")
	rr := do(mux, "PUT", "/fs/content?path=/x", bytes.NewReader([]byte("x")), nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("read-only put: %d", rr.Code)
	}
}

func itoa(n int64) string {
	var b []byte
	b = appendInt(b, n)
	return string(b)
}

func appendInt(b []byte, n int64) []byte {
	if n == 0 {
		return append(b, '0')
	}
	var tmp [20]byte
	i := len(tmp)
	for n > 0 {
		i--
		tmp[i] = byte('0' + n%10)
		n /= 10
	}
	return append(b, tmp[i:]...)
}
