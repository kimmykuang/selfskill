// api.js — central fetch wrapper. Other modules use window.api.
window.api = (function () {
  function toQuery(params) {
    if (!params) return "";
    const u = new URLSearchParams();
    for (const [k, v] of Object.entries(params)) {
      if (v === "" || v === null || v === undefined) continue;
      u.set(k, v);
    }
    const s = u.toString();
    return s ? "?" + s : "";
  }

  async function call(method, path, body, params) {
    const opts = { method, headers: {} };
    if (body !== undefined) {
      opts.body = JSON.stringify(body);
      opts.headers["Content-Type"] = "application/json";
    }
    const res = await fetch(path + toQuery(params), opts);
    const text = await res.text();
    let data = null;
    if (text) {
      try { data = JSON.parse(text); } catch (e) { data = text; }
    }
    if (!res.ok) {
      const err = new Error((data && data.error) || res.statusText);
      err.status = res.status;
      err.body = data;
      throw err;
    }
    return data;
  }

  return {
    listSkills: (filter) => call("GET", "/api/skills", undefined, filter),
    getSkill: (name) => call("GET", "/api/skills/" + encodeURIComponent(name)),
    updateSkill: (name, body) => call("PUT", "/api/skills/" + encodeURIComponent(name), body),
    skillFiles: (name) => call("GET", "/api/skills/" + encodeURIComponent(name) + "/files"),
    skillFile: (name, p) => fetch("/api/skills/" + encodeURIComponent(name) + "/files/" + p).then(r => r.text()),
    skillBody: (name) => fetch("/api/skills/" + encodeURIComponent(name) + "/files/SKILL.md").then(r => r.text()),
    skillDiff: (name) => fetch("/api/skills/" + encodeURIComponent(name) + "/diff").then(r => r.text()),

    listPrompts: (filter) => call("GET", "/api/prompts", undefined, filter),
    getPrompt: (id) => call("GET", "/api/prompts/" + encodeURIComponent(id)),
    createPrompt: (body) => call("POST", "/api/prompts", body),
    updatePrompt: (id, body) => call("PUT", "/api/prompts/" + encodeURIComponent(id), body),
    deletePrompt: (id) => call("DELETE", "/api/prompts/" + encodeURIComponent(id)),

    listPlugins: (filter) => call("GET", "/api/plugins", undefined, filter),
    getPlugin: (name) => call("GET", "/api/plugins/" + encodeURIComponent(name)),
    pluginSkills: (name) => call("GET", "/api/plugins/" + encodeURIComponent(name) + "/skills"),
    loadPlugin: (name) => call("POST", "/api/plugins/" + encodeURIComponent(name) + "/load"),
    unloadPlugin: (name) => call("POST", "/api/plugins/" + encodeURIComponent(name) + "/unload"),

    listGroups: () => call("GET", "/api/groups"),
    getGroup: (name) => call("GET", "/api/groups/" + encodeURIComponent(name)),
    createGroup: (body) => call("POST", "/api/groups", body),
    updateGroup: (name, body) => call("PUT", "/api/groups/" + encodeURIComponent(name), body),
    deleteGroup: (name) => call("DELETE", "/api/groups/" + encodeURIComponent(name)),
    loadGroup: (name) => call("POST", "/api/groups/" + encodeURIComponent(name) + "/load"),
    unloadGroup: (name) => call("POST", "/api/groups/" + encodeURIComponent(name) + "/unload"),

    listMarketplaces: () => call("GET", "/api/marketplaces"),
    addMarketplace: (url) => call("POST", "/api/marketplaces", { url }),
    listMarketplacePlugins: (mp) => call("GET", "/api/marketplaces/" + encodeURIComponent(mp) + "/plugins"),
    installPlugin: (plugin, marketplace) => call("POST", "/api/plugins/install", { plugin, marketplace }),

    install: (type, source, force) => call("POST", "/api/install", { type, source, force }),
    search: (q) => call("GET", "/api/search", undefined, { q }),
    status: () => call("GET", "/api/status"),
  };
})();
