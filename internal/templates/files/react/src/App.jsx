import { useEffect, useState } from "react";
// A copy of the koru JavaScript module. In the APK it talks to Android;
// under `vite` or `koru dev` it uses web implementations of the same API.
import { device, platform, storage } from "./koru/core.mjs";

// koru errors carry a stable `code`, such as PERMISSION_DENIED or TIMEOUT.
const describe = (error) => (error?.code ? `${error.code}: ${error.message}` : String(error?.message ?? error));

export default function App() {
  const [phone, setPhone] = useState("…");
  const [count, setCount] = useState(0);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    (async () => {
      const info = await device.getInfo();
      const saved = Number((await storage.get("count")) ?? 0);
      if (!active) return;
      setPhone(info.sdk ? `${info.manufacturer} ${info.model} · Android API ${info.sdk}` : "Running in a browser");
      setCount(saved);
      setReady(true);
    })().catch((e) => active && setError(describe(e)));
    return () => {
      active = false;
    };
  }, []);

  async function increment() {
    const next = count + 1;
    setCount(next);
    try {
      await storage.set("count", String(next));
    } catch (e) {
      setError(describe(e));
    }
  }

  return (
    <main>
      <p className="badge">{platform}</p>
      <h1>__APP_NAME__</h1>
      <p className="lead">
        Vite + React + koru. Edit <code>src/App.jsx</code> and save.
      </p>
      <section className="card">
        <h2>Saved counter</h2>
        <button type="button" disabled={!ready} onClick={increment}>
          Count: {count}
        </button>
      </section>
      <section className="card">
        <h2>This device</h2>
        <p>{phone}</p>
      </section>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
    </main>
  );
}
