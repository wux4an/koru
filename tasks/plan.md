# Plan: experiencia de desarrollo (ver docs/specs/developer-experience.md)

Orden por dependencias; cada tarea deja el árbol verde.

- [x] 1. JS: modo web en `core.js`, `platform`, tipos `.d.ts`; tests node.
- [x] 2. Config: `$schema` permitido, JSON Schema, mensajes de error con pista.
- [x] 3. Plantillas embebidas + `koru create` (basic, vite) con prompts en TTY.
- [x] Checkpoint: create → build → verify en ambas plantillas.
- [x] 4. `internal/devserver` + `koru dev` (estático, SPA, SSE, proxy).
- [x] 5. `internal/adb` + `koru devices` + `koru run`; ADR 0003; doctor.
- [x] Checkpoint: go test/vet/race, gofmt, node --test.
- [x] 6. Sitio: landing page en/es, guías nuevas (dev, run, frameworks,
      TypeScript, comparación), CLI/config/API actualizadas, README.
- [x] 7. milestones.md, tests/README.md, resultados; zip final.
