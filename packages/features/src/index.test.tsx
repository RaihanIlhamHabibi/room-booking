// @vitest-environment jsdom
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AppShell, DemoLogin } from "./index";

const push = vi.fn();
const clearCache = vi.fn();
const logout = vi.fn(async () => ({ success: true }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: push, refresh: vi.fn() }) }));
vi.mock("@room-booking/api", () => ({ api: { users: vi.fn(async () => ({ data: [{ id: "person-a", name: "Ari Moss", email: "ari@example.test" }] })), login: vi.fn(async () => ({ id: "person-a" })) } }));
vi.mock("@refinedev/core", () => ({ useGetIdentity: () => ({ data: { id: "person-a", name: "Ari Moss", email: "ari@example.test" }, isLoading: false }), useLogout: () => ({ mutateAsync: logout, isPending: false }) }));
vi.mock("@tanstack/react-query", () => ({ useQueryClient: () => ({ clear: clearCache }) }));

afterEach(() => { push.mockClear(); clearCache.mockClear(); logout.mockClear(); });

describe("demo login", () => {
  it("loads a fictional person and signs in through the shared API", async () => {
    render(<DemoLogin />);
    await screen.findByRole("option", { name: "Ari Moss" });
    fireEvent.click(screen.getByTestId("login-button"));
    await waitFor(() => expect(push).toHaveBeenCalledWith("/rooms"));
  });

  it("clears cached identity and navigates after logout completes", async () => {
    render(<AppShell><p>Room directory</p></AppShell>);
    fireEvent.click(screen.getByTestId("logout-button"));
    await waitFor(() => expect(push).toHaveBeenCalledWith("/login"));
    expect(logout).toHaveBeenCalledOnce();
    expect(clearCache).toHaveBeenCalledOnce();
    expect(clearCache.mock.invocationCallOrder[0]).toBeLessThan(push.mock.invocationCallOrder[0]);
  });
});
