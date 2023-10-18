const {Given, When, Then} = require('@cucumber/cucumber')
// import expect for assertion
const { expect } = require("@playwright/test");

//import assert
const assert = require("assert")

//import page
const CreateFilePage = require("../pageObjects/CreateFilePage.js");
const createFilePage = new CreateFilePage;

When('user has added file {string} with content {string}', async function (filename,content) {
    await createFilePage.createNewFile(filename,content)
});

Then('for user there should contain files {string}', async function (string) {

});