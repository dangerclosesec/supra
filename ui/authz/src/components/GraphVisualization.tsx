import React, { useEffect, useRef, useState } from "react";
import * as d3 from "d3";
import { GraphData, Link, Node } from "@/types/graph";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface GraphVisualizationProps {
  data: GraphData;
  onNodeClick?: (node: Node) => void;
  highlightPath?: Link[];
}

const GraphVisualization: React.FC<GraphVisualizationProps> = ({
  data,
  onNodeClick,
  highlightPath,
}) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });

  // Entity colors mapping
  const entityColors: Record<string, string> = {
    user: "#4285F4",
    organization: "#EA4335",
    group: "#FBBC05",
    project: "#34A853",
    task: "#FF6D01",
    default: "#9AA0A6",
  };

  // Relation colors mapping
  const relationColors: Record<string, string> = {
    owner: "#AA00FF",
    admin: "#00C853",
    member: "#2979FF",
    manager: "#FF6D00",
    contributor: "#FFC400",
    default: "#9E9E9E",
  };

  // Get color for entity type
  const getEntityColor = (type: string) => {
    return entityColors[type] || entityColors.default;
  };

  // Get color for relation type
  const getRelationColor = (type: string) => {
    return relationColors[type] || relationColors.default;
  };

  // Update dimensions when window resizes
  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        setDimensions({
          width: containerRef.current.clientWidth,
          height: containerRef.current.clientHeight,
        });
      }
    };
    updateDimensions();
    window.addEventListener("resize", updateDimensions);
    return () => window.removeEventListener("resize", updateDimensions);
  }, []);

  // Create and update visualization
  useEffect(() => {
    if (!svgRef.current || !data.nodes.length) return;

    // Clear existing svg content
    d3.select(svgRef.current).selectAll("*").remove();

    const svg = d3.select(svgRef.current);
    const width = dimensions.width;
    const height = dimensions.height;

    // Create zoom behavior
    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 8])
      .on("zoom", (event) => {
        g.attr("transform", event.transform);
      });

    svg.call(zoom);

    // Create main group for zooming
    const g = svg.append("g");

    // Create arrow markers for links
    svg
      .append("defs")
      .selectAll("marker")
      .data(["default", ...Object.keys(relationColors)])
      .enter()
      .append("marker")
      .attr("id", (d) => `arrow-${d}`)
      .attr("viewBox", "0 -5 10 10")
      .attr("refX", 20)
      .attr("refY", 0)
      .attr("markerWidth", 6)
      .attr("markerHeight", 6)
      .attr("orient", "auto")
      .append("path")
      .attr("d", "M0,-5L10,0L0,5")
      .attr("fill", (d) => (d === "default" ? "#999" : getRelationColor(d)));

    // Prepare link data with proper source/target objects
    const links = (data.links || []).map((link) => ({
      ...link,
      source: typeof link.source === "string" ? link.source : link.source.id,
      target: typeof link.target === "string" ? link.target : link.target.id,
    }));

    // Create force simulation
    const simulation = d3
      .forceSimulation(data.nodes as d3.SimulationNodeDatum[])
      .force(
        "link",
        d3
          .forceLink<
            d3.SimulationNodeDatum,
            d3.SimulationLinkDatum<d3.SimulationNodeDatum>
          >(links as any)
          .id((d) => (d as any).id)
          .distance(250)
      )
      .force("charge", d3.forceManyBody().strength(-500))
      .force("center", d3.forceCenter(width / 2, height / 2).strength(0.1))
      .force("x", d3.forceX(width / 2).strength(0.1))
      .force("y", d3.forceY(height / 2).strength(0.1))
      .force("collision", d3.forceCollide().radius(80));

    // Create links
    const link = g
      .append("g")
      .attr("class", "links")
      .selectAll("line")
      .data(links)
      .enter()
      .append("line")
      .attr("stroke", (d) => getRelationColor(d.type))
      .attr("stroke-width", 2)
      .attr("marker-end", (d) => `url(#arrow-${d.type || "default"})`);

    // Highlight path if provided
    if (highlightPath && highlightPath.length > 0) {
      // First dim all links
      link.attr("stroke-opacity", 0.2);

      // Then highlight the path links
      link
        .filter((d) => {
          return highlightPath.some((pathLink) => {
            const sourceMatch =
              d.source === pathLink.source ||
              (typeof d.source === "object" &&
                (d.source as any).id === pathLink.source);
            const targetMatch =
              d.target === pathLink.target ||
              (typeof d.target === "object" &&
                (d.target as any).id === pathLink.target);
            return sourceMatch && targetMatch;
          });
        })
        .attr("stroke-opacity", 1)
        .attr("stroke-width", 4);
    }

    // Create link labels
    const linkLabels = g
      .append("g")
      .attr("class", "link-labels")
      .selectAll("text")
      .data(links)
      .enter()
      .append("text")
      .attr("dy", -5)
      .attr("text-anchor", "middle")
      .attr("fill", (d) => getRelationColor(d.type))
      .text((d) => d.label);

    // Create nodes group
    const node = g
      .append("g")
      .attr("class", "nodes")
      .selectAll("g")
      .data(data.nodes)
      .enter()
      .append("g")
      .attr("class", "node")
      .call(
        d3
          .drag<SVGGElement, any>()
          .on("start", dragstarted)
          .on("drag", dragged)
          .on("end", dragended)
      )
      .on("click", (event, d) => {
        event.stopPropagation();
        if (onNodeClick) onNodeClick(d);
      });

    // Create node circles
    node
      .append("circle")
      .attr("r", 10)
      .attr("fill", (d) => getEntityColor(d.type))
      .attr("stroke", (d) => {
        // Highlight nodes in the path
        if (highlightPath && highlightPath.length > 0) {
          const nodeIsInPath = highlightPath.some((link) => {
            return link.source === d.id || link.target === d.id;
          });
          return nodeIsInPath ? "#ff0000" : "none";
        }
        return "none";
      })
      .attr("stroke-width", (d) => {
        if (highlightPath && highlightPath.length > 0) {
          const nodeIsInPath = highlightPath.some((link) => {
            return link.source === d.id || link.target === d.id;
          });
          return nodeIsInPath ? 3 : 0;
        }
        return 0;
      });

    // Create node labels
    node
      .append("text")
      .attr("dx", 15)
      .attr("dy", 5)
      .text((d) => d.label)
      .attr("fill", "#efefef");

    // Update node and link positions on simulation tick
    simulation.on("tick", () => {
      link
        .attr("x1", (d) => (d.source as any).x)
        .attr("y1", (d) => (d.source as any).y)
        .attr("x2", (d) => (d.target as any).x)
        .attr("y2", (d) => (d.target as any).y);

      linkLabels
        .attr("x", (d) => ((d.source as any).x + (d.target as any).x) / 2)
        .attr("y", (d) => ((d.source as any).y + (d.target as any).y) / 2);

      node.attr("transform", (d) => `translate(${d.x},${d.y})`);
    });

    // Drag functions
    function dragstarted(event: any, d: any) {
      if (!event.active) simulation.alphaTarget(0.3).restart();
      d.fx = d.x;
      d.fy = d.y;
    }

    function dragged(event: any, d: any) {
      d.fx = event.x;
      d.fy = event.y;
    }

    function dragended(event: any, d: any) {
      if (!event.active) simulation.alphaTarget(0);
      d.fx = null;
      d.fy = null;
    }

    // Center the graph initially
    const initialTransform = d3.zoomIdentity
      .translate(width / 2, height / 2)
      .scale(0.7);
    svg.call(zoom.transform, initialTransform);

    // Cleanup on unmount
    return () => {
      simulation.stop();
    };
  }, [data, dimensions, highlightPath]);

  return (
    <div ref={containerRef} className="w-full h-full">
      <svg
        ref={svgRef}
        // width={dimensions.width}
        // height={dimensions.height}
        className="bg-background h-full w-full"
      />
    </div>
  );
};

export default GraphVisualization;
