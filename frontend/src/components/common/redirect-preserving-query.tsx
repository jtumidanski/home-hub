import React from "react";
import { Navigate, useLocation } from "react-router-dom";

export function RedirectPreservingQuery({ to }: { to: string }): React.JSX.Element {
  const { search, hash } = useLocation();
  return <Navigate to={{ pathname: to, search, hash }} replace />;
}
