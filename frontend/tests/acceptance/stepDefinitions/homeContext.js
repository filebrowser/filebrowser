import { Given, When, Then } from "@cucumber/cucumber";
import { equal } from "assert";
import { format } from "util";
import { expect } from "@playwright/test";

import { HomePage } from "../PageObject/HomePage.js";
import { LoginPage } from "../PageObject/LoginPage.js";
import { createFile } from "../../helper/file_helper.js";

const login = new LoginPage();
const homepage = new HomePage();

Given("{string} has logged in", async function (role) {
  await login.loginBasedOnRole(role);
});

Given("admin has navigated to the homepage", async function () {
  await expect(global.page).toHaveURL(login.baseURL + "files/");
});

When("admin creates a new folder named {string}", async function (folderName) {
  await homepage.createNewFolder(folderName);
});

Then(
  "admin should be able to see a folder named {string}",
  async function (folderName) {
    const userCreatedFolderName = await global.page.innerHTML(
      homepage.lastNavaigatedFolderSelector
    );
    equal(
      userCreatedFolderName,
      folderName,
      `Expected "${folderName}" but recieved message "${userCreatedFolderName}" from UI`
    );
  }
);

Given(
  "admin has created a file named {string} with content {string}",
  async function (filename, content) {
    await homepage.createFileWithContent(filename, content);
    await expect(
      global.page.locator(format(homepage.fileSelector, filename))
    ).toBeVisible();
  }
);

When(
  "admin creates a new file named {string} with content {string}",
  async function (filename, content) {
    await homepage.createFileWithContent(filename, content);
  }
);

Given(
  "admin creates a new file named {string} using API",
  async function (filename) {
    await createFile(filename);
    await global.page.reload();
    await expect(
      global.page.locator(format(homepage.fileSelector, filename))
    ).toBeVisible();
  }
);

Then(
  "admin should be able to see a file named {string} with content {string}",
  async function (filename, content) {
    await expect(
      global.page.locator(format(homepage.fileSelector, filename))
    ).toBeVisible();
    await global.page.dblclick(format(homepage.fileSelector, filename));
    const fileContent = await global.page.innerHTML(homepage.editorContent);
    equal(
      fileContent,
      content,
      `Expected content as "${content}" but recieved "${fileContent}"`
    );
  }
);

When(
  "admin renames a file {string} to {string}",
  async function (oldfileName, newfileName) {
    await homepage.renameFile(oldfileName, newfileName);
  }
);

Then(
  "admin should be able to see file with {string} name",
  async function (newfileName) {
    await expect(
      global.page.locator(format(homepage.fileSelector, newfileName))
    ).toBeVisible();
  }
);

When("admin deletes a file named {string}", async function (filename) {
  await homepage.deleteFile(filename);
});

Then("admin shouln't see {string} in the UI", async function (filename) {
  await expect(
    global.page.locator(format(homepage.fileSelector, filename))
  ).toBeHidden();
});
