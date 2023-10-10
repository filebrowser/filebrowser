const {Given, When, Then} = require('@cucumber/cucumber')
// import expect for assertion
const { expect } = require("@playwright/test");

//launch url
const url = 'http://localhost:8080'

//define selectors
const usernameSelector = '//input[@placeholder="Username"]'
const passwordSelector = '//input[@placeholder="Password"]'
const loginSelector = '//input[@type="submit"]'
const loginUrl = 'http://localhost:8080/login?redirect=%2Ffiles%2F'
const fileUrl = 'http://localhost:8080/files/'

Given('the user has browsed to the login page', async function () {
    await page.goto(url);
    await expect(page).toHaveURL(loginUrl);
});

When('user logs in with username {string} and password {string}', async function (username, password) {
    await page.fill(usernameSelector, username);
    await page.fill(passwordSelector, password);
    await page.click(loginSelector);
});

Then('user should redirect to the homepage', async function () {
    await expect(page).toHaveURL(fileUrl);
});