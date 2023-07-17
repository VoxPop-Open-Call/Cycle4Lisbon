import { admin } from "../../support/util";

describe("login", () => {
  beforeEach(() => {
    cy.visit("/");
  });

  it("redirects to /login from the root path", () => {
    cy.location("pathname").should("eq", "/login");
  });

  it("redirects to the dashboard after a successful login", () => {
    cy.get('input[placeholder="example@email.com"]').type(admin().email);
    cy.get('input[placeholder="Password"]').type(admin().password);
    cy.get('button:contains("Continue")').click();
    cy.location("pathname").should("eq", "/dashboard");
  });

  it("displays errors from the API", () => {
    cy.get('input[placeholder="example@email.com"]').type(admin().email);
    cy.get('input[placeholder="Password"]').type("WrongPassword!");
    cy.get('button:contains("Continue")').click();
    cy.contains("Invalid username or password");
  });
});

describe("cypress login command", () => {
  it("works", () => {
    cy.login(admin());
    cy.location("pathname").should("eq", "/dashboard");
  });
});
