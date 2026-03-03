import React, { useState } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

interface SortableItemProps extends React.HTMLAttributes<HTMLButtonElement> {
  id: string;
  children: React.ReactNode;
  flashing: boolean;
}

export const RoomButton = React.memo(
  ({ id, children, flashing, ...props }: SortableItemProps) => {
    // const [isFlashing, setIsFlashing] = useState<boolean>(false);

    const {
      attributes,
      listeners,
      setNodeRef,
      transform,
      transition,
      isDragging,
    } = useSortable({ id });

    const style = {
      transform: CSS.Transform.toString(transform),
      transition,
      opacity: isDragging ? 0.5 : 1,
      ...props.style,
    };
    
    return (
      <button
        ref={setNodeRef}
        style={style}
        className={"room-icon" + " " + (flashing ? "warning-flashing" : "")}
        onClick={() => {
          // console.log("flashing")
          // setIsFlashing(true)
        }}
        {...listeners}
        {...attributes}
        {...props}
      >
        {children}
      </button>
    );
  }
);

RoomButton.displayName = "RoomButton";