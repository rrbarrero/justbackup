import type { NextConfig } from "next";
// Use the internal URL for server-side rewrites; fail fast if not provided.
const API_URL = process.env.BACKEND_INTERNAL_URL;
if (!API_URL) {
  throw new Error("BACKEND_INTERNAL_URL is required for Next.js rewrites");
}

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/v1/:path*",
        destination: API_URL + "/api/v1/:path*",
      },
      {
        source: "/api/:path*",
        destination: API_URL + "/api/v1/:path*",
      },
    ];
  },
};

export default nextConfig;
