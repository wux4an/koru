import { resolve, sep, extname } from "node:path";
import { readFile, stat } from "node:fs/promises";
import { root } from "./build";
const dist = resolve(root, process.env.DIST_DIR || "dist");
const config = JSON.parse(await readFile(resolve(dist, "site.json"), "utf8"));
const port = Number(process.env.PORT || 4173);
const server = Bun.serve({
  hostname: "127.0.0.1",
  port,
  async fetch(request) {
    const pathname = new URL(request.url).pathname;
    if (!pathname.startsWith(config.base))
      return new Response("Not found", { status: 404 });
    let decoded: string;
    try {
      decoded = decodeURIComponent(pathname.slice(config.base.length));
    } catch {
      return new Response("Invalid URL", { status: 400 });
    }
    const filePath = resolve(dist, decoded || "index.html");
    if (filePath !== dist && !filePath.startsWith(dist + sep))
      return new Response("Not found", { status: 404 });
    let isDirectory = false;
    try {
      isDirectory = (await stat(filePath)).isDirectory();
    } catch (error) {
      if (
        !["ENOENT", "ENOTDIR"].includes(
          (error as NodeJS.ErrnoException).code || "",
        )
      )
        throw error;
    }
    const file = Bun.file(
      isDirectory ? resolve(filePath, "index.html") : filePath,
    );
    if (!(await file.exists()))
      return new Response(Bun.file(resolve(dist, "404.html")), {
        status: 404,
        headers: { "content-type": "text/html; charset=utf-8" },
      });
    if (isDirectory && !pathname.endsWith("/"))
      return Response.redirect(new URL(pathname + "/", request.url), 308);
    // Archived example HTML is source material, not another executable app.
    const archive = decoded.startsWith("original/");
    const headers: Record<string, string> = {
      "cache-control": "no-cache",
      "x-content-type-options": "nosniff",
    };
    if (archive) headers["x-robots-tag"] = "noindex";
    if (
      archive &&
      ["", ".md", ".go", ".tape"].includes(extname(file.name || ""))
    )
      headers["content-type"] = "text/plain; charset=utf-8";
    return new Response(file.size === 0 ? "" : file, { status: 200, headers });
  },
});
console.log(
  `koru documentation: http://${server.hostname}:${server.port}${config.base}en/`,
);
