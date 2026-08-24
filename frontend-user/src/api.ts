const TOKEN_KEY = "cp_token";

export function token(): string {
  return localStorage.getItem(TOKEN_KEY) || "";
}

export function setToken(t: string) {
  localStorage.setItem(TOKEN_KEY, t);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

type Envelope<T> = { ok: boolean; data?: T; error?: { code: string; message: string } };

export class ApiError extends Error {
  status: number;
  code: string;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function req<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (token()) headers.set("Authorization", `Bearer ${token()}`);
  if (init.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const res = await fetch(path, { ...init, headers });
  const text = await res.text();
  let body: Envelope<T> | null = null;
  try {
    body = text ? JSON.parse(text) : null;
  } catch {
    throw new ApiError(res.status, "parse", "响应不是 JSON");
  }
  if (!res.ok || !body?.ok) {
    throw new ApiError(res.status, body?.error?.code || "error", body?.error?.message || "请求失败");
  }
  return body.data as T;
}

export const api = {
  login: (username: string, password: string) =>
    req<{ token: string; username: string }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),
  health: () => req<{ status: string }>("/api/v1/health"),
  list: (q: URLSearchParams) =>
    req<{ items: Asset[]; total: number; page: number; page_size: number }>(`/api/v1/assets?${q}`),
  get: (id: number) => req<Asset>(`/api/v1/assets/${id}`),
  patch: (id: number, body: { rating?: number; tags?: string[] }) =>
    req<Asset>(`/api/v1/assets/${id}`, { method: "PATCH", body: JSON.stringify(body) }),
  del: (id: number) => req<{ deleted: number }>(`/api/v1/assets/${id}`, { method: "DELETE" }),
  upload: (files: File[]) => {
    const fd = new FormData();
    files.forEach((f) => fd.append("files", f));
    return req<{ items: Asset[] }>("/api/v1/assets/upload", { method: "POST", body: fd });
  },
  histogram: (id: number) =>
    req<{ r: number[]; g: number[]; b: number[]; y: number[]; clip_shadow: number; clip_highlight: number }>(
      `/api/v1/assets/${id}/histogram`,
    ),
  report: () => req<Report>("/api/v1/reports/golden-lens"),
};

export type Asset = {
  id: number;
  filename: string;
  format: string;
  size_bytes: number;
  extraction_mode: string;
  camera_make: string;
  camera_model: string;
  lens_model: string;
  lens_spec: string;
  aperture: number;
  shutter_text: string;
  iso: number;
  focal_length: number;
  focal_length_35mm: number;
  datetime_original: string;
  rating: number;
  tags: string[];
  tile_status: string;
  tile_max_z: number;
  width: number;
  height: number;
  fidelity_label: string;
  thumb_url: string;
  preview_url: string;
  sharpness?: number | null;
  noise?: number | null;
  clip_shadow?: number | null;
  clip_highlight?: number | null;
  white_balance?: string;
  exposure_bias?: number;
  exif_raw?: Record<string, unknown>;
};

export type Report = {
  window_from: string;
  window_to: string;
  total: number;
  golden_lens: string;
  recommended_combo: string;
  lenses: Array<{
    lens: string;
    count: number;
    insufficient_data: boolean;
    score: number;
    factors: Record<string, { value: number; confidence: number; samples: number; excluded_count: number }>;
    derivation: string[];
    focal_hist: Record<string, number>;
    aperture_hist: Record<string, number>;
    iso_hist: Record<string, number>;
    month_hist: Record<string, number>;
  }>;
  focal_global: Record<string, number>;
  aperture_global: Record<string, number>;
};
