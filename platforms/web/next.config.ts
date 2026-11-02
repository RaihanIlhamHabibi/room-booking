import type { NextConfig } from "next";

const proxy = process.env.API_PROXY_URL || "http://localhost:18080";
const nextConfig: NextConfig = {
  agentRules: false,
  transpilePackages: ["@room-booking/api", "@room-booking/features", "@room-booking/ui"],
  async rewrites() { return [{ source: "/api/:path*", destination: `${proxy}/api/:path*` }]; },
};
export default nextConfig;
