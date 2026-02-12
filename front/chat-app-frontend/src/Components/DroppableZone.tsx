import React from "react";
import { useDroppable } from "@dnd-kit/core";

interface SortableItemProps extends React.HTMLAttributes<HTMLDivElement> {
  id: string;
  children: React.ReactNode;
  className?: string;
  hoveredRegion?: REGION;
}

export type REGION = "top" | "bottom" | "left" | "right" | null;

export function DroppableZone({ 
  id, 
  children, 
  hoveredRegion, 
  className,
  style,
  ...props 
}: SortableItemProps) {
  const { setNodeRef, isOver } = useDroppable({ id });

  const regionStyle: React.CSSProperties = {
    position: "absolute",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontSize: "14px",
    fontWeight: "600",
    color: "#fff",
    backgroundColor: "rgba(88, 101, 242, 0.3)",
    border: "2px dashed rgba(88, 101, 242, 0.8)",
    backdropFilter: "blur(2px)",
    zIndex: 10,
    pointerEvents: "none",
    transition: "all 0.2s ease",
  };

  return (
    <div
      ref={setNodeRef}
      id={id}
      style={{
        position: "relative",
        ...style,
      }}
      className={className}
      {...props}
    >
      {isOver && hoveredRegion && (
        <>
          {hoveredRegion === "top" && (
            <div
              style={{
                ...regionStyle,
                top: 0,
                left: 0,
                right: 0,
                height: "30%",
                borderLeft: "none",
                borderRight: "none",
                borderTop: "none",
              }}
            >
              <div
                style={{
                  padding: "8px 16px",
                  backgroundColor: "rgba(88, 101, 242, 0.9)",
                  borderRadius: "4px",
                  boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
                }}
              >
                ⬆️ Drop here to open above
              </div>
            </div>
          )}
          {hoveredRegion === "bottom" && (
            <div
              style={{
                ...regionStyle,
                bottom: 0,
                left: 0,
                right: 0,
                height: "30%",
                borderLeft: "none",
                borderRight: "none",
                borderBottom: "none",
              }}
            >
              <div
                style={{
                  padding: "8px 16px",
                  backgroundColor: "rgba(88, 101, 242, 0.9)",
                  borderRadius: "4px",
                  boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
                }}
              >
                ⬇️ Drop here to open below
              </div>
            </div>
          )}
          {hoveredRegion === "left" && (
            <div
              style={{
                ...regionStyle,
                top: 0,
                left: 0,
                bottom: 0,
                width: "30%",
                borderTop: "none",
                borderBottom: "none",
                borderLeft: "none",
              }}
            >
              <div
                style={{
                  padding: "8px 16px",
                  backgroundColor: "rgba(88, 101, 242, 0.9)",
                  borderRadius: "4px",
                  boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
                  transform: "rotate(-90deg)",
                }}
              >
                ⬅️ Drop here to open left
              </div>
            </div>
          )}
          {hoveredRegion === "right" && (
            <div
              style={{
                ...regionStyle,
                top: 0,
                right: 0,
                bottom: 0,
                width: "30%",
                borderTop: "none",
                borderBottom: "none",
                borderRight: "none",
              }}
            >
              <div
                style={{
                  padding: "8px 16px",
                  backgroundColor: "rgba(88, 101, 242, 0.9)",
                  borderRadius: "4px",
                  boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
                  transform: "rotate(90deg)",
                }}
              >
                ➡️ Drop here to open right
              </div>
            </div>
          )}
        </>
      )}
      {children}
    </div>
  );
}