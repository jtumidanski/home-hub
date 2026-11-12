import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Routes, Route, useLocation } from "react-router-dom";
import { RedirectPreservingQuery } from "../redirect-preserving-query";
import { LocationProbe } from "@/test/location-probe";

function HashProbe() {
  return <div data-testid="hash">{useLocation().hash}</div>;
}

describe("RedirectPreservingQuery", () => {
  it("redirects with replace, preserving search and hash", () => {
    render(
      <MemoryRouter initialEntries={["/old?a=1#h"]}>
        <Routes>
          <Route path="/old" element={<RedirectPreservingQuery to="/new" />} />
          <Route path="/new" element={<div>new</div>} />
        </Routes>
        <LocationProbe />
        <HashProbe />
      </MemoryRouter>,
    );
    expect(screen.getByTestId("location").textContent).toBe("/new?a=1");
    expect(screen.getByTestId("hash").textContent).toBe("#h");
    expect(screen.getByTestId("nav-type").textContent).toBe("REPLACE");
  });
});
