import {
  type Locator,
  type Page,
  test as base,
  expect,
} from "@playwright/test";
import { AuthPage } from "./auth";

type SettingsType = "profile" | "shares" | "global" | "users";

export class SettingsPage {
  public readonly hideDotfiles: Locator; // checkbox
  public readonly singleClick: Locator; // checkbox
  public readonly dateFormat: Locator; // checkbox
  private readonly languages: Locator; // selection
  private readonly submitProfile: Locator; // submit
  private readonly submitPassword: Locator; // submit

  constructor(public readonly page: Page) {
    this.hideDotfiles = this.page.locator('input[name="hideDotfiles"]');
    this.singleClick = this.page.locator('input[name="singleClick"]');
    this.dateFormat = this.page.locator('input[name="dateFormat"]');
    this.languages = this.page.locator('select[name="selectLanguage"]');
    this.submitProfile = this.page.locator('input[name="submitProfile"]');
    this.submitPassword = this.page.locator('input[name="submitPassword"]');
  }

  async goto(type: SettingsType = "profile") {
    await this.page.goto(`/settings/${type}`);
  }

  async setLanguage(locale: string = "en") {
    await this.languages.selectOption(locale);
  }

  async saveProfile() {
    await this.submitProfile.click();
  }

  async savePassword() {
    await this.submitPassword.click();
  }
}

const test = base.extend<{ settingsPage: SettingsPage }>({
  page: async ({ page }, use) => {
    // Sign in with our account.
    const authPage = new AuthPage(page);
    await authPage.goto();
    await authPage.loginAs();
    await expect(page).toHaveTitle(/.*Files - File Browser$/);
    // Use signed-in page in the test.
    await use(page);
  },
  settingsPage: async ({ page }, use) => {
    const settingsPage = new SettingsPage(page);
    await use(settingsPage);
  },
});

export { test, expect };
