const { defineConfig } = require("cypress");

module.exports = defineConfig({
  e2e: {
    baseUrl: "http://localhost:3000",
    env: {
      apiUrl: "http://localhost:8080/api",
      dexUrl: "http://localhost:8080/dex",
      apiClientId: "dashboard", // client id and secret from the dexrc example
      apiClientSecret: "KlNQCgzZEGwcXErsxNSZlKzH",
    },
  },
});
