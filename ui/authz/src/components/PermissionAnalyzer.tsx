// src/components/PermissionAnalyzer.tsx
import React, { useState } from "react";
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
import GraphVisualization from "./GraphVisualization";
import { Link, Node, PermissionPathResponse } from "@/types/graph";
import {
  Loader2,
  ShieldAlert,
  ShieldCheck,
  Network,
  ArrowDownLeft,
  ChevronRight,
  Lightbulb,
  Fingerprint,
  Lock,
  Target,
} from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";

interface PermissionAnalyzerProps {
  className?: string;
}

const PermissionAnalyzer: React.FC<PermissionAnalyzerProps> = ({
  className,
}) => {
  const [subjectType, setSubjectType] = useState<string>("user");
  const [subjectId, setSubjectId] = useState<string>("");
  const [permission, setPermission] = useState<string>("");
  const [objectType, setObjectType] = useState<string>("organization");
  const [objectId, setObjectId] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<PermissionPathResponse | null>(null);
  const [selectedPath, setSelectedPath] = useState<Link[] | null>(null);

  const analyzePermission = async () => {
    if (!subjectId || !permission || !objectId) {
      setError("All fields are required");
      return;
    }

    setLoading(true);
    setError(null);
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
        console.log("response", response);
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      setResult(data);
      setSelectedPath(null);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to analyze permission"
      );
      console.error("Error analyzing permission:", err);
    } finally {
      setLoading(false);
    }
  };

  const getExpressionHighlighted = (expr: string): React.ReactNode => {
    // Replace 'and' with styled version
    const andReplaced = expr.replace(
      / and /g,
      ' <span class="text-red-400 font-semibold">and</span> '
    );

    // Replace 'or' with styled version
    const orReplaced = andReplaced.replace(
      / or /g,
      ' <span class="text-green-400 font-semibold">or</span> '
    );

    // Highlight relation references
    const relationReplaced = orReplaced.replace(
      /([a-zA-Z_]+)\.([a-zA-Z_]+)/g,
      '<span class="text-blue-400">$1</span>.<span class="text-violet-400">$2</span>'
    );

    return <div dangerouslySetInnerHTML={{ __html: relationReplaced }} />;
  };

  return (
    <div className={`flex flex-col gap-6 ${className}`}>
      <Card className="border border-border/40 bg-card/30 backdrop-blur">
        <CardHeader>
          <CardTitle className="flex items-center text-xl gap-2">
            <Fingerprint className="w-5 h-5 text-accent" />
            Permission Analysis
          </CardTitle>
          <CardDescription>
            Analyze permission paths between subjects and objects
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Subject Type
              </label>
              <Select value={subjectType} onValueChange={setSubjectType}>
                <SelectTrigger className="bg-secondary/50 border-border/60">
                  <SelectValue placeholder="User" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="organization">Organization</SelectItem>
                  <SelectItem value="group">Group</SelectItem>
                  <SelectItem value="project">Project</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Subject ID
              </label>
              <Input
                type="text"
                placeholder="e.g. alice"
                value={subjectId}
                onChange={(e) => setSubjectId(e.target.value)}
                className="bg-secondary/50 border-border/60"
              />
            </div>

            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Permission
              </label>
              <Input
                type="text"
                placeholder="e.g. manage_settings"
                value={permission}
                onChange={(e) => setPermission(e.target.value)}
                className="bg-secondary/50 border-border/60"
              />
            </div>

            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Object Type
              </label>
              <Select value={objectType} onValueChange={setObjectType}>
                <SelectTrigger className="bg-secondary/50 border-border/60">
                  <SelectValue placeholder="Organization" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="organization">Organization</SelectItem>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="group">Group</SelectItem>
                  <SelectItem value="project">Project</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="md:col-span-1">
              <label className="block text-sm font-medium mb-1 text-muted-foreground">
                Object ID
              </label>
              <Input
                type="text"
                placeholder="e.g. acme"
                value={objectId}
                onChange={(e) => setObjectId(e.target.value)}
                className="bg-secondary/50 border-border/60"
              />
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex justify-end">
          <Button
            onClick={analyzePermission}
            disabled={loading}
            size="lg"
            className="gap-2"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Analyzing...
              </>
            ) : (
              <>
                <Lock className="h-4 w-4" />
                Analyze Permission
              </>
            )}
          </Button>
        </CardFooter>
      </Card>

      {error && (
        <Card className="bg-destructive/10 border-destructive/30">
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <ShieldAlert className="text-destructive h-5 w-5" />
              <p className="font-medium text-destructive">{error}</p>
            </div>
          </CardContent>
        </Card>
      )}

      {result && (
        <div className="space-y-6">
          <div
            className={`p-4 rounded-md flex items-center gap-3 border ${
              result.allowed
                ? "bg-accent/10 text-accent-foreground border-accent/30"
                : "bg-destructive/10 text-destructive border-destructive/30"
            }`}
          >
            {result.allowed ? (
              <ShieldCheck className="h-5 w-5 text-accent" />
            ) : (
              <ShieldAlert className="h-5 w-5 text-destructive" />
            )}
            <span className="font-medium">
              {result.allowed
                ? "✅ Permission Granted"
                : "❌ Permission Denied"}
            </span>
          </div>

          <Tabs defaultValue="paths" className="w-full">
            <TabsList className="w-full grid grid-cols-2">
              <TabsTrigger value="paths" className="gap-2">
                <Network className="h-4 w-4" />
                Permission Paths
              </TabsTrigger>
              <TabsTrigger value="expression" className="gap-2">
                <Lightbulb className="h-4 w-4" />
                Expression Details
              </TabsTrigger>
            </TabsList>

            <TabsContent value="paths" className="mt-6">
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Target className="h-5 w-5 text-accent" />
                    {result.paths && result.paths.length > 0
                      ? `Found ${result.paths.length} Permission Path${
                          result.paths.length !== 1 ? "s" : ""
                        }`
                      : "No Permission Paths Found"}
                  </CardTitle>
                  <CardDescription>
                    Select a path to highlight it in the graph
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {result.paths && result.paths.length > 0 ? (
                    <div className="space-y-3">
                      {result.paths.map((path, index) => (
                        <div
                          key={index}
                          className={`p-3 rounded-md cursor-pointer transition-all hover:bg-primary/5 ${
                            selectedPath === path
                              ? "bg-primary/10 border border-primary/20"
                              : "bg-secondary/50 hover:bg-primary/5"
                          }`}
                          onClick={() => setSelectedPath(path)}
                        >
                          <div className="flex justify-between items-center mb-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              <ArrowDownLeft className="h-3 w-3 text-accent" />
                              Path {index + 1}
                            </div>
                            <span className="text-xs text-muted-foreground">
                              {path.length} hop{path.length !== 1 ? "s" : ""}
                            </span>
                          </div>
                          <div className="flex items-center flex-wrap gap-1 text-sm">
                            {path.map((link, linkIndex) => {
                              const sourceId =
                                typeof link.source === "string"
                                  ? link.source.split(":")[1]
                                  : (link.source as any).id.split(":")[1];

                              const targetId =
                                typeof link.target === "string"
                                  ? link.target.split(":")[1]
                                  : (link.target as any).id.split(":")[1];

                              return (
                                <React.Fragment key={linkIndex}>
                                  <span className="font-medium">
                                    {sourceId}
                                  </span>
                                  <span className="text-primary mx-1 flex items-center">
                                    <ChevronRight className="h-3 w-3" />
                                    <span className="text-xs italic">
                                      {link.type}
                                    </span>
                                  </span>
                                  {linkIndex === path.length - 1 && (
                                    <span className="font-medium">
                                      {targetId}
                                    </span>
                                  )}
                                </React.Fragment>
                              );
                            })}
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-muted-foreground flex items-center gap-2 bg-secondary/30 p-4 rounded-md">
                      <ShieldAlert className="h-4 w-4" />
                      No permission paths found that would grant access.
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="expression" className="mt-6">
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Lightbulb className="h-5 w-5 text-accent" />
                    Permission Expression
                  </CardTitle>
                  <CardDescription>
                    Detailed breakdown of the permission rule
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="font-mono text-sm bg-secondary/40 p-4 rounded-md overflow-x-auto">
                    {getExpressionHighlighted(result.expression)}
                  </div>

                  <div className="mt-4 text-sm">
                    <h4 className="font-medium">How expressions work:</h4>
                    <ul className="mt-2 space-y-1 list-disc pl-5 text-muted-foreground">
                      <li>
                        <span className="text-green-400 font-semibold">or</span>{" "}
                        - If any condition is true, access is granted
                      </li>
                      <li>
                        <span className="text-red-400 font-semibold">and</span>{" "}
                        - All conditions must be true for access
                      </li>
                      <li>
                        <span className="text-blue-400">entity</span>.
                        <span className="text-violet-400">relation</span> -
                        Follows a relation path
                      </li>
                    </ul>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>

          <div className="h-[500px] graph-container">
            {result.nodes.length > 0 ? (
              <GraphVisualization
                data={{ nodes: result.nodes, links: result.links }}
                highlightPath={selectedPath || undefined}
              />
            ) : (
              <div className="h-full flex items-center justify-center text-muted-foreground">
                <div className="text-center">
                  <Network className="h-16 w-16 mb-4 text-muted-foreground/40 mx-auto" />
                  <p>No graph data available for visualization</p>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default PermissionAnalyzer;
