import { test, expect } from "@playwright/test";

test("redirect to login", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveURL(/\/login/);

  await page.goto("/files/");
  await expect(page).toHaveURL(/\/login\?redirect=\/files\//);
});

test("login and logout", async ({ page, context }) => {
  await page.goto("/login");
  await expect(page).toHaveTitle("Login - File Browser");

  await page.getByRole("button", { name: "Login" }).click();
  await expect(
    page.getByText("Wrong credentials", { exact: true })
  ).toBeVisible();

  await page.getByPlaceholder("Username").fill("admin");
  await page.getByPlaceholder("Password").fill("admin");
  await page.getByRole("button", { name: "Login" }).click();
  await page.waitForURL("**/files/", { timeout: 5000 });
  await expect(page).toHaveTitle(/.*Files - File Browser/);

  let cookies = await context.cookies();
  await expect(cookies.find((c) => c.name == "auth")?.value).toBeDefined();

  await page.getByRole("button", { name: "Logout" }).click();
  await page.waitForURL("**/login", { timeout: 5000 });
  await expect(page).toHaveTitle("Login - File Browser");

  cookies = await context.cookies();
  await expect(cookies.find((c) => c.name == "auth")?.value).toBeUndefined();
});
