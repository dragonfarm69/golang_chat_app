import React from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

interface SortableItemProps extends React.HTMLAttributes<HTMLButtonElement> {
  id: string;
  children: React.ReactNode;
}

export const RoomListBar = React.memo(
  ({ id, children, ...props }: SortableItemProps) => {
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
        {...listeners}
        {...attributes}
        {...props}
      >
        {children}
      </button>
    );
  }
);

RoomListBar.displayName = "RoomListBar";