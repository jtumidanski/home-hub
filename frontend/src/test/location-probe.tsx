import React from "react";
import { useLocation, useNavigate, useNavigationType } from "react-router-dom";

export function LocationProbe(): React.JSX.Element {
  const loc = useLocation();
  const navType = useNavigationType();
  const navigate = useNavigate();
  return (
    <>
      <div data-testid="location">{loc.pathname + loc.search}</div>
      <div data-testid="nav-type">{navType}</div>
      <button onClick={() => navigate(-1)}>probe-back</button>
    </>
  );
}
