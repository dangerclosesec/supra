// src/components/layout/Layout.tsx
import React from "react";
import { Database, Network, Shield, Share2 } from "lucide-react";
import Link from "next/link";
import ThemeToggle from "./ThemeToggle";
import { useRouter } from "next/router";

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const router = useRouter();
  return (
    <div
      style={{ minHeight: "100vh", display: "flex", flexDirection: "column" }}
    >
      <header className="app-header">
        <div
          className="container-custom"
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div className={"flex items-center space-x-4"}>
            <Share2 className="icon-md text-teal" />
            <h1 style={{ fontSize: "1.25rem", fontWeight: "bold", margin: 0 }}>
              Identity Graph
            </h1>
          </div>

          <nav className="flex items-center space-x-4 w-full">
            <Link
              href="/"
              className={`nav-item ${
                router.pathname === "/" ? "nav-item-active" : ""
              }`}
            >
              <Database className="icon-sm" />
              <span>Entities</span>
            </Link>
            <Link
              href="/permissions"
              className={`nav-item ${
                router.pathname === "/permissions" ? "nav-item-active" : ""
              }`}
            >
              <Shield className="icon-sm" />
              <span>Permissions</span>
            </Link>
            <Link
              href="/health"
              className={`nav-item ${
                router.pathname === "/health" ? "nav-item-active" : ""
              }`}
            >
              <Shield className="icon-sm" />
              <span>Health</span>
            </Link>
          </nav>
          <div className="ml-auto flex items-center space-x-4">
            <ThemeToggle />
          </div>
        </div>
      </header>

      <main style={{ flex: "1", paddingTop: "2rem", paddingBottom: "2rem" }}>
        <div className="container-custom">{children}</div>
      </main>

      <footer className="app-footer">
        <div className="container-custom">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              flexWrap: "wrap",
            }}
          >
            <p style={{ margin: 0 }}>Identity Graph Visualization</p>
            <p style={{ margin: 0 }}>Built with Next.js and Tailwind CSS</p>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Layout;
