import { Given, When, Then } from "@cucumber/cucumber";
import { expect } from "@playwright/test";
import { equal } from "assert";

import { LoginPage } from "../PageObject/LoginPage.js";

const login = new LoginPage();

Given("admin has browsed to the login page", async () => {
  await login.navigateToLoginPage();
  await expect(global.page).toHaveURL(login.baseURL + "login");
});

When(
  "admin logs in with username as {string} and password as {string}",
  async (username, password) => {
    await login.loginWithUsernameAndPassword(username, password);
  }
);

Then("admin should be navigated to homescreen", async function () {
  await expect(global.page).toHaveURL(login.baseURL + "files/");
});

Then("admin should see {string} message", async function (expectedMessage) {
  const errorMessage = await global.page.innerHTML(
    login.wrongCredentialsDivSelector
  );
  equal(
    errorMessage,
    expectedMessage,
    `Expected message string "${expectedMessage}" but received message "${errorMessage}" from UI`
  );
});
