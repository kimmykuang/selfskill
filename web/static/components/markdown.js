window.appComponents = window.appComponents || {};
window.appComponents.renderMarkdown = function (text) {
  if (!text) return "";
  if (typeof window.marked === "undefined") {
    return "<pre class='whitespace-pre-wrap text-xs'>" +
      text.replace(/&/g, "&amp;").replace(/</g, "&lt;") + "</pre>";
  }
  return window.marked.parse(text, { breaks: true, mangle: false, headerIds: false });
};
