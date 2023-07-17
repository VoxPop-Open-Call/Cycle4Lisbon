import { faker } from "@faker-js/faker";
import { admin } from "./util";

Cypress.Commands.add("login", ({ email, password }) => {
  cy.visit("/");
  cy.get('input[placeholder="example@email.com"]').type(email);
  cy.get('input[placeholder="Password"]').type(password);
  cy.get('button:contains("Continue")').click();
});

Cypress.Commands.add("apiLogin", ({ email, password }) => {
  return cy.request({
    method: "POST",
    url: `${Cypress.env("dexUrl")}/token`,
    body: {
      username: email,
      password,
      grant_type: "password",
      client_id: Cypress.env("apiClientId"),
      client_secret: Cypress.env("apiClientSecret"),
      scope: "openid profile email offline_access",
    },
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
  });
});

Cypress.Commands.add("authorizedRequest", (options, credentials) => {
  return cy.apiLogin(credentials).then((res) => {
    cy.expect(res.status).to.eq(200);
    return cy.request({
      ...options,
      headers: {
        Authorization: `bearer ${res.body.access_token}`,
      },
    });
  });
});

Cypress.Commands.add("createUser", (user) => {
  return cy.request("POST", `${Cypress.env("apiUrl")}/users`, user);
});

Cypress.Commands.add("updateUser", (user) => {
  return cy.authorizedRequest(
    {
      method: "PUT",
      url: `${Cypress.env("apiUrl")}/users/${user.id}`,
      body: user,
    },
    {
      email: user.email,
      password: user.password,
    }
  );
});

Cypress.Commands.add("createRandomUser", () => {
  const firstName = faker.person.firstName();
  const lastName = faker.person.lastName();
  const user = {
    email: faker.internet.email({ firstName, lastName }),
    name: firstName + " " + lastName,
    password: faker.internet.password(),
  };

  return cy.createUser(user).then((res) => {
    cy.expect(res.status).to.eq(201);
    return cy.updateUser({
      ...res.body,
      password: user.password,
      username: faker.internet.displayName({ firstName, lastName }),
    });
  });
});

Cypress.Commands.add("seedUsers", () => {
  return cy
    .authorizedRequest(
      {
        method: "GET",
        url: `${Cypress.env("apiUrl")}/users`,
      },
      admin()
    )
    .then((res) => {
      cy.expect(res.status).to.eq(200);

      // Create users if the database is sparsely populated.
      if (res.body.length < 20) {
        for (let i = 0; i < 20; i++) {
          cy.createRandomUser().then((res) => {
            cy.expect(res.status).to.eq(200);
          });
        }
      } else {
        cy.log("Skipping creation of users");
      }
    });
});
