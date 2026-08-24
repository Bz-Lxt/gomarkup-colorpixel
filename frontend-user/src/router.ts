import { createRouter, createWebHistory } from "vue-router";
import { token } from "./api";
import Login from "./views/Login.vue";
import Library from "./views/Library.vue";
import Detail from "./views/Detail.vue";
import Compare from "./views/Compare.vue";
import Histogram from "./views/Histogram.vue";
import Report from "./views/Report.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", component: Login },
    { path: "/", component: Library },
    { path: "/assets/:id", component: Detail },
    { path: "/compare", component: Compare },
    { path: "/histogram", component: Histogram },
    { path: "/report", component: Report },
  ],
});

router.beforeEach((to) => {
  if (to.path !== "/login" && !token()) return "/login";
  if (to.path === "/login" && token()) return "/";
  return true;
});

export default router;
