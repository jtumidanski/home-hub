import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { ErrorPage, Error404Page, Error403Page } from "../error-page";

function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

describe("ErrorPage", () => {
  it("renders default title for 404", () => {
    renderWithRouter(<ErrorPage statusCode={404} />);
    expect(screen.getByText("Page Not Found")).toBeInTheDocument();
  });

  it("renders default title for 403", () => {
    renderWithRouter(<ErrorPage statusCode={403} />);
    expect(screen.getByText("Access Denied")).toBeInTheDocument();
  });

  it("renders default title for 500", () => {
    renderWithRouter(<ErrorPage statusCode={500} />);
    expect(screen.getByText("Server Error")).toBeInTheDocument();
  });

  it("renders custom title and message", () => {
    renderWithRouter(
      <ErrorPage statusCode={418} title="I'm a Teapot" message="Cannot brew coffee" />
    );
    expect(screen.getByText("I'm a Teapot")).toBeInTheDocument();
    expect(screen.getByText("Cannot brew coffee")).toBeInTheDocument();
  });

  it("renders retry button when showRetryButton and onRetry provided", () => {
    const onRetry = vi.fn();
    renderWithRouter(
      <ErrorPage statusCode={500} showRetryButton onRetry={onRetry} />
    );
    const button = screen.getByRole("button", { name: /try again/i });
    fireEvent.click(button);
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("does not render retry button without onRetry", () => {
    renderWithRouter(<ErrorPage statusCode={500} showRetryButton />);
    expect(screen.queryByRole("button", { name: /try again/i })).not.toBeInTheDocument();
  });

  it("renders back button when showBackButton is true", () => {
    renderWithRouter(<ErrorPage statusCode={404} showBackButton />);
    expect(screen.getByRole("button", { name: /go back/i })).toBeInTheDocument();
  });

  it("does not render back button by default", () => {
    renderWithRouter(<ErrorPage statusCode={404} />);
    expect(screen.queryByRole("button", { name: /go back/i })).not.toBeInTheDocument();
  });
});

describe("Error404Page", () => {
  it("renders 404 with back button", () => {
    renderWithRouter(<Error404Page />);
    expect(screen.getByText("Page Not Found")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /go back/i })).toBeInTheDocument();
  });
});

describe("Error403Page", () => {
  it("renders 403 with back button", () => {
    renderWithRouter(<Error403Page />);
    expect(screen.getByText("Access Denied")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /go back/i })).toBeInTheDocument();
  });
});
