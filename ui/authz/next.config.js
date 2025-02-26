module.exports = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:4780/api/:path*", // Proxy API requests to the Go backend
      },
    ];
  },
};
