import {
  Before,
  BeforeAll,
  AfterAll,
  After,
  setDefaultTimeout,
} from "@cucumber/cucumber";
import { chromium } from "@playwright/test";
import { cleanUpTempFiles } from "./helper/file_helper.js";

setDefaultTimeout(60000);

BeforeAll(async function () {
  global.browser = await chromium.launch({
    headless: true,
  });
});

AfterAll(async function () {
  await global.browser.close();
});

Before(async function () {
  global.context = await global.browser.newContext();
  global.page = await global.context.newPage();
});

After(async function () {
  await global.page.close();
  await global.context.close();
  await cleanUpTempFiles();
});
