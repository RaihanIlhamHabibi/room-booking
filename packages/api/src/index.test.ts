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
});
