// src/components/DynamicPermissionAnalyzer.tsx
import React, { useState, useEffect, useMemo, useCallback } from "react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
  Info,
} from "lucide-react";

// Define interface for permission definition data
interface PermissionDefinition {
  id: number;
  entity_type: string;
  permission_name: string;
  condition_expression: string;
  created_at: string;
}

interface DynamicPermissionAnalyzerProps {
  className?: string;
}

const DynamicPermissionAnalyzer: React.FC<DynamicPermissionAnalyzerProps> = ({
  className,
}) => {
  // State for form fields
  const [subjectType, setSubjectType] = useState<string>("");
  const [subjectId, setSubjectId] = useState<string>("");
  const [permission, setPermission] = useState<string>("");
  const [objectType, setObjectType] = useState<string>("");
  const [objectId, setObjectId] = useState<string>("");

  // State for API data
  const [permissionDefinitions, setPermissionDefinitions] = useState<
    PermissionDefinition[]
  >([]);
  const [initialLoading, setInitialLoading] = useState<boolean>(true);

  // State for loading/error handling
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // State for results
  const [result, setResult] = useState<PermissionPathResponse | null>(null);
  const [selectedPath, setSelectedPath] = useState<Link[] | null>(null);

  // Fetch permission definitions on component mount
  useEffect(() => {
    fetchPermissionDefinitions();
  }, []);

  // Memoized fetch function to prevent double renders
  const fetchPermissionDefinitions = useCallback(async () => {
    setInitialLoading(true);
    setError(null);

    try {
      const response = await fetch(`/api/permission-definitions`);
      if (!response.ok && response.status !== 404) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      if (data.error) {
        throw new Error(data.error);
      }

      setPermissionDefinitions(data);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to fetch permission definitions"
      );
      console.error("Error fetching permission definitions:", err);
    } finally {
      setInitialLoading(false);
    }
  }, []);

  // All available entity types from API data
  const entityTypes = useMemo(() => {
    const types = new Set(permissionDefinitions.map((def) => def.entity_type));
    return Array.from(types).sort();
  }, [permissionDefinitions]);

  // All available permissions from API data
  const availablePermissions = useMemo(() => {
    if (!objectType) return [];

    // Find permissions where the entity_type matches the target type
    const permissions = permissionDefinitions
      .filter((def) => def.entity_type === objectType)
      .map((def) => def.permission_name);

    return Array.from(new Set(permissions)).sort();
  }, [permissionDefinitions, objectType]);

  // Get the current permission definition being used
  const currentPermissionDefinition = useMemo(() => {
    if (!subjectType || !permission) return null;

    return permissionDefinitions.find(
      (def) =>
        def.entity_type === subjectType && def.permission_name === permission
    );
  }, [permissionDefinitions, subjectType, permission]);

  // Handle form field changes
  const handleObjectTypeChange = (value: string) => {
    setObjectType(value);
    setPermission(""); // Reset permission when target type changes
    setResult(null);
  };

  const handlePermissionChange = (value: string) => {
    setPermission(value);
    setResult(null);
  };

  const handleSubjectTypeChange = (value: string) => {
    setSubjectType(value);
    setResult(null);
  };

  // Analyze permission
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

      if (!response.ok && response.status !== 404) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();

      if (data.error) {
        throw new Error(data.error);
      }

      setResult(data);
      setSelectedPath(null);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to analyze permission"
      );
      console.error("Error analyzing permission:", err);
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  // Format expression for display
  const getExpressionHighlighted = (expr: string): React.ReactNode => {
    if (!expr) return null;

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
      <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
        <CardHeader className="border-b border-slate-700/40">
          <div className="flex items-center gap-2">
            <Fingerprint className="h-5 w-5 text-teal-400" />
            <CardTitle className="text-xl">Permission Analysis</CardTitle>
          </div>
          <CardDescription className="text-gray-400">
            Analyze permission paths between subjects and objects
          </CardDescription>
        </CardHeader>

        <CardContent className="pt-6">
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-400 mb-4 flex items-center gap-2">
              <ShieldAlert className="h-4 w-4" />
              <p>{error}</p>
            </div>
          )}

          {initialLoading ? (
            <div className="flex justify-center items-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-teal-400 mr-3" />
              <span className="text-gray-400">
                Loading permission definitions...
              </span>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
              <div className="bg-[#2a2a3c] rounded-md p-3 mb-4 text-sm text-gray-300 border border-slate-700/40">
                <span className="font-medium text-teal-400">
                  How Permissions Work:
                </span>
                <p className="mt-1">
                  Permissions determine if an{" "}
                  <span className="text-violet-400">Actor</span> (user, system)
                  can perform an <span className="text-violet-400">Action</span>{" "}
                  (view, edit) on a{" "}
                  <span className="text-violet-400">Target</span> (organization,
                  project).
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1 text-gray-400">
                  Target Type
                  <span className="text-xs block text-gray-500 mt-0.5">
                    <b>Object Type:</b> What is being acted upon
                  </span>
                </label>
                <Select
                  value={objectType}
                  onValueChange={handleObjectTypeChange}
                  disabled={loading}
                >
                  <SelectTrigger className="bg-[#2a2a3c] border-slate-700/40">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    {entityTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {type}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1 text-gray-400">
                  Target ID
                  <span className="text-xs block text-gray-500 mt-0.5">
                    <b>Object ID:</b> Specific ID of the target
                  </span>
                </label>
                <Input
                  type="text"
                  placeholder="e.g. acme"
                  value={objectId}
                  onChange={(e) => setObjectId(e.target.value)}
                  className="bg-[#2a2a3c] border-slate-700/40"
                  disabled={loading}
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1 text-gray-400">
                  Permission
                </label>
                <Select
                  value={permission}
                  onValueChange={handlePermissionChange}
                  disabled={loading || !objectType}
                >
                  <SelectTrigger className="bg-[#2a2a3c] border-slate-700/40">
                    <SelectValue placeholder="Select permission" />
                  </SelectTrigger>
                  <SelectContent>
                    {availablePermissions.map((perm) => (
                      <SelectItem key={perm} value={perm}>
                        {perm}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                {objectType && availablePermissions.length > 0 && (
                  <div className="mt-4 bg-teal-500/10 border border-teal-500/20 rounded-md p-3 text-teal-400">
                    <div className="flex items-center gap-2 text-sm">
                      <Info className="h-4 w-4" />
                      <p>
                        Found {availablePermissions.length} possible{" "}
                        {availablePermissions.length === 1
                          ? "action"
                          : "actions"}
                        that can be performed on a {objectType}.
                      </p>
                    </div>
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium mb-1 text-gray-400">
                  Actor Type
                  <span className="text-xs block text-gray-500 mt-0.5">
                    <b>Subject:</b> Who is performing the action
                  </span>
                </label>
                <Select
                  value={subjectType}
                  onValueChange={handleSubjectTypeChange}
                  disabled={loading || entityTypes.length === 0}
                >
                  <SelectTrigger className="bg-[#2a2a3c] border-slate-700/40">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    {entityTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {type}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1 text-gray-400">
                  Actor ID
                  <span className="text-xs block text-gray-500 mt-0.5">
                    <b>Subject ID:</b> Specific ID of the actor
                  </span>
                </label>
                <Input
                  type="text"
                  placeholder="e.g. alice"
                  value={subjectId}
                  onChange={(e) => setSubjectId(e.target.value)}
                  className="bg-[#2a2a3c] border-slate-700/40"
                  disabled={loading || !subjectType}
                />
              </div>
            </div>
          )}

          {currentPermissionDefinition && (
            <div className="mt-4 p-3 bg-[#2a2a3c] rounded-md border border-slate-700/40">
              <div className="flex items-center gap-2 text-sm">
                <Info className="h-4 w-4 text-teal-400" />
                <span className="text-gray-300">Permission rule:</span>
                <span className="font-mono text-xs bg-[#374151] px-2 py-1 rounded">
                  {getExpressionHighlighted(
                    currentPermissionDefinition.condition_expression
                  )}
                </span>
              </div>
            </div>
          )}
        </CardContent>

        <CardFooter className="flex justify-end">
          <Button
            onClick={analyzePermission}
            disabled={
              loading ||
              !subjectType ||
              !subjectId ||
              !permission ||
              !objectType ||
              !objectId
            }
            size="lg"
            className="bg-violet-500 hover:bg-violet-600 text-white"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Analyzing...
              </>
            ) : (
              <>
                <Lock className="h-4 w-4 mr-2" />
                Analyze Permission
              </>
            )}
          </Button>
        </CardFooter>
      </Card>

      {result && (
        <div className="space-y-6">
          <div
            className={`p-4 rounded-md flex items-center gap-3 border ${
              result.allowed
                ? "bg-green-500/10 text-green-400 border-green-500/30"
                : "bg-red-500/10 text-red-400 border-red-500/30"
            }`}
          >
            {result.allowed ? (
              <ShieldCheck className="h-5 w-5" />
            ) : (
              <ShieldAlert className="h-5 w-5" />
            )}
            <span className="font-medium">
              {result.allowed
                ? "✅ Permission Granted"
                : "❌ Permission Denied"}
            </span>
          </div>

          <Tabs defaultValue="paths" className="w-full">
            <TabsList className="w-full grid grid-cols-2 bg-[#1e1e2e]/50 border border-slate-700/30">
              <TabsTrigger
                value="paths"
                className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
              >
                <Network className="h-4 w-4" />
                Permission Paths
              </TabsTrigger>
              <TabsTrigger
                value="expression"
                className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
              >
                <Lightbulb className="h-4 w-4" />
                Expression Details
              </TabsTrigger>
            </TabsList>

            <TabsContent value="paths" className="mt-6">
              <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Target className="h-5 w-5 text-teal-400" />
                    {result.paths && result.paths.length > 0
                      ? `Found ${result.paths.length} Permission Path${
                          result.paths.length !== 1 ? "s" : ""
                        }`
                      : "No Permission Paths Found"}
                  </CardTitle>
                  <CardDescription className="text-gray-400">
                    Select a path to highlight it in the graph
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {result.paths && result.paths.length > 0 ? (
                    <div className="space-y-3">
                      {result.paths.map((path, index) => (
                        <div
                          key={index}
                          className={`p-3 rounded-md cursor-pointer transition-all hover:bg-violet-500/5 ${
                            selectedPath === path
                              ? "bg-violet-500/10 border border-violet-500/20"
                              : "bg-[#2a2a3c] hover:bg-violet-500/5"
                          }`}
                          onClick={() => setSelectedPath(path)}
                        >
                          <div className="flex justify-between items-center mb-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              <ArrowDownLeft className="h-3 w-3 text-teal-400" />
                              Path {index + 1}
                            </div>
                            <span className="text-xs text-gray-400">
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
                                  <span className="text-violet-400 mx-1 flex items-center">
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
                    <div className="text-gray-400 flex items-center gap-2 bg-[#2a2a3c] p-4 rounded-md">
                      <ShieldAlert className="h-4 w-4" />
                      No permission paths found that would grant access.
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="expression" className="mt-6">
              <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Lightbulb className="h-5 w-5 text-teal-400" />
                    Permission Expression
                  </CardTitle>
                  <CardDescription className="text-gray-400">
                    Detailed breakdown of the permission rule
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="font-mono text-sm bg-[#2a2a3c] p-4 rounded-md overflow-x-auto">
                    {getExpressionHighlighted(result.expression)}
                  </div>

                  <div className="mt-4 text-sm">
                    <h4 className="font-medium text-white">
                      How expressions work:
                    </h4>
                    <ul className="mt-2 space-y-1 list-disc pl-5 text-gray-400">
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
            {result && result.nodes && result.nodes.length > 0 ? (
              <GraphVisualization
                data={{
                  nodes: result.nodes || [],
                  links: result.links || [],
                }}
                highlightPath={selectedPath || undefined}
              />
            ) : (
              <div className="h-full flex items-center justify-center text-gray-400">
                <div className="text-center">
                  <Network className="h-16 w-16 mb-4 text-gray-600 mx-auto" />
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

export default DynamicPermissionAnalyzer;
