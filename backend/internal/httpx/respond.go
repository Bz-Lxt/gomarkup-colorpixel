package httpx

import (
	"encoding/json"
	"net/http"

	"colorpixel/internal/logger"
)

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error *apiError       `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, `{"ok":false,"error":{"code":"encode","message":"encode failed"}}`, http.StatusInternalServerError)
		return
	}
	raw, _ := json.Marshal(envelope{OK: status < 400, Data: body})
	if status >= 400 {
		raw, _ = json.Marshal(envelope{OK: false, Error: &apiError{Code: http.StatusText(status), Message: "error"}})
	}
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, err := json.Marshal(data)
	if err != nil {
		logger.L().Error("json encode", "err", err)
		writeErr(w, http.StatusInternalServerError, "encode", "encode failed")
		return
	}
	raw, _ := json.Marshal(envelope{OK: true, Data: body})
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	raw, _ := json.Marshal(envelope{OK: false, Error: &apiError{Code: code, Message: msg}})
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}
