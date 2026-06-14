window.appPages = window.appPages || {};

window.appPages.renderSkills = async function (filterOverride) {
  const app = this;
  const root = document.getElementById("page-content");
  const filter = filterOverride || app._skillFilter || { origin: "" };
  app._skillFilter = filter;

  let skills = [];
  try { skills = await window.api.listSkills(filter); }
  catch (e) { app.showError(e); skills = []; }

  root.innerHTML = `
    <div class="p-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-base font-semibold">Skills</h2>
        <span class="text-xs text-gh-subtle">${skills.length} total</span>
      </div>
      <div id="skill-filters" class="flex gap-2 mb-3"></div>
      <table class="gh-table">
        <thead><tr><th>Name</th><th>Origin</th><th>Source</th><th>Version</th></tr></thead>
        <tbody id="skill-rows"></tbody>
      </table>
    </div>
  `;
  window.appComponents.renderFilterBar(
    document.getElementById("skill-filters"),
    filter.origin || "",
    [
      { key: "", label: "All" },
      { key: "local", label: "Local" },
      { key: "github", label: "GitHub" },
    ],
    (k) => window.appPages.renderSkills.call(app, { origin: k })
  );
  const tbody = document.getElementById("skill-rows");
  tbody.innerHTML = "";
  skills.forEach(sk => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td><a class="gh-link">${sk.Name}</a></td>
      <td><span class="pill">${sk.Origin || "local"}</span></td>
      <td class="text-gh-subtle">${sk.Source || ""}</td>
      <td>${sk.Version || ""}</td>
    `;
    tr.addEventListener("click", async () => {
      tbody.querySelectorAll("tr").forEach(r => r.classList.remove("active"));
      tr.classList.add("active");
      let body = "";
      try { body = await window.api.skillBody(sk.Name); } catch (e) { /* ignore — show without body */ }
      const stripped = stripFrontmatter(body);
      app.setDetail("skill", { ...sk, _origin: sk.Origin, _body: stripped });
    });
    tbody.appendChild(tr);
  });
};

function stripFrontmatter(md) {
  if (!md) return "";
  const m = md.match(/^---\n[\s\S]*?\n---\n?/);
  return m ? md.slice(m[0].length) : md;
}
