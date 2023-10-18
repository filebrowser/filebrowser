class LoginPage {
    constructor() {
        //url
        this.url = 'http://localhost:8080'
        this.loginUrl = 'http://localhost:8080/login?redirect=%2Ffiles%2F'
        this.fileUrl = this.url + '/files/'

        //define selectors
        this.usernameSelector = '//input[@placeholder="Username"]'
        this.passwordSelector = '//input[@placeholder="Password"]'
        this.loginSelector = '//input[@type="submit"]'
    }

    async goToLoginUrl() {
        await page.goto(this.url);
    }

    async login(username, password) {
        await page.fill(this.usernameSelector, username);
        await page.fill(this.passwordSelector, password);
        await page.click(this.loginSelector);
    }

    // async assertLoginPageIsOpen() {
    //     await expect(this.page).toHaveURL(this.loginUrl);
    //   }
}

module.exports = LoginPage