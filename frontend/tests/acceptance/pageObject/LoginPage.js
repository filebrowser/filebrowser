export class LoginPage {
  constructor() {
    this.usernameSelector = '//input[@type="text"]';
    this.passwordSelector = '//input[@type="password"]';
    this.loginButton = '//input[@type="submit"]';
    this.wrongCredentialsDivSelector = '//div[@class="wrong"]';
    this.baseURL = "http://localhost:8080/";
  }

  async navigateToLoginPage() {
    await global.page.goto(this.baseURL + "login");
  }

  async loginWithUsernameAndPassword(username, password) {
    await global.page.fill(this.usernameSelector, username);
    await global.page.fill(this.passwordSelector, password);
    await global.page.click(this.loginButton);
  }

  async loginBasedOnRole(role) {
    this.navigateToLoginPage();
    switch (role) {
      case "admin":
        await this.loginWithUsernameAndPassword("admin", "admin");
        break;
      case "user":
        await this.loginWithUsernameAndPassword("user", "user");
        break;
      default:
        throw new Error(`Invalid role ${role} passed`);
    }
  }
}
