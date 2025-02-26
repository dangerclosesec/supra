// src/components/GraphExplorer.tsx
import React, { useState } from "react";
import GraphVisualization from "./GraphVisualization";
import { Button } from "./ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Input } from "./ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";
import { GraphData, Node } from "@/types/graph";
import {
  Loader2,
  RefreshCw,
  Search,
  Eye,
  Share2,
  Zap,
  Filter,
} from "lucide-react";

interface GraphExplorerProps {
  className?: string;
}

const GraphExplorer: React.FC<GraphExplorerProps> = ({ className }) => {
  const [entityType, setEntityType] = useState<string>("");
  const [entityId, setEntityId] = useState<string>("");
  const [relationType, setRelationType] = useState<string>("");
  const [depth, setDepth] = useState<string>("2");
  const [graphData, setGraphData] = useState<GraphData>({
    nodes: [],
    links: [],
  });
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);

  const fetchGraphData = async () => {
    setLoading(true);
    setError(null);
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
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to fetch graph data"
      );
      console.error("Error fetching graph data:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleNodeClick = (node: Node) => {
    setSelectedNode(node);
  };

  return (
    <div className={`flex flex-col gap-6 ${className}`}>
      <Card className="border border-border/40 bg-card/30 backdrop-blur">
        <CardHeader>
          <CardTitle className="flex items-center text-xl gap-2">
            <Filter className="w-5 h-5 text-accent" />
            Filters
          </CardTitle>
          <CardDescription>
            Select criteria to explore the identity graph
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Entity Type
              </label>
              <Select value={entityType} onValueChange={setEntityType}>
                <SelectTrigger className="bg-secondary/50 border-border/60">
                  <SelectValue placeholder="All Types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Types</SelectItem>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="organization">Organization</SelectItem>
                  <SelectItem value="group">Group</SelectItem>
                  <SelectItem value="project">Project</SelectItem>
                  <SelectItem value="task">Task</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Entity ID
              </label>
              <Input
                type="text"
                placeholder="e.g. alice"
                value={entityId}
                onChange={(e) => setEntityId(e.target.value)}
                className="bg-secondary/50 border-border/60"
              />
            </div>

            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Relation Type
              </label>
              <Input
                type="text"
                placeholder="e.g. owner"
                value={relationType}
                onChange={(e) => setRelationType(e.target.value)}
                className="bg-secondary/50 border-border/60"
              />
            </div>

            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Depth
              </label>
              <Select value={depth} onValueChange={setDepth}>
                <SelectTrigger className="bg-secondary/50 border-border/60">
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

            <div className="md:col-span-1 flex items-end">
              <Button
                onClick={fetchGraphData}
                className="w-full gap-2"
                disabled={loading}
              >
                {loading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Loading...
                  </>
                ) : (
                  <>
                    <Search className="h-4 w-4" />
                    Load Graph
                  </>
                )}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {error && (
        <div className="bg-destructive/10 text-destructive p-4 rounded-md border border-destructive/20">
          {error}
        </div>
      )}

      <div className="flex gap-6 h-[calc(100vh-24rem)] relative">
        <div className="flex-1 graph-container">
          {graphData.nodes.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-muted-foreground">
              {loading ? (
                <>
                  <Loader2 className="h-8 w-8 animate-spin mb-4 text-accent" />
                  <p>Loading graph data...</p>
                </>
              ) : (
                <>
                  <Share2 className="h-16 w-16 mb-4 text-muted-foreground/40" />
                  <p>No graph data loaded</p>
                  <p className="text-sm mt-2">
                    Use the filters above to explore the identity graph
                  </p>
                </>
              )}
            </div>
          ) : (
            <GraphVisualization
              data={graphData}
              onNodeClick={handleNodeClick}
            />
          )}
        </div>

        {selectedNode && (
          <Card className="w-80 h-fit shadow-lg border-primary/20">
            <CardHeader className="pb-2">
              <CardTitle className="text-lg flex items-center gap-2">
                <Eye className="h-4 w-4 text-accent" />
                {selectedNode.label}
              </CardTitle>
              <CardDescription>Entity Details</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">ID:</span>
                  <span className="font-mono bg-secondary/50 px-2 py-0.5 rounded text-xs">
                    {selectedNode.id}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Type:</span>
                  <span className="bg-primary/10 text-primary px-2 py-0.5 rounded-full text-xs">
                    {selectedNode.type}
                  </span>
                </div>
                <div className="flex justify-between pt-2">
                  <span className="text-muted-foreground">Connections:</span>
                  <span className="font-medium">
                    {
                      graphData.links.filter(
                        (link) =>
                          link.source === selectedNode.id ||
                          link.target === selectedNode.id
                      ).length
                    }
                  </span>
                </div>
              </div>
            </CardContent>
            <CardFooter className="flex justify-between pt-2">
              <Button
                size="sm"
                variant="outline"
                className="gap-1"
                onClick={() => {
                  setEntityType(selectedNode.type);
                  setEntityId(selectedNode.id.split(":")[1]);
                  setSelectedNode(null);
                  setTimeout(fetchGraphData, 100);
                }}
              >
                <Zap className="h-3 w-3" />
                Focus
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setSelectedNode(null)}
              >
                Close
              </Button>
            </CardFooter>
          </Card>
        )}
      </div>

      {graphData.nodes.length > 0 && (
        <Card className="border-border/40 bg-card/30 backdrop-blur">
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <Zap className="h-4 w-4 text-accent" />
              Statistics & Legend
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <h4 className="text-sm font-medium mb-2 text-muted-foreground">
                  Entity Types
                </h4>
                <div className="space-y-2">
                  <div className="flex items-center">
                    <div className="w-3 h-3 rounded-full bg-[#4285F4] mr-2"></div>
                    <span className="text-xs">User</span>
                  </div>
                  <div className="flex items-center">
                    <div className="w-3 h-3 rounded-full bg-[#EA4335] mr-2"></div>
                    <span className="text-xs">Organization</span>
                  </div>
                  <div className="flex items-center">
                    <div className="w-3 h-3 rounded-full bg-[#FBBC05] mr-2"></div>
                    <span className="text-xs">Group</span>
                  </div>
                  <div className="flex items-center">
                    <div className="w-3 h-3 rounded-full bg-[#34A853] mr-2"></div>
                    <span className="text-xs">Project</span>
                  </div>
                </div>
              </div>

              <div>
                <h4 className="text-sm font-medium mb-2 text-muted-foreground">
                  Relation Types
                </h4>
                <div className="space-y-2">
                  <div className="flex items-center">
                    <div className="w-3 h-3 bg-[#AA00FF] mr-2"></div>
                    <span className="text-xs">Owner</span>
                  </div>
                  <div className="flex items-center">
                    <div className="w-3 h-3 bg-[#00C853] mr-2"></div>
                    <span className="text-xs">Admin</span>
                  </div>
                  <div className="flex items-center">
                    <div className="w-3 h-3 bg-[#2979FF] mr-2"></div>
                    <span className="text-xs">Member</span>
                  </div>
                  <div className="flex items-center">
                    <div className="w-3 h-3 bg-[#FF6D00] mr-2"></div>
                    <span className="text-xs">Manager</span>
                  </div>
                </div>
              </div>

              <div className="border-l border-border/40 pl-6">
                <h4 className="text-sm font-medium mb-2 text-muted-foreground">
                  Graph Summary
                </h4>
                <dl className="space-y-2 text-xs">
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Nodes:</dt>
                    <dd className="font-medium">{graphData.nodes.length}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Connections:</dt>
                    <dd className="font-medium">{graphData.links.length}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Entity Types:</dt>
                    <dd className="font-medium">
                      {new Set(graphData.nodes.map((n) => n.type)).size}
                    </dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Relation Types:</dt>
                    <dd className="font-medium">
                      {new Set(graphData.links.map((l) => l.type)).size}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>
          </CardContent>
          <CardFooter className="text-xs text-muted-foreground border-t border-border/40 mt-2">
            <div className="flex items-center gap-2">
              <RefreshCw className="h-3 w-3" />
              Drag nodes to rearrange • Double-click to zoom
            </div>
          </CardFooter>
        </Card>
      )}
    </div>
  );
};

export default GraphExplorer;
