window.appComponents = window.appComponents || {};
// renderFilterBar: container, current state, options [{key, label}], onPick(key)
window.appComponents.renderFilterBar = function (container, current, options, onPick) {
  container.innerHTML = "";
  options.forEach(opt => {
    const span = document.createElement("span");
    span.textContent = opt.label;
    span.className = "chip" + (current === opt.key ? " chip-active" : "");
    span.addEventListener("click", () => onPick(opt.key));
    container.appendChild(span);
  });
};
