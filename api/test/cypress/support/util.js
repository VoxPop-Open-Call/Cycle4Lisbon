export const admin = () => ({
  email: Cypress.env("ADMIN_EMAIL"),
  password: Cypress.env("ADMIN_PASSWD"),
});
