import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Dialog, DialogContent, DialogTitle, DialogDescription } from "../dialog";

describe("Dialog", () => {
  it("restyles backdrop, popup and title onto Kitchen Table tokens", () => {
    render(
      <Dialog open>
        <DialogContent>
          <DialogTitle>Heading</DialogTitle>
          <DialogDescription>Body</DialogDescription>
        </DialogContent>
      </Dialog>,
    );
    const overlay = document.querySelector("[data-slot='dialog-overlay']")!;
    expect(overlay.className).toContain("bg-overlay");
    expect(overlay.className).not.toContain("bg-black/30");
    expect(screen.getByRole("dialog").className).toContain("shadow-overlay");
    const title = screen.getByText("Heading").className;
    expect(title).toContain("font-heading");
    expect(title).toContain("text-lg");
    expect(title).toContain("leading-tight");
    expect(title).toContain("pr-8");
    const popup = screen.getByRole("dialog").className;
    expect(popup).toContain("max-h-[calc(100dvh-2rem)]");
    expect(popup).toContain("overflow-y-auto");
  });
});
