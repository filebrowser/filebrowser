import { test, expect } from "./fixtures/settings";
import { Toast } from "./fixtures/toast";

// test.describe("profile settings", () => {
test("settings button", async ({ page }) => {
  const button = page.getByLabel("Settings", { exact: true });
  await expect(button).toBeVisible();
  await button.click();
  await expect(page).toHaveTitle(/^Profile Settings/);
  await expect(
    page.getByRole("heading", { name: "Profile Settings" })
  ).toBeVisible();
});

test("set locale", async ({ settingsPage, page }) => {
  const toast = new Toast(page);

  await settingsPage.goto("profile");
  await expect(page).toHaveTitle(/^Profile Settings/);
  // await settingsPage.saveProfile();
  // await toast.isSuccess();
  // await expect(
  //   page.getByText("Settings updated!", { exact: true })
  // ).toBeVisible();

  await settingsPage.setLanguage("hu");
  await settingsPage.saveProfile();
  await toast.isSuccess();
  await expect(
    page.getByText("Beállítások frissítve!", { exact: true })
  ).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Profilbeállítások" })
  ).toBeVisible();

  await settingsPage.setLanguage("en");
  await settingsPage.saveProfile();
  await toast.isSuccess();
  await expect(
    page.getByText("Settings updated!", { exact: true })
  ).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Profile Settings" })
  ).toBeVisible();
});
// });
