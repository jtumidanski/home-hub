import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import * as colorContrast from "@/lib/color-contrast";
import { UserAvatar } from "../user-avatar";

describe("UserAvatar", () => {
  const baseProps = {
    displayName: "Jane Doe",
    userId: "550e8400-e29b-41d4-a716-446655440000",
  };

  describe("DiceBear avatar", () => {
    it("renders DiceBear SVG when avatarUrl starts with dicebear:", () => {
      render(
        <UserAvatar
          {...baseProps}
          avatarUrl="dicebear:adventurer:seed123"
        />,
      );
      const img = screen.getByRole("img", { name: "Jane Doe" });
      expect(img).toHaveAttribute(
        "src",
        "https://api.dicebear.com/9.x/adventurer/svg?seed=seed123",
      );
    });
  });

  describe("provider image", () => {
    it("renders provider image when no user-selected avatar", () => {
      render(
        <UserAvatar
          {...baseProps}
          providerAvatarUrl="https://google.com/photo.jpg"
        />,
      );
      const img = screen.getByRole("img", { name: "Jane Doe" });
      expect(img).toHaveAttribute("src", "https://google.com/photo.jpg");
    });

    it("falls back to initials on image error", () => {
      render(
        <UserAvatar
          {...baseProps}
          providerAvatarUrl="https://broken.com/photo.jpg"
        />,
      );
      const img = screen.getByRole("img", { name: "Jane Doe" });
      fireEvent.error(img);
      expect(screen.getByText("JD")).toBeInTheDocument();
    });
  });

  describe("initials fallback", () => {
    it("renders initials when no avatar URLs provided", () => {
      render(<UserAvatar {...baseProps} />);
      expect(screen.getByText("JD")).toBeInTheDocument();
    });

    it("uses first two chars for single-word name", () => {
      render(<UserAvatar {...baseProps} displayName="Jane" />);
      expect(screen.getByText("JA")).toBeInTheDocument();
    });

    it("has deterministic color from userId", () => {
      const { container } = render(<UserAvatar {...baseProps} />);
      const div = container.querySelector("[role='img']");
      expect(div).toHaveStyle({ backgroundColor: expect.stringContaining("hsl(") });
    });
  });

  describe("sizes", () => {
    it("applies sm size class", () => {
      render(<UserAvatar {...baseProps} size="sm" />);
      const el = screen.getByRole("img", { name: "Jane Doe" });
      expect(el.className).toContain("h-8");
    });

    it("applies md size class by default", () => {
      render(<UserAvatar {...baseProps} />);
      const el = screen.getByRole("img", { name: "Jane Doe" });
      expect(el.className).toContain("h-10");
    });

    it("applies lg size class", () => {
      render(<UserAvatar {...baseProps} size="lg" />);
      const el = screen.getByRole("img", { name: "Jane Doe" });
      expect(el.className).toContain("h-20");
    });
  });

  describe("priority", () => {
    it("prefers user-selected avatar over provider", () => {
      render(
        <UserAvatar
          {...baseProps}
          avatarUrl="dicebear:bottts:myseed"
          providerAvatarUrl="https://google.com/photo.jpg"
        />,
      );
      const img = screen.getByRole("img", { name: "Jane Doe" });
      expect(img).toHaveAttribute(
        "src",
        "https://api.dicebear.com/9.x/bottts/svg?seed=myseed",
      );
    });
  });

  describe("foreground colour", () => {
    it("uses the ink token on a light generated colour and never text-white", () => {
      const { container } = render(<UserAvatar {...baseProps} />);
      const div = container.querySelector("[role='img']") as HTMLElement;
      expect(div.style.color).toBe("var(--on-color-ink)");
      expect(div.className).not.toContain("text-white");
    });

    it("uses the paper token when the readable foreground is paper", () => {
      const spy = vi.spyOn(colorContrast, "readableForeground").mockReturnValue("paper");
      const { container } = render(<UserAvatar {...baseProps} />);
      const div = container.querySelector("[role='img']") as HTMLElement;
      expect(div.style.color).toBe("var(--on-color-paper)");
      spy.mockRestore();
    });
  });
});
