import "./style.css";
// A copy of the koru JavaScript module. In the APK it talks to Android;
// under `vite` or `koru dev` it uses web implementations of the same API.
import { device, platform, storage } from "./koru/core.mjs";

/** @param {string} selector */
const $ = (selector) => document.querySelector(selector);

async function main() {
  $("#platform").textContent = platform;
  const phone = await device.getInfo();
  $("#device").textContent = phone.sdk
    ? `${phone.manufacturer} ${phone.model} · Android API ${phone.sdk}`
    : "Running in a browser";

  let count = Number((await storage.get("count")) ?? 0);
  const button = $("#count");
  const render = () => (button.textContent = `Count: ${count}`);
  render();
  button.addEventListener("click", async () => {
    count += 1;
    render();
    await storage.set("count", String(count));
  });
}

main().catch((error) => ($("#device").textContent = error.message));
