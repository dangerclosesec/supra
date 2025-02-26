// src/components/PermissionTableVisualizer.tsx
import React, { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ShieldCheck,
  Network,
  Database,
  RefreshCw,
  Loader2,
  Search,
  AlertCircle,
  Info,
} from "lucide-react";

const PermissionTableVisualizer = () => {
  // State for data
  const [entityTypes, setEntityTypes] = useState([]);
  const [permissions, setPermissions] = useState([]);
  const [relations, setRelations] = useState([]);
  const [loading, setLoading] = useState({
    entities: true,
    permissions: true,
    relations: true,
  });
  const [error, setError] = useState(null);

  // Filtering state
  const [entityFilter, setEntityFilter] = useState("");
  const [permissionFilter, setPermissionFilter] = useState("");
  const [relationFilter, setRelationFilter] = useState("");

  // Fetch entity types
  const fetchEntityTypes = async () => {
    setLoading((prev) => ({ ...prev, entities: true }));
    setError(null);
    try {
      const response = await fetch("/api/entity-types");
      if (!response.ok)
        throw new Error(`HTTP error! Status: ${response.status}`);

      const data = await response.json();
      setEntityTypes(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error("Error fetching entity types:", err);
      setError(`Failed to fetch entity types: ${err.message}`);
    } finally {
      setLoading((prev) => ({ ...prev, entities: false }));
    }
  };

  // Fetch permission definitions
  const fetchPermissions = async () => {
    setLoading((prev) => ({ ...prev, permissions: true }));
    setError(null);
    try {
      const response = await fetch("/api/permission-definitions");
      if (!response.ok)
        throw new Error(`HTTP error! Status: ${response.status}`);

      const data = await response.json();
      setPermissions(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error("Error fetching permissions:", err);
      setError(`Failed to fetch permissions: ${err.message}`);
    } finally {
      setLoading((prev) => ({ ...prev, permissions: false }));
    }
  };

  // Fetch relations
  const fetchRelations = async () => {
    setLoading((prev) => ({ ...prev, relations: true }));
    setError(null);
    try {
      const response = await fetch("/api/relations?limit=100");
      if (!response.ok)
        throw new Error(`HTTP error! Status: ${response.status}`);

      const data = await response.json();
      setRelations(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error("Error fetching relations:", err);
      setError(`Failed to fetch relations: ${err.message}`);
    } finally {
      setLoading((prev) => ({ ...prev, relations: false }));
    }
  };

  // Fetch all data when component mounts
  useEffect(() => {
    fetchEntityTypes();
    fetchPermissions();
    fetchRelations();
  }, []);

  // Filter functions
  const filteredEntities = entityTypes.filter((entity) => {
    const searchTerm = entityFilter.toLowerCase();
    return (
      entity.type?.toLowerCase().includes(searchTerm) ||
      entity.display_name?.toLowerCase().includes(searchTerm)
    );
  });

  const filteredPermissions = permissions.filter((permission) => {
    const searchTerm = permissionFilter.toLowerCase();
    return (
      permission.entity_type?.toLowerCase().includes(searchTerm) ||
      permission.permission_name?.toLowerCase().includes(searchTerm) ||
      permission.condition_expression?.toLowerCase().includes(searchTerm) ||
      permission.description?.toLowerCase().includes(searchTerm)
    );
  });

  const filteredRelations = relations.filter((relation) => {
    const searchTerm = relationFilter.toLowerCase();
    return (
      relation.subject_type?.toLowerCase().includes(searchTerm) ||
      relation.subject_id?.toLowerCase().includes(searchTerm) ||
      relation.relation?.toLowerCase().includes(searchTerm) ||
      relation.object_type?.toLowerCase().includes(searchTerm) ||
      relation.object_id?.toLowerCase().includes(searchTerm)
    );
  });

  // Helper function to get an expression with highlighted keywords
  const getExpressionHighlighted = (expr) => {
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
    <div className="flex flex-col gap-6">
      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-400 mb-4 flex items-center gap-2">
          <AlertCircle className="h-4 w-4" />
          <p>{error}</p>
        </div>
      )}

      <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
        <CardHeader className="border-b border-slate-700/40">
          <div className="flex items-center gap-2">
            <ShieldCheck className="w-5 h-5 text-teal-400" />
            <CardTitle className="text-xl">
              Permission Schema Explorer
            </CardTitle>
          </div>
          <CardDescription className="text-gray-400">
            Explore your identity graph schema including entities, permissions,
            and relations
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-6">
          <Tabs defaultValue="entities" className="w-full">
            <TabsList className="w-full grid grid-cols-3 bg-[#1e1e2e]/50 border border-slate-700/30">
              <TabsTrigger
                value="entities"
                className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
              >
                <Database className="h-4 w-4" />
                Entities
              </TabsTrigger>
              <TabsTrigger
                value="permissions"
                className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
              >
                <ShieldCheck className="h-4 w-4" />
                Permissions
              </TabsTrigger>
              <TabsTrigger
                value="relations"
                className="gap-2 data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
              >
                <Network className="h-4 w-4" />
                Relations
              </TabsTrigger>
            </TabsList>

            {/* Entities Tab */}
            <TabsContent value="entities" className="mt-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="relative w-full max-w-sm">
                  <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
                  <Input
                    type="text"
                    placeholder="Filter entities..."
                    className="w-full pl-8 pr-4 bg-[#2a2a3c] border-slate-700/40"
                    value={entityFilter}
                    onChange={(e) => setEntityFilter(e.target.value)}
                  />
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="border-slate-700/40 text-gray-300 hover:bg-slate-700/20"
                  onClick={fetchEntityTypes}
                  disabled={loading.entities}
                >
                  {loading.entities ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <RefreshCw className="h-4 w-4" />
                  )}
                  Refresh
                </Button>
              </div>

              {loading.entities ? (
                <div className="flex justify-center items-center py-8">
                  <Loader2 className="h-8 w-8 animate-spin text-teal-400" />
                </div>
              ) : (
                <div className="rounded-md border border-slate-700/40 overflow-hidden">
                  <Table>
                    <TableHeader className="bg-[#2a2a3c]">
                      <TableRow>
                        <TableHead className="text-gray-400 font-medium">
                          Entity Type
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium">
                          Display Name
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium">
                          Count
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredEntities.length > 0 ? (
                        filteredEntities.map((entity, index) => (
                          <TableRow
                            key={index}
                            className="border-t border-slate-700/40"
                          >
                            <TableCell>
                              <Badge
                                variant="outline"
                                className="font-mono bg-violet-500/10 text-violet-400 border-violet-500/20"
                              >
                                {entity.type}
                              </Badge>
                            </TableCell>
                            <TableCell className="font-medium">
                              {entity.display_name || entity.type}
                            </TableCell>
                            <TableCell className="text-gray-400">
                              {entity.count || 0} instances
                            </TableCell>
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell
                            colSpan={3}
                            className="text-center text-gray-400 py-8"
                          >
                            No entities found
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              )}
            </TabsContent>

            {/* Permissions Tab */}
            <TabsContent value="permissions" className="mt-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="relative w-full max-w-sm">
                  <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
                  <Input
                    type="text"
                    placeholder="Filter permissions..."
                    className="w-full pl-8 pr-4 bg-[#2a2a3c] border-slate-700/40"
                    value={permissionFilter}
                    onChange={(e) => setPermissionFilter(e.target.value)}
                  />
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="border-slate-700/40 text-gray-300 hover:bg-slate-700/20"
                  onClick={fetchPermissions}
                  disabled={loading.permissions}
                >
                  {loading.permissions ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <RefreshCw className="h-4 w-4" />
                  )}
                  Refresh
                </Button>
              </div>

              {loading.permissions ? (
                <div className="flex justify-center items-center py-8">
                  <Loader2 className="h-8 w-8 animate-spin text-teal-400" />
                </div>
              ) : (
                <div className="rounded-md border border-slate-700/40 overflow-hidden">
                  <Table>
                    <TableHeader className="bg-[#2a2a3c]">
                      <TableRow>
                        <TableHead className="text-gray-400 font-medium">
                          Entity Type
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium">
                          Permission
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium">
                          Expression
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium">
                          Description
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredPermissions.length > 0 ? (
                        filteredPermissions.map((permission, index) => (
                          <TableRow
                            key={index}
                            className="border-t border-slate-700/40"
                          >
                            <TableCell>
                              <Badge
                                variant="outline"
                                className="font-mono bg-violet-500/10 text-violet-400 border-violet-500/20"
                              >
                                {permission.entity_type}
                              </Badge>
                            </TableCell>
                            <TableCell className="font-medium">
                              {permission.permission_name}
                            </TableCell>
                            <TableCell className="font-mono">
                              {getExpressionHighlighted(
                                permission.condition_expression
                              )}
                            </TableCell>
                            <TableCell className="text-gray-400">
                              {permission.description || "-"}
                            </TableCell>
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell
                            colSpan={4}
                            className="text-center text-gray-400 py-8"
                          >
                            No permissions found
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              )}
            </TabsContent>

            {/* Relations Tab */}
            <TabsContent value="relations" className="mt-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="relative w-full max-w-sm">
                  <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
                  <Input
                    type="text"
                    placeholder="Filter relations..."
                    className="w-full pl-8 pr-4 bg-[#2a2a3c] border-slate-700/40"
                    value={relationFilter}
                    onChange={(e) => setRelationFilter(e.target.value)}
                  />
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="border-slate-700/40 text-gray-300 hover:bg-slate-700/20"
                  onClick={fetchRelations}
                  disabled={loading.relations}
                >
                  {loading.relations ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <RefreshCw className="h-4 w-4" />
                  )}
                  Refresh
                </Button>
              </div>

              {loading.relations ? (
                <div className="flex justify-center items-center py-8">
                  <Loader2 className="h-8 w-8 animate-spin text-teal-400" />
                </div>
              ) : (
                <div className="rounded-md border border-slate-700/40 overflow-hidden">
                  <Table>
                    <TableHeader className="bg-[#2a2a3c]">
                      <TableRow>
                        <TableHead className="text-gray-400 font-medium">
                          Subject
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium">
                          Relation
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium">
                          Object
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredRelations.length > 0 ? (
                        filteredRelations.map((relation, index) => (
                          <TableRow
                            key={index}
                            className="border-t border-slate-700/40"
                          >
                            <TableCell>
                              <div className="flex flex-col gap-1">
                                <Badge
                                  variant="outline"
                                  className="font-mono text-xs bg-violet-500/10 text-violet-400 border-violet-500/20 w-fit"
                                >
                                  {relation.subject_type}
                                </Badge>
                                <span className="font-medium">
                                  {relation.subject_id}
                                </span>
                              </div>
                            </TableCell>
                            <TableCell className="font-semibold text-teal-400">
                              {relation.relation}
                            </TableCell>
                            <TableCell>
                              <div className="flex flex-col gap-1">
                                <Badge
                                  variant="outline"
                                  className="font-mono text-xs bg-violet-500/10 text-violet-400 border-violet-500/20 w-fit"
                                >
                                  {relation.object_type}
                                </Badge>
                                <span className="font-medium">
                                  {relation.object_id}
                                </span>
                              </div>
                            </TableCell>
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell
                            colSpan={3}
                            className="text-center text-gray-400 py-8"
                          >
                            No relations found
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              )}

              {filteredRelations.length > 0 && relations.length >= 100 && (
                <div className="flex items-center justify-center p-2 text-xs text-gray-400 gap-1">
                  <Info className="h-3 w-3" />
                  <span>
                    Showing first 100 relations. Apply filters to see more
                    specific results.
                  </span>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
        <CardFooter className="text-sm text-gray-400 border-t border-slate-700/40 mt-2 flex justify-between">
          <div className="flex items-center gap-2">
            <Database className="h-3 w-3" />
            {entityTypes.length} Entity Types
          </div>
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-3 w-3" />
            {permissions.length} Permissions
          </div>
          <div className="flex items-center gap-2">
            <Network className="h-3 w-3" />
            {relations.length} Relations
          </div>
        </CardFooter>
      </Card>
    </div>
  );
};

export default PermissionTableVisualizer;
