// The koru runtime serves this module inside the APK. Under `koru dev` and in
// a desktop browser the same calls use web implementations.
import { app, device, platform, share, storage, vibration } from "/__koru/core.mjs";

/** @param {string} selector */
const $ = (selector) => document.querySelector(selector);
const status = (text) => ($("#status").textContent = text);

async function main() {
  const info = await app.getInfo();
  $("#platform").textContent = `${platform} · v${info.version}`;

  const phone = await device.getInfo();
  $("#device").innerHTML = "";
  for (const [label, value] of [
    ["Model", phone.model],
    ["Manufacturer", phone.manufacturer || "—"],
    ["Android API", phone.sdk ? String(phone.sdk) : "not Android"],
  ]) {
    const dt = document.createElement("dt");
    const dd = document.createElement("dd");
    dt.textContent = label;
    dd.textContent = String(value);
    $("#device").append(dt, dd);
  }

  let count = Number((await storage.get("count")) ?? 0);
  const button = $("#count");
  button.textContent = `Count: ${count}`;
  button.addEventListener("click", async () => {
    count += 1;
    button.textContent = `Count: ${count}`;
    await storage.set("count", String(count));
  });

  $("#vibrate").addEventListener("click", () =>
    vibration.vibrate(80).then(() => status("Buzz!"), (error) => status(error.message)),
  );
  $("#share").addEventListener("click", () =>
    share.text(`I'm using ${info.name}`).then(() => status("Shared."), (error) => status(error.message)),
  );
}

main().catch((error) => status(`${error.code ?? "ERROR"}: ${error.message}`));
