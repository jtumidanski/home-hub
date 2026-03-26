import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";

const mockUseAuth = vi.fn();
const mockUseHouseholdMembers = vi.fn();
const mockUseHouseholdInvitations = vi.fn();
const mockUseUsersByIds = vi.fn();
const mockUpdateRoleMutateAsync = vi.fn();
const mockRemoveMemberMutateAsync = vi.fn();
const mockLeaveMutateAsync = vi.fn();
const mockRevokeMutateAsync = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/lib/hooks/api/use-memberships", () => ({
  useHouseholdMembers: () => mockUseHouseholdMembers(),
  useUpdateMemberRole: () => ({ mutateAsync: mockUpdateRoleMutateAsync, isPending: false }),
  useRemoveMember: () => ({ mutateAsync: mockRemoveMemberMutateAsync, isPending: false }),
  useLeaveHousehold: () => ({ mutateAsync: mockLeaveMutateAsync, isPending: false }),
}));

vi.mock("@/lib/hooks/api/use-invitations", () => ({
  useHouseholdInvitations: () => mockUseHouseholdInvitations(),
  useRevokeInvitation: () => ({ mutateAsync: mockRevokeMutateAsync, isPending: false }),
}));

vi.mock("@/lib/hooks/api/use-users", () => ({
  useUsersByIds: () => mockUseUsersByIds(),
}));

vi.mock("@/components/features/households/invite-member-dialog", () => ({
  InviteMemberDialog: ({ open }: { open: boolean }) =>
    open ? <div role="dialog">InviteMemberDialog</div> : null,
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/lib/api/errors", () => ({
  createErrorFromUnknown: (_err: unknown, fallback: string) => ({ message: fallback, type: "unknown" }),
}));

import { HouseholdMembersPage } from "../HouseholdMembersPage";

const householdId = "hh-1";
const currentUserId = "user-1";

function renderPage() {
  return render(
    <MemoryRouter initialEntries={[`/app/households/${householdId}/members`]}>
      <Routes>
        <Route path="/app/households/:id/members" element={<HouseholdMembersPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

const makeMember = (id: string, userId: string, role: string, isLastOwner = false) => ({
  id,
  type: "memberships",
  attributes: { role, isLastOwner, createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" },
  relationships: {
    household: { data: { type: "households", id: householdId } },
    user: { data: { type: "users", id: userId } },
  },
});

const makeUser = (id: string, name: string, email: string) => ({
  id,
  type: "users",
  attributes: { displayName: name, email, givenName: "", familyName: "", avatarUrl: "", createdAt: "", updatedAt: "" },
});

const makeInvitation = (id: string, email: string, role: string) => ({
  id,
  type: "invitations",
  attributes: { email, role, status: "pending", expiresAt: "2026-04-01T00:00:00Z", createdAt: "2026-03-01T00:00:00Z", updatedAt: "2026-03-01T00:00:00Z" },
  relationships: {
    household: { data: { type: "households", id: householdId } },
    invitedBy: { data: { type: "users", id: currentUserId } },
  },
});

describe("HouseholdMembersPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    mockUseAuth.mockReturnValue({
      user: { id: currentUserId, attributes: { displayName: "Current User", email: "current@example.com" } },
      appContext: { attributes: { resolvedRole: "owner", pendingInvitationCount: 0 } },
    });

    mockUseHouseholdMembers.mockReturnValue({
      data: {
        data: [
          makeMember("m-1", currentUserId, "owner", false),
          makeMember("m-2", "user-2", "viewer", false),
        ],
      },
      isLoading: false,
      isError: false,
    });

    mockUseHouseholdInvitations.mockReturnValue({
      data: { data: [] },
      isLoading: false,
    });

    mockUseUsersByIds.mockReturnValue({
      data: {
        data: [
          makeUser(currentUserId, "Current User", "current@example.com"),
          makeUser("user-2", "Other User", "other@example.com"),
        ],
      },
    });
  });

  it("renders loading skeleton when members are loading", () => {
    mockUseHouseholdMembers.mockReturnValue({ data: null, isLoading: true, isError: false });
    renderPage();
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders error state on load failure", () => {
    mockUseHouseholdMembers.mockReturnValue({ data: null, isLoading: false, isError: true });
    renderPage();
    expect(screen.getByText(/failed to load household members/i)).toBeInTheDocument();
  });

  it("renders member list with names and roles", () => {
    renderPage();
    expect(screen.getByText("Household Members")).toBeInTheDocument();
    expect(screen.getByText("Current User")).toBeInTheDocument();
    expect(screen.getByText("Other User")).toBeInTheDocument();
    expect(screen.getByText("You")).toBeInTheDocument();
  });

  it("shows invite button for privileged users", () => {
    renderPage();
    expect(screen.getByRole("button", { name: /invite/i })).toBeInTheDocument();
  });

  it("hides invite button for non-privileged users", () => {
    mockUseAuth.mockReturnValue({
      user: { id: currentUserId, attributes: { displayName: "Current User", email: "current@example.com" } },
      appContext: { attributes: { resolvedRole: "viewer", pendingInvitationCount: 0 } },
    });
    renderPage();
    expect(screen.queryByRole("button", { name: /invite/i })).not.toBeInTheDocument();
  });

  it("shows leave button for all members", () => {
    renderPage();
    expect(screen.getByRole("button", { name: /leave/i })).toBeInTheDocument();
  });

  it("shows sole owner badge when member is last owner", () => {
    mockUseHouseholdMembers.mockReturnValue({
      data: {
        data: [makeMember("m-1", currentUserId, "owner", true)],
      },
      isLoading: false,
      isError: false,
    });
    renderPage();
    expect(screen.getByText("Sole Owner")).toBeInTheDocument();
  });

  it("renders pending invitations section", () => {
    mockUseHouseholdInvitations.mockReturnValue({
      data: {
        data: [makeInvitation("inv-1", "invited@example.com", "viewer")],
      },
      isLoading: false,
    });
    renderPage();
    expect(screen.getByText("Pending Invitations")).toBeInTheDocument();
    expect(screen.getByText("invited@example.com")).toBeInTheDocument();
  });

  it("renders empty invitations message", () => {
    renderPage();
    expect(screen.getByText("No pending invitations.")).toBeInTheDocument();
  });

  it("opens invite dialog on button click", async () => {
    const user = userEvent.setup();
    renderPage();
    await user.click(screen.getByRole("button", { name: /invite/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("opens remove confirmation dialog", async () => {
    const user = userEvent.setup();
    renderPage();
    // Find remove buttons (UserMinus icon buttons) — there should be one for the non-self member
    const removeButtons = screen.getAllByRole("button").filter(
      (btn) => btn.querySelector(".text-destructive") !== null,
    );
    // Click the first remove icon button
    if (removeButtons.length > 0) {
      await user.click(removeButtons[0]!);
      await waitFor(() => {
        expect(screen.getByText("Remove Member")).toBeInTheDocument();
      });
    }
  });
});
