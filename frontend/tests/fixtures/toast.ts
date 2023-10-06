//classes: Vue-Toastification__toast Vue-Toastification__toast--success bottom-center
import { type Page, type Locator, expect } from "@playwright/test";

export class Toast {
  private readonly success: Locator;
  private readonly error: Locator;

  constructor(public readonly page: Page) {
    this.success = this.page.locator("div.Vue-Toastification__toast--success");
    this.error = this.page.locator("div.Vue-Toastification__toast--error");
  }

  async isSuccess() {
    await expect(this.success).toBeVisible();
  }

  async isError() {
    await expect(this.error).toBeVisible();
  }
}
