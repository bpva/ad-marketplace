import { fileURLToPath } from "url";
import { dirname, resolve } from "path";
import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const __dirname = dirname(fileURLToPath(import.meta.url));

function tonConnectManifest(): Plugin {
  const manifest = () => {
    const appUrl = process.env.VITE_APP_URL || "http://localhost:1313";
    return JSON.stringify({
      url: appUrl,
      name: "AD\u00d7CHANGE",
      iconUrl: `${appUrl}/logo.png`,
    });
  };

  return {
    name: "tonconnect-manifest",
    configureServer(server) {
      server.middlewares.use("/tonconnect-manifest.json", (_req, res) => {
        res.setHeader("Content-Type", "application/json");
        res.end(manifest());
      });
    },
    generateBundle() {
      this.emitFile({
        type: "asset",
        fileName: "tonconnect-manifest.json",
        source: manifest(),
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), tailwindcss(), tonConnectManifest()],
  resolve: {
    alias: {
      "@": resolve(__dirname, "./src"),
    },
  },
  server: {
    host: true,
    port: 1313,
    allowedHosts: true,
  },
});
