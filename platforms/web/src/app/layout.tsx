import type { Metadata } from "next";
import { Providers } from "./providers";
import "./style.css";

export const metadata: Metadata = { title: "Common Room", description: "A fictional campus room directory" };
export default function Layout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="en"><body><Providers>{children}</Providers></body></html>;
}
