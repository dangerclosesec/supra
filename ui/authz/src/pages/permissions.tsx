// src/pages/permissions.js
import React, { useState } from "react";
import Layout from "@/components/layout/Layout";
import PermissionTableVisualizer from "@/components/PermissionTableVisualizer";
import RelationshipCreator from "@/components/RelationshipCreator";
import {
  Shield,
  Book,
  GitCompare,
  TreesIcon,
  Antenna,
  Link2,
} from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const classNames = (...classes: string[]) => classes.filter(Boolean).join(" ");

export default function PermissionsPage() {
  const [activeTab, setActiveTab] = useState("explorer");

  return (
    <Layout>
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-4xl font-bold">Permission Management</h1>
          <p className="text-gray-400 mt-2">
            Explore and manage permissions, entities, and relations in your
            identity graph system
          </p>
        </div>

        <div className="mb-6">
          <div className="flex w-full space-x-1 rounded-lg bg-[#1e1e2e]/50 p-1 border border-slate-700/30">
            <button
              className={classNames(
                activeTab === "explorer" ? "bg-violet-500/10" : "",
                "w-full cursor-pointer rounded-md px-3 py-1.5 text-sm font-medium text-violet-400 outline-none"
              )}
              onClick={() => setActiveTab("explorer")}
            >
              <div className="flex items-center justify-center gap-2">
                <TreesIcon className="h-4 w-4" />
                <span>Explorer</span>
              </div>
            </button>
            <button
              className={classNames(
                activeTab === "documentation" ? "bg-violet-500/10" : "",
                "w-full cursor-pointer rounded-md px-3 py-1.5 text-sm font-medium text-violet-400 outline-none"
              )}
              onClick={() => setActiveTab("documentation")}
            >
              <div className="flex items-center justify-center gap-2">
                <Antenna className="h-4 w-4" />
                <span>Documentation</span>
              </div>
            </button>

            <button
              className={`flex items-center px-4 py-2 rounded-md text-sm font-medium transition-all duration-150 ease-in gap-2 ${
                activeTab === "relationship"
                  ? "bg-violet-500/10 text-violet-400 border border-violet-500/30"
                  : "text-gray-300 border border-slate-700/30 hover:bg-gray-600/30 hover:text-gray-100"
              }`}
              onClick={() => setActiveTab("relationship")}
            >
              <Link2 className="h-4 w-4" />
              <span>Add Relationship</span>
            </button>
          </div>
        </div>

        {/* Tab content */}
        {activeTab === "explorer" ? (
          <PermissionTableVisualizer />
        ) : activeTab === "documentation" ? (
          <div id="docs-content">
            <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
              <CardHeader className="border-b border-slate-700/40">
                <div className="flex items-center gap-2">
                  <Book className="h-5 w-5 text-teal-400" />
                  <CardTitle className="text-xl">
                    Permission System Documentation
                  </CardTitle>
                </div>
                <CardDescription className="text-gray-400">
                  Learn about our permission model and how to use it
                </CardDescription>
              </CardHeader>
              <CardContent className="pt-6 space-y-6">
                <section>
                  <h3 className="text-lg font-semibold mb-2">
                    Permission Model Overview
                  </h3>
                  <p className="text-gray-400">
                    Our permission system is based on a flexible
                    relationship-based model that allows defining complex access
                    control rules. Permissions are evaluated using expressions
                    that can reference direct and indirect relationships between
                    entities.
                  </p>
                </section>

                <section>
                  <h3 className="text-lg font-semibold mb-2">Key Concepts</h3>
                  <ul className="list-disc pl-6 space-y-2 text-gray-400">
                    <li>
                      <strong className="text-white">Entities</strong> - Objects
                      in the system that can have permissions or relationships
                      (e.g., users, organizations, projects)
                    </li>
                    <li>
                      <strong className="text-white">Relations</strong> -
                      Connections between entities (e.g., owner, member, admin)
                    </li>
                    <li>
                      <strong className="text-white">Permissions</strong> -
                      Access control rules defined using expressions
                    </li>
                    <li>
                      <strong className="text-white">Expressions</strong> -
                      Logic that defines permission rules using relations and
                      operators
                    </li>
                  </ul>
                </section>

                <section>
                  <h3 className="text-lg font-semibold mb-2">
                    Expression Syntax
                  </h3>
                  <div className="bg-[#2a2a3c] p-4 rounded-md border border-slate-700/40">
                    <p className="mb-2 font-medium">Examples:</p>
                    <ul className="list-disc pl-6 space-y-1 text-gray-400">
                      <li>
                        <code className="text-sm bg-violet-500/10 px-1 py-0.5 rounded text-violet-400">
                          owner
                        </code>{" "}
                        - Direct relation check
                      </li>
                      <li>
                        <code className="text-sm bg-violet-500/10 px-1 py-0.5 rounded text-violet-400">
                          organization.admin
                        </code>{" "}
                        - Indirect relation check
                      </li>
                      <li>
                        <code className="text-sm bg-violet-500/10 px-1 py-0.5 rounded text-violet-400">
                          owner or admin
                        </code>{" "}
                        - Logical OR
                      </li>
                      <li>
                        <code className="text-sm bg-violet-500/10 px-1 py-0.5 rounded text-violet-400">
                          (manager and organization.billing_manager)
                        </code>{" "}
                        - Logical AND with grouping
                      </li>
                    </ul>
                  </div>
                </section>
              </CardContent>
            </Card>
          </div>
        ) : (
          <RelationshipCreator />
        )}
      </div>
    </Layout>
  );
}
