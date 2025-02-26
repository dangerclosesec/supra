// src/pages/health.tsx
import Layout from "@/components/layout/Layout";
import OrphanedRelationshipsView from "@/components/OrphanedRelationshipsView";
import { ActivitySquare, Database, Shield, Terminal } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function HealthPage() {
  return (
    <Layout>
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-4xl font-bold">System Health</h1>
          <p className="text-gray-400 mt-2">
            Monitor and maintain your identity graph system integrity
          </p>
        </div>

        <Tabs defaultValue="orphaned" className="w-full">
          <TabsList className="w-full grid grid-cols-3 mb-6 bg-[#1e1e2e]/50 border border-slate-700/30">
            <TabsTrigger
              value="orphaned"
              className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
            >
              <Terminal className="h-4 w-4" />
              Orphaned Relationships
            </TabsTrigger>
            <TabsTrigger
              value="stats"
              className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
            >
              <ActivitySquare className="h-4 w-4" />
              System Stats
            </TabsTrigger>
            <TabsTrigger
              value="logs"
              className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
            >
              <Terminal className="h-4 w-4" />
              System Logs
            </TabsTrigger>
          </TabsList>

          <TabsContent value="orphaned">
            <OrphanedRelationshipsView />
          </TabsContent>

          <TabsContent value="stats">
            <div className="text-center py-24 border border-slate-700/40 rounded-md bg-[#2a2a3c]/50">
              <p className="text-gray-400 text-lg">System Stats Coming Soon</p>
              <p className="text-gray-500 text-sm mt-2">
                This feature is under development
              </p>
            </div>
          </TabsContent>

          <TabsContent value="logs">
            <div className="text-center py-24 border border-slate-700/40 rounded-md bg-[#2a2a3c]/50">
              <p className="text-gray-400 text-lg">System Logs Coming Soon</p>
              <p className="text-gray-500 text-sm mt-2">
                This feature is under development
              </p>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </Layout>
  );
}
