import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Input } from "../input";
import { Textarea } from "../textarea";
import { Label } from "../label";
import { Button } from "../button";
import { RadioGroup, RadioGroupItem } from "../radio-group";

describe("Input", () => {
  it("has a coarse-pointer target and default size", () => {
    render(<Input aria-label="a" />);
    const cls = screen.getByLabelText("a").className;
    expect(cls).toContain("coarse:min-h-11");
    expect(cls).toContain("h-8");
  });

  it("sm size is compact on fine pointers and keeps 16px text on coarse", () => {
    render(<Input aria-label="a" size="sm" />);
    const cls = screen.getByLabelText("a").className;
    expect(cls).toContain("h-7");
    expect(cls).toContain("coarse:text-base");
    expect(cls).toContain("coarse:min-h-11");
  });
});

describe("Textarea", () => {
  it("follows the Input recipe without a shadow", () => {
    render(<Textarea aria-label="t" />);
    const cls = screen.getByLabelText("t").className;
    expect(cls).toContain("aria-invalid:border-destructive");
    expect(cls).toContain("dark:bg-input/30");
    expect(cls).toContain("coarse:min-h-11");
    expect(cls).not.toContain("shadow-sm");
  });
});

describe("Label", () => {
  it("dims disabled text via colour token, not opacity", () => {
    render(<Label>x</Label>);
    const cls = screen.getByText("x").className;
    expect(cls).toContain("peer-disabled:text-muted-foreground");
    expect(cls).not.toContain("opacity");
  });
});

describe("Button destructive", () => {
  it("is a solid fill with the foreground token", () => {
    render(<Button variant="destructive">Delete</Button>);
    const cls = screen.getByRole("button").className;
    expect(cls).toContain("bg-destructive");
    expect(cls).toContain("text-destructive-foreground");
    expect(cls).not.toContain("bg-destructive/10");
  });
});

describe("RadioGroupItem", () => {
  it("renders a token-drawn 20px radio with a coarse row", () => {
    render(
      <RadioGroup value="a" onValueChange={() => {}}>
        <RadioGroupItem value="a">Alpha</RadioGroupItem>
      </RadioGroup>
    );
    const radio = screen.getByRole("radio", { name: "Alpha" });
    expect(radio.className).toContain("size-5");
    expect(radio.className).toContain("appearance-none");
    expect(radio.className).toContain("border-input");
    expect(radio.className).toContain("checked:border-primary");
    expect(radio.className).not.toContain("accent-primary");
    expect(radio.closest("label")!.className).toContain("coarse:min-h-11");
  });

  it("selects an item when its label is clicked", () => {
    const Harness = () => {
      const [v, setV] = useState("a");
      return (
        <RadioGroup value={v} onValueChange={setV}>
          <RadioGroupItem value="a">Alpha</RadioGroupItem>
          <RadioGroupItem value="b">Beta</RadioGroupItem>
        </RadioGroup>
      );
    };
    render(<Harness />);
    expect(screen.getByRole("radio", { name: "Beta" })).not.toBeChecked();
    fireEvent.click(screen.getByText("Beta"));
    expect(screen.getByRole("radio", { name: "Beta" })).toBeChecked();
    expect(screen.getByRole("radio", { name: "Alpha" })).not.toBeChecked();
  });
});
