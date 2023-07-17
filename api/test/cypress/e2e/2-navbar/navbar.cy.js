import { admin } from "../../support/util";

describe("navbar", () => {
  beforeEach(() => {
    cy.login(admin());
  });

  it("redirects to Users page", () => {
    cy.get('a:contains("Users")').click();
    cy.location("pathname").should("eq", "/users");
  });
});
