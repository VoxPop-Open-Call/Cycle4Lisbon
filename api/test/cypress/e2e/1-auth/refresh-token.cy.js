import { admin } from "../../support/util";

describe("refresh token", () => {
  beforeEach(() => {
    cy.clock(Date.now());
    cy.login(admin());
  });

  it("is used to get a new token before making a request, if the current one has expired", () => {
    cy.intercept("POST", "**/dex/token").as("refreshToken");

    const second = 1000;
    const day = second * 60 * 60 * 24;
    cy.tick(2 * day);

    cy.get('a:contains("Users")').click();
    cy.location("pathname").should("eq", "/users");

    cy.wait("@refreshToken");
  });
});
