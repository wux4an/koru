import "./style.css";
// A copy of the koru JavaScript module, typed by the core.d.mts next to it.
// In the APK it talks to Android; under `vite` or `koru dev` it uses web
// implementations of the same API.
import { device, platform, storage } from "./koru/core.mjs";

function $<T extends HTMLElement>(selector: string): T {
  const element = document.querySelector<T>(selector);
  if (!element) throw new Error(`Missing element ${selector}`);
  return element;
}

// koru errors carry a stable `code`, such as PERMISSION_DENIED or TIMEOUT.
function describe(error: unknown): string {
  if (error instanceof Error) {
    const code = (error as { code?: string }).code;
    return code ? `${code}: ${error.message}` : error.message;
  }
  return String(error);
}

async function main(): Promise<void> {
  $("#platform").textContent = platform;
  const phone = await device.getInfo();
  $("#device").textContent = phone.sdk
    ? `${phone.manufacturer} ${phone.model} · Android API ${phone.sdk}`
    : "Running in a browser";

  let count = Number((await storage.get("count")) ?? 0);
  const button = $<HTMLButtonElement>("#count");
  const render = (): void => {
    button.textContent = `Count: ${count}`;
  };
  render();
  button.disabled = false;
  button.addEventListener("click", async () => {
    count += 1;
    render();
    try {
      await storage.set("count", String(count));
    } catch (e) {
      $("#error").textContent = describe(e);
    }
  });
}

main().catch((e: unknown) => {
  $("#error").textContent = describe(e);
});
