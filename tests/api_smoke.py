#!/usr/bin/env python3
"""API smoke in mock/offline mode. Expected cost: ¥0."""
import json
import os
import urllib.error
import urllib.request

API = os.environ.get("API_URL", "http://backend:8080").rstrip("/")


def req(method, path, data=None, token=None, form=False):
    headers = {}
    body = None
    if token:
        headers["Authorization"] = "Bearer " + token
    if data is not None and not form:
        headers["Content-Type"] = "application/json"
        body = json.dumps(data).encode()
    r = urllib.request.Request(API + path, data=body, headers=headers, method=method)
    with urllib.request.urlopen(r, timeout=30) as resp:
        raw = resp.read()
        if resp.headers.get_content_type() == "application/json" or raw[:1] == b"{":
            return json.loads(raw)
        return {"ok": True, "bytes": len(raw)}


def main():
    h = req("GET", "/api/v1/health")
    assert h["ok"] and h["data"]["status"] == "ok", h
    login = req("POST", "/api/v1/auth/login", {"username": "photographer", "password": "colorpixel"})
    assert login["ok"], login
    tok = login["data"]["token"]
    listing = req("GET", "/api/v1/assets?page=1&page_size=5", token=tok)
    assert listing["ok"], listing
    assert listing["data"]["total"] > 0
    item = listing["data"]["items"][0]
    assert item["extraction_mode"] in ("stream", "deferred", "none")
    detail = req("GET", f"/api/v1/assets/{item['id']}", token=tok)
    assert detail["ok"]
    assert "exif_raw" in detail["data"]
    hist = req("GET", f"/api/v1/assets/{item['id']}/histogram", token=tok)
    assert hist["ok"] and isinstance(hist["data"]["r"], list)
    report = req("GET", "/api/v1/reports/golden-lens", token=tok)
    assert report["ok"]
    assert "lenses" in report["data"]
    print("[PASS] Health Check")
    print("[PASS] Auth")
    print("[PASS] Asset list + EXIF")
    print("[PASS] Histogram")
    print("[PASS] Golden lens report")
    print("COST ¥0")


if __name__ == "__main__":
    main()
