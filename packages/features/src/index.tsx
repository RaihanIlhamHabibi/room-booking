"use client";

import { api, type Booking, type Person, type Room } from "@room-booking/api";
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
    {isLoading ? <Notice data-testid="rooms-loading">Loading rooms…</Notice> : isError ? <Card><Notice tone="error" data-testid="rooms-error">{error?.message || "Could not load rooms."}</Notice><Button type="button" onClick={() => refetch()} data-testid="rooms-retry">Try again</Button></Card> : rooms.length === 0 ? <Notice data-testid="rooms-empty">No rooms are available.</Notice> : <div className="room-grid" data-testid="room-list">{rooms.map((room) => <Link className="room-link" href={`/rooms/${encodeURIComponent(room.id)}`} key={room.id} data-testid={`room-${room.id}`}><Card className="room-card"><span className="eyebrow">SPACE · {room.capacity} PEOPLE</span><h2>{room.name}</h2><p className="muted">{room.description || "No description yet."}</p><div className="amenities">{room.amenities.map((amenity) => <span className="chip" key={amenity}>{amenity}</span>)}</div><span className="text-link">View room →</span></Card></Link>)}</div>}
  </Stack></AppShell>;
}

function Reservations({ roomId }: { roomId: string }) {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => { let active = true; api.bookings(roomId).then(({ data }) => { if (active) setBookings(data); }).catch((cause) => { if (active) setError(cause.message); }).finally(() => { if (active) setLoading(false); }); return () => { active = false; }; }, [roomId]);
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

export function RoomDetail({ id }: { id: string }) {
  const { query, result: room } = useOne<Room>({ resource: "rooms", id });
  const { isLoading, isError, error, refetch } = query;
  return <AppShell><Stack><Link className="back-link" href="/rooms" data-testid="back-to-rooms">← All rooms</Link>{isLoading ? <Notice data-testid="room-loading">Loading room…</Notice> : isError || !room ? <Card><Notice tone="error" data-testid="room-error">{error?.message || "Room unavailable."}</Notice><Button type="button" onClick={() => refetch()} data-testid="room-retry">Try again</Button></Card> : <><div className="page-heading"><span className="eyebrow">SHARED SPACE</span><h1>{room.name}</h1><p className="muted">Seats {room.capacity} people</p><div className="amenities">{room.amenities.map((amenity) => <span className="chip" key={amenity}>{amenity}</span>)}</div></div><div className="detail-grid"><RoomDescription room={room} /><Reservations roomId={room.id} /></div></>}</Stack></AppShell>;
}
