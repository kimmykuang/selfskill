window.appPages = window.appPages || {};

window.appPages.renderPlugins = async function (filterOverride) {
  const app = this;
  const root = document.getElementById("page-content");
  const filter = filterOverride || app._pluginFilter || { state: "" };
  app._pluginFilter = filter;
  let plugins = [];
  try { plugins = await window.api.listPlugins(filter); } catch (e) { app.showError(e); }

  root.innerHTML = `
    <div class="p-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-base font-semibold">Plugins</h2>
        <span class="text-xs text-gh-subtle">${plugins.length} total</span>
      </div>
      <div id="plugin-filters" class="flex gap-2 mb-3"></div>
      <table class="gh-table">
        <thead><tr><th>Name</th><th>Marketplace</th><th>Version</th><th>State</th></tr></thead>
        <tbody id="plugin-rows"></tbody>
      </table>
    </div>
  `;
  window.appComponents.renderFilterBar(
    document.getElementById("plugin-filters"),
    filter.state || "",
    [{ key: "", label: "All" }, { key: "loaded", label: "Loaded" }, { key: "installed", label: "Installed" }],
    (k) => window.appPages.renderPlugins.call(app, { state: k })
  );
  const tbody = document.getElementById("plugin-rows");
  plugins.forEach(p => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td><a class="gh-link">${p.name}</a></td>
      <td class="text-gh-subtle">${p.marketplace}</td>
      <td>${p.version}</td>
      <td><span class="pill ${p.state === "loaded" ? "pill-loaded" : "pill-installed"}">${p.state}</span></td>
    `;
    tr.addEventListener("click", () => {
      tbody.querySelectorAll("tr").forEach(r => r.classList.remove("active"));
      tr.classList.add("active");
      app.setDetail("plugin", p);
    });
    tbody.appendChild(tr);
  });
};
