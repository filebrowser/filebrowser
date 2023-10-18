const {Given, When, Then} = require('@cucumber/cucumber')
// import expect for assertion
const { expect } = require("@playwright/test");

//import assert
const assert = require("assert")

//import page
const LoginPage = require("../pageObjects/LoginPage.js")
const loginPage = new LoginPage;

Given('the user has browsed to the login page', async function () {
    await loginPage.goToLoginUrl();
    await expect(page).toHaveURL(loginPage.loginUrl);
});

Given('the user has logged in with username {string} and password {string}', async function (username, password) {
    await loginPage.login(username,password);
    await expect(page).toHaveURL(loginPage.fileUrl);
});

When('user logs in with username {string} and password {string}', async function (username, password) {
    await loginPage.login(username,password);
});

Then('user should redirect to the homepage', async function () {
    await expect(page).toHaveURL(loginPage.fileUrl);
});