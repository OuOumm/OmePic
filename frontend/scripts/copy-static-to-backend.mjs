import { cp, rm } from "node:fs/promises";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const outDir = join(__dirname, "..", "out");
const targetDir = join(__dirname, "..", "..", "backend", "web");

try {
  await rm(targetDir, { recursive: true, force: true });
  await cp(outDir, targetDir, { recursive: true });
  console.log("Static files copied to backend/web/");
} catch (err) {
  console.error("Failed to copy static files:", err);
  process.exit(1);
}
