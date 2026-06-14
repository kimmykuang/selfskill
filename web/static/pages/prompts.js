window.appPages = window.appPages || {};

window.appPages.renderPrompts = async function () {
  const app = this;
  const root = document.getElementById("page-content");
  let prompts = [];
  try { prompts = await window.api.listPrompts(); } catch (e) { app.showError(e); }

  root.innerHTML = `
    <div class="p-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-base font-semibold">Prompts</h2>
        <a href="#/prompts/new" class="btn btn-primary">New</a>
      </div>
      <table class="gh-table">
        <thead><tr><th>ID</th><th>Description</th><th>Tags</th></tr></thead>
        <tbody id="prompt-rows"></tbody>
      </table>
    </div>
  `;
  const tbody = document.getElementById("prompt-rows");
  prompts.forEach(p => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td><a class="gh-link">${p.ID}</a></td>
      <td class="text-gh-subtle">${p.Description || ""}</td>
      <td>${(p.Tags || []).map(t => `<span class="pill">${t}</span>`).join(" ")}</td>
    `;
    tr.addEventListener("click", () => {
      tbody.querySelectorAll("tr").forEach(r => r.classList.remove("active"));
      tr.classList.add("active");
      app.setDetail("prompt", p);
    });
    tbody.appendChild(tr);
  });
};

window.appPages.renderPromptForm = async function (id) {
  const app = this;
  const root = document.getElementById("page-content");
  let p = { ID: "", Description: "", Tags: [], Body: "" };
  let isNew = true;
  if (id) {
    isNew = false;
    try { p = await window.api.getPrompt(id); } catch (e) { app.showError(e); window.location.hash = "#/prompts"; return; }
  }
  root.innerHTML = `
    <div class="p-4">
      <h2 class="text-base font-semibold mb-3">${isNew ? "New Prompt" : "Edit: " + p.ID}</h2>
      <form id="prompt-form" class="space-y-3 max-w-2xl">
        <div><label class="label">ID</label>
          <input class="input w-full" id="f-id" value="${p.ID || ""}" ${isNew ? "" : "disabled"} required></div>
        <div><label class="label">Description</label>
          <input class="input w-full" id="f-desc" value="${p.Description || ""}"></div>
        <div><label class="label">Tags (comma)</label>
          <input class="input w-full" id="f-tags" value="${(p.Tags || []).join(", ")}"></div>
        <div><label class="label">Body</label>
          <textarea class="input w-full font-mono text-xs h-72" id="f-body">${p.Body || ""}</textarea></div>
        <div class="flex gap-2">
          <button type="submit" class="btn btn-primary">${isNew ? "Create" : "Save"}</button>
          ${isNew ? "" : '<button type="button" class="btn btn-danger" id="f-del">Delete</button>'}
          <a href="#/prompts" class="btn">Cancel</a>
        </div>
      </form>
    </div>
  `;
  document.getElementById("prompt-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const tags = (document.getElementById("f-tags").value || "")
      .split(",").map(s => s.trim()).filter(s => s);
    const body = {
      ID: document.getElementById("f-id").value,
      Description: document.getElementById("f-desc").value,
      Tags: tags,
      Body: document.getElementById("f-body").value,
    };
    try {
      if (isNew) await window.api.createPrompt(body);
      else await window.api.updatePrompt(body.ID, body);
      app.showToast("Saved");
      window.location.hash = "#/prompts";
    } catch (err) { app.showError(err); }
  });
  const del = document.getElementById("f-del");
  if (del) del.addEventListener("click", async () => {
    if (!confirm("Delete prompt " + p.ID + "?")) return;
    try { await window.api.deletePrompt(p.ID); window.location.hash = "#/prompts"; }
    catch (err) { app.showError(err); }
  });
};
