<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useSession } from "../stores/session";

const username = ref("photographer");
const password = ref("colorpixel");
const err = ref("");
const loading = ref(false);
const session = useSession();
const router = useRouter();

async function submit() {
  err.value = "";
  if (!username.value || !password.value) {
    err.value = "请填写用户名与密码";
    return;
  }
  loading.value = true;
  try {
    await session.login(username.value, password.value);
    router.push("/");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "登录失败";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="min-h-full flex items-center justify-center px-4">
    <form class="w-full max-w-md border border-line bg-panel/80 p-8 space-y-5" @submit.prevent="submit">
      <p class="font-mono text-[11px] tracking-[0.3em] text-tungsten uppercase">Mini Pixel Wall</p>
      <h1 class="font-display text-4xl">暗房观测台</h1>
      <p class="text-mute text-sm">高并发 RAW 资产、同步画幅比对与镜头 EXIF 审计。</p>
      <label class="block space-y-1">
        <span class="label">用户名 *</span>
        <input v-model="username" class="field" autocomplete="username" />
      </label>
      <label class="block space-y-1">
        <span class="label">密码 *</span>
        <input v-model="password" type="password" class="field" autocomplete="current-password" />
      </label>
      <p v-if="err" class="text-alert text-sm">{{ err }}</p>
      <button class="btn w-full" :disabled="loading">{{ loading ? "进入中…" : "进入工作台" }}</button>
    </form>
  </div>
</template>
