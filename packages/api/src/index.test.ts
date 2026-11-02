import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError } from "./index";

afterEach(() => vi.unstubAllGlobals());

describe("room API", () => {
  it("updates a description through the shared request path", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ id: "room-a", description: "Quiet corner" }) });
    vi.stubGlobal("fetch", fetchMock);
    await expect(api.updateRoom("room-a", "Quiet corner")).resolves.toMatchObject({ description: "Quiet corner" });
    expect(fetchMock).toHaveBeenCalledWith("/api/rooms/room-a", expect.objectContaining({ method: "PATCH", body: JSON.stringify({ description: "Quiet corner" }), credentials: "same-origin" }));
  });

  it("surfaces an API failure to the caller", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 403, json: async () => ({ message: "Demo session expired" }) }));
    await expect(api.room("room-a")).rejects.toEqual(new ApiError(403, "Demo session expired"));
  });

  it("creates a booking through the authenticated API path", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 201, json: async () => ({ id: "booking-a" }) });
    vi.stubGlobal("fetch", fetchMock);
    await expect(api.createBooking("room/a", { date: "2030-05-13", startTime: "09:00", endTime: "10:00", purpose: "Study group" })).resolves.toMatchObject({ id: "booking-a" });
    expect(fetchMock).toHaveBeenCalledWith("/api/rooms/room%2Fa/bookings", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ date: "2030-05-13", startTime: "09:00", endTime: "10:00", purpose: "Study group" }),
      credentials: "same-origin",
    }));
  });

  it("encodes availability criteria in the shared request path", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ data: [] }) });
    vi.stubGlobal("fetch", fetchMock);
    await api.availableRooms({ date: "2030-05-13", startTime: "09:00", endTime: "10:30", attendeeCount: 4 });
    expect(fetchMock).toHaveBeenCalledWith("/api/rooms/available?date=2030-05-13&startTime=09%3A00&endTime=10%3A30&attendeeCount=4", expect.objectContaining({ credentials: "same-origin" }));
  });

  it("posts the selected attendees and duration to the calendar scheduler", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ data: [] }) });
    vi.stubGlobal("fetch", fetchMock);
    await api.scheduleRooms({ date: "2030-05-13", durationMinutes: 60, attendeeIds: ["u-bima", "u-citra"] });
    expect(fetchMock).toHaveBeenCalledWith("/api/rooms/schedule", expect.objectContaining({
      method: "POST",
      body: JSON.stringify({ date: "2030-05-13", durationMinutes: 60, attendeeIds: ["u-bima", "u-citra"] }),
      credentials: "same-origin",
    }));
  });
});
