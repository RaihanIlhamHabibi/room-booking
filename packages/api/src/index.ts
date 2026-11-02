import type { AuthProvider, BaseRecord, DataProvider } from "@refinedev/core";

export type Person = { id: string; name: string; email: string };
export type Room = { id: string; name: string; capacity: number; amenities: string[]; description: string };
export type Booking = { id: string; roomId: string; organizerId: string; startsAt: string; endsAt: string; purpose: string };
export type CalendarEvent = { id: string; userId: string; title: string; startsAt: string; endsAt: string };

export class ApiError extends Error {
  constructor(public status: number, message: string) { super(message); }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    credentials: "same-origin",
    headers: { ...(init?.body ? { "Content-Type": "application/json" } : {}), ...init?.headers },
  });
  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try { const body = await response.json(); message = body.message || body.error || message; } catch { /* retain status message */ }
    throw new ApiError(response.status, message);
  }
  return response.status === 204 ? undefined as T : response.json() as Promise<T>;
}

export const api = {
  users: () => request<{ data: Person[] }>("/users"),
  session: () => request<Person>("/session"),
  login: (userId: string) => request<Person>("/session", { method: "POST", body: JSON.stringify({ userId }) }),
  logout: () => request<void>("/session", { method: "DELETE" }),
  rooms: () => request<{ data: Room[] }>("/rooms"),
  room: (id: string) => request<Room>(`/rooms/${encodeURIComponent(id)}`),
  updateRoom: (id: string, description: string) => request<Room>(`/rooms/${encodeURIComponent(id)}`, { method: "PATCH", body: JSON.stringify({ description }) }),
  bookings: (roomId: string) => request<{ data: Booking[] }>(`/rooms/${encodeURIComponent(roomId)}/bookings`),
  calendarEvents: (date?: string) => request<{ data: CalendarEvent[] }>(`/calendar-events${date ? `?date=${encodeURIComponent(date)}` : ""}`),
};

export const authProvider: AuthProvider = {
  login: async ({ userId }) => { await api.login(userId); return { success: true, redirectTo: "/rooms" }; },
  logout: async () => { await api.logout(); return { success: true, redirectTo: "/login" }; },
  check: async () => { try { await api.session(); return { authenticated: true }; } catch { return { authenticated: false, redirectTo: "/login" }; } },
  getIdentity: api.session,
  onError: async (error) => ({ logout: error?.status === 401 }),
};

export const dataProvider: DataProvider = {
  getList: async <TData extends BaseRecord = BaseRecord>({ resource }: { resource: string }) => {
    if (resource !== "rooms") throw new Error(`Unknown resource: ${resource}`);
    const { data } = await api.rooms();
    return { data: data as unknown as TData[], total: data.length };
  },
  getOne: async <TData extends BaseRecord = BaseRecord>({ resource, id }: { resource: string; id: string | number }) => {
    if (resource !== "rooms") throw new Error(`Unknown resource: ${resource}`);
    return { data: await api.room(String(id)) as unknown as TData };
  },
  update: async <TData extends BaseRecord = BaseRecord, TVariables = {}>({ resource, id, variables }: { resource: string; id: string | number; variables: TVariables }) => {
    if (resource !== "rooms") throw new Error(`Unknown resource: ${resource}`);
    const { description } = variables as { description?: unknown };
    return { data: await api.updateRoom(String(id), String(description ?? "")) as unknown as TData };
  },
  create: async () => { throw new Error("This resource does not support creation."); },
  deleteOne: async () => { throw new Error("This resource does not support deletion."); },
  getApiUrl: () => "/api",
};
