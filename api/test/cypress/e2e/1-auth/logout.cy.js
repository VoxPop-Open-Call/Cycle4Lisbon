import { admin } from "../../support/util";

describe("logout", () => {
  beforeEach(() => {
    cy.login(admin());
  });

  it("redirects to login page", () => {
    cy.get('a:contains("Log Out")').click();
    cy.location("pathname").should("eq", "/login");
  });
});
