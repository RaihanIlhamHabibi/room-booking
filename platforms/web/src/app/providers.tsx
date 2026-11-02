"use client";

import { authProvider, dataProvider } from "@room-booking/api";
import { Refine } from "@refinedev/core";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({ defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } } }));
  return <QueryClientProvider client={queryClient}><Refine authProvider={authProvider} dataProvider={dataProvider} resources={[{ name: "rooms", list: "/rooms", show: "/rooms/:id", edit: "/rooms/:id" }]} options={{ syncWithLocation: false, disableTelemetry: true }}>{children}</Refine></QueryClientProvider>;
}
