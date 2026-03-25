import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CardActionMenu } from "../card-action-menu";

const makeActions = (overrides?: { onClick?: () => void; onDelete?: () => void }) => [
  {
    icon: <span data-testid="edit-icon" />,
    label: "Edit",
    onClick: overrides?.onClick ?? vi.fn(),
  },
  {
    icon: <span data-testid="delete-icon" />,
    label: "Delete",
    onClick: overrides?.onDelete ?? vi.fn(),
    variant: "destructive" as const,
  },
];

describe("CardActionMenu", () => {
  it("opens dropdown menu when action button is clicked", async () => {
    const user = userEvent.setup();
    render(<CardActionMenu actions={makeActions()} />);

    expect(screen.queryByText("Edit")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Actions" }));

    expect(screen.getByText("Edit")).toBeInTheDocument();
    expect(screen.getByText("Delete")).toBeInTheDocument();
  });

  it("renders all action items", async () => {
    const user = userEvent.setup();
    const actions = [
      { icon: <span />, label: "Edit", onClick: vi.fn() },
      { icon: <span />, label: "Delete", onClick: vi.fn(), variant: "destructive" as const },
      { icon: <span />, label: "Archive", onClick: vi.fn() },
    ];
    render(<CardActionMenu actions={actions} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));

    expect(screen.getByText("Edit")).toBeInTheDocument();
    expect(screen.getByText("Delete")).toBeInTheDocument();
    expect(screen.getByText("Archive")).toBeInTheDocument();
  });

  it("calls action callback and closes menu when action item is clicked", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(<CardActionMenu actions={makeActions({ onClick })} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));
    await user.click(screen.getByText("Edit"));

    expect(onClick).toHaveBeenCalledOnce();
    expect(screen.queryByText("Edit")).not.toBeInTheDocument();
  });

  it("closes menu when clicking outside", async () => {
    const user = userEvent.setup();
    render(
      <div>
        <span>Outside</span>
        <CardActionMenu actions={makeActions()} />
      </div>,
    );

    await user.click(screen.getByRole("button", { name: "Actions" }));
    expect(screen.getByText("Edit")).toBeInTheDocument();

    await user.click(screen.getByText("Outside"));

    expect(screen.queryByText("Edit")).not.toBeInTheDocument();
  });
});
