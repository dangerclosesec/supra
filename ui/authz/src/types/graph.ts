// src/types/graph.ts
export interface Node {
  id: string;
  type: string;
  label: string;
  group: string;
}

export interface Link {
  source: string;
  target: string;
  type: string;
  label: string;
}

export interface GraphData {
  nodes: Node[];
  links: Link[];
}

export interface PermissionPathResponse {
  nodes: Node[];
  links: Link[];
  paths?: Link[][];
  expression: string;
  allowed: boolean;
}
