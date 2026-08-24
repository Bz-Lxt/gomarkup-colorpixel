self.onmessage = (ev: MessageEvent<{ id: number; bitmap: ImageBitmap }>) => {
  const { id, bitmap } = ev.data;
  const canvas = new OffscreenCanvas(bitmap.width, bitmap.height);
  const ctx = canvas.getContext("2d");
  if (!ctx) {
    self.postMessage({ id, r: [], g: [], b: [], y: [] });
    return;
  }
  ctx.drawImage(bitmap, 0, 0);
  const img = ctx.getImageData(0, 0, bitmap.width, bitmap.height);
  const r = new Array(256).fill(0);
  const g = new Array(256).fill(0);
  const b = new Array(256).fill(0);
  const y = new Array(256).fill(0);
  const d = img.data;
  const step = d.length > 4_000_000 ? 16 : 4;
  for (let i = 0; i < d.length; i += step) {
    r[d[i]]++;
    g[d[i + 1]]++;
    b[d[i + 2]]++;
    const yy = Math.min(255, (d[i] * 77 + d[i + 1] * 150 + d[i + 2] * 29) >> 8);
    y[yy]++;
  }
  bitmap.close();
  self.postMessage({ id, r, g, b, y });
};
