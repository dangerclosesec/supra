import { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { DatePickerWithRange } from "@/components/ui/date-range-picker";
import {
  Search,
  AlertCircle,
  CheckCircle,
  XCircle,
  Filter,
  RefreshCw,
  ClipboardList,
} from "lucide-react";
import { Pagination } from "@/components/ui/pagination";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { formatDistance } from "date-fns";

// Type definitions for the audit log entries
type AuditLog = {
  id: string;
  timestamp: string;
  action_type: string;
  result?: boolean;
  entity_type: string;
  entity_id: string;
  subject_type?: string;
  subject_id?: string;
  relation?: string;
  permission?: string;
  context?: Record<string, any>;
  request_id?: string;
  client_ip?: string;
  user_agent?: string;
};

type AuditLogResponse = {
  logs: AuditLog[];
  total: number;
};

const actionTypeOptions = [
  { value: "permission_check", label: "Permission Check" },
  { value: "entity_create", label: "Entity Create" },
  { value: "entity_delete", label: "Entity Delete" },
  { value: "relation_create", label: "Relation Create" },
  { value: "relation_delete", label: "Relation Delete" },
];

const resultOptions = [
  { value: "true", label: "Allowed" },
  { value: "false", label: "Denied" },
];

export default function AuditLogViewer() {
  // Filter state
  const [filters, setFilters] = useState({
    action_type: "",
    entity_type: "",
    entity_id: "",
    subject_type: "",
    subject_id: "",
    result: "",
    date_range: { from: undefined, to: undefined },
  });

  // Pagination state
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Data state
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Detail view state
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);

  // Fetch logs from the API
  const fetchLogs = async () => {
    try {
      setLoading(true);
      setError(null);

      // Calculate offset from page and page size
      const offset = (page - 1) * pageSize;

      // Build query parameters
      const params = new URLSearchParams();
      if (filters.action_type)
        params.append("action_type", filters.action_type);
      if (filters.entity_type)
        params.append("entity_type", filters.entity_type);
      if (filters.entity_id) params.append("entity_id", filters.entity_id);
      if (filters.subject_type)
        params.append("subject_type", filters.subject_type);
      if (filters.subject_id) params.append("subject_id", filters.subject_id);
      if (filters.result) params.append("result", filters.result);
      if (filters.date_range.from)
        params.append("start_time", filters.date_range.from.toISOString());
      if (filters.date_range.to)
        params.append("end_time", filters.date_range.to.toISOString());
      params.append("limit", pageSize.toString());
      params.append("offset", offset.toString());

      // Call the API
      const response = await fetch(`/api/audit/logs?${params.toString()}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch audit logs: ${response.statusText}`);
      }

      const data: AuditLogResponse = await response.json();
      setLogs(data.logs);
      setTotal(data.total);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "An error occurred while fetching audit logs"
      );
      console.error("Error fetching audit logs:", err);
    } finally {
      setLoading(false);
    }
  };

  // Fetch logs when filters or pagination change
  useEffect(() => {
    fetchLogs();
  }, [filters, page, pageSize]);

  // Handle filter changes
  const handleFilterChange = (key: string, value: any) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setPage(1); // Reset to first page when filters change
  };

  // Reset all filters
  const resetFilters = () => {
    setFilters({
      action_type: "",
      entity_type: "",
      entity_id: "",
      subject_type: "",
      subject_id: "",
      result: "",
      date_range: { from: undefined, to: undefined },
    });
    setPage(1);
  };

  // Format action type for display
  const formatActionType = (actionType: string) => {
    return actionType
      .split("_")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ");
  };

  // Get badge color based on action type
  const getActionBadgeColor = (actionType: string) => {
    switch (actionType) {
      case "permission_check":
        return "bg-blue-500/10 text-blue-500 border-blue-500/20";
      case "entity_create":
        return "bg-green-500/10 text-green-500 border-green-500/20";
      case "entity_delete":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      case "relation_create":
        return "bg-purple-500/10 text-purple-500 border-purple-500/20";
      case "relation_delete":
        return "bg-orange-500/10 text-orange-500 border-orange-500/20";
      default:
        return "bg-gray-500/10 text-gray-500 border-gray-500/20";
    }
  };

  // Get badge color based on result
  const getResultBadgeColor = (result: boolean | undefined) => {
    if (result === undefined) return "";
    return result
      ? "bg-green-500/10 text-green-500 border-green-500/20"
      : "bg-red-500/10 text-red-500 border-red-500/20";
  };

  // Show log details
  const showLogDetails = (log: AuditLog) => {
    setSelectedLog(log);
    setDetailDialogOpen(true);
  };

  return (
    <div className="space-y-4">
      <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
        <CardHeader className="border-b border-slate-700/40">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ClipboardList className="h-5 w-5 text-violet-400" />
              <CardTitle className="text-xl">
                Authorization Audit Logs
              </CardTitle>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={fetchLogs}
              className="border-slate-700 hover:bg-slate-800"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh
            </Button>
          </div>
          <CardDescription className="text-gray-400">
            Track and monitor authorization-related activities in the system
          </CardDescription>
        </CardHeader>

        {/* Filters */}
        <CardContent className="pt-6">
          <div className="space-y-4">
            <div className="flex items-center gap-2 mb-4">
              <Filter className="h-4 w-4 text-violet-400" />
              <h3 className="text-sm font-medium">Filters</h3>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="space-y-2">
                <label className="text-xs text-gray-400">Action Type</label>
                <Select
                  value={filters.action_type}
                  onValueChange={(value) =>
                    handleFilterChange("action_type", value)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="All Actions" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">All Actions</SelectItem>
                    {actionTypeOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-xs text-gray-400">Entity Type</label>
                <Input
                  placeholder="e.g. user, organization"
                  value={filters.entity_type}
                  onChange={(e) =>
                    handleFilterChange("entity_type", e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs text-gray-400">Entity ID</label>
                <Input
                  placeholder="e.g. user123, org456"
                  value={filters.entity_id}
                  onChange={(e) =>
                    handleFilterChange("entity_id", e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs text-gray-400">Subject Type</label>
                <Input
                  placeholder="e.g. user"
                  value={filters.subject_type}
                  onChange={(e) =>
                    handleFilterChange("subject_type", e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs text-gray-400">Subject ID</label>
                <Input
                  placeholder="e.g. user456"
                  value={filters.subject_id}
                  onChange={(e) =>
                    handleFilterChange("subject_id", e.target.value)
                  }
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs text-gray-400">Result</label>
                <Select
                  value={filters.result}
                  onValueChange={(value) => handleFilterChange("result", value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="All Results" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">All Results</SelectItem>
                    {resultOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={resetFilters}
                className="border-slate-700 hover:bg-slate-800"
              >
                Reset Filters
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Results Table */}
      <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
        <CardContent className="pt-6">
          {error && (
            <div className="flex items-center gap-2 p-4 mb-4 border border-red-500/30 rounded-md bg-red-500/10 text-red-400">
              <AlertCircle className="h-4 w-4" />
              <p>{error}</p>
            </div>
          )}

          {loading ? (
            // Loading state
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex items-center gap-4">
                  <Skeleton className="h-12 w-full" />
                </div>
              ))}
            </div>
          ) : logs && logs.length === 0 ? (
            // Empty state
            <div className="text-center py-8">
              <p className="text-gray-400">
                No audit logs found matching the current filters.
              </p>
            </div>
          ) : (
            // Results table
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[180px]">Timestamp</TableHead>
                    <TableHead>Action</TableHead>
                    <TableHead>Entity</TableHead>
                    <TableHead>Subject</TableHead>
                    <TableHead>Result</TableHead>
                    <TableHead className="text-right">Details</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs &&
                    logs.map((log) => (
                      <TableRow
                        key={log.id}
                        className="hover:bg-slate-800/50 cursor-pointer"
                        onClick={() => showLogDetails(log)}
                      >
                        <TableCell className="text-xs text-gray-400">
                          {new Date(log.timestamp).toLocaleString()}
                          <div className="text-xs text-gray-500 mt-1">
                            {formatDistance(
                              new Date(log.timestamp),
                              new Date(),
                              {
                                addSuffix: true,
                              }
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge
                            className={`px-2 py-1 border ${getActionBadgeColor(
                              log.action_type
                            )}`}
                          >
                            {formatActionType(log.action_type)}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <span className="text-gray-400">
                              {log.entity_type}
                            </span>
                            <span className="text-xs text-gray-500 ml-1">
                              :{log.entity_id}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          {log.subject_type && log.subject_id ? (
                            <div className="text-sm">
                              <span className="text-gray-400">
                                {log.subject_type}
                              </span>
                              <span className="text-xs text-gray-500 ml-1">
                                :{log.subject_id}
                              </span>
                            </div>
                          ) : (
                            <span className="text-gray-500 text-xs">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          {log.result !== undefined ? (
                            <Badge
                              className={`px-2 py-0.5 border ${getResultBadgeColor(
                                log.result
                              )}`}
                            >
                              {log.result ? "Allowed" : "Denied"}
                            </Badge>
                          ) : (
                            <span className="text-gray-500 text-xs">-</span>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              showLogDetails(log);
                            }}
                          >
                            View
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </div>
          )}

          {/* Pagination */}
          {!loading && logs && logs.length > 0 && (
            <div className="flex items-center justify-between mt-4">
              <div className="text-sm text-gray-400">
                Showing {(page - 1) * pageSize + 1} to{" "}
                {Math.min(page * pageSize, total)} of {total} entries
              </div>
              <Pagination
                page={page}
                pageCount={Math.ceil(total / pageSize)}
                onPageChange={setPage}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Detail Dialog */}
      <Dialog open={detailDialogOpen} onOpenChange={setDetailDialogOpen}>
        <DialogContent className="max-w-3xl bg-[#1e1e2e] shadow-lg border border-slate-700/40">
          <DialogHeader>
            <DialogTitle>Audit Log Details</DialogTitle>
            <DialogDescription>
              Detailed information about the selected audit log entry
            </DialogDescription>
          </DialogHeader>

          {selectedLog && (
            <div className="space-y-4 mt-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-xs text-gray-400">ID</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                    {selectedLog.id}
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs text-gray-400">Timestamp</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                    {new Date(selectedLog.timestamp).toLocaleString()}
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs text-gray-400">Action Type</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                    <Badge
                      className={`px-2 py-1 border ${getActionBadgeColor(
                        selectedLog.action_type
                      )}`}
                    >
                      {formatActionType(selectedLog.action_type)}
                    </Badge>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs text-gray-400">Result</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                    {selectedLog.result !== undefined ? (
                      <Badge
                        className={`px-2 py-0.5 border ${getResultBadgeColor(
                          selectedLog.result
                        )}`}
                      >
                        {selectedLog.result ? "Allowed" : "Denied"}
                      </Badge>
                    ) : (
                      <span className="text-gray-500">N/A</span>
                    )}
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs text-gray-400">Entity</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                    <span className="text-violet-400">
                      {selectedLog.entity_type}
                    </span>
                    <span className="text-gray-300">
                      :{selectedLog.entity_id}
                    </span>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs text-gray-400">Subject</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                    {selectedLog.subject_type && selectedLog.subject_id ? (
                      <>
                        <span className="text-violet-400">
                          {selectedLog.subject_type}
                        </span>
                        <span className="text-gray-300">
                          :{selectedLog.subject_id}
                        </span>
                      </>
                    ) : (
                      <span className="text-gray-500">N/A</span>
                    )}
                  </div>
                </div>

                {selectedLog.relation && (
                  <div className="space-y-2">
                    <label className="text-xs text-gray-400">Relation</label>
                    <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                      {selectedLog.relation}
                    </div>
                  </div>
                )}

                {selectedLog.permission && (
                  <div className="space-y-2">
                    <label className="text-xs text-gray-400">Permission</label>
                    <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40">
                      {selectedLog.permission}
                    </div>
                  </div>
                )}
              </div>

              {selectedLog.context &&
                Object.keys(selectedLog.context).length > 0 && (
                  <div className="space-y-2">
                    <label className="text-xs text-gray-400">Context</label>
                    <pre className="text-xs bg-slate-800/50 p-3 rounded-md border border-slate-700/40 overflow-auto whitespace-pre-wrap max-h-60">
                      {JSON.stringify(selectedLog.context, null, 2)}
                    </pre>
                  </div>
                )}

              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <label className="text-xs text-gray-400">Request ID</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40 truncate">
                    {selectedLog.request_id || "N/A"}
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs text-gray-400">Client IP</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40 truncate">
                    {selectedLog.client_ip || "N/A"}
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs text-gray-400">User Agent</label>
                  <div className="text-sm bg-slate-800/50 p-2 rounded-md border border-slate-700/40 truncate">
                    {selectedLog.user_agent || "N/A"}
                  </div>
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
