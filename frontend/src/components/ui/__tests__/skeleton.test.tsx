import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Skeleton } from "../skeleton";

describe("Skeleton", () => {
  it("pulses but stops under reduced motion", () => {
    render(<Skeleton role="status" aria-label="Loading" />);
    const el = screen.getByRole("status");
    expect(el).toHaveClass("animate-pulse");
    expect(el).toHaveClass("motion-reduce:animate-none");
  });
});
