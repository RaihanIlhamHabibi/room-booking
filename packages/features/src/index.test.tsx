// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { api } from "@room-booking/api";
import { AppShell, AvailabilitySearch, CalendarScheduler, DemoLogin, RoomBookingForm } from "./index";

const push = vi.fn();
const clearCache = vi.fn();
const logout = vi.fn(async () => ({ success: true }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: push, refresh: vi.fn() }) }));
vi.mock("@room-booking/api", () => ({ api: {
  users: vi.fn(async () => ({ data: [
    { id: "person-a", name: "Ari Moss", email: "ari@example.test" },
    { id: "person-b", name: "Bea Santos", email: "bea@example.test" },
    { id: "person-c", name: "Caro Lee", email: "caro@example.test" },
  ] })),
  login: vi.fn(async () => ({ id: "person-a" })),
  availableRooms: vi.fn(),
  scheduleRooms: vi.fn(),
  createBooking: vi.fn(),
} }));
vi.mock("@refinedev/core", () => ({ useGetIdentity: () => ({ data: { id: "person-a", name: "Ari Moss", email: "ari@example.test" }, isLoading: false }), useLogout: () => ({ mutateAsync: logout, isPending: false }) }));
vi.mock("@tanstack/react-query", () => ({ useQueryClient: () => ({ clear: clearCache }) }));

afterEach(() => { cleanup(); push.mockClear(); clearCache.mockClear(); logout.mockClear(); vi.mocked(api.availableRooms).mockReset(); vi.mocked(api.scheduleRooms).mockReset(); vi.mocked(api.createBooking).mockReset(); });

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

  it("searches by date, time, and total attendees and carries the criteria to booking", async () => {
    vi.mocked(api.availableRooms).mockResolvedValue({ data: [{ id: "room-a", name: "Garden Room", capacity: 4, amenities: [], description: "" }] });
    render(<AvailabilitySearch />);
    fireEvent.change(screen.getByTestId("availability-attendees"), { target: { value: "3" } });
    fireEvent.click(screen.getByTestId("availability-search"));
    const continueLink = await screen.findByTestId("book-available-room-a");
    expect(api.availableRooms).toHaveBeenCalledWith({ date: "2030-05-13", startTime: "09:00", endTime: "10:00", attendeeCount: 3 });
    expect(continueLink.getAttribute("href")).toContain("date=2030-05-13");
    expect(continueLink.getAttribute("href")).toContain("startTime=09%3A00");
    expect(continueLink.getAttribute("href")).toContain("endTime=10%3A00");
  });

  it("submits a booking with search defaults and refreshes the reservation list", async () => {
    vi.mocked(api.createBooking).mockResolvedValue({ id: "booking-a", roomId: "room-a", organizerId: "person-a", startsAt: "2030-05-13T09:00:00+07:00", endsAt: "2030-05-13T10:00:00+07:00", purpose: "Study group" });
    const onCreated = vi.fn();
    render(<RoomBookingForm roomId="room-a" defaults={{ date: "2030-05-14", startTime: "11:00", endTime: "12:00" }} onCreated={onCreated} />);
    expect((screen.getByTestId("booking-date") as HTMLInputElement).value).toBe("2030-05-14");
    expect((screen.getByTestId("booking-start") as HTMLInputElement).value).toBe("11:00");
    expect((screen.getByTestId("booking-end") as HTMLInputElement).value).toBe("12:00");
    fireEvent.change(screen.getByTestId("booking-purpose"), { target: { value: "Study group" } });
    fireEvent.click(screen.getByTestId("create-booking"));
    await screen.findByText(/Booking created/);
    expect(api.createBooking).toHaveBeenCalledWith("room-a", { date: "2030-05-14", startTime: "11:00", endTime: "12:00", purpose: "Study group" });
    expect(onCreated).toHaveBeenCalledOnce();
  });

  it("shows a booking conflict returned by the API", async () => {
    vi.mocked(api.createBooking).mockRejectedValue(new Error("room is already booked for that time"));
    const onCreated = vi.fn();
    render(<RoomBookingForm roomId="room-a" onCreated={onCreated} />);
    fireEvent.change(screen.getByTestId("booking-purpose"), { target: { value: "Study group" } });
    fireEvent.click(screen.getByTestId("create-booking"));
    expect(await screen.findByText("room is already booked for that time")).toBeTruthy();
    expect(onCreated).not.toHaveBeenCalled();
  });

  it("finds calendar-compatible slots and carries the chosen time to booking", async () => {
    vi.mocked(api.scheduleRooms).mockResolvedValue({ data: [{
      startsAt: "2030-05-13T10:00:00+07:00",
      endsAt: "2030-05-13T11:00:00+07:00",
      rooms: [{ id: "room-a", name: "Garden Room", capacity: 4, amenities: [], description: "" }],
    }] });
    render(<CalendarScheduler />);
    const participant = await screen.findByTestId("schedule-person-person-b");
    fireEvent.click(participant);
    fireEvent.click(screen.getByTestId("schedule-person-person-c"));
    expect(screen.getByTestId("schedule-organizer").textContent).toContain("Ari Moss");
    fireEvent.click(screen.getByTestId("schedule-search"));
    const bookLink = await screen.findByRole("link", { name: "Garden Room · 4 seats · Book" });
    expect(api.scheduleRooms).toHaveBeenCalledWith({ date: "2030-05-13", durationMinutes: 60, attendeeIds: ["person-b", "person-c"] });
    expect(bookLink.getAttribute("href")).toContain("date=2030-05-13");
    expect(bookLink.getAttribute("href")).toContain("startTime=10%3A00");
    expect(bookLink.getAttribute("href")).toContain("endTime=11%3A00");
    expect(bookLink.getAttribute("href")).toContain("attendeeCount=3");
  });

  it("shows an error when calendar scheduling fails", async () => {
    vi.mocked(api.scheduleRooms).mockRejectedValue(new Error("Schedule unavailable"));
    render(<CalendarScheduler />);
    fireEvent.click(await screen.findByTestId("schedule-person-person-b"));
    fireEvent.click(screen.getByTestId("schedule-search"));
    expect((await screen.findByTestId("schedule-error")).textContent).toBe("Schedule unavailable");
  });
});
