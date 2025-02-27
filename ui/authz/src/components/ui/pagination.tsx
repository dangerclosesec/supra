import React from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "./button";

interface PaginationProps {
  page: number;
  pageCount: number;
  onPageChange: (page: number) => void;
}

export function Pagination({ page, pageCount, onPageChange }: PaginationProps) {
  // Don't render pagination if there's only one page
  if (pageCount <= 1) return null;

  // Generate page numbers to display
  const getPageNumbers = () => {
    const pages = [];
    const showEllipsis = pageCount > 7;
    
    if (showEllipsis) {
      // Always show first page
      pages.push(1);
      
      // Show ellipsis if we're not near the beginning
      if (page > 3) {
        pages.push("ellipsis-start");
      }
      
      // Show pages around the current page
      const start = Math.max(2, page - 1);
      const end = Math.min(pageCount - 1, page + 1);
      
      for (let i = start; i <= end; i++) {
        pages.push(i);
      }
      
      // Show ellipsis if we're not near the end
      if (page < pageCount - 2) {
        pages.push("ellipsis-end");
      }
      
      // Always show last page
      pages.push(pageCount);
    } else {
      // Show all pages if there are few
      for (let i = 1; i <= pageCount; i++) {
        pages.push(i);
      }
    }
    
    return pages;
  };

  return (
    <div className="flex items-center space-x-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => onPageChange(page - 1)}
        disabled={page === 1}
        className="h-8 w-8 p-0 border-slate-700 hover:bg-slate-800"
      >
        <ChevronLeft className="h-4 w-4" />
        <span className="sr-only">Previous Page</span>
      </Button>
      
      {getPageNumbers().map((pageNum, i) => {
        // Handle ellipsis
        if (pageNum === "ellipsis-start" || pageNum === "ellipsis-end") {
          return (
            <span key={`ellipsis-${i}`} className="px-2 text-gray-500">
              ...
            </span>
          );
        }
        
        // Handle numbered buttons
        return (
          <Button
            key={pageNum}
            variant={page === pageNum ? "default" : "outline"}
            size="sm"
            onClick={() => onPageChange(Number(pageNum))}
            className={`h-8 w-8 p-0 ${
              page === pageNum 
                ? "bg-violet-600 hover:bg-violet-700 text-white" 
                : "border-slate-700 hover:bg-slate-800"
            }`}
          >
            {pageNum}
          </Button>
        );
      })}
      
      <Button
        variant="outline"
        size="sm"
        onClick={() => onPageChange(page + 1)}
        disabled={page === pageCount}
        className="h-8 w-8 p-0 border-slate-700 hover:bg-slate-800"
      >
        <ChevronRight className="h-4 w-4" />
        <span className="sr-only">Next Page</span>
      </Button>
    </div>
  );
}