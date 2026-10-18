import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Frontend CBT SMK v2.
// Di produksi, hasil build di-embed ke binary Go (go:embed).
// Dev: backend di :9988, request /api di-proxy ke backend.
// Plugin tailwindcss WAJIB: tanpa ini, `@import "tailwindcss"` di
// index.css tidak diproses dan TIDAK ADA utility class yang dihasilkan
// (aplikasi tampil tanpa styling — temuan integrasi Fase 1).
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:9988",
    },
  },
});
