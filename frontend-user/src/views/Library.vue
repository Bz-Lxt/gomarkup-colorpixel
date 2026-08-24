<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { api, type Asset } from "../api";
import { useSession } from "../stores/session";

const items = ref<Asset[]>([]);
const total = ref(0);
const page = ref(1);
const q = ref("");
const camera = ref("");
const lens = ref("");
const loading = ref(false);
const session = useSession();
const router = useRouter();
const confirmDel = ref<Asset | null>(null);
const selected = ref<number[]>([]);

async function load() {
  loading.value = true;
  try {
    const p = new URLSearchParams();
    p.set("page", String(page.value));
    p.set("page_size", "40");
    if (q.value) p.set("q", q.value);
    if (camera.value) p.set("camera", camera.value);
    if (lens.value) p.set("lens", lens.value);
    const r = await api.list(p);
    items.value = r.items || [];
    total.value = r.total;
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "加载失败", "error");
  } finally {
    loading.value = false;
  }
}

onMounted(load);

async function onUpload(ev: Event) {
  const input = ev.target as HTMLInputElement;
  if (!input.files?.length) return;
  try {
    await api.upload(Array.from(input.files));
    session.flash("上传完成，预览抽取中");
    await load();
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "上传失败", "error");
  }
  input.value = "";
}

function toggle(id: number) {
  const i = selected.value.indexOf(id);
  if (i >= 0) selected.value.splice(i, 1);
  else if (selected.value.length < 4) selected.value.push(id);
}

function goCompare() {
  if (selected.value.length < 2) {
    session.flash("请至少勾选 2 张进行比对", "error");
    return;
  }
  router.push({ path: "/compare", query: { ids: selected.value.join(",") } });
}

async function doDelete() {
  if (!confirmDel.value) return;
  try {
    await api.del(confirmDel.value.id);
    session.flash("已移入回收");
    confirmDel.value = null;
    await load();
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "删除失败", "error");
  }
}
</script>

<template>
  <div class="w-full px-4 md:px-6 py-5 space-y-5">
    <div class="flex flex-col md:flex-row md:items-end gap-3">
      <div class="flex-1 grid grid-cols-1 sm:grid-cols-3 gap-3">
        <label class="space-y-1">
          <span class="label">检索</span>
          <input v-model="q" class="field" placeholder="文件 / 机身 / 镜头" @keyup.enter="load" />
        </label>
        <label class="space-y-1">
          <span class="label">机身</span>
          <input v-model="camera" class="field" @keyup.enter="load" />
        </label>
        <label class="space-y-1">
          <span class="label">镜头</span>
          <input v-model="lens" class="field" @keyup.enter="load" />
        </label>
      </div>
      <div class="flex flex-wrap gap-2">
        <button class="btn-ghost" type="button" @click="load">筛选</button>
        <label class="btn cursor-pointer">
          上传 RAW
          <input type="file" multiple class="hidden" accept=".cr3,.cr2,.nef,.arw,.dng" @change="onUpload" />
        </label>
        <button class="btn-ghost" type="button" @click="goCompare">比对已选 ({{ selected.length }})</button>
      </div>
    </div>
    <p class="font-mono text-xs text-mute">{{ total }} 张 · 第 {{ page }} 页</p>
    <p v-if="loading" class="text-mute">载入资产墙…</p>
    <div v-else class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-6 gap-3">
      <article v-for="a in items" :key="a.id" class="border border-line bg-panel group relative">
        <button class="block w-full text-left" type="button" @click="router.push('/assets/' + a.id)">
          <div class="aspect-[3/2] overflow-hidden bg-ink">
            <img :src="a.thumb_url" :alt="a.filename" class="w-full h-full object-cover" />
          </div>
          <div class="p-2 space-y-1">
            <p class="font-mono text-[11px] truncate">{{ a.filename }}</p>
            <p class="font-mono text-[10px] text-mute truncate">
              {{ a.camera_model }} · {{ a.shutter_text }} · f/{{ a.aperture.toFixed?.(1) || a.aperture }} · ISO {{ a.iso }}
            </p>
          </div>
        </button>
        <div class="absolute top-2 left-2 flex gap-1">
          <button
            type="button"
            class="text-[10px] font-mono px-1.5 py-0.5 border"
            :class="selected.includes(a.id) ? 'border-tungsten text-tungsten' : 'border-line text-mute bg-ink/60'"
            :aria-label="'选择 ' + a.filename"
            @click.stop="toggle(a.id)"
          >选</button>
        </div>
        <button type="button" class="absolute top-2 right-2 text-mute hover:text-alert text-sm" aria-label="删除资产" @click.stop="confirmDel = a">×</button>
      </article>
    </div>
    <div v-if="confirmDel" class="fixed inset-0 z-40 bg-ink/70 flex items-center justify-center p-4">
      <div class="bg-panel border border-line p-6 max-w-md w-full space-y-4">
        <h2 class="font-display text-2xl">移出资产墙？</h2>
        <p class="text-mute text-sm">将软删除 {{ confirmDel.filename }}，可在后续清理任务中回收磁盘。</p>
        <div class="flex justify-end gap-2">
          <button class="btn-ghost" type="button" @click="confirmDel = null">取消</button>
          <button class="btn" type="button" @click="doDelete">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
