window.appComponents = window.appComponents || {};

function kv(label, value) {
  return `<div class="kv"><span>${label}</span><b>${value || ""}</b></div>`;
}

window.appComponents.renderDetail = function (app, el, kind, data) {
  if (!el) return;
  if (!kind) { el.innerHTML = ""; return; }
  if (kind === "skill") {
    const md = window.appComponents.renderMarkdown(data._body || "");
    el.innerHTML = `
      <div class="font-semibold text-base mb-1">${data.Name}</div>
      <div class="text-xs text-gh-subtle mb-3">${data.Description || ""}</div>
      <div class="border border-gh-border rounded mb-3">
        ${kv("Origin", `<span class="pill">${data._origin || "local"}</span>`)}
        ${kv("Version", data.Version || "")}
        ${kv("Source", data.Source ? `<a href="${data.Source}" target="_blank" class="gh-link">${data.Source}</a>` : "(local)")}
        ${kv("Path", `<span class="text-gh-subtle text-xs">${data.DirPath || ""}</span>`)}
      </div>
      <div class="flex gap-2 mb-3">
        <button class="btn" data-act="compare" ${data.Source ? "" : "disabled title='No upstream source'"}>Compare</button>
        <button class="btn" data-act="edit">Edit</button>
        <button class="btn" data-act="reload">Refresh</button>
      </div>
      <div class="prose-md" data-prose>${md || '<p class="text-gh-subtle text-xs">(empty body)</p>'}</div>
    `;
    el.querySelector('[data-act="compare"]').onclick = async () => {
      const proseEl = el.querySelector("[data-prose]");
      try {
        const diff = await window.api.skillDiff(data.Name);
        proseEl.innerHTML = `<pre class="whitespace-pre-wrap text-xs">${(diff || "(no differences)").replace(/[<&]/g, c => c === '<' ? '&lt;' : '&amp;')}</pre>`;
      } catch (e) { app.showError(e); }
    };
    el.querySelector('[data-act="edit"]').onclick = () => {
      const newDesc = prompt("New description (cancel to keep)", data.Description || "");
      if (newDesc === null) return;
      window.api.updateSkill(data.Name, { description: newDesc })
        .then(() => { app.showToast("Updated"); window.appPages.renderSkills.call(app); })
        .catch(e => app.showError(e));
    };
    el.querySelector('[data-act="reload"]').onclick = async () => {
      try {
        const fresh = await window.api.skillBody(data.Name);
        const proseEl = el.querySelector("[data-prose]");
        proseEl.innerHTML = window.appComponents.renderMarkdown(fresh.replace(/^---\n[\s\S]*?\n---\n?/, "")) || "(empty body)";
      } catch (e) { app.showError(e); }
    };
    return;
  }
  if (kind === "prompt") {
    el.innerHTML = `
      <div class="font-semibold mb-2">${data.ID}</div>
      ${kv("Description", data.Description || "")}
      ${kv("Tags", (data.Tags || []).join(", "))}
      <div class="mt-2"><a class="gh-link text-xs" href="#/prompts/${encodeURIComponent(data.ID)}">Open editor →</a></div>
      <pre class="mt-3 whitespace-pre-wrap text-xs bg-white p-2 rounded border border-gh-border">${(data.Body || "").substring(0, 2000)}</pre>
    `;
    return;
  }
  if (kind === "plugin") {
    el.innerHTML = `
      <div class="font-semibold mb-2">${data.name}</div>
      ${kv("Marketplace", data.marketplace)}
      ${kv("Version", data.version)}
      ${kv("State", data.state)}
      <div class="mt-2 flex gap-2">
        ${data.state === "loaded"
          ? '<button class="btn btn-danger" data-act="unload">Unload</button>'
          : '<button class="btn btn-primary" data-act="load">Load</button>'}
      </div>
    `;
    const btnLoad = el.querySelector('[data-act="load"]');
    const btnUnload = el.querySelector('[data-act="unload"]');
    if (btnLoad) btnLoad.onclick = () => window.api.loadPlugin(data.name).then(() => { app.showToast("Loaded"); window.appPages.renderPlugins.call(app); }).catch(e => app.showError(e));
    if (btnUnload) btnUnload.onclick = () => window.api.unloadPlugin(data.name).then(() => { app.showToast("Unloaded"); window.appPages.renderPlugins.call(app); }).catch(e => app.showError(e));
    return;
  }
  if (kind === "group") {
    el.innerHTML = `
      <div class="font-semibold mb-2">${data.Name}</div>
      ${kv("Skills", (data.Skills || []).length)}
      ${kv("Plugins", (data.Plugins || []).length)}
      ${kv("Prompts", (data.Prompts || []).length)}
      <div class="mt-2 flex gap-2">
        <button class="btn btn-primary" data-act="load">Load</button>
        <button class="btn btn-danger" data-act="unload">Unload</button>
      </div>
    `;
    el.querySelector('[data-act="load"]').onclick = () =>
      window.api.loadGroup(data.Name).then(() => app.showToast("Loaded")).catch(e => app.showError(e));
    el.querySelector('[data-act="unload"]').onclick = () =>
      window.api.unloadGroup(data.Name).then(() => app.showToast("Unloaded")).catch(e => app.showError(e));
    return;
  }
  el.innerHTML = `<pre class="whitespace-pre-wrap text-xs">${JSON.stringify(data, null, 2)}</pre>`;
};
