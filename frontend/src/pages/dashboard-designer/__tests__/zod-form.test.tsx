import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { ZodFormFields } from "../zod-form";

const schema = z.object({ done: z.boolean(), title: z.string() });
type Values = z.infer<typeof schema>;

function Harness() {
  const form = useForm<Values>({ defaultValues: { done: false, title: "" } });
  return <ZodFormFields schema={schema} form={form} />;
}

describe("ZodFormFields", () => {
  it("renders booleans as an accessible checkbox that toggles", () => {
    render(<Harness />);
    const box = screen.getByRole("checkbox", { name: "Done" });
    expect(box).toHaveAttribute("aria-checked", "false");
    fireEvent.click(box);
    expect(screen.getByRole("checkbox", { name: "Done" })).toHaveAttribute("aria-checked", "true");
  });

  it("renders strings with the shared Input primitive", () => {
    render(<Harness />);
    expect(screen.getByTestId("zod-field-title")).toHaveAttribute("data-slot", "input");
  });

  it("associates labels with text inputs and radio groups (DA-052)", () => {
    const rich = z.object({ title: z.string(), mode: z.enum(["a", "b"]) });
    function RichHarness() {
      const form = useForm<z.infer<typeof rich>>({ defaultValues: { title: "", mode: "a" } });
      return <ZodFormFields schema={rich} form={form} />;
    }
    render(<RichHarness />);
    expect(screen.getByRole("textbox", { name: "Title" })).toBe(screen.getByTestId("zod-field-title"));
    expect(screen.getByRole("radiogroup", { name: "Mode" })).toBeInTheDocument();
  });
});
