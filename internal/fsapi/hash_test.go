package fsapi

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"testing"
)

func TestFingerprint(t *testing.T) {
	small := []byte("hello")
	want := md5.Sum(small)
	got, err := Fingerprint(bytes.NewReader(small), int64(len(small)), "md5", true)
	if err != nil || got != hex.EncodeToString(want[:]) {
		t.Fatalf("small file must be hashed whole: %q %v", got, err)
	}
	big := bytes.Repeat([]byte{7}, 3*QuickHashWindow)
	q, _ := Fingerprint(bytes.NewReader(big), int64(len(big)), "md5", true)
	full, _ := Fingerprint(bytes.NewReader(big), int64(len(big)), "md5", false)
	if q == full {
		t.Fatal("quick and full fingerprints of a big file must differ (the quick one mixes in the size)")
	}
	// The quick fingerprint is the md5 of head + tail + decimal size.
	h := md5.New()
	h.Write(big[:QuickHashWindow])
	h.Write(big[len(big)-QuickHashWindow:])
	h.Write([]byte("3145728"))
	if q != hex.EncodeToString(h.Sum(nil)) {
		t.Fatalf("quick fingerprint layout changed: %s", q)
	}
	if _, err := Fingerprint(bytes.NewReader(small), 5, "crc", true); err == nil {
		t.Fatal("unknown algo must fail")
	}
}
