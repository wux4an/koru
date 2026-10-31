import "./core.js";
const invoke = (method, params, options) =>
  window.Koru.invoke(method, params, options);
export { invoke };
/** "android" inside the APK, "web" in a browser or `koru dev`. */
export const platform = window.Koru.platform;
export const isNative = () => platform === "android";
export const device = Object.freeze({
  getInfo: (options) => invoke("device.info", {}, options),
});
export const app = Object.freeze({
  getInfo: (options) => invoke("app.info", {}, options),
});
export const storage = Object.freeze({
  get: (key, options) =>
    invoke("storage.get", { key }, options).then((result) =>
      result.value === undefined ? null : result.value,
    ),
  set: (key, value, options) => invoke("storage.set", { key, value }, options),
  remove: (key, options) => invoke("storage.remove", { key }, options),
  clear: (options) => invoke("storage.clear", {}, options),
});
export const vibration = Object.freeze({
  vibrate: (duration = 100, options) =>
    invoke("vibration.vibrate", { duration }, options),
});
export const share = Object.freeze({
  text: (text, options) => invoke("share.text", { text }, options),
});
export const notifications = Object.freeze({
  show: (params, options) => invoke("notifications.show", params, options),
  cancel: (id, options) => invoke("notifications.cancel", { id }, options),
  permission: (options) => invoke("notifications.permission", {}, options),
  requestPermission: (options) =>
    invoke(
      "notifications.requestPermission",
      {},
      Object.assign({ timeout: 120000 }, options),
    ),
});
export const alarms = Object.freeze({
  schedule: ({ id, title, body, at }, options) =>
    invoke(
      "alarms.schedule",
      { id, title, body, at: at instanceof Date ? at.getTime() : at },
      options,
    ),
  cancel: (id, options) => invoke("alarms.cancel", { id }, options),
  list: (options) =>
    invoke("alarms.list", {}, options).then((result) => [...result.ids]),
});
export const browser = Object.freeze({
  open: (url, options) => invoke("browser.open", { url: String(url) }, options),
});
export const screen = Object.freeze({
  keepAwake: (enabled = true, options) =>
    invoke("screen.keepAwake", { enabled }, options),
});
export const clipboard = Object.freeze({
  writeText: (text, options) =>
    invoke("clipboard.writeText", { text }, options),
  readText: (options) =>
    invoke("clipboard.readText", {}, options).then((result) =>
      result.value === undefined ? null : result.value,
    ),
});
