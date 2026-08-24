<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from "vue";
import { api, type Asset } from "../api";
import { useSession } from "../stores/session";

const session = useSession();
const list = ref<Asset[]>([]);
const current = ref<Asset | null>(null);
const canvas = ref<HTMLCanvasElement | null>(null);
const clip = ref({ shadow: 0, highlight: 0 });
const worker = new Worker(new URL("../workers/hist.worker.ts", import.meta.url), { type: "module" });

worker.onmessage = (ev: MessageEvent<{ r: number[]; g: number[]; b: number[] }>) => {
  draw(ev.data.r, ev.data.g, ev.data.b);
};

async function load() {
  const r = await api.list(new URLSearchParams({ page: "1", page_size: "24" }));
  list.value = r.items || [];
  if (list.value[0]) select(list.value[0]);
}

async function select(a: Asset) {
  current.value = a;
  try {
    const h = await api.histogram(a.id);
    clip.value = { shadow: h.clip_shadow, highlight: h.clip_highlight };
    draw(h.r, h.g, h.b);
    const img = new Image();
    img.crossOrigin = "anonymous";
    img.src = a.preview_url;
    img.onload = async () => {
      const bmp = await createImageBitmap(img);
      worker.postMessage({ id: a.id, bitmap: bmp }, [bmp]);
    };
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "直方图失败", "error");
  }
}

function draw(r: number[], g: number[], b: number[]) {
  const el = canvas.value;
  if (!el) return;
  const ctx = el.getContext("2d");
  if (!ctx) return;
  const w = el.width;
  const h = el.height;
  ctx.fillStyle = "#0B0C0E";
  ctx.fillRect(0, 0, w, h);
  const max = Math.max(1, ...r, ...g, ...b);
  const paint = (arr: number[], color: string) => {
    ctx.beginPath();
    ctx.moveTo(0, h);
    arr.forEach((v, i) => {
      const x = (i / 255) * w;
      const y = h - (v / max) * (h - 8);
      ctx.lineTo(x, y);
    });
    ctx.lineTo(w, h);
    ctx.closePath();
    ctx.fillStyle = color;
    ctx.fill();
  };
  paint(r, "rgba(212,83,59,0.35)");
  paint(g, "rgba(111,158,122,0.35)");
  paint(b, "rgba(80,140,200,0.35)");
}

onMounted(async () => {
  await load();
  await nextTick();
});
watch(canvas, () => {
  if (current.value) select(current.value);
});
</script>

<template>
  <div class="w-full px-4 md:px-6 py-5 grid lg:grid-cols-[280px_1fr] gap-4">
    <aside class="border border-line bg-panel p-3 space-y-2 max-h-[70vh] overflow-auto">
      <p class="label">当前选中跟随刷新</p>
      <button
        v-for="a in list"
        :key="a.id"
        type="button"
        class="block w-full text-left font-mono text-[11px] px-2 py-1 border border-transparent hover:border-line"
        :class="current?.id === a.id ? 'text-tungsten' : 'text-mute'"
        @click="select(a)"
      >
        {{ a.filename }}
      </button>
    </aside>
    <section class="space-y-3">
      <div class="flex items-center gap-3">
        <h1 class="font-display text-3xl">RGB 通道大屏</h1>
        <span v-if="clip.highlight > 0.05" class="font-mono text-[10px] text-alert">高光裁切 {{ (clip.highlight * 100).toFixed(1) }}%</span>
        <span v-if="clip.shadow > 0.05" class="font-mono text-[10px] text-alert">暗部裁切 {{ (clip.shadow * 100).toFixed(1) }}%</span>
      </div>
      <canvas ref="canvas" width="1200" height="420" class="w-full border border-line bg-ink"></canvas>
      <p class="text-mute text-sm">前端 Worker 流式重算叠加服务端直方图。主线程仅负责描边。</p>
    </section>
  </div>
</template>
