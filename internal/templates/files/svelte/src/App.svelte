<script>
  import { onMount } from "svelte";
  // A copy of the koru JavaScript module. In the APK it talks to Android;
  // under `vite` or `koru dev` it uses web implementations of the same API.
  import { device, platform, storage } from "./koru/core.mjs";

  // koru errors carry a stable `code`, such as PERMISSION_DENIED or TIMEOUT.
  const describe = (error) => (error?.code ? `${error.code}: ${error.message}` : String(error?.message ?? error));

  let phone = $state("…");
  let count = $state(0);
  let ready = $state(false);
  let error = $state("");

  onMount(async () => {
    try {
      const info = await device.getInfo();
      phone = info.sdk ? `${info.manufacturer} ${info.model} · Android API ${info.sdk}` : "Running in a browser";
      count = Number((await storage.get("count")) ?? 0);
      ready = true;
    } catch (e) {
      error = describe(e);
    }
  });

  async function increment() {
    count += 1;
    try {
      await storage.set("count", String(count));
    } catch (e) {
      error = describe(e);
    }
  }
</script>

<main>
  <p class="badge">{platform}</p>
  <h1>__APP_NAME__</h1>
  <p class="lead">Vite + Svelte + koru. Edit <code>src/App.svelte</code> and save.</p>
  <section class="card">
    <h2>Saved counter</h2>
    <button type="button" disabled={!ready} onclick={increment}>Count: {count}</button>
  </section>
  <section class="card">
    <h2>This device</h2>
    <p>{phone}</p>
  </section>
  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
</main>
