import React from "react";
import { useDraggable, useDroppable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";

interface DraggableProps extends React.HTMLAttributes<HTMLButtonElement> {
  id: string;
  children: React.ReactNode;
  isBeingDragged?: boolean;
}

export function Draggable({ id, children, isBeingDragged, ...props }: DraggableProps) {
  const { attributes, listeners, setNodeRef: setDraggableRef, transform } = useDraggable({
    id: id,
  });

  const { isOver, setNodeRef: setDroppableRef } = useDroppable({ id });

  const setNodeRef = (node: HTMLElement | null) => {
    setDraggableRef(node);
    setDroppableRef(node);
  };

  const style = {
    transform: CSS.Transform.toString(transform),
    opacity: isBeingDragged ? 0.3 : (isOver ? 0.5 : 1),
    ...props.style
  };

  return (
    <button
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      {...props}
    >
      {/* <span
        {...listeners}
        {...attributes}
        style={{ cursor: "grab", padding: "0 4px" }}
      >
        ⋮⋮
      </span> */}
      {children}
    </button>
  );
}
