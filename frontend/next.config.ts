import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Strict TypeScript verification enabled (Zero build errors allowed)
  typescript: {
    ignoreBuildErrors: false,
  },
};

export default nextConfig;