<script setup>
import { onMounted, ref } from "vue";
// A copy of the koru JavaScript module. In the APK it talks to Android;
// under `vite` or `koru dev` it uses web implementations of the same API.
import { device, platform, storage } from "./koru/core.mjs";

// koru errors carry a stable `code`, such as PERMISSION_DENIED or TIMEOUT.
const describe = (error) => (error?.code ? `${error.code}: ${error.message}` : String(error?.message ?? error));

const phone = ref("…");
const count = ref(0);
const ready = ref(false);
const error = ref("");

onMounted(async () => {
  try {
    const info = await device.getInfo();
    phone.value = info.sdk ? `${info.manufacturer} ${info.model} · Android API ${info.sdk}` : "Running in a browser";
    count.value = Number((await storage.get("count")) ?? 0);
    ready.value = true;
  } catch (e) {
    error.value = describe(e);
  }
});

async function increment() {
  count.value += 1;
  try {
    await storage.set("count", String(count.value));
  } catch (e) {
    error.value = describe(e);
  }
}
</script>

<template>
  <main>
    <p class="badge">{{ platform }}</p>
    <h1>__APP_NAME__</h1>
    <p class="lead">Vite + Vue + koru. Edit <code>src/App.vue</code> and save.</p>
    <section class="card">
      <h2>Saved counter</h2>
      <button type="button" :disabled="!ready" @click="increment">Count: {{ count }}</button>
    </section>
    <section class="card">
      <h2>This device</h2>
      <p>{{ phone }}</p>
    </section>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
  </main>
</template>
