window.appPages = window.appPages || {};

window.appPages.renderGroups = async function () {
  const app = this;
  const root = document.getElementById("page-content");
  const groups = (await window.api.listGroups()) || [];
  root.innerHTML = `
    <div class="p-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-base font-semibold">Groups</h2>
        <button class="btn btn-primary" id="new-group">New group</button>
      </div>
      <table class="gh-table">
        <thead><tr><th>Name</th><th>Skills</th><th>Plugins</th><th>Prompts</th></tr></thead>
        <tbody id="group-rows"></tbody>
      </table>
    </div>
  `;
  document.getElementById("new-group").onclick = async () => {
    const name = prompt("Group name");
    if (!name) return;
    try { await window.api.createGroup({ name }); app.refreshGroups(); window.appPages.renderGroups.call(app); }
    catch (e) { app.showError(e); }
  };
  const tbody = document.getElementById("group-rows");
  groups.forEach(g => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td><a class="gh-link" href="#/groups/${encodeURIComponent(g.Name)}">${g.Name}</a></td>
      <td>${(g.Skills || []).length}</td>
      <td>${(g.Plugins || []).length}</td>
      <td>${(g.Prompts || []).length}</td>
    `;
    tr.addEventListener("click", () => {
      tbody.querySelectorAll("tr").forEach(r => r.classList.remove("active"));
      tr.classList.add("active");
      app.setDetail("group", g);
    });
    tbody.appendChild(tr);
  });
};

window.appPages.renderGroup = async function (name) {
  const app = this;
  const root = document.getElementById("page-content");
  let g; try { g = await window.api.getGroup(name); } catch (e) { app.showError(e); window.location.hash = "#/groups"; return; }
  app.setDetail("group", g);
  root.innerHTML = `
    <div class="p-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-base font-semibold">${g.Name}</h2>
        <div class="flex gap-2">
          <button class="btn btn-primary" id="g-load">Load</button>
          <button class="btn btn-danger" id="g-unload">Unload</button>
          <button class="btn" id="g-delete">Delete</button>
        </div>
      </div>
      <p class="text-sm text-gh-subtle mb-3">${g.Description || "(no description)"}</p>
      ${section("Skills", g.Skills, "skill")}
      ${section("Plugins", g.Plugins, "plugin")}
      ${section("Prompts", g.Prompts, "prompt")}
    </div>
  `;
  document.getElementById("g-load").onclick = () => window.api.loadGroup(name).then(() => app.showToast("Loaded")).catch(e => app.showError(e));
  document.getElementById("g-unload").onclick = () => window.api.unloadGroup(name).then(() => app.showToast("Unloaded")).catch(e => app.showError(e));
  document.getElementById("g-delete").onclick = async () => {
    if (!confirm("Delete group " + name + "?")) return;
    try { await window.api.deleteGroup(name); app.refreshGroups(); window.location.hash = "#/groups"; }
    catch (e) { app.showError(e); }
  };

  function section(label, items, kind) {
    const list = (items || []).map(i => `<li class="py-1 px-2 border-b border-gh-border text-sm">${i}</li>`).join("");
    return `<div class="mb-4 border border-gh-border rounded">
      <div class="bg-gh-muted px-3 py-1 text-[11px] uppercase text-gh-subtle font-semibold">${label}</div>
      <ul>${list || '<li class="py-2 px-2 text-gh-subtle text-xs">(none)</li>'}</ul>
    </div>`;
  }
};
