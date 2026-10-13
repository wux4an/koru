// Load before CSS so a stored theme is applied before the first paint.
(() => {
  const key = "koru-theme";
  const choices = ["system", "light", "dark"];
  const system = window.matchMedia("(prefers-color-scheme: dark)");
  let preference = "system";
  try {
    const saved = localStorage.getItem(key);
    if (choices.includes(saved)) preference = saved;
  } catch {
    // Preference storage is optional in private/restricted browsers.
  }
  const apply = () => {
    const dark =
      preference === "dark" || (preference === "system" && system.matches);
    const theme = dark ? "dark" : "light";
    document.documentElement.dataset.theme = theme;
    document.documentElement.style.colorScheme = theme;
    const color = document.querySelector('meta[name="theme-color"]');
    if (color) color.content = dark ? "#171F32" : "#F7F9FD";
    const control = document.querySelector("[data-theme-button]");
    if (control) control.dataset.preference = preference;
    document.querySelectorAll("[data-theme-choice]").forEach((item) => {
      const checked = item.dataset.themeChoice === preference;
      item.setAttribute("aria-checked", String(checked));
      if (checked)
        document.querySelector("[data-theme-label]").textContent =
          item.dataset.label;
    });
    document.querySelectorAll("[data-theme-icon]").forEach((icon) => {
      icon.className = `ph-duotone ph-${preference === "system" ? "desktop" : dark ? "moon" : "sun"}`;
    });
  };
  document.addEventListener("DOMContentLoaded", () => {
    const control = document.querySelector(".theme-control");
    const button = control.querySelector("[data-theme-button]");
    const menu = control.querySelector("[data-theme-menu]");
    const items = [...menu.querySelectorAll("[data-theme-choice]")];
    const close = (restoreFocus = false) => {
      menu.hidden = true;
      button.setAttribute("aria-expanded", "false");
      if (restoreFocus) button.focus();
    };
    const open = (
      index = items.findIndex(
        (item) => item.dataset.themeChoice === preference,
      ),
    ) => {
      menu.hidden = false;
      button.setAttribute("aria-expanded", "true");
      items[Math.max(0, index)].focus();
    };
    button.addEventListener("click", () =>
      menu.hidden ? open() : close(true),
    );
    button.addEventListener("keydown", (event) => {
      if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;
      event.preventDefault();
      open(event.key === "ArrowDown" ? 0 : items.length - 1);
    });
    menu.addEventListener("keydown", (event) => {
      const current = items.indexOf(document.activeElement);
      let next;
      if (event.key === "ArrowDown") next = (current + 1) % items.length;
      if (event.key === "ArrowUp")
        next = (current + items.length - 1) % items.length;
      if (event.key === "Home") next = 0;
      if (event.key === "End") next = items.length - 1;
      if (next !== undefined) {
        event.preventDefault();
        items[next].focus();
      }
      if (event.key === "Escape") {
        event.preventDefault();
        event.stopPropagation();
        close(true);
      }
      if (event.key === "Tab") close();
    });
    items.forEach((item) =>
      item.addEventListener("click", () => {
        preference = item.dataset.themeChoice;
        try {
          localStorage.setItem(key, preference);
        } catch {
          /* Keep the choice for this document. */
        }
        apply();
        close(true);
      }),
    );
    document.addEventListener("pointerdown", (event) => {
      if (!control.contains(event.target)) close();
    });
    control.addEventListener("focusout", (event) => {
      if (!control.contains(event.relatedTarget)) close();
    });
    apply();
  });
  system.addEventListener("change", () => {
    if (preference === "system") apply();
  });
  window.addEventListener("storage", (event) => {
    if (event.key !== key && event.key !== null) return;
    preference = choices.includes(event.newValue) ? event.newValue : "system";
    apply();
  });
  apply();
})();
