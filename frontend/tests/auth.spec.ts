import { test as base, expect } from "@playwright/test";
import { AuthPage } from "./auth-page";

const test = base.extend<{ authPage: AuthPage }>({
  authPage: async ({ page }, use) => {
    const authPage = new AuthPage(page);
    // await authPage.goto();
    // await authPage.loginAs();
    await use(authPage);
    // await authPage.logout();
  },
});

test("redirect to login", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveURL(/\/login/);

  await page.goto("/files/");
  await expect(page).toHaveURL(/\/login\?redirect=\/files\//);
});

test("login and logout", async ({ authPage, page, context }) => {
  await authPage.goto();
  await expect(page).toHaveTitle(/Login - File Browser$/);

  await authPage.loginAs("fake", "fake");
  await expect(authPage.wrongCredentials).toBeVisible();

  await authPage.loginAs();
  await expect(authPage.wrongCredentials).toBeHidden();
  await page.waitForURL("**/files/", { timeout: 5000 });
  await expect(page).toHaveTitle(/.*Files - File Browser$/);

  let cookies = await context.cookies();
  expect(cookies.find((c) => c.name == "auth")?.value).toBeDefined();

  await authPage.logout();
  await page.waitForURL("**/login", { timeout: 5000 });
  await expect(page).toHaveTitle(/Login - File Browser$/);

  cookies = await context.cookies();
  expect(cookies.find((c) => c.name == "auth")?.value).toBeUndefined();
});
