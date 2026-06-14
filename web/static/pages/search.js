window.appPages = window.appPages || {};

window.appPages.renderSearch = async function (q) {
  const app = this;
  const root = document.getElementById("page-content");
  let hits = [];
  try { hits = await window.api.search(q || ""); } catch (e) { app.showError(e); }
  root.innerHTML = `
    <div class="p-4">
      <h2 class="text-base font-semibold mb-3">Search "${q || ""}"</h2>
      <ul class="space-y-2">
        ${(hits || []).map(h => `
          <li class="border border-gh-border rounded p-2 hover:bg-gh-muted">
            <span class="text-xs text-gh-subtle uppercase mr-2">${h.Kind}</span>
            <span class="font-medium">${h.Name}</span>
            <span class="text-gh-subtle text-xs ml-2">${h.Description || ""}</span>
          </li>`).join("") || '<li class="text-gh-subtle text-xs">No results.</li>'}
      </ul>
    </div>
  `;
};

window.appPages.renderStatus = async function () {
  const app = this;
  const root = document.getElementById("page-content");
  let st = { userLinks: [], projectLinks: [] };
  try { st = await window.api.status(); } catch (e) { app.showError(e); }
  root.innerHTML = `
    <div class="p-4 grid grid-cols-2 gap-4">
      <section>
        <h2 class="text-base font-semibold mb-2">User scope</h2>
        <ul class="space-y-1">
          ${(st.userLinks || []).map(l => `<li class="text-sm border border-gh-border rounded px-2 py-1">${l.skillName}</li>`).join("") || '<li class="text-gh-subtle text-xs">(none)</li>'}
        </ul>
      </section>
      <section>
        <h2 class="text-base font-semibold mb-2">Project scope</h2>
        <ul class="space-y-1">
          ${(st.projectLinks || []).map(l => `<li class="text-sm border border-gh-border rounded px-2 py-1">${l.skillName}</li>`).join("") || '<li class="text-gh-subtle text-xs">(none)</li>'}
        </ul>
      </section>
    </div>
  `;
};
