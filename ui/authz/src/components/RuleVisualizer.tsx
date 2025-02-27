// src/components/RuleVisualizer.tsx
import React, { useState, useEffect } from "react";
import RuleTestModal from "./RuleTestModal";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from "@/components/ui/card";
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
  Search,
  RefreshCw,
  Loader2,
  AlertCircle,
  BookOpen,
  Code,
  Play,
} from "lucide-react";

// Types for rule definitions
interface ParameterDefinition {
  name: string;
  data_type: string;
}

interface RuleDefinition {
  id: number;
  name: string;
  parameters: ParameterDefinition[];
  expression: string;
  description?: string;
  created_at?: string;
}

const RuleVisualizer = () => {
  // State for data
  const [ruleDefinitions, setRuleDefinitions] = useState<RuleDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Filtering state
  const [ruleFilter, setRuleFilter] = useState("");
  
  // Test modal state
  const [testModalOpen, setTestModalOpen] = useState(false);
  const [selectedRule, setSelectedRule] = useState<RuleDefinition | null>(null);

  // Fetch rule definitions
  const fetchRuleDefinitions = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/rule-definitions");
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      
      const data = await response.json();
      setRuleDefinitions(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error("Error fetching rule definitions:", err);
      setError(`Failed to fetch rule definitions: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Fetch data when component mounts
  useEffect(() => {
    fetchRuleDefinitions();
  }, []);

  // Filter rules based on search term
  const filteredRules = ruleDefinitions.filter((rule) => {
    const searchTerm = ruleFilter.toLowerCase();
    return (
      rule.name?.toLowerCase().includes(searchTerm) ||
      rule.expression?.toLowerCase().includes(searchTerm) ||
      rule.description?.toLowerCase().includes(searchTerm)
    );
  });

  // Format parameter type for display
  const formatParameterType = (type: string) => {
    switch (type.toLowerCase()) {
      case "string":
        return "String";
      case "boolean":
        return "Boolean";
      case "integer":
        return "Integer";
      case "float":
        return "Float";
      case "decimal":
        return "Decimal";
      default:
        return type;
    }
  };

  // Helper function to get an expression with highlighted keywords
  const getExpressionHighlighted = (expr: string) => {
    if (!expr) return null;

    // Replace comparison operators
    const operatorsReplaced = expr
      .replace(/ == /g, ' <span class="text-yellow-400 font-semibold">==</span> ')
      .replace(/ != /g, ' <span class="text-yellow-400 font-semibold">!=</span> ')
      .replace(/ > /g, ' <span class="text-yellow-400 font-semibold">></span> ')
      .replace(/ >= /g, ' <span class="text-yellow-400 font-semibold">>=</span> ')
      .replace(/ < /g, ' <span class="text-yellow-400 font-semibold"><</span> ')
      .replace(/ <= /g, ' <span class="text-yellow-400 font-semibold"><=</span> ');

    // Replace 'and' with styled version
    const andReplaced = operatorsReplaced.replace(
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
            <BookOpen className="w-5 h-5 text-teal-400" />
            <CardTitle className="text-xl">Rule Definitions</CardTitle>
          </div>
          <CardDescription className="text-gray-400">
            View and manage rules that can be used in permission expressions
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between mb-4">
            <div className="relative w-full max-w-sm">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
              <Input
                type="text"
                placeholder="Filter rules..."
                className="w-full pl-8 pr-4 bg-[#2a2a3c] border-slate-700/40"
                value={ruleFilter}
                onChange={(e) => setRuleFilter(e.target.value)}
              />
            </div>
            <Button
              variant="outline"
              size="sm"
              className="border-slate-700/40 text-gray-300 hover:bg-slate-700/20"
              onClick={fetchRuleDefinitions}
              disabled={loading}
            >
              {loading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4" />
              )}
              Refresh
            </Button>
          </div>

          {loading ? (
            <div className="flex justify-center items-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-teal-400" />
            </div>
          ) : (
            <div className="rounded-md border border-slate-700/40 overflow-hidden">
              <Table>
                <TableHeader className="bg-[#2a2a3c]">
                  <TableRow>
                    <TableHead className="text-gray-400 font-medium">
                      Rule Name
                    </TableHead>
                    <TableHead className="text-gray-400 font-medium">
                      Parameters
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
                  {filteredRules.length > 0 ? (
                    filteredRules.map((rule) => (
                      <TableRow
                        key={rule.id}
                        className="border-t border-slate-700/40"
                      >
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Badge
                              variant="outline"
                              className="font-mono bg-teal-500/10 text-teal-400 border-teal-500/20"
                            >
                              {rule.name}
                            </Badge>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0 text-gray-400 hover:text-teal-400 hover:bg-teal-500/10"
                              onClick={() => {
                                setSelectedRule(rule);
                                setTestModalOpen(true);
                              }}
                              title="Test Rule"
                            >
                              <Play className="h-4 w-4" />
                            </Button>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col gap-1">
                            {rule.parameters.map((param, idx) => (
                              <div key={idx} className="flex items-center gap-1.5">
                                <span className="font-mono text-sm text-violet-400">
                                  {param.name}
                                </span>
                                <span className="text-xs text-gray-400">
                                  {formatParameterType(param.data_type)}
                                </span>
                              </div>
                            ))}
                            {rule.parameters.length === 0 && (
                              <span className="text-gray-400 text-sm">None</span>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {getExpressionHighlighted(rule.expression)}
                        </TableCell>
                        <TableCell className="text-gray-400">
                          {rule.description || "-"}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell
                        colSpan={4}
                        className="text-center text-gray-400 py-8"
                      >
                        No rules found
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
        <CardFooter className="text-sm text-gray-400 border-t border-slate-700/40 mt-2">
          <div className="flex items-center gap-2">
            <Code className="h-3 w-3" />
            {ruleDefinitions.length} Rule Definitions
          </div>
        </CardFooter>
      </Card>
      
      {/* Test Rule Modal */}
      <RuleTestModal 
        rule={selectedRule}
        isOpen={testModalOpen}
        onClose={() => setTestModalOpen(false)}
      />
    </div>
  );
};

export default RuleVisualizer;