import users from "../../fixtures/users.json";
import { admin } from "../../support/util";

describe("users table", () => {
  before(cy.seedUsers);

  it("renders", () => {
    cy.intercept("GET", "**/users*").as("listUsers");

    cy.login(admin());
    cy.get('a:contains("Users")').click();

    cy.wait("@listUsers");
    cy.contains("No items found").should("not.exist");

    cy.get('div:contains("Name")');
    cy.get('div:contains("Email")');
    cy.get('div:contains("Verified")');

    cy.get("select").select("5");
    cy.get("table").get("tbody").find("tr").should("have.length", 5);
  });

  it('shows "No items found" message', () => {
    cy.intercept("GET", "**/users*", []).as("listUsers");

    cy.login(admin());
    cy.get('a:contains("Users")').click();

    cy.wait("@listUsers");
    cy.contains("No items found");
  });

  it("lists users in response", () => {
    cy.intercept("GET", "**/users?limit=10&offset=0*", users.slice(0, 10)).as(
      "listUsers"
    );
    cy.intercept("GET", "**/users?limit=10&offset=10*", users.slice(10, 20)).as(
      "listUsers"
    );

    cy.login(admin());
    cy.get('a:contains("Users")').click();

    cy.wait("@listUsers");
    cy.contains("No items found").should("not.exist");

    for (let i = 0; i < 10; i++) {
      cy.get("tbody").find("tr").eq(i).contains(users[i].name);
      cy.get("tbody").find("tr").eq(i).contains(users[i].email);
      cy.get("tbody")
        .find("tr")
        .eq(i)
        .contains(users[i].verified ? "Verified" : "Not Verified");
    }

    // page 2
    cy.get("table").get('.page-link:contains("2")').click();

    for (let i = 0; i < 10; i++) {
      cy.get("tbody")
        .find("tr")
        .eq(i)
        .contains(users[i + 10].name);
      cy.get("tbody")
        .find("tr")
        .eq(i)
        .contains(users[i + 10].email);
      cy.get("tbody")
        .find("tr")
        .eq(i)
        .contains(users[i + 10].verified ? "Verified" : "Not Verified");
    }
  });
});
