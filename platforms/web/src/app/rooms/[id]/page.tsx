import { RoomDetail } from "@room-booking/features";
export default async function RoomPage({ params }: { params: Promise<{ id: string }> }) { const { id } = await params; return <RoomDetail id={id} />; }
