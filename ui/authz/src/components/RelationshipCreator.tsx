// src/components/RelationshipCreator.tsx
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
import {
  Loader2,
  ShieldAlert,
  Link2,
  Info,
  Plus,
  Check,
  ArrowRight,
} from "lucide-react";

interface RelationshipCreatorProps {
  className?: string;
}

const RelationshipCreator: React.FC<RelationshipCreatorProps> = ({
  className,
}) => {
  // State for form fields
  const [actorType, setActorType] = useState<string>("");
  const [actorId, setActorId] = useState<string>("");
  const [relationType, setRelationType] = useState<string>("");
  const [targetType, setTargetType] = useState<string>("");
  const [targetId, setTargetId] = useState<string>("");

  // State for API data
  const [entityTypes, setEntityTypes] = useState<string[]>([]);
  const [relationTypes, setRelationTypes] = useState<string[]>([]);
  const [initialLoading, setInitialLoading] = useState<boolean>(true);

  // State for loading/error handling
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<boolean>(false);

  // Fetch entity types and relation types on component mount
  useEffect(() => {
    Promise.all([fetchEntityTypes(), fetchRelationTypes()])
      .then(() => {
        setInitialLoading(false);
      })
      .catch((err) => {
        console.error("Error initializing:", err);
        setInitialLoading(false);
      });
  }, []);

  // Function to fetch entity types
  const fetchEntityTypes = async () => {
    try {
      const response = await fetch("/api/entity-types");
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      if (data.error) {
        throw new Error(data.error);
      }

      // Extract entity types from the response
      // Assuming the API returns an array of objects with a 'type' field
      const types = Array.isArray(data)
        ? data.map((item) => item.type || item)
        : [];

      setEntityTypes(types);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to fetch entity types"
      );
      console.error("Error fetching entity types:", err);
    }
  };

  // Function to fetch relation types from existing relations
  const fetchRelationTypes = async () => {
    try {
      // Fall back to extracting unique relation types from relations
      const response = await fetch("/api/relations?limit=100");
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      if (data.error) {
        throw new Error(data.error);
      }

      // Extract unique relation types
      const types = Array.isArray(data)
        ? [...new Set(data.map((item) => item.relation || item.type))]
        : [];

      setRelationTypes(types);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to fetch relation types"
      );
      console.error("Error fetching relation types:", err);
    }
  };

  // Update the createRelationship function to use the correct endpoint and field names
  const createRelationship = async () => {
    if (!actorType || !actorId || !relationType || !targetType || !targetId) {
      setError("All fields are required");
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      const response = await fetch("/api/relation", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          // Change field names to match API
          subject_type: actorType,
          subject_id: actorId,
          relation: relationType,
          object_type: targetType,
          object_id: targetId,
        }),
      });

      if (!response.ok && response.status > 500) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const data = await response.json();
      if (data.error) {
        throw new Error(data.error);
      }

      setSuccess(true);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to create relationship"
      );
      console.error("Error creating relationship:", err);
    } finally {
      setLoading(false);
    }
  };

  // Reset the form
  const resetForm = () => {
    setActorType("");
    setActorId("");
    setRelationType("");
    setTargetType("");
    setTargetId("");
    setError(null);
    setSuccess(false);
  };

  return (
    <div className={`flex flex-col gap-6 ${className}`}>
      <Card className="bg-[#1e1e2e] shadow-md border border-slate-700/40">
        <CardHeader className="border-b border-slate-700/40">
          <div className="flex items-center gap-2">
            <Link2 className="h-5 w-5 text-teal-400" />
            <CardTitle className="text-xl">Create Relationship</CardTitle>
          </div>
          <CardDescription className="text-gray-400">
            Define a new relationship between entities in the system
          </CardDescription>
        </CardHeader>

        <CardContent className="pt-6">
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-400 mb-4 flex items-center gap-2">
              <ShieldAlert className="h-4 w-4" />
              <p>{error}</p>
            </div>
          )}

          {success && (
            <div className="p-3 bg-green-500/10 border border-green-500/20 rounded-md text-green-400 mb-4 flex items-center gap-2">
              <Check className="h-4 w-4" />
              <p>Relationship created successfully!</p>
            </div>
          )}

          <div className="bg-[#2a2a3c] rounded-md p-3 mb-4 text-sm text-gray-300 border border-slate-700/40">
            <span className="font-medium text-teal-400">
              How Relationships Work:
            </span>
            <p className="mt-1">
              Relationships connect one entity to another. For example, a{" "}
              <span className="text-violet-400">User</span> can be an
              <span className="text-violet-400"> Owner</span> of an
              <span className="text-violet-400"> Organization</span>. These
              connections define permissions in the system.
            </p>
          </div>

          {initialLoading ? (
            <div className="flex justify-center items-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-teal-400 mr-3" />
              <span className="text-gray-400">Loading entity types...</span>
            </div>
          ) : (
            <div className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1 text-gray-400">
                    Actor Entity Type
                    <span className="text-xs block text-gray-500 mt-0.5">
                      <b>Subject Type:</b> Entity that owns the relationship
                    </span>
                  </label>
                  <Select
                    value={actorType}
                    onValueChange={setActorType}
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
                    Actor Entity ID
                    <span className="text-xs block text-gray-500 mt-0.5">
                      <b>Subject ID:</b> Specific ID of the source
                    </span>
                  </label>
                  <Input
                    type="text"
                    placeholder="e.g. alice"
                    value={actorId}
                    onChange={(e) => setActorId(e.target.value)}
                    className="bg-[#2a2a3c] border-slate-700/40"
                    disabled={loading || !actorType}
                  />
                </div>
              </div>

              <div className="flex justify-center items-center py-2">
                <div className="w-20 h-0.5 bg-slate-700/40"></div>
                <div className="mx-4 p-2 rounded-full bg-[#2a2a3c] border border-slate-700/40">
                  <ArrowRight className="h-5 w-5 text-teal-400" />
                </div>
                <div className="w-20 h-0.5 bg-slate-700/40"></div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1 text-gray-400">
                  Relationship Type
                  <span className="text-xs block text-gray-500 mt-0.5">
                    Type of connection between entities
                  </span>
                </label>
                <Select
                  value={relationType}
                  onValueChange={setRelationType}
                  disabled={loading || !actorType || relationTypes.length === 0}
                >
                  <SelectTrigger className="bg-[#2a2a3c] border-slate-700/40">
                    <SelectValue placeholder="Select relationship" />
                  </SelectTrigger>
                  <SelectContent>
                    {relationTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {type}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex justify-center items-center py-2">
                <div className="w-20 h-0.5 bg-slate-700/40"></div>
                <div className="mx-4 p-2 rounded-full bg-[#2a2a3c] border border-slate-700/40">
                  <ArrowRight className="h-5 w-5 text-teal-400" />
                </div>
                <div className="w-20 h-0.5 bg-slate-700/40"></div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1 text-gray-400">
                    Target Entity Type
                    <span className="text-xs block text-gray-500 mt-0.5">
                      <b>Object Type:</b> Entity that receives the relationship
                    </span>
                  </label>
                  <Select
                    value={targetType}
                    onValueChange={setTargetType}
                    disabled={
                      loading || !relationType || entityTypes.length === 0
                    }
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
                    Target Entity ID
                    <span className="text-xs block text-gray-500 mt-0.5">
                      <b>Object ID:</b> Specific ID of the target
                    </span>
                  </label>
                  <Input
                    type="text"
                    placeholder="e.g. acme"
                    value={targetId}
                    onChange={(e) => setTargetId(e.target.value)}
                    className="bg-[#2a2a3c] border-slate-700/40"
                    disabled={loading || !targetType}
                  />
                </div>
              </div>
            </div>
          )}
        </CardContent>

        <CardFooter className="flex justify-between pt-4 border-t border-slate-700/40 mt-6">
          <Button
            onClick={resetForm}
            variant="outline"
            className="border-slate-700/40 text-gray-300 hover:bg-slate-700/20"
            disabled={loading}
          >
            Reset Form
          </Button>
          <Button
            onClick={createRelationship}
            disabled={
              loading ||
              !actorType ||
              !actorId ||
              !relationType ||
              !targetType ||
              !targetId
            }
            className="bg-teal-600 hover:bg-teal-700 text-white"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Creating...
              </>
            ) : (
              <>
                <Plus className="h-4 w-4 mr-2" />
                Create Relationship
              </>
            )}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
};

export default RelationshipCreator;
