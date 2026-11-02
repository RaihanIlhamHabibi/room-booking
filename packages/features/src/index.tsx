"use client";

import { api, type Booking, type BookingInput, type Person, type Room, type ScheduleOption } from "@room-booking/api";
import { Button, Card, Field, Notice, Stack } from "@room-booking/ui";
import { useGetIdentity, useList, useLogout, useOne, useUpdate } from "@refinedev/core";
import { useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

function formatTime(value: string) {
  return new Intl.DateTimeFormat("en", { dateStyle: "medium", timeStyle: "short", timeZone: "Asia/Jakarta" }).format(new Date(value));
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: person, isLoading } = useGetIdentity<Person>();
  const { mutateAsync: logout, isPending: loggingOut } = useLogout();
  const [logoutError, setLogoutError] = useState("");
  async function signOut() {
    setLogoutError("");
    try {
      await logout();
      queryClient.clear();
      router.replace("/login");
    } catch {
      setLogoutError("Could not sign out. Try again.");
    }
  }
  useEffect(() => { if (!isLoading && !person) router.replace("/login"); }, [isLoading, person, router]);
  if (isLoading || !person) return <main className="shell" data-testid="session-loading"><Notice>Checking demo session…</Notice></main>;
  return <>
    <header className="app-header"><div className="header-inner"><Link className="brand" href="/rooms">Common Room</Link><div className="header-actions"><span className="session-name">{person.name}</span><Button variant="quiet" type="button" disabled={loggingOut} onClick={signOut} data-testid="logout-button">Sign out</Button></div></div>{logoutError && <div className="header-inner"><Notice tone="error" data-testid="logout-error">{logoutError}</Notice></div>}</header>
    <main className="shell">{children}</main>
  </>;
}

export function DemoLogin() {
  const router = useRouter();
  const [people, setPeople] = useState<Person[]>([]);
  const [selected, setSelected] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => { let active = true; api.users().then(({ data }) => { if (active) { setPeople(data); setSelected(data[0]?.id ?? ""); } }).catch((cause) => { if (active) setError(cause.message); }).finally(() => { if (active) setLoading(false); }); return () => { active = false; }; }, []);
  async function signIn(event: React.FormEvent) {
    event.preventDefault(); setSubmitting(true); setError("");
    try { await api.login(selected); router.replace("/rooms"); router.refresh(); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Sign in failed."); }
    finally { setSubmitting(false); }
  }
  return <main className="login-screen"><Card className="login-card"><Stack><span className="eyebrow">CAMPUS COMMUNITY SPACES</span><h1>Find a room for your next gathering.</h1><p className="muted">Explore shared rooms and their reservations using a demo profile.</p>
    {loading ? <Notice data-testid="people-loading">Loading demo people…</Notice> : error ? <Notice tone="error" data-testid="login-error">{error}</Notice> : people.length === 0 ? <Notice data-testid="people-empty">No demo people are available.</Notice> : <form className="stack" onSubmit={signIn}><Field id="demo-person" label="Continue as"><select id="demo-person" value={selected} onChange={(event) => setSelected(event.target.value)} data-testid="person-select">{people.map((person) => <option key={person.id} value={person.id}>{person.name}</option>)}</select></Field><Button type="submit" disabled={!selected || submitting} data-testid="login-button">{submitting ? "Signing in…" : "Enter the space"}</Button></form>}
    <p className="fine-print">Demo dates: May 13–15, 2030 · Asia/Jakarta</p></Stack></Card></main>;
}

export function RoomDirectory() {
  const { query, result } = useList<Room>({ resource: "rooms", pagination: { mode: "off" } });
  const { isLoading, isError, error, refetch } = query;
  const rooms = result.data ?? [];
  return <AppShell><Stack><div className="page-heading"><span className="eyebrow">YOUR SHARED SPACES</span><h1>Rooms for every gathering</h1><p className="muted">Choose a room to see its details and existing reservations.</p></div>
    <Notice>Fixture reservations are shown for May 13–15, 2030 in Asia/Jakarta.</Notice>
    <AvailabilitySearch />
    <CalendarScheduler />
    {isLoading ? <Notice data-testid="rooms-loading">Loading rooms…</Notice> : isError ? <Card><Notice tone="error" data-testid="rooms-error">{error?.message || "Could not load rooms."}</Notice><Button type="button" onClick={() => refetch()} data-testid="rooms-retry">Try again</Button></Card> : rooms.length === 0 ? <Notice data-testid="rooms-empty">No rooms are available.</Notice> : <div className="room-grid" data-testid="room-list">{rooms.map((room) => <Link className="room-link" href={`/rooms/${encodeURIComponent(room.id)}`} key={room.id} data-testid={`room-${room.id}`}><Card className="room-card"><span className="eyebrow">SPACE · {room.capacity} PEOPLE</span><h2>{room.name}</h2><p className="muted">{room.description || "No description yet."}</p><div className="amenities">{room.amenities.map((amenity) => <span className="chip" key={amenity}>{amenity}</span>)}</div><span className="text-link">View room →</span></Card></Link>)}</div>}
  </Stack></AppShell>;
}

export function AvailabilitySearch() {
  const [date, setDate] = useState("2030-05-13");
  const [startTime, setStartTime] = useState("09:00");
  const [endTime, setEndTime] = useState("10:00");
  const [attendeeCount, setAttendeeCount] = useState("1");
  const [rooms, setRooms] = useState<Room[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function search(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setRooms(null);
    setLoading(true);
    try {
      const result = await api.availableRooms({ date, startTime, endTime, attendeeCount: Number(attendeeCount) });
      setRooms(result.data);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not search for available rooms.");
    } finally {
      setLoading(false);
    }
  }

  return <Card><Stack><div><h2>Find an available room</h2><p className="muted">Search by time and total attendees, including yourself.</p></div>
    <form className="stack" onSubmit={search}><div className="form-grid">
      <Field id="availability-date" label="Date" required><input id="availability-date" type="date" required value={date} onChange={(event) => setDate(event.target.value)} data-testid="availability-date" /></Field>
      <Field id="availability-start" label="Start time" required><input id="availability-start" type="time" required value={startTime} onChange={(event) => setStartTime(event.target.value)} data-testid="availability-start" /></Field>
      <Field id="availability-end" label="End time" required><input id="availability-end" type="time" min={startTime} required value={endTime} onChange={(event) => setEndTime(event.target.value)} data-testid="availability-end" /></Field>
      <Field id="availability-attendees" label="Total attendees" required><input id="availability-attendees" type="number" min="1" step="1" required value={attendeeCount} onChange={(event) => setAttendeeCount(event.target.value)} data-testid="availability-attendees" /></Field>
    </div><Button type="submit" disabled={loading} data-testid="availability-search">{loading ? "Searching…" : "Search rooms"}</Button></form>
    {error && <Notice tone="error" data-testid="availability-error">{error}</Notice>}
    {rooms && (rooms.length === 0 ? <Notice data-testid="availability-empty">No rooms fit the group and time you selected.</Notice> : <div className="availability-results" data-testid="availability-results">{rooms.map((room) => {
      const query = new URLSearchParams({ date, startTime, endTime, attendeeCount });
      return <Card key={room.id} className="availability-result"><div><h3>{room.name}</h3><p className="muted">Capacity: {room.capacity}</p></div><Link className="button button-primary" href={`/rooms/${encodeURIComponent(room.id)}?${query.toString()}`} data-testid={`book-available-${room.id}`}>Continue to booking</Link></Card>;
    })}</div>)}
  </Stack></Card>;
}

function formatClock(value: string) {
  return new Intl.DateTimeFormat("en-GB", { hour: "2-digit", minute: "2-digit", hour12: false, timeZone: "Asia/Jakarta" }).format(new Date(value));
}

export function CalendarScheduler() {
  const { data: currentPerson } = useGetIdentity<Person>();
  const [people, setPeople] = useState<Person[]>([]);
  const [date, setDate] = useState("2030-05-13");
  const [durationMinutes, setDurationMinutes] = useState("60");
  const [attendeeIds, setAttendeeIds] = useState<string[]>([]);
  const [options, setOptions] = useState<ScheduleOption[] | null>(null);
  const [loadingPeople, setLoadingPeople] = useState(true);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    api.users().then(({ data }) => {
      if (active) setPeople(data);
    }).catch((cause) => {
      if (active) setError(cause instanceof Error ? cause.message : "Could not load meeting participants.");
    }).finally(() => {
      if (active) setLoadingPeople(false);
    });
    return () => { active = false; };
  }, []);

  function toggleAttendee(id: string) {
    setAttendeeIds((selected) => selected.includes(id) ? selected.filter((item) => item !== id) : [...selected, id]);
  }

  async function search(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setOptions(null);
    setSearching(true);
    try {
      const result = await api.scheduleRooms({ date, durationMinutes: Number(durationMinutes), attendeeIds });
      setOptions(result.data);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not find a time for everyone.");
    } finally {
      setSearching(false);
    }
  }

  const otherPeople = people.filter((person) => person.id !== currentPerson?.id);
  const attendeeCount = attendeeIds.length + (currentPerson ? 1 : 0);

  return <Card><Stack><div><h2>Find a time everyone can attend</h2><p className="muted">We include you automatically and check everyone’s calendar and room bookings between 09:00 and 17:00.</p></div>
    {loadingPeople ? <Notice data-testid="schedule-people-loading">Loading participants…</Notice> : <form className="stack" onSubmit={search}>
      <div className="form-grid">
        <Field id="schedule-date" label="Date" required><input id="schedule-date" type="date" required value={date} onChange={(event) => setDate(event.target.value)} data-testid="schedule-date" /></Field>
        <Field id="schedule-duration" label="Meeting duration (minutes)" required><input id="schedule-duration" type="number" min="15" max="480" step="15" required value={durationMinutes} onChange={(event) => setDurationMinutes(event.target.value)} data-testid="schedule-duration" /></Field>
      </div>
      <fieldset className="attendee-list"><legend>Participants ({attendeeCount} including you)</legend>
        {currentPerson && <p className="muted" data-testid="schedule-organizer">Organizer: {currentPerson.name} (included)</p>}
        {otherPeople.length === 0 ? <Notice>No other participants are available.</Notice> : otherPeople.map((person) => <label className="attendee-option" key={person.id}><input type="checkbox" checked={attendeeIds.includes(person.id)} onChange={() => toggleAttendee(person.id)} data-testid={`schedule-person-${person.id}`} /><span>{person.name}</span></label>)}
      </fieldset>
      <Button type="submit" disabled={searching || loadingPeople || attendeeIds.length === 0} data-testid="schedule-search">{searching ? "Finding times…" : "Find times and rooms"}</Button>
    </form>}
    {error && <Notice tone="error" data-testid="schedule-error">{error}</Notice>}
    {options && (options.length === 0 ? <Notice data-testid="schedule-empty">No time and room combination is available for everyone on this date.</Notice> : <div className="schedule-results" data-testid="schedule-results">{options.map((option) => {
      const startTime = formatClock(option.startsAt);
      const endTime = formatClock(option.endsAt);
      return <section className="schedule-slot" key={option.startsAt}>
        <h3>{startTime}–{endTime}</h3>
        <div className="schedule-rooms">{option.rooms.map((room) => {
          const query = new URLSearchParams({ date, startTime, endTime, attendeeCount: String(attendeeCount) });
          return <Link className="button button-quiet" href={`/rooms/${encodeURIComponent(room.id)}?${query.toString()}`} key={room.id} data-testid={`schedule-book-${option.startsAt}-${room.id}`}>{room.name} · {room.capacity} seats · Book</Link>;
        })}</div>
      </section>;
    })}</div>)}
  </Stack></Card>;
}

export function RoomBookingForm({ roomId, defaults, onCreated }: {
  roomId: string;
  defaults?: Pick<BookingInput, "date" | "startTime" | "endTime">;
  onCreated: () => void;
}) {
  const [date, setDate] = useState(defaults?.date || "2030-05-13");
  const [startTime, setStartTime] = useState(defaults?.startTime || "09:00");
  const [endTime, setEndTime] = useState(defaults?.endTime || "10:00");
  const [purpose, setPurpose] = useState("");
  const [pending, setPending] = useState(false);
  const [feedback, setFeedback] = useState<{ message: string; tone: "error" | "success" } | null>(null);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setPending(true);
    setFeedback(null);
    try {
      await api.createBooking(roomId, { date, startTime, endTime, purpose });
      setFeedback({ message: "Booking created. It now appears in this room’s reservations.", tone: "success" });
      setPurpose("");
      onCreated();
    } catch (cause) {
      setFeedback({ message: cause instanceof Error ? cause.message : "Could not create the booking.", tone: "error" });
    } finally {
      setPending(false);
    }
  }

  return <Card><h2>Reserve this room</h2><form className="stack" onSubmit={submit}>
    <Field id="booking-date" label="Date" required><input id="booking-date" type="date" required value={date} onChange={(event) => setDate(event.target.value)} data-testid="booking-date" /></Field>
    <div className="form-grid">
      <Field id="booking-start" label="Start time" required><input id="booking-start" type="time" required value={startTime} onChange={(event) => setStartTime(event.target.value)} data-testid="booking-start" /></Field>
      <Field id="booking-end" label="End time" required><input id="booking-end" type="time" min={startTime} required value={endTime} onChange={(event) => setEndTime(event.target.value)} data-testid="booking-end" /></Field>
    </div>
    <Field id="booking-purpose" label="Purpose" required hint="Up to 200 characters."><textarea id="booking-purpose" rows={3} maxLength={200} required value={purpose} onChange={(event) => setPurpose(event.target.value)} data-testid="booking-purpose" /></Field>
    <Button type="submit" disabled={pending || !purpose.trim()} data-testid="create-booking">{pending ? "Booking…" : "Create booking"}</Button>
    {feedback && <Notice tone={feedback.tone} data-testid="booking-feedback">{feedback.message}</Notice>}
  </form></Card>;
}

function Reservations({ roomId, refreshToken }: { roomId: string; refreshToken: number }) {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => { let active = true; setLoading(true); setError(""); api.bookings(roomId).then(({ data }) => { if (active) setBookings(data); }).catch((cause) => { if (active) setError(cause.message); }).finally(() => { if (active) setLoading(false); }); return () => { active = false; }; }, [roomId, refreshToken]);
  return <Card><h2>Existing reservations</h2><p className="muted">All times shown in Asia/Jakarta.</p>{loading ? <Notice data-testid="bookings-loading">Loading reservations…</Notice> : error ? <Notice tone="error" data-testid="bookings-error">{error}</Notice> : bookings.length === 0 ? <Notice data-testid="bookings-empty">No reservations for this room yet.</Notice> : <ul className="booking-list" data-testid="booking-list">{bookings.map((booking) => <li key={booking.id} data-testid={`booking-${booking.id}`}><strong>{booking.purpose}</strong><span>{formatTime(booking.startsAt)} – {formatTime(booking.endsAt)}</span></li>)}</ul>}</Card>;
}

function RoomDescription({ room }: { room: Room }) {
  const [editing, setEditing] = useState(false);
  const [description, setDescription] = useState(room.description);
  const [feedback, setFeedback] = useState("");
  const [isPending, setIsPending] = useState(false);
  const { mutateAsync } = useUpdate<Room>();
  useEffect(() => setDescription(room.description), [room.description]);
  async function save(event: React.FormEvent) {
    event.preventDefault(); setFeedback(""); setIsPending(true);
    try { await mutateAsync({ resource: "rooms", id: room.id, values: { description } }); setEditing(false); setFeedback("Description saved."); }
    catch (cause) { setFeedback(cause instanceof Error ? cause.message : "Could not save the description."); }
    finally { setIsPending(false); }
  }
  return <Card><div className="section-head"><div><h2>About this room</h2><p className="muted">A short note to help members choose.</p></div>{!editing && <Button type="button" variant="quiet" onClick={() => { setFeedback(""); setEditing(true); }} data-testid="edit-description">Edit description</Button>}</div>
    {editing ? <form className="stack" onSubmit={save}><Field id="room-description" label="Description" required><textarea id="room-description" rows={4} maxLength={500} required value={description} onChange={(event) => setDescription(event.target.value)} data-testid="description-input" /></Field><div className="button-row"><Button type="submit" disabled={isPending || !description.trim()} data-testid="save-description">{isPending ? "Saving…" : "Save"}</Button><Button type="button" variant="quiet" disabled={isPending} onClick={() => { setDescription(room.description); setEditing(false); setFeedback(""); }} data-testid="cancel-description">Cancel</Button></div></form> : <p>{room.description || "No description yet."}</p>}
    {feedback && <Notice tone={editing ? "error" : "success"} data-testid="description-feedback">{feedback}</Notice>}
  </Card>;
}

export function RoomDetail({ id, bookingDefaults }: { id: string; bookingDefaults?: Pick<BookingInput, "date" | "startTime" | "endTime"> }) {
  const [refreshToken, setRefreshToken] = useState(0);
  const { query, result: room } = useOne<Room>({ resource: "rooms", id });
  const { isLoading, isError, error, refetch } = query;
  return <AppShell><Stack><Link className="back-link" href="/rooms" data-testid="back-to-rooms">← All rooms</Link>{isLoading ? <Notice data-testid="room-loading">Loading room…</Notice> : isError || !room ? <Card><Notice tone="error" data-testid="room-error">{error?.message || "Room unavailable."}</Notice><Button type="button" onClick={() => refetch()} data-testid="room-retry">Try again</Button></Card> : <><div className="page-heading"><span className="eyebrow">SHARED SPACE</span><h1>{room.name}</h1><p className="muted">Seats {room.capacity} people</p><div className="amenities">{room.amenities.map((amenity) => <span className="chip" key={amenity}>{amenity}</span>)}</div></div><div className="detail-grid"><RoomDescription room={room} /><RoomBookingForm key={room.id} roomId={room.id} defaults={bookingDefaults} onCreated={() => setRefreshToken((value) => value + 1)} /><Reservations roomId={room.id} refreshToken={refreshToken} /></div></>}</Stack></AppShell>;
}
