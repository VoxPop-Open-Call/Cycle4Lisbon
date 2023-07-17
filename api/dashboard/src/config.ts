const config = {
  API_URL: import.meta.env.VITE_API_URL ?? "http://localhost:8080/api",
  ISSUER_URL: import.meta.env.VITE_ISSUER_URL ?? "http://localhost:8080/dex",
  CLIENT_ID: import.meta.env.VITE_CLIENT_ID ?? "dashboard",
  SECRET: import.meta.env.VITE_CLIENT_SECRET ?? "KlNQCgzZEGwcXErsxNSZlKzH",

  SENTRY_DSN: import.meta.env.VITE_SENTRY_DSN ?? "",
};

export default config;
