// src/components/OrphanedRelationshipsView.tsx
import React, { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  AlertCircle,
  Loader2,
  RefreshCw,
  Plus,
  Shield,
  Wrench,
  Unlink,
  Check,
  X,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

// Interface for orphaned relationship
interface OrphanedRelationship {
  id: number;
  subject_type: string;
  subject_id: string;
  relation: string;
  object_type: string;
  object_id: string;
  created_at: string;
  missing_type: "subject" | "object" | "both";
}

// Interface for summary statistics
interface OrphanedSummary {
  total_orphaned: number;
  both_missing: number;
  subject_missing: number;
  object_missing: number;
}

const OrphanedRelationshipsView: React.FC = () => {
  const [orphanedRelationships, setOrphanedRelationships] = useState<
    OrphanedRelationship[]
  >([]);
  const [summary, setSummary] = useState<OrphanedSummary | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<"all" | "subject" | "object" | "both">(
    "all"
  );
  const [fixingRelations, setFixingRelations] = useState<
    Record<number, boolean>
  >({});

  // Fetch orphaned relationships data
  const fetchOrphanedRelationships = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/health/orphaned-relationships");
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      setOrphanedRelationships(data.orphaned_relationships || []);
      setSummary(data.summary || null);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to fetch orphaned relationships"
      );
      console.error("Error fetching orphaned relationships:", err);
    } finally {
      setLoading(false);
    }
  };

  // Fix an orphaned relationship by creating missing entities
  const fixOrphanedRelationship = async (relationId: number) => {
    setFixingRelations((prev) => ({ ...prev, [relationId]: true }));
    try {
      const response = await fetch("/api/health/fix-orphaned-relationship", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          relation_id: relationId,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      if (data.success) {
        // Remove the fixed relationship from the list
        setOrphanedRelationships((prev) =>
          prev.filter((rel) => rel.id !== relationId)
        );

        // Update summary counts
        if (summary) {
          const fixedRel = orphanedRelationships.find(
            (rel) => rel.id === relationId
          );
          if (fixedRel) {
            const newSummary = { ...summary };
            newSummary.total_orphaned--;

            if (fixedRel.missing_type === "both") {
              newSummary.both_missing--;
            } else if (fixedRel.missing_type === "subject") {
              newSummary.subject_missing--;
            } else if (fixedRel.missing_type === "object") {
              newSummary.object_missing--;
            }

            setSummary(newSummary);
          }
        }
      }
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to fix orphaned relationship"
      );
      console.error("Error fixing orphaned relationship:", err);
    } finally {
      setFixingRelations((prev) => ({ ...prev, [relationId]: false }));
    }
  };

  // Fix all orphaned relationships with a batch operation
  const fixAllOrphanedRelationships = async () => {
    const visibleRelationships = getFilteredRelationships();
    if (visibleRelationships.length === 0) return;

    const confirmMessage = `This will auto-create ${visibleRelationships.length} missing entities. Continue?`;
    if (!window.confirm(confirmMessage)) return;

    // Mark all as fixing
    const fixing: Record<number, boolean> = {};
    visibleRelationships.forEach((rel) => {
      fixing[rel.id] = true;
    });
    setFixingRelations(fixing);

    // Fix each one sequentially
    for (const rel of visibleRelationships) {
      try {
        await fixOrphanedRelationship(rel.id);
      } catch (err) {
        console.error(`Failed to fix relation ${rel.id}:`, err);
      }
    }

    // Refresh the list after fixing all
    fetchOrphanedRelationships();
  };

  // Apply the selected filter to the orphaned relationships
  const getFilteredRelationships = () => {
    if (filter === "all") {
      return orphanedRelationships;
    }
    return orphanedRelationships.filter((rel) => rel.missing_type === filter);
  };

  // Format the date into a readable string
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  // Load data on initial render
  useEffect(() => {
    fetchOrphanedRelationships();
  }, []);

  // Get filtered relationships
  const filteredRelationships = getFilteredRelationships();

  return (
    <div className="flex flex-col gap-6">
      <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
        <CardHeader className="border-b border-slate-700/40">
          <div className="flex items-center gap-2">
            <Unlink className="h-5 w-5 text-teal-400" />
            <CardTitle className="text-xl">Orphaned Relationships</CardTitle>
          </div>
          <CardDescription className="text-gray-400">
            Identify and fix relationships with missing entity records
          </CardDescription>
        </CardHeader>

        <CardContent className="pt-6">
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-400 mb-4 flex items-center gap-2">
              <AlertCircle className="h-4 w-4" />
              <p>{error}</p>
            </div>
          )}

          {loading && !orphanedRelationships.length ? (
            <div className="flex justify-center items-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-teal-400 mr-3" />
              <span className="text-gray-400">
                Loading orphaned relationships...
              </span>
            </div>
          ) : (
            <>
              {/* Summary Cards */}
              {summary && (
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                  <Card className="bg-[#2a2a3c] border-slate-700/40">
                    <CardContent className="p-4">
                      <div className="text-sm text-gray-400 mb-1">
                        Total Orphaned
                      </div>
                      <div className="text-2xl font-bold">
                        {summary.total_orphaned}
                      </div>
                    </CardContent>
                  </Card>

                  <Card className="bg-[#2a2a3c] border-slate-700/40">
                    <CardContent className="p-4">
                      <div className="text-sm text-gray-400 mb-1">
                        Subject Missing
                      </div>
                      <div className="text-2xl font-bold">
                        {summary.subject_missing}
                      </div>
                    </CardContent>
                  </Card>

                  <Card className="bg-[#2a2a3c] border-slate-700/40">
                    <CardContent className="p-4">
                      <div className="text-sm text-gray-400 mb-1">
                        Object Missing
                      </div>
                      <div className="text-2xl font-bold">
                        {summary.object_missing}
                      </div>
                    </CardContent>
                  </Card>

                  <Card className="bg-[#2a2a3c] border-slate-700/40">
                    <CardContent className="p-4">
                      <div className="text-sm text-gray-400 mb-1">
                        Both Missing
                      </div>
                      <div className="text-2xl font-bold">
                        {summary.both_missing}
                      </div>
                    </CardContent>
                  </Card>
                </div>
              )}

              {/* Filters */}
              <div className="mb-6">
                <Tabs
                  defaultValue="all"
                  value={filter}
                  onValueChange={(v) => setFilter(v as any)}
                >
                  <TabsList className="grid grid-cols-4 bg-[#1e1e2e]/50 border border-slate-700/30">
                    <TabsTrigger
                      value="all"
                      className="data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
                    >
                      All
                    </TabsTrigger>
                    <TabsTrigger
                      value="subject"
                      className="data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
                    >
                      Subject Missing
                    </TabsTrigger>
                    <TabsTrigger
                      value="object"
                      className="data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
                    >
                      Object Missing
                    </TabsTrigger>
                    <TabsTrigger
                      value="both"
                      className="data-[state=active]:bg-violet-500/10 data-[state=active]:text-violet-400"
                    >
                      Both Missing
                    </TabsTrigger>
                  </TabsList>
                </Tabs>
              </div>

              {/* Relationship Table */}
              {filteredRelationships.length > 0 ? (
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
                        <TableHead className="text-gray-400 font-medium">
                          Missing
                        </TableHead>
                        <TableHead className="text-gray-400 font-medium text-right">
                          Actions
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredRelationships.map((rel) => (
                        <TableRow
                          key={rel.id}
                          className="border-t border-slate-700/40"
                        >
                          <TableCell>
                            <div className="flex flex-col">
                              <code
                                className={`text-xs ${
                                  rel.missing_type === "subject" ||
                                  rel.missing_type === "both"
                                    ? "text-red-400"
                                    : ""
                                }`}
                              >
                                {rel.subject_type}:{rel.subject_id}
                              </code>
                              {rel.missing_type === "subject" ||
                              rel.missing_type === "both" ? (
                                <span className="text-xs text-red-400 flex items-center mt-1">
                                  <X className="h-3 w-3 mr-1" /> Missing
                                </span>
                              ) : (
                                <span className="text-xs text-green-400 flex items-center mt-1">
                                  <Check className="h-3 w-3 mr-1" /> Exists
                                </span>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge
                              variant="outline"
                              className="bg-violet-500/10 text-violet-400 border-violet-500/20"
                            >
                              {rel.relation}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <div className="flex flex-col">
                              <code
                                className={`text-xs ${
                                  rel.missing_type === "object" ||
                                  rel.missing_type === "both"
                                    ? "text-red-400"
                                    : ""
                                }`}
                              >
                                {rel.object_type}:{rel.object_id}
                              </code>
                              {rel.missing_type === "object" ||
                              rel.missing_type === "both" ? (
                                <span className="text-xs text-red-400 flex items-center mt-1">
                                  <X className="h-3 w-3 mr-1" /> Missing
                                </span>
                              ) : (
                                <span className="text-xs text-green-400 flex items-center mt-1">
                                  <Check className="h-3 w-3 mr-1" /> Exists
                                </span>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>
                            {rel.missing_type === "both" ? (
                              <Badge className="bg-red-500/10 text-red-400 border-red-500/20">
                                Both
                              </Badge>
                            ) : rel.missing_type === "subject" ? (
                              <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/20">
                                Subject
                              </Badge>
                            ) : (
                              <Badge className="bg-amber-500/10 text-amber-400 border-amber-500/20">
                                Object
                              </Badge>
                            )}
                          </TableCell>
                          <TableCell className="text-right">
                            <Button
                              variant="outline"
                              size="sm"
                              className="border-teal-500/30 text-teal-400 hover:bg-teal-500/10"
                              onClick={() => fixOrphanedRelationship(rel.id)}
                              disabled={fixingRelations[rel.id]}
                            >
                              {fixingRelations[rel.id] ? (
                                <Loader2 className="h-3 w-3 animate-spin mr-1" />
                              ) : (
                                <Wrench className="h-3 w-3 mr-1" />
                              )}
                              Auto Fix
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <div className="text-center py-12 border border-slate-700/40 rounded-md bg-[#2a2a3c]/50">
                  {orphanedRelationships.length > 0 ? (
                    <div className="text-gray-400">
                      <Shield className="h-12 w-12 mx-auto mb-4 text-gray-600" />
                      <p className="text-lg font-medium">
                        No relationships match the current filter
                      </p>
                      <p className="text-sm mt-1">
                        Try changing the filter or refreshing the data
                      </p>
                    </div>
                  ) : (
                    <div className="text-gray-400">
                      <Shield className="h-12 w-12 mx-auto mb-4 text-green-400" />
                      <p className="text-lg font-medium">
                        No orphaned relationships found
                      </p>
                      <p className="text-sm mt-1">
                        All relationships have valid entity records
                      </p>
                    </div>
                  )}
                </div>
              )}
            </>
          )}
        </CardContent>

        <CardFooter className="flex justify-between border-t border-slate-700/40 pt-4">
          <Button
            variant="outline"
            className="border-slate-700/40 text-gray-300 hover:bg-slate-700/20"
            onClick={fetchOrphanedRelationships}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <RefreshCw className="h-4 w-4 mr-2" />
            )}
            Refresh
          </Button>

          {filteredRelationships.length > 0 && (
            <Button
              className="bg-teal-600 hover:bg-teal-700 text-white"
              onClick={fixAllOrphanedRelationships}
              disabled={Object.values(fixingRelations).some((v) => v)}
            >
              <Plus className="h-4 w-4 mr-2" />
              Auto-Create All Missing Entities
            </Button>
          )}
        </CardFooter>
      </Card>
    </div>
  );
};

export default OrphanedRelationshipsView;
