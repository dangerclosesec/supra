// src/components/RuleTestModal.tsx
import React, { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { AlertCircle, CheckCircle2, XCircle } from "lucide-react";

// Types
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
}

interface RuleTestModalProps {
  rule: RuleDefinition | null;
  isOpen: boolean;
  onClose: () => void;
}

const RuleTestModal = ({ rule, isOpen, onClose }: RuleTestModalProps) => {
  const [paramValues, setParamValues] = useState<Record<string, string>>({});
  const [testResult, setTestResult] = useState<boolean | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Reset form when rule changes
  React.useEffect(() => {
    if (rule) {
      const initialValues: Record<string, string> = {};
      rule.parameters.forEach((param) => {
        initialValues[param.name] = "";
      });
      setParamValues(initialValues);
      setTestResult(null);
      setError(null);
    }
  }, [rule]);

  const handleInputChange = (paramName: string, value: string) => {
    setParamValues((prev) => ({
      ...prev,
      [paramName]: value,
    }));
  };

  const handleTestRule = async () => {
    if (!rule) return;

    setIsLoading(true);
    setError(null);
    setTestResult(null);

    try {
      const response = await fetch("/api/test-rule", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          rule_name: rule.name,
          parameters: paramValues,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.error || "Failed to test rule");
        return;
      }

      setTestResult(data.result);
    } catch (err) {
      setError("An error occurred while testing the rule");
      console.error("Error testing rule:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const getInputTypeForParam = (dataType: string) => {
    switch (dataType.toLowerCase()) {
      case "integer":
        return "number";
      case "double":
      case "float":
        return "number";
      case "boolean":
        return "checkbox";
      default:
        return "text";
    }
  };

  // Format parameter placeholder based on type
  const getPlaceholderForParam = (dataType: string) => {
    switch (dataType.toLowerCase()) {
      case "integer":
        return "Enter a whole number";
      case "double":
      case "float":
        return "Enter a decimal number";
      case "boolean":
        return "true/false";
      case "string":
        return "Enter text value";
      default:
        return `Enter ${dataType} value`;
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="bg-[#1e1e2e] border border-slate-700/40 text-white max-w-md">
        <DialogHeader>
          <DialogTitle className="text-xl flex items-center gap-2">
            Test Rule: {rule?.name}
          </DialogTitle>
          <DialogDescription className="text-gray-400">
            Enter parameter values to test the rule expression
          </DialogDescription>
        </DialogHeader>

        <div className="py-4">
          {/* Rule expression info */}
          <div className="mb-4 p-3 bg-slate-800/50 rounded-md">
            <Label className="text-sm text-gray-400 mb-1 block">Expression</Label>
            <div className="font-mono text-sm text-teal-400">{rule?.expression}</div>
          </div>

          {/* Parameters form */}
          <div className="space-y-4">
            {rule?.parameters.map((param) => (
              <div key={param.name} className="grid gap-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor={param.name} className="text-sm">
                    {param.name}
                  </Label>
                  <Badge
                    variant="outline"
                    className="font-mono text-xs bg-violet-500/10 text-violet-400 border-violet-500/20"
                  >
                    {param.data_type}
                  </Badge>
                </div>
                <Input
                  id={param.name}
                  type={getInputTypeForParam(param.data_type)}
                  className="bg-[#2a2a3c] border-slate-700/40"
                  placeholder={getPlaceholderForParam(param.data_type)}
                  value={paramValues[param.name] || ""}
                  onChange={(e) => handleInputChange(param.name, e.target.value)}
                />
              </div>
            ))}
          </div>

          {/* Test result */}
          {testResult !== null && (
            <div
              className={`mt-4 p-3 rounded-md flex items-center gap-2 ${
                testResult
                  ? "bg-green-500/10 border border-green-500/20 text-green-400"
                  : "bg-red-500/10 border border-red-500/20 text-red-400"
              }`}
            >
              {testResult ? (
                <CheckCircle2 className="h-5 w-5" />
              ) : (
                <XCircle className="h-5 w-5" />
              )}
              <span>
                Rule evaluation result: <strong>{testResult ? "TRUE" : "FALSE"}</strong>
              </span>
            </div>
          )}

          {/* Error message */}
          {error && (
            <div className="mt-4 p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-400 flex items-center gap-2">
              <AlertCircle className="h-5 w-5" />
              <span>{error}</span>
            </div>
          )}
        </div>

        <DialogFooter className="flex justify-between items-center gap-3 pt-2">
          <Button
            variant="outline"
            onClick={onClose}
            className="border-slate-700/40 text-gray-300 hover:bg-slate-700/20"
          >
            Close
          </Button>
          <Button
            onClick={handleTestRule}
            disabled={isLoading}
            className="bg-violet-600 hover:bg-violet-700 text-white"
          >
            {isLoading ? "Testing..." : "Test Rule"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default RuleTestModal;