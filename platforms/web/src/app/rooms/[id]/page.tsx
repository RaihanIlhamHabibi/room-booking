import { RoomDetail } from "@room-booking/features";
type SearchParams = { date?: string | string[]; startTime?: string | string[]; endTime?: string | string[] };
function first(value?: string | string[]) { return Array.isArray(value) ? value[0] ?? "" : value ?? ""; }

export default async function RoomPage({ params, searchParams }: { params: Promise<{ id: string }>; searchParams: Promise<SearchParams> }) {
  const [{ id }, search] = await Promise.all([params, searchParams]);
  return <RoomDetail id={id} bookingDefaults={{ date: first(search.date), startTime: first(search.startTime), endTime: first(search.endTime) }} />;
}
