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
