import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
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

    expect(await screen.findByText("Edit")).toBeInTheDocument();
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

    expect(await screen.findByText("Edit")).toBeInTheDocument();
    expect(screen.getByText("Delete")).toBeInTheDocument();
    expect(screen.getByText("Archive")).toBeInTheDocument();
  });

  it("calls action callback and closes menu when action item is clicked", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(<CardActionMenu actions={makeActions({ onClick })} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));
    await user.click(await screen.findByText("Edit"));

    expect(onClick).toHaveBeenCalledOnce();
    await waitFor(() => expect(screen.queryByText("Edit")).not.toBeInTheDocument());
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
    expect(await screen.findByText("Edit")).toBeInTheDocument();

    await user.click(screen.getByText("Outside"));

    await waitFor(() => expect(screen.queryByText("Edit")).not.toBeInTheDocument());
  });

  it("exposes menu and menuitem roles", async () => {
    const user = userEvent.setup();
    render(<CardActionMenu actions={makeActions()} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));

    expect(await screen.findByRole("menu")).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Edit" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Delete" })).toBeInTheDocument();
  });

  it("moves focus into the menu, supports arrow keys, and closes on Escape returning focus", async () => {
    const user = userEvent.setup();
    render(<CardActionMenu actions={makeActions()} />);
    const trigger = screen.getByRole("button", { name: "Actions" });

    trigger.focus();
    await user.keyboard("{ArrowDown}");
    expect(await screen.findByRole("menu")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByRole("menuitem", { name: "Edit" })).toHaveFocus());

    await user.keyboard("{ArrowDown}");
    await waitFor(() => expect(screen.getByRole("menuitem", { name: "Delete" })).toHaveFocus());

    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("menu")).not.toBeInTheDocument());
    await waitFor(() => expect(trigger).toHaveFocus());
  });

  it("does not bubble item and trigger clicks to the parent", async () => {
    const user = userEvent.setup();
    const parentClick = vi.fn();
    render(
      <div onClick={parentClick}>
        <CardActionMenu actions={makeActions()} />
      </div>,
    );
    await user.click(screen.getByRole("button", { name: "Actions" }));
    await user.click(await screen.findByRole("menuitem", { name: "Edit" }));
    expect(parentClick).not.toHaveBeenCalled();
  });
});
