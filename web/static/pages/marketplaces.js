window.appPages = window.appPages || {};

window.appPages.renderMarketplaces = async function () {
  const app = this;
  const root = document.getElementById("page-content");
  const mps = (await window.api.listMarketplaces()) || [];
  root.innerHTML = `
    <div class="p-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-base font-semibold">Marketplaces</h2>
        <button class="btn btn-primary" id="add-mp">Add marketplace</button>
      </div>
      <ul id="mp-list" class="space-y-2"></ul>
    </div>
  `;
  document.getElementById("add-mp").onclick = async () => {
    const url = prompt("Marketplace git URL");
    if (!url) return;
    try { await window.api.addMarketplace(url); window.appPages.renderMarketplaces.call(app); }
    catch (e) { app.showError(e); }
  };
  const list = document.getElementById("mp-list");
  for (const mp of mps) {
    const li = document.createElement("li");
    li.className = "border border-gh-border rounded";
    li.innerHTML = `
      <div class="bg-gh-muted px-3 py-1 flex justify-between items-center">
        <span class="font-medium text-sm">${mp}</span>
      </div>
      <ul class="text-sm" data-mp="${mp}"></ul>
    `;
    list.appendChild(li);
    try {
      const plugins = await window.api.listMarketplacePlugins(mp);
      const ul = li.querySelector("ul");
      plugins.forEach(p => {
        const item = document.createElement("li");
        item.className = "px-3 py-2 border-t border-gh-border flex justify-between items-center";
        item.innerHTML = `<span><b>${p.name}</b> <span class="text-gh-subtle text-xs">${p.description || ""}</span></span>
                          <button class="btn">Install</button>`;
        item.querySelector("button").onclick = async () => {
          try { await window.api.installPlugin(p.name, mp); app.showToast("Installed " + p.name); }
          catch (e) { app.showError(e); }
        };
        ul.appendChild(item);
      });
    } catch (e) { /* ignore per-mp errors */ }
  }
};
