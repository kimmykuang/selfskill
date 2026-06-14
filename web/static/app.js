// app.js — Alpine root state + router + ⌘K palette.
function app() {
  return {
    route: "skills",
    loading: false,
    counts: { skills: "", prompts: "", plugins: "" },
    groups: [],
    detail: { kind: null, data: null },
    globalQuery: "",
    cmdkHits: [],
    cmdkActive: 0,
    toast: { visible: false, kind: "info", message: "" },

    init() {
      window.appInstance = this;
      window.addEventListener("hashchange", () => this.handleRoute());
      window.addEventListener("keydown", (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === "k") {
          e.preventDefault();
          this.openCmdK();
        }
      });
      this.refreshGroups();
      this.refreshCounts();
      this.handleRoute();
    },

    async refreshGroups() {
      try { this.groups = (await window.api.listGroups()) || []; }
      catch (e) { this.groups = []; }
    },
    async refreshCounts() {
      try {
        const [s, p, pl] = await Promise.all([
          window.api.listSkills(), window.api.listPrompts(), window.api.listPlugins(),
        ]);
        this.counts.skills = "(" + (s || []).length + ")";
        this.counts.prompts = "(" + (p || []).length + ")";
        this.counts.plugins = "(" + (pl || []).length + ")";
      } catch (e) { /* ignore */ }
    },

    handleRoute() {
      const hash = (window.location.hash || "#/skills").slice(2);
      this.route = hash || "skills";
      this.detail = { kind: null, data: null };
      const page = window.appPages || {};

      const dispatch = (name, ...args) => {
        const fn = page[name];
        if (!fn) return;
        this.loading = true;
        Promise.resolve(fn.call(this, ...args)).finally(() => { this.loading = false; });
      };

      if (this.route === "skills") return dispatch("renderSkills");
      if (this.route === "prompts") return dispatch("renderPrompts");
      if (this.route === "prompts/new") return dispatch("renderPromptForm", null);
      if (this.route.startsWith("prompts/")) return dispatch("renderPromptForm", this.route.slice("prompts/".length));
      if (this.route === "plugins") return dispatch("renderPlugins");
      if (this.route === "groups") return dispatch("renderGroups");
      if (this.route.startsWith("groups/")) return dispatch("renderGroup", this.route.slice("groups/".length));
      if (this.route === "marketplaces") return dispatch("renderMarketplaces");
      if (this.route === "status") return dispatch("renderStatus");
      if (this.route === "search") return dispatch("renderSearch", this.globalQuery);

      document.getElementById("page-content").innerHTML = "<div class='p-6 text-gh-subtle text-sm'>Page not found.</div>";
    },

    showToast(message, kind) {
      this.toast = { visible: true, kind: kind || "info", message };
      setTimeout(() => { this.toast.visible = false; }, 2500);
    },
    showError(e) {
      const msg = (e && e.message) || String(e);
      this.showToast(msg, "error");
    },

    setDetail(kind, data) {
      this.detail = { kind, data };
      const el = document.getElementById("detail-content");
      if (window.appComponents && window.appComponents.renderDetail) {
        window.appComponents.renderDetail(this, el, kind, data);
      }
    },

    runSearch() {
      if (!this.globalQuery) return;
      window.location.hash = "#/search";
      const fn = (window.appPages || {}).renderSearch;
      if (fn) fn.call(this, this.globalQuery);
    },

    openCmdK() {
      const dlg = document.getElementById("cmdk");
      if (!dlg) return;
      this.cmdkHits = []; this.cmdkActive = 0;
      const input = document.getElementById("cmdk-input");
      if (input) input.value = "";
      dlg.showModal();
      setTimeout(() => input && input.focus(), 0);
    },
    async runCmdKQuery(q) {
      if (!q) { this.cmdkHits = []; return; }
      try { this.cmdkHits = (await window.api.search(q)) || []; }
      catch (e) { this.cmdkHits = []; }
    },
    runCmdKChoice(h) {
      const dlg = document.getElementById("cmdk");
      if (dlg && dlg.open) dlg.close();
      if (h.Kind === "skill") window.location.hash = "#/skills";
      else if (h.Kind === "prompt") window.location.hash = "#/prompts/" + encodeURIComponent(h.Name);
      else if (h.Kind === "plugin") window.location.hash = "#/plugins";
      // Refresh after navigation so detail shows.
      this.handleRoute();
    },
  };
}
