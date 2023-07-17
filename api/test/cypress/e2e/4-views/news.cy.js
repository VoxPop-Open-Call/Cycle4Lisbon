import news from "../../fixtures/news.json";
import { admin } from "../../support/util";

describe("news page", () => {
  it("renders", () => {
    cy.intercept("GET", "**/external*", { fixture: "news.json" }).as(
      "listNews"
    );

    cy.login(admin());
    cy.get('a:contains("News")').click();

    cy.wait("@listNews");

    cy.contains("No items found").should("not.exist");
    cy.get("tbody")
      .find("tr")
      .eq(0)
      .find('div:contains("Show")')
      .click({ force: true });

    cy.location("pathname").should("match", /\/news\/.*/);
  });

  it("contains news data", () => {
    cy.intercept(
      "GET",
      "**/external?limit=10&offset=0*&type=news",
      news.slice(0, 10)
    ).as("listNews");

    cy.login(admin());
    cy.get('a:contains("News")').click();

    cy.wait("@listNews");
    cy.contains("No items found").should("not.exist");

    cy.get("tbody")
      .find("tr")
      .eq(0)
      .find('div:contains("Show")')
      .click({ force: true });
    cy.location("pathname").should("match", /\/news\/.*/);

    cy.contains("Approve");
    cy.contains("Reject");

    cy.contains(news[0].title);
    cy.contains(news[0].subtitle);
    cy.contains(news[0].description.replace("\n\n", " "));
    cy.get("img").invoke("attr", "src").should("eq", news[0].imageUrl);

    cy.contains("Source");
    cy.get('a:contains("click here")')
      .invoke("attr", "href")
      .should("eq", news[0].articleUrl);
  });
});
