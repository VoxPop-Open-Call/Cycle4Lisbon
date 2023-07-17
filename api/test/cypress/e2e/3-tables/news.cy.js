import moment from "moment";
import news from "../../fixtures/news";
import { admin } from "../../support/util";

describe("news table", () => {
  it("renders", () => {
    cy.intercept("GET", "**/external*", {
      fixture: "news.json",
    }).as("listNews");

    cy.login(admin());
    cy.get('a:contains("News")').click();

    cy.wait("@listNews");
    cy.contains("No items found").should("not.exist");

    cy.get('div:contains("Title")');
    cy.get('div:contains("Date")');
    cy.get('div:contains("Status")');

    cy.get("select").select("5");
    cy.get("table").get("tbody").find("tr").should("have.length", 5);
  });

  it('shows "No items found" message', () => {
    cy.intercept("GET", "**/external*", []).as("listNews");

    cy.login(admin());
    cy.get('a:contains("News")').click();

    cy.wait("@listNews");
    cy.contains("No items found");
  });

  it("lists news in response", () => {
    cy.intercept(
      "GET",
      "**/external?limit=10&offset=0*&type=news",
      news.slice(0, 10)
    ).as("listNews");
    cy.intercept(
      "GET",
      "**/external?limit=10&offset=10*&type=news",
      news.slice(10, 20)
    ).as("listNews");

    cy.login(admin());
    cy.get('a:contains("News")').click();

    cy.wait("@listNews");
    cy.contains("No items found").should("not.exist");

    for (let i = 0; i < 10; i++) {
      cy.get("tbody").find("tr").eq(i).contains(news[i].title);
      cy.get("tbody")
        .find("tr")
        .eq(i)
        .contains(moment(news[i].date).format("MMMM DD, YYYY"));
      cy.get("tbody").find("tr").eq(i).contains("Pending");
    }

    // page 2
    cy.get("table").get('.page-link:contains("2")').click();

    for (let i = 0; i < 10; i++) {
      cy.get("tbody").find("tr").eq(i).contains(news[i].title);
      cy.get("tbody")
        .find("tr")
        .eq(i)
        .contains(moment(news[i].date).format("MMMM DD, YYYY"));
      cy.get("tbody").find("tr").eq(i).contains("Pending");
    }
  });
});
