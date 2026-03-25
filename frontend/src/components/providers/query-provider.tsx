import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { isAuthError } from "@/lib/api/errors";
import { api } from "@/lib/api/client";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      gcTime: 10 * 60 * 1000,
      retry: (failureCount, error) => {
        if (isAuthError(error)) return false;
        return failureCount < 3;
      },
    },
    mutations: {
      retry: 1,
    },
  },
});

api.onAuthFailure = () => {
  queryClient.clear();
};

export function QueryProvider({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
