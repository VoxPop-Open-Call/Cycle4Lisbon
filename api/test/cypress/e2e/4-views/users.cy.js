import users from "../../fixtures/users.json";
import { admin } from "../../support/util";

const mockProfilepic =
  "https://upload.wikimedia.org/wikipedia/en/a/a4/Hide_the_Pain_Harold_%28Andr%C3%A1s_Arat%C3%B3%29.jpg";

describe("user page", () => {
  before(cy.seedUsers);

  it("renders", () => {
    cy.intercept("GET", "**/users*").as("listUsers");
    cy.intercept("GET", "**/users/*/picture-get-url", {
      url: mockProfilepic,
      method: "GET",
    }).as("getProfilepic");

    cy.login(admin());
    cy.get('a:contains("Users")').click();

    cy.wait("@listUsers");

    cy.contains("No items found").should("not.exist");
    cy.get("tbody")
      .find("tr")
      .eq(0)
      .find('div:contains("Show")')
      .click({ force: true });

    cy.location("pathname").should("match", /\/user\/.*/);
    cy.wait("@getProfilepic");

    cy.contains("Total Rides");
    cy.contains("Total Distance");
  });

  it("contains user data", () => {
    cy.intercept("GET", "**/users?limit=10&offset=0*", users.slice(0, 10)).as(
      "listUsers"
    );
    cy.intercept("GET", `**/users/${users[0].id}`, users[0]).as("getUser");
    cy.intercept("GET", "**/users/*/picture-get-url", {
      url: mockProfilepic,
      method: "GET",
    }).as("getProfilepic");

    cy.login(admin());
    cy.get('a:contains("Users")').click();

    cy.wait("@listUsers");
    cy.contains("No items found").should("not.exist");

    cy.get("tbody")
      .find("tr")
      .eq(0)
      .find('div:contains("Show")')
      .click({ force: true });
    cy.location("pathname").should("match", /\/user\/.*/);
    cy.wait("@getUser");
    cy.wait("@getProfilepic");

    cy.contains(users[0].name);
    cy.contains(users[0].username);
    cy.contains(users[0].email);
    cy.contains(users[0].tripCount);
    cy.contains(users[0].totalDist);
    cy.contains(users[0].credits);
    cy.get("img").invoke("attr", "src").should("eq", mockProfilepic);
  });
});
