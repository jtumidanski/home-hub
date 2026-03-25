import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { type ColumnDef } from "@tanstack/react-table";
import { DataTable } from "../data-table";

interface TestRow {
  id: string;
  name: string;
}

const columns: ColumnDef<TestRow, unknown>[] = [
  { accessorKey: "id", header: "ID" },
  { accessorKey: "name", header: "Name" },
];

const sampleData: TestRow[] = [
  { id: "1", name: "Alice" },
  { id: "2", name: "Bob" },
];

describe("DataTable", () => {
  it("renders loading skeletons when isLoading is true", () => {
    render(<DataTable columns={columns} data={[]} isLoading />);
    expect(screen.queryByText("ID")).not.toBeInTheDocument();
    expect(screen.queryByText("Alice")).not.toBeInTheDocument();
  });

  it("renders configurable number of skeleton rows", () => {
    const { container } = render(
      <DataTable columns={columns} data={[]} isLoading skeletonRows={2} />,
    );
    const skeletons = container.querySelectorAll("[class*='animate-pulse'], [data-slot='skeleton']");
    expect(skeletons.length).toBeGreaterThanOrEqual(2);
  });

  it("renders empty message when data is empty", () => {
    render(<DataTable columns={columns} data={[]} emptyMessage="Nothing here" />);
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
  });

  it("renders default empty message when none provided", () => {
    render(<DataTable columns={columns} data={[]} />);
    expect(screen.getByText("No results.")).toBeInTheDocument();
  });

  it("renders table headers and rows when data is provided", () => {
    render(<DataTable columns={columns} data={sampleData} />);
    expect(screen.getByText("ID")).toBeInTheDocument();
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("Bob")).toBeInTheDocument();
  });

  it("calls onRowClick when a row is clicked", async () => {
    const user = userEvent.setup();
    const onRowClick = vi.fn();
    render(<DataTable columns={columns} data={sampleData} onRowClick={onRowClick} />);

    await user.click(screen.getByText("Alice"));
    expect(onRowClick).toHaveBeenCalledWith(sampleData[0]);
  });

  it("does not call onRowClick when not provided", async () => {
    const user = userEvent.setup();
    render(<DataTable columns={columns} data={sampleData} />);
    await user.click(screen.getByText("Alice"));
  });
});
