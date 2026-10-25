document.addEventListener("DOMContentLoaded", () => {
    if (window.hljs) {
        hljs.highlightAll();
    }
});

document.addEventListener("htmx:afterSettle", () => {
    if (!window.hljs) return;

    document.querySelectorAll("pre code:not(.hljs)").forEach((block) => {
        hljs.highlightElement(block);
    });
});
