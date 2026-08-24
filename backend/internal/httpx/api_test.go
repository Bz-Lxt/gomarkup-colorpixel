package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteOKAndErr(t *testing.T) {
	rec := httptest.NewRecorder()
	writeOK(rec, map[string]any{"k": 1})
	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil || !env.OK {
		t.Fatalf("%s", rec.Body.String())
	}
	rec = httptest.NewRecorder()
	writeErr(rec, 400, "invalid", "bad")
	if rec.Code != 400 {
		t.Fatal(rec.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	h := cors([]string{"http://localhost:28381"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:28381")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatal(rec.Code)
	}
}

func TestSameOrigin(t *testing.T) {
	if !sameOrigin("http://localhost:28381", "localhost:28381") {
		t.Fatal("same")
	}
	if sameOrigin("http://evil.test", "localhost:28381") {
		t.Fatal("cross")
	}
}

func TestSanitizeName(t *testing.T) {
	if sanitizeName("../etc/passwd") != "passwd" && sanitizeName("../etc/passwd") != "upload.bin" {
		got := sanitizeName("../etc/passwd")
		if bytes.Contains([]byte(got), []byte("..")) {
			t.Fatal(got)
		}
	}
	if sanitizeName("ok.CR3") != "ok.CR3" {
		t.Fatal(sanitizeName("ok.CR3"))
	}
}
