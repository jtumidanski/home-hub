import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { WeekSelector } from "../week-selector";

describe("WeekSelector", () => {
  it("names the week navigation controls", async () => {
    const onWeekChange = vi.fn();
    render(<WeekSelector startsOn={new Date(2026, 2, 2)} onWeekChange={onWeekChange} />);

    expect(screen.getByRole("button", { name: /^Choose week, Mar 2/ })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Next week" }));
    expect(onWeekChange).toHaveBeenCalledWith(new Date(2026, 2, 9));
    await userEvent.click(screen.getByRole("button", { name: "Previous week" }));
    expect(onWeekChange).toHaveBeenLastCalledWith(new Date(2026, 1, 23));
  });
});
