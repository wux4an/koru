// Copy buttons on the landing page; the commands stay visible without JavaScript.
document.querySelectorAll("[data-copy]").forEach((button) => {
  const label = button.querySelector("span");
  const original = label.textContent;
  button.addEventListener("click", async () => {
    try {
      await navigator.clipboard.writeText(button.dataset.copy);
      label.textContent = button.dataset.copied;
    } catch {
      label.textContent = "Ctrl+C";
    }
    setTimeout(() => (label.textContent = original), 1600);
  });
});
