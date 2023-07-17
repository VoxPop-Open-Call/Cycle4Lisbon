import { sentryVitePlugin } from "@sentry/vite-plugin";
import react from "@vitejs/plugin-react-swc";
import { defineConfig, loadEnv } from "vite";
import svgr from "vite-plugin-svgr";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  return {
    build: {
      sourcemap: true,
      outDir: "build",
    },
    plugins: [
      sentryVitePlugin({
        org: "pensarmais",
        project: "cycle-for-lisbon-dashboard",
        authToken: env.VITE_SENTRY_AUTH_TOKEN,
        release: {
          // Release name, defaults to the commit hash.
          //
          // This environment variable is only used when deploying from a
          // container, because the plugin cannot automatically generate a
          // release name if not inside a git repository.
          name: env.VITE_SENTRY_RELEASE,
        },
        sourcemaps: {
          assets: "./build/**",
        },
      }),
      react(),
      svgr(),
    ],
  };
});
