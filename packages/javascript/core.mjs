import "./core.js";
const invoke = (method, params, options) =>
  window.Koru.invoke(method, params, options);
export { invoke };
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
