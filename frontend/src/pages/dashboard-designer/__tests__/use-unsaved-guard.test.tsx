import { describe, it, expect, vi, beforeEach } from "vitest";
import { render } from "@testing-library/react";
import { useUnsavedGuard } from "@/pages/dashboard-designer/use-unsaved-guard";

function Probe({ dirty }: { dirty: boolean }) {
  useUnsavedGuard(dirty);
  return <div data-testid="probe" />;
}

beforeEach(() => {
  vi.restoreAllMocks();
});

describe("useUnsavedGuard beforeunload listener", () => {
  it("adds a beforeunload listener when dirty is true", () => {
    const addSpy = vi.spyOn(window, "addEventListener");
    render(<Probe dirty />);
    const calls = addSpy.mock.calls.filter(([event]) => event === "beforeunload");
    expect(calls.length).toBeGreaterThanOrEqual(1);
  });

  it("removes the beforeunload listener once dirty returns to false", () => {
    const removeSpy = vi.spyOn(window, "removeEventListener");
    const { rerender } = render(<Probe dirty />);
    rerender(<Probe dirty={false} />);
    const calls = removeSpy.mock.calls.filter(([event]) => event === "beforeunload");
    expect(calls.length).toBeGreaterThanOrEqual(1);
  });

  it("does not add a beforeunload listener when dirty starts false", () => {
    const addSpy = vi.spyOn(window, "addEventListener");
    render(<Probe dirty={false} />);
    const calls = addSpy.mock.calls.filter(([event]) => event === "beforeunload");
    expect(calls.length).toBe(0);
  });
});
