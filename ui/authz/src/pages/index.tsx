// src/pages/index.tsx
import Layout from "@/components/layout/Layout";
import { useEffect, useState } from "react";
import {
  Network,
  Lock,
  Share2,
  Loader2,
  Eye,
  Search,
  Link2,
  RefreshCw,
  ShieldCheck,
  ShieldAlert,
  Filter,
  Zap,
} from "lucide-react";
import { GraphData, PermissionPathResponse, Link, Node } from "@/types/graph";
import GraphVisualization from "@/components/GraphVisualization";
import DynamicPermissionAnalyzer from "@/components/DynamicPermissionAnalyzer";
import RelationshipCreator from "@/components/RelationshipCreator";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const classNames = (...classes: string[]) => classes.filter(Boolean).join(" ");
export default function Home() {
  const [activeTab, setActiveTab] = useState("graph");

  // Graph Explorer states
  const [entityType, setEntityType] = useState<string>("");
  const [entityId, setEntityId] = useState<string>("");
  const [relationType, setRelationType] = useState<string>("");
  const [relationTypes, setRelationTypes] = useState<string[]>([]);
  const [depth, setDepth] = useState<string>("2");
  const [graphData, setGraphData] = useState<GraphData>({
    nodes: [],
    links: [],
  });
  const [graphLoading, setGraphLoading] = useState<boolean>(false);
  const [graphError, setGraphError] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);

  // Permission Analyzer states
  const [subjectType, setSubjectType] = useState<string>("user");
  const [subjectId, setSubjectId] = useState<string>("");
  const [permission, setPermission] = useState<string>("");
  const [objectType, setObjectType] = useState<string>("organization");
  const [objectId, setObjectId] = useState<string>("");
  const [permissionLoading, setPermissionLoading] = useState<boolean>(false);
  const [permissionError, setPermissionError] = useState<string | null>(null);
  const [permissionResult, setPermissionResult] =
    useState<PermissionPathResponse | null>(null);
  const [selectedPath, setSelectedPath] = useState<Link[] | null>(null);

  useEffect(() => {
    fetchRelations();
  }, []);

  const fetchRelations = async () => {
    setGraphLoading(true);
    setGraphError(null);

    try {
      const response = await fetch(`/api/permission-definitions`);
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      console.log(data);
      setRelationTypes(data);
      setSelectedNode(null);
    } catch (err) {
      setGraphError(
        err instanceof Error ? err.message : "Failed to fetch graph data"
      );
      console.error("Error fetching graph data:", err);
    } finally {
      setGraphLoading(false);
    }
  };

  // Graph Explorer functionality
  const fetchGraphData = async () => {
    setGraphLoading(true);
    setGraphError(null);

    try {
      const params = new URLSearchParams();
      if (entityType) params.append("entity_type", entityType);
      if (entityId) params.append("entity_id", entityId);
      if (relationType) params.append("relation_type", relationType);
      params.append("depth", depth);

      const response = await fetch(`/api/graph?${params.toString()}`);
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      setGraphData(data);
      setSelectedNode(null);
    } catch (err) {
      setGraphError(
        err instanceof Error ? err.message : "Failed to fetch graph data"
      );
      console.error("Error fetching graph data:", err);
    } finally {
      setGraphLoading(false);
    }
  };

  // Permission Analyzer functionality
  const analyzePermission = async () => {
    if (!subjectId || !permission || !objectId) {
      setPermissionError("All fields are required");
      return;
    }

    setPermissionLoading(true);
    setPermissionError(null);

    try {
      const response = await fetch("/api/permission-path", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          subject_type: subjectType,
          subject_id: subjectId,
          permission: permission,
          object_type: objectType,
          object_id: objectId,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      setPermissionResult(data);
      setSelectedPath(null);
    } catch (err) {
      setPermissionError(
        err instanceof Error ? err.message : "Failed to analyze permission"
      );
      console.error("Error analyzing permission:", err);
    } finally {
      setPermissionLoading(false);
    }
  };

  // Handle node selection in the graph
  const handleNodeClick = (node: Node) => {
    setSelectedNode(node);
  };

  return (
    <Layout>
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-4xl font-bold">Identity Graph Visualization</h1>
          <p className="text-gray-400 mt-2">
            Explore identity relationships and analyze permission paths in your
            system
          </p>
        </div>

        {/* Custom styled tabs */}
        <div className="mb-6">
          <div className="flex w-full space-x-1 rounded-lg bg-[#1e1e2e]/50 p-1 border border-slate-700/30">
            <button
              className={classNames(
                activeTab === "graph" ? "bg-violet-500/10" : "",
                "w-full cursor-pointer rounded-md px-3 py-1.5 text-sm font-medium text-violet-400 outline-none"
              )}
              onClick={() => setActiveTab("graph")}
            >
              <div className="flex items-center justify-center gap-2">
                <Network className="h-4 w-4" />
                <span>Graph Explorer</span>
              </div>
            </button>
            <button
              className={classNames(
                activeTab === "permission" ? "bg-violet-500/10" : "",
                "w-full cursor-pointer rounded-md px-3 py-1.5 text-sm font-medium text-violet-400 outline-none"
              )}
              onClick={() => setActiveTab("permission")}
            >
              <div className="flex items-center justify-center gap-2">
                <Lock className="h-4 w-4" />
                <span>Permission Analysis</span>
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
        {activeTab === "graph" ? (
          <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
            <CardHeader className="border-b border-slate-700/40">
              <div className="flex items-center gap-2">
                <Network className="h-5 w-5 text-teal-400" />
                <CardTitle className="text-xl">Graph Explorer</CardTitle>
              </div>
              <CardDescription className="text-gray-400">
                Use the filters below to explore the identity graph
              </CardDescription>
            </CardHeader>
            <CardContent className="pt-6">
              {graphError && (
                <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-400 mb-4">
                  {graphError}
                </div>
              )}

              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1 text-gray-400">
                    Entity Type
                  </label>
                  <Select value={entityType} onValueChange={setEntityType}>
                    <SelectTrigger className="bg-[#2a2a3c] border-slate-700/40">
                      <SelectValue placeholder="All Types" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">All Types</SelectItem>
                      <SelectItem value="user">User</SelectItem>
                      <SelectItem value="organization">Organization</SelectItem>
                      <SelectItem value="group">Group</SelectItem>
                      <SelectItem value="project">Project</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1 text-gray-400">
                    Entity ID
                  </label>
                  <Input
                    type="text"
                    placeholder="e.g. alice"
                    value={entityId}
                    onChange={(e) => setEntityId(e.target.value)}
                    className="bg-[#2a2a3c] border-slate-700/40"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1 text-gray-400">
                    Relation Type
                  </label>
                  <Input
                    type="text"
                    placeholder="e.g. owner"
                    value={relationType}
                    onChange={(e) => setRelationType(e.target.value)}
                    className="bg-[#2a2a3c] border-slate-700/40"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1 text-gray-400">
                    Depth
                  </label>
                  <Select value={depth} onValueChange={setDepth}>
                    <SelectTrigger className="bg-[#2a2a3c] border-slate-700/40">
                      <SelectValue placeholder="2" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">1</SelectItem>
                      <SelectItem value="2">2</SelectItem>
                      <SelectItem value="3">3</SelectItem>
                      <SelectItem value="4">4</SelectItem>
                      <SelectItem value="5">5</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="mt-6">
                <Button
                  onClick={fetchGraphData}
                  disabled={graphLoading}
                  className="bg-violet-500 hover:bg-violet-600 text-white"
                >
                  {graphLoading ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      <span>Loading...</span>
                    </>
                  ) : (
                    <>
                      <Network className="h-4 w-4 mr-2" />
                      <span>Load Graph</span>
                    </>
                  )}
                </Button>
              </div>

              <div className="flex mt-8">
                {/* Graph container */}
                <div className="flex-1 graph-container h-[500px]">
                  {graphData.nodes.length > 0 ? (
                    <GraphVisualization
                      data={graphData}
                      onNodeClick={handleNodeClick}
                    />
                  ) : (
                    <div className="h-full flex flex-col items-center justify-center text-gray-400">
                      {graphLoading ? (
                        <>
                          <Loader2 className="h-8 w-8 animate-spin mb-4 text-teal-400" />
                          <p>Loading graph data...</p>
                        </>
                      ) : (
                        <>
                          <p className=" font-bold text-teal-500">
                            No graph data loaded
                          </p>
                          <p className="text-sm text-gray-500 mt-2">
                            Use the filters above to explore the identity graph
                          </p>
                        </>
                      )}
                    </div>
                  )}
                </div>

                {/* Node details panel */}
                {selectedNode && (
                  <div className="w-64 bg-[#1e1e2e] rounded-md border border-[#4b5563] p-4 h-fit ml-4">
                    <div className="flex items-center gap-2 mb-4">
                      <Eye className="h-4 w-4 text-teal-400" />
                      <h3 className="m-0 text-base font-semibold">
                        {selectedNode.label}
                      </h3>
                    </div>

                    <div className="text-sm mb-2">
                      <span className="text-gray-400">ID: </span>
                      <span className="bg-[#374151] px-2 py-1 rounded font-mono text-xs">
                        {selectedNode.id}
                      </span>
                    </div>

                    <div className="text-sm mb-2">
                      <span className="text-gray-400">Type: </span>
                      <span className="bg-violet-500/10 text-violet-400 px-2 py-1 rounded-full text-xs">
                        {selectedNode.type}
                      </span>
                    </div>

                    <Button
                      size="sm"
                      className="w-full mt-4 bg-teal-500 hover:bg-teal-600 text-white"
                      onClick={() => {
                        setEntityType(selectedNode.type);
                        setEntityId(selectedNode.id.split(":")[1]);
                        setSelectedNode(null);
                        setTimeout(fetchGraphData, 100);
                      }}
                    >
                      <Zap className="h-3 w-3 mr-1" />
                      Focus on this entity
                    </Button>
                  </div>
                )}
              </div>

              {graphData.nodes.length > 0 && (
                <div className="mt-4 p-4 bg-[#1e1e2e] rounded-md text-xs text-gray-400 border border-slate-700/40">
                  <div className="flex justify-between mb-2">
                    <span>
                      Nodes:{" "}
                      <strong className="text-white">
                        {graphData.nodes.length}
                      </strong>
                    </span>
                    <span>
                      Links:{" "}
                      <strong className="text-white">
                        {graphData.links.length}
                      </strong>
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span>
                      Entity Types:{" "}
                      <strong className="text-white">
                        {new Set(graphData.nodes.map((n) => n.type)).size}
                      </strong>
                    </span>
                    <span>
                      Relation Types:{" "}
                      <strong className="text-white">
                        {new Set(graphData.links.map((l) => l.type)).size}
                      </strong>
                    </span>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        ) : activeTab === "permission" ? (
          <DynamicPermissionAnalyzer />
        ) : (
          <RelationshipCreator />
        )}
      </div>
    </Layout>
  );
}
