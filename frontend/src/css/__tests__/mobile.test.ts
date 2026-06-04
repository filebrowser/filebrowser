import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const mobileCss = readFileSync(resolve(__dirname, "../mobile.css"), "utf8");

const normalizedCss = mobileCss.replace(/\s+/g, " ");

describe("mobile file listing styles", () => {
  it("hides file row metadata without hiding list header sort controls", () => {
    expect(normalizedCss).toContain(
      "#listing.list .item:not(.header) .size { display: none;"
    );
    expect(normalizedCss).toContain(
      "#listing.list .item:not(.header) .modified { display: none;"
    );

    expect(normalizedCss).not.toMatch(
      /#listing\.list \.item \.(size|modified) \{ display: none;/
    );
  });
});
