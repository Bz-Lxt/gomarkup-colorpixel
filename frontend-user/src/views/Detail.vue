<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api, type Asset } from "../api";
import { useSession } from "../stores/session";

const route = useRoute();
const router = useRouter();
const session = useSession();
const asset = ref<Asset | null>(null);
const rating = ref(0);
const tagText = ref("");

async function load() {
  const id = Number(route.params.id);
  try {
    asset.value = await api.get(id);
    rating.value = asset.value.rating;
    tagText.value = (asset.value.tags || []).join(",");
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "加载失败", "error");
  }
}

onMounted(load);
watch(() => route.params.id, load);

async function save() {
  if (!asset.value) return;
  if (rating.value < 0 || rating.value > 5) {
    session.flash("评分须为 0-5", "error");
    return;
  }
  try {
    asset.value = await api.patch(asset.value.id, {
      rating: rating.value,
      tags: tagText.value.split(",").map((s) => s.trim()).filter(Boolean),
    });
    session.flash("已保存评级");
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "保存失败", "error");
  }
}
</script>

<template>
  <div v-if="asset" class="w-full px-4 md:px-6 py-6 grid lg:grid-cols-2 gap-6">
    <div class="border border-line bg-ink">
      <img :src="asset.preview_url" :alt="asset.filename" class="w-full h-auto" />
    </div>
    <div class="space-y-4">
      <p class="font-mono text-[10px] text-tungsten tracking-widest">{{ asset.fidelity_label }} · {{ asset.extraction_mode }}</p>
      <h1 class="font-display text-3xl">{{ asset.filename }}</h1>
      <dl class="grid grid-cols-2 gap-3 font-mono text-sm">
        <div><dt class="label">机身</dt><dd>{{ asset.camera_make }} {{ asset.camera_model }}</dd></div>
        <div><dt class="label">镜头</dt><dd>{{ asset.lens_model }}</dd></div>
        <div><dt class="label">光圈</dt><dd>f/{{ asset.aperture }}</dd></div>
        <div><dt class="label">快门</dt><dd>{{ asset.shutter_text }}</dd></div>
        <div><dt class="label">ISO</dt><dd>{{ asset.iso }}</dd></div>
        <div><dt class="label">焦距</dt><dd>{{ asset.focal_length }}mm / {{ asset.focal_length_35mm }}eq</dd></div>
        <div><dt class="label">拍摄</dt><dd>{{ asset.datetime_original }}</dd></div>
        <div><dt class="label">白平衡</dt><dd>{{ asset.white_balance }}</dd></div>
      </dl>
      <div class="space-y-2">
        <label class="label">人工评级 * (1-5)</label>
        <input v-model.number="rating" type="number" min="0" max="5" class="field max-w-[8rem]" />
        <p v-if="rating < 0 || rating > 5" class="text-alert text-xs">评分必须在 0–5</p>
        <label class="label">标签</label>
        <input v-model="tagText" class="field" placeholder="keeper,hero" />
        <button class="btn" type="button" @click="save">保存</button>
        <button class="btn-ghost ml-2" type="button" @click="router.push({ path: '/compare', query: { ids: String(asset.id) } })">加入比对</button>
      </div>
      <details class="border border-line p-3">
        <summary class="font-mono text-xs text-mute">原始 EXIF tag 树</summary>
        <pre class="mt-3 text-[11px] font-mono text-mute overflow-auto">{{ JSON.stringify(asset.exif_raw, null, 2) }}</pre>
      </details>
    </div>
  </div>
</template>
