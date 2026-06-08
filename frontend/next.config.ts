import type { NextConfig } from 'next';
import path from 'path';

const nextConfig: NextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  // Local monorepo: Playwright root lockfile requires tracing from repo root.
  // Docker sets DOCKER_BUILD=1 so standalone output is flat (.next/standalone/server.js).
  ...(process.env.DOCKER_BUILD !== '1'
    ? { outputFileTracingRoot: path.join(__dirname, '..') }
    : {}),
};

export default nextConfig;
