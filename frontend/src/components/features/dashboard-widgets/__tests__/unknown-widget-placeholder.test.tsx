import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { UnknownWidgetPlaceholder } from "@/components/features/dashboard-widgets/unknown-widget-placeholder";

describe("UnknownWidgetPlaceholder", () => {
  it("renders the widget type in the body", () => {
    render(<UnknownWidgetPlaceholder type="future-widget" />);
    expect(screen.getByText("Unknown widget")).toBeInTheDocument();
    expect(screen.getByText("future-widget")).toBeInTheDocument();
  });
});
