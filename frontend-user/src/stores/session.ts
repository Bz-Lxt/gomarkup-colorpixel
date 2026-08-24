import { defineStore } from "pinia";
import { api, clearToken, setToken, token } from "../api";

export const useSession = defineStore("session", {
  state: () => ({
    username: localStorage.getItem("cp_user") || "",
    toast: "" as string,
    toastKind: "info" as "info" | "error",
  }),
  getters: {
    authed: (s) => !!s.username && !!token(),
  },
  actions: {
    async login(u: string, p: string) {
      const r = await api.login(u, p);
      setToken(r.token);
      this.username = r.username;
      localStorage.setItem("cp_user", r.username);
    },
    logout() {
      clearToken();
      localStorage.removeItem("cp_user");
      this.username = "";
    },
    flash(msg: string, kind: "info" | "error" = "info") {
      this.toast = msg;
      this.toastKind = kind;
      window.setTimeout(() => {
        if (this.toast === msg) this.toast = "";
      }, 5000);
    },
  },
});
