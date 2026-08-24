<script setup lang="ts">
import { onMounted, ref } from "vue";
import { api, type Report } from "../api";
import { useSession } from "../stores/session";

const session = useSession();
const report = ref<Report | null>(null);

onMounted(async () => {
  try {
    report.value = await api.report();
  } catch (e) {
    session.flash(e instanceof Error ? e.message : "报告失败", "error");
  }
});

function bars(m: Record<string, number> | undefined) {
  const entries = Object.entries(m || {});
  const max = Math.max(1, ...entries.map(([, n]) => n));
  return entries.map(([k, n]) => ({ k, n, w: (n / max) * 100 }));
}

function exportJSON() {
  if (!report.value) return;
  const blob = new Blob([JSON.stringify(report.value, null, 2)], { type: "application/json" });
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = "golden-lens.json";
  a.click();
}
</script>

<template>
  <div v-if="report" class="w-full px-4 md:px-6 py-6 space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <p class="label">{{ report.window_from }} — {{ report.window_to }} · {{ report.total }} 张</p>
        <h1 class="font-display text-4xl mt-1">黄金挂机镜</h1>
        <p class="text-tungsten text-lg mt-2">{{ report.golden_lens || "样本量不足，暂不排名" }}</p>
        <p class="text-mute">推荐组合 {{ report.recommended_combo || "—" }}</p>
      </div>
      <button class="btn" type="button" @click="exportJSON">导出 JSON</button>
    </div>
    <div class="grid md:grid-cols-2 gap-4">
      <section class="border border-line p-4">
        <h2 class="font-display text-xl mb-3">焦段分布</h2>
        <div v-for="b in bars(report.focal_global)" :key="b.k" class="flex items-center gap-2 mb-1 font-mono text-xs">
          <span class="w-16 text-mute">{{ b.k }}</span>
          <div class="flex-1 h-2 bg-ink"><div class="h-2 bg-tungsten" :style="{ width: b.w + '%' }" /></div>
          <span>{{ b.n }}</span>
        </div>
      </section>
      <section class="border border-line p-4">
        <h2 class="font-display text-xl mb-3">光圈分布</h2>
        <div v-for="b in bars(report.aperture_global)" :key="b.k" class="flex items-center gap-2 mb-1 font-mono text-xs">
          <span class="w-16 text-mute">{{ b.k }}</span>
          <div class="flex-1 h-2 bg-ink"><div class="h-2 bg-film/70" :style="{ width: b.w + '%' }" /></div>
          <span>{{ b.n }}</span>
        </div>
      </section>
    </div>
    <section v-for="l in report.lenses" :key="l.lens" class="border border-line p-4 space-y-2">
      <div class="flex justify-between">
        <h3 class="font-display text-2xl">{{ l.lens }}</h3>
        <p class="font-mono text-tungsten">{{ l.score.toFixed(1) }}</p>
      </div>
      <p v-if="l.insufficient_data" class="text-alert text-sm">insufficient_data · 样本 {{ l.count }} &lt; 30，不参与排名</p>
      <div class="grid sm:grid-cols-4 gap-2 font-mono text-[11px]">
        <div v-for="(f, k) in l.factors" :key="k" :class="f.confidence < 0.4 ? 'opacity-40' : ''">
          {{ k }} {{ f.value.toFixed(2) }} · c={{ f.confidence.toFixed(2) }} · n={{ f.samples }} · ex={{ f.excluded_count }}
        </div>
      </div>
      <ul class="text-mute text-sm list-disc pl-5">
        <li v-for="d in l.derivation" :key="d">{{ d }}</li>
      </ul>
    </section>
  </div>
</template>
