<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { api, type Asset } from "../api";
import { useSession } from "../stores/session";

const route = useRoute();
const session = useSession();
const assets = ref<Asset[]>([]);
const layout = ref<2 | 4>(2);
const transform = reactive({ x: 0, y: 0, scale: 1 });
const locked = ref(true);
const oneToOne = ref(false);
const pick = ref("");

const panes = computed(() => assets.value.slice(0, layout.value));

async function loadFromQuery() {
  const ids = String(route.query.ids || "")
    .split(",")
    .map((s) => Number(s))
    .filter((n) => n > 0);
  if (!ids.length) {
    const r = await api.list(new URLSearchParams({ page: "1", page_size: "8" }));
    assets.value = (r.items || []).slice(0, 4);
    return;
  }
  const got: Asset[] = [];
  for (const id of ids.slice(0, 4)) {
    try {
      got.push(await api.get(id));
    } catch {
      /* skip */
    }
  }
  assets.value = got;
}

onMounted(loadFromQuery);
watch(() => route.query.ids, loadFromQuery);

function onWheel(e: WheelEvent) {
  e.preventDefault();
  const next = Math.min(8, Math.max(0.2, transform.scale * (e.deltaY < 0 ? 1.08 : 0.92)));
  transform.scale = next;
  oneToOne.value = Math.abs(next - 1) < 0.02;
}

let drag: { x: number; y: number } | null = null;
function down(e: MouseEvent) {
  drag = { x: e.clientX - transform.x, y: e.clientY - transform.y };
}
function move(e: MouseEvent) {
  if (!drag || !locked.value) return;
  transform.x = e.clientX - drag.x;
  transform.y = e.clientY - drag.y;
}
function up() {
  drag = null;
}

function reset() {
  transform.x = 0;
  transform.y = 0;
  transform.scale = 1;
  oneToOne.value = true;
}

function setOne() {
  transform.scale = 1;
  oneToOne.value = true;
}

async function addPick() {
  const id = Number(pick.value);
  if (!id) return;
  if (assets.value.length >= 4) {
    session.flash("最多四路", "error");
    return;
  }
  try {
    const a = await api.get(id);
    assets.value.push(a);
    pick.value = "";
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "加入失败", "error");
  }
}

const diffs = computed(() => {
  if (panes.value.length < 2) return [];
  const keys = ["aperture", "shutter_text", "iso", "focal_length"] as const;
  return keys.filter((k) => new Set(panes.value.map((a) => String(a[k]))).size > 1);
});
</script>

<template>
  <div class="w-full h-[calc(100vh-64px)] flex flex-col">
    <div class="px-4 py-2 border-b border-line flex flex-wrap items-center gap-2">
      <span class="font-mono text-[10px] text-tungsten tracking-widest">预览级 (Embedded JPEG)</span>
      <button class="btn-ghost" type="button" @click="layout = 2">双镜</button>
      <button class="btn-ghost" type="button" @click="layout = 4">四镜</button>
      <button class="btn-ghost" type="button" @click="locked = !locked">{{ locked ? "同步开" : "同步关" }}</button>
      <button class="btn-ghost" type="button" @click="setOne">1:1</button>
      <button class="btn-ghost" type="button" @click="reset">复位</button>
      <span class="font-mono text-xs text-mute">{{ transform.scale.toFixed(2) }}× · {{ Math.round(transform.x) }},{{ Math.round(transform.y) }}</span>
      <input v-model="pick" class="field w-28" placeholder="资产 ID" />
      <button class="btn-ghost" type="button" @click="addPick">加入</button>
    </div>
    <div
      class="flex-1 grid"
      :class="layout === 2 ? 'grid-cols-1 md:grid-cols-2' : 'grid-cols-1 md:grid-cols-2'"
      @wheel="onWheel"
      @mousedown="down"
      @mousemove="move"
      @mouseup="up"
      @mouseleave="up"
    >
      <section v-for="a in panes" :key="a.id" class="relative overflow-hidden border border-line bg-ink min-h-[240px]">
        <img
          :src="a.preview_url"
          :alt="a.filename"
          class="max-w-none select-none pointer-events-none"
          :style="{
            transform: `translate(${transform.x}px, ${transform.y}px) scale(${transform.scale})`,
            transformOrigin: 'center center',
          }"
        />
        <div class="absolute bottom-2 left-2 right-2 bg-ink/70 p-2 font-mono text-[10px] space-y-0.5">
          <p>{{ a.filename }}</p>
          <p :class="diffs.includes('aperture') ? 'text-tungsten' : ''">f/{{ a.aperture }} · {{ a.shutter_text }} · ISO {{ a.iso }} · {{ a.focal_length }}mm</p>
        </div>
      </section>
    </div>
  </div>
</template>
