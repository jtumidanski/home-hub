import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { TenantProvider, useTenant } from "../tenant-context";
import { contextKeys } from "@/lib/hooks/api/use-context";

// --- Mocks ---

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: vi.fn(),
}));

vi.mock("@/lib/api/client", () => ({
  api: {
    setTenant: vi.fn(),
    clearTenant: vi.fn(),
  },
}));

vi.mock("@/services/api/account", () => ({
  accountService: {
    setActiveHousehold: vi.fn().mockResolvedValue(undefined),
  },
}));

// Import mocked modules to get references
import { useAuth } from "@/components/providers/auth-provider";
import { api } from "@/lib/api/client";
import { accountService } from "@/services/api/account";

const mockUseAuth = vi.mocked(useAuth);
const mockApi = vi.mocked(api);
const mockAccountService = vi.mocked(accountService);

// --- Helpers ---

function buildAppContext(overrides?: {
  tenantId?: string;
  householdId?: string;
  preferenceId?: string;
}) {
  const tenantId = overrides?.tenantId ?? "tenant-1";
  const householdId = overrides?.householdId ?? "household-1";
  const preferenceId = overrides?.preferenceId ?? "preference-1";

  return {
    id: "current" as const,
    type: "contexts" as const,
    attributes: {
      resolvedTheme: "light" as const,
      resolvedRole: "member",
      canCreateHousehold: true,
    },
    relationships: {
      tenant: { data: { id: tenantId, type: "tenants" as const } },
      activeHousehold: { data: { id: householdId, type: "households" as const } },
      preference: { data: { id: preferenceId, type: "preferences" as const } },
      memberships: { data: [] },
    },
  };
}

function buildIncludedTenant(id: string) {
  return {
    type: "tenants",
    id,
    attributes: {
      name: "Test Tenant",
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-02T00:00:00Z",
    },
  };
}

function buildIncludedHousehold(id: string) {
  return {
    type: "households",
    id,
    attributes: {
      name: "Test Household",
      timezone: "America/Chicago",
      units: "imperial",
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-02T00:00:00Z",
    },
  };
}

function createWrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <TenantProvider>{children}</TenantProvider>
      </QueryClientProvider>
    );
  };
}

function mockAuthValue(appContext: ReturnType<typeof buildAppContext> | null) {
  return {
    user: appContext ? { id: "user-1", type: "users" as const, attributes: { email: "test@example.com", displayName: "Test", givenName: "Test", familyName: "User", avatarUrl: "", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" } } : null,
    appContext,
    isLoading: false,
    isAuthenticated: !!appContext,
    needsOnboarding: false,
  };
}

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
}

// --- Tests ---

describe("TenantProvider / useTenant", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("throws when useTenant is used outside TenantProvider", () => {
    const queryClient = makeQueryClient();

    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    expect(() => renderHook(() => useTenant(), { wrapper })).toThrow(
      "useTenant must be used within TenantProvider",
    );
  });

  it("provides tenant and household objects from appContext relationships", () => {
    mockUseAuth.mockReturnValue(mockAuthValue(buildAppContext()));
    const queryClient = makeQueryClient();

    const { result } = renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.tenant).not.toBeNull();
    expect(result.current.tenant?.id).toBe("tenant-1");
    expect(result.current.household).not.toBeNull();
    expect(result.current.household?.id).toBe("household-1");
  });

  it("returns null tenant and household when appContext is null", () => {
    mockUseAuth.mockReturnValue(mockAuthValue(null));
    const queryClient = makeQueryClient();

    const { result } = renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.tenant).toBeNull();
    expect(result.current.household).toBeNull();
  });

  it("calls api.setTenant with tenant object when tenant is available", async () => {
    mockUseAuth.mockReturnValue(mockAuthValue(buildAppContext()));
    const queryClient = makeQueryClient();

    renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => {
      expect(mockApi.setTenant).toHaveBeenCalledWith(
        expect.objectContaining({ id: "tenant-1" }),
      );
    });
  });

  it("calls api.clearTenant when tenant is null", async () => {
    mockUseAuth.mockReturnValue(mockAuthValue(null));
    const queryClient = makeQueryClient();

    renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => {
      expect(mockApi.clearTenant).toHaveBeenCalled();
    });
  });

  it("setActiveHousehold calls accountService with tenant object and invalidates context queries", async () => {
    mockUseAuth.mockReturnValue(mockAuthValue(buildAppContext({ preferenceId: "pref-42" })));
    const queryClient = makeQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const { result } = renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.setActiveHousehold("household-99");
    });

    expect(mockAccountService.setActiveHousehold).toHaveBeenCalledWith(
      expect.objectContaining({ id: "tenant-1" }),
      "pref-42",
      "household-99",
    );
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: contextKeys.current(),
    });
  });

  it("extracts tenant from included resources in query cache", () => {
    mockUseAuth.mockReturnValue(mockAuthValue(buildAppContext()));
    const queryClient = makeQueryClient();

    queryClient.setQueryData(contextKeys.current(), {
      included: [buildIncludedTenant("tenant-1"), buildIncludedHousehold("household-1")],
    });

    const { result } = renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.tenant).toEqual({
      id: "tenant-1",
      type: "tenants",
      attributes: {
        name: "Test Tenant",
        createdAt: "2026-01-01T00:00:00Z",
        updatedAt: "2026-01-02T00:00:00Z",
      },
    });
  });

  it("extracts household from included resources in query cache", () => {
    mockUseAuth.mockReturnValue(mockAuthValue(buildAppContext()));
    const queryClient = makeQueryClient();

    queryClient.setQueryData(contextKeys.current(), {
      included: [buildIncludedTenant("tenant-1"), buildIncludedHousehold("household-1")],
    });

    const { result } = renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.household).toEqual({
      id: "household-1",
      type: "households",
      attributes: {
        name: "Test Household",
        timezone: "America/Chicago",
        units: "imperial",
        createdAt: "2026-01-01T00:00:00Z",
        updatedAt: "2026-01-02T00:00:00Z",
      },
    });
  });

  it("returns fallback tenant when included resource is not found", () => {
    mockUseAuth.mockReturnValue(mockAuthValue(buildAppContext()));
    const queryClient = makeQueryClient();

    queryClient.setQueryData(contextKeys.current(), { included: [] });

    const { result } = renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.tenant).toEqual({
      id: "tenant-1",
      type: "tenants",
      attributes: { name: "", createdAt: "", updatedAt: "" },
    });
  });

  it("returns fallback household when included resource is not found", () => {
    mockUseAuth.mockReturnValue(mockAuthValue(buildAppContext()));
    const queryClient = makeQueryClient();

    queryClient.setQueryData(contextKeys.current(), { included: [] });

    const { result } = renderHook(() => useTenant(), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.household).toEqual({
      id: "household-1",
      type: "households",
      attributes: {
        name: "",
        timezone: "",
        units: "imperial",
        createdAt: "",
        updatedAt: "",
      },
    });
  });
});
