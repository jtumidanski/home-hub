import { describe, it, expect } from "vitest";
import { toasterStyle, toastOptions } from "../toast-options";

describe("toast options", () => {
  it("themes the base surface via Sonner variables referencing tokens", () => {
    expect(toasterStyle).toMatchObject({
      "--normal-bg": "var(--popover)",
      "--normal-text": "var(--popover-foreground)",
      "--normal-border": "var(--border)",
    });
  });

  it("uses important per-type borders and a 24px close button", () => {
    const c = toastOptions.classNames;
    expect(c.success).toBe("border-success!");
    expect(c.error).toBe("border-destructive!");
    expect(c.warning).toBe("border-warning!");
    expect(c.info).toBe("border-info!");
    expect(c.closeButton).toBe("size-6!");
  });
});
