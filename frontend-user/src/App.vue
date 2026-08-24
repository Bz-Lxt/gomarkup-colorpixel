<script setup lang="ts">
import { useRouter, useRoute } from "vue-router";
import { useSession } from "./stores/session";

const router = useRouter();
const route = useRoute();
const session = useSession();

const links = [
  { to: "/", label: "资产墙" },
  { to: "/compare", label: "比对墙" },
  { to: "/histogram", label: "直方图" },
  { to: "/report", label: "挂机镜" },
];

function logout() {
  session.logout();
  router.push("/login");
}
</script>

<template>
  <div class="grain min-h-full flex flex-col">
    <header v-if="route.path !== '/login'" class="border-b border-line px-4 md:px-6 py-3 flex items-center gap-4 w-full">
      <router-link to="/" class="font-display text-xl tracking-tight text-film">ColorPixel</router-link>
      <nav class="flex flex-wrap gap-3 font-mono text-[11px] uppercase tracking-[0.16em] text-mute">
        <router-link v-for="l in links" :key="l.to" :to="l.to" class="hover:text-tungsten" active-class="text-tungsten">
          {{ l.label }}
        </router-link>
      </nav>
      <div class="ml-auto flex items-center gap-3">
        <span class="hidden sm:inline font-mono text-[10px] tracking-widest text-tungsten border border-tungsten/40 px-2 py-0.5">
          预览级 (Embedded JPEG)
        </span>
        <span class="font-mono text-xs text-mute">{{ session.username }}</span>
        <button class="btn-ghost" type="button" @click="logout">退出</button>
      </div>
    </header>
    <main class="flex-1 w-full">
      <router-view />
    </main>
    <div
      v-if="session.toast"
      class="fixed bottom-6 right-6 z-[60] border px-4 py-3 max-w-sm"
      :class="session.toastKind === 'error' ? 'border-alert bg-panel text-film' : 'border-tungsten bg-panel'"
    >
      <div class="flex items-start gap-3">
        <p class="text-sm">{{ session.toast }}</p>
        <button class="text-mute" type="button" aria-label="关闭提示" @click="session.toast = ''">×</button>
      </div>
    </div>
  </div>
</template>
