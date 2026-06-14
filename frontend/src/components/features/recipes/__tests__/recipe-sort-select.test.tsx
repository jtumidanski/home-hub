import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RecipeSortSelect } from "../recipe-sort-select";

describe("RecipeSortSelect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls onChange with -usageCount when 'Most cooked' is picked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RecipeSortSelect value={undefined} onChange={onChange} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /most cooked/i }));

    expect(onChange).toHaveBeenCalledWith("-usageCount");
  });

  it("calls onChange with usageCount when 'Least cooked' is picked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RecipeSortSelect value={undefined} onChange={onChange} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /least cooked/i }));

    expect(onChange).toHaveBeenCalledWith("usageCount");
  });

  it("calls onChange with undefined when 'Default' is picked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RecipeSortSelect value="-usageCount" onChange={onChange} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /default/i }));

    expect(onChange).toHaveBeenCalledWith(undefined);
  });
});
