import { DragStartEvent } from "@dnd-kit/core";
import { DragEndEvent, DragOverEvent, DragMoveEvent } from "@dnd-kit/core";
import { REGION } from "../../../Components/DroppableZone";
import { RoomLitePayload, UserInfo } from "../../../bindings";
import { handleRoomSelect, MessageMap } from "./useRooms";
import { Dispatch, SetStateAction } from "react";
import { arrayMove } from "@dnd-kit/sortable";

export interface DragEndContext {
  userData: UserInfo;
  allRooms: RoomLitePayload[];
  selectedRooms: RoomLitePayload[];
  currentRegion: REGION;
}

export interface DragEndSetters {
  setAllMessages: Dispatch<SetStateAction<MessageMap>>;
  setCurrentRegion: Dispatch<SetStateAction<REGION>>;
  setSelectedRooms: Dispatch<SetStateAction<RoomLitePayload[]>>;
  setActiveRoomIndex: Dispatch<SetStateAction<number>>;
  setActiveId: Dispatch<SetStateAction<string | null>>;
  setRooms: Dispatch<SetStateAction<RoomLitePayload[]>>;
}

export const handleDragStart = (
  event: DragStartEvent,
  setActiveId: Dispatch<SetStateAction<string | null>>,
) => {
  setActiveId(event.active.id.toString());
};

export const getRoomById = (roomId: string, rooms: RoomLitePayload[]) => {
  return rooms.find((r) => r.id.toString() === roomId);
};

export const detectRegion = (event: DragEndEvent | DragOverEvent): REGION => {
  const { over, activatorEvent, delta } = event;
  if (!over) {
    return null;
  }

  const element = document.querySelector(`[id="${over.id}"]`);
  if (!element) {
    return null;
  }

  const rect = element.getBoundingClientRect();

  const pointerEvent = activatorEvent as PointerEvent;
  if (!pointerEvent) {
    return null;
  }

  const currentX = pointerEvent.clientX + delta.x;
  const currentY = pointerEvent.clientY + delta.y;

  const x = currentX - rect.left;
  const y = currentY - rect.top;

  const width = rect.width;
  const height = rect.height;

  const threshold = 0.3;

  if (y < height * threshold) {
    return "top";
  } else if (y > height * (1 - threshold)) {
    return "bottom";
  } else if (x < width * threshold) {
    return "left";
  } else if (x > width * (1 - threshold)) {
    return "right";
  }

  return null;
};

export const handleDragMove = (
  event: DragMoveEvent,
  setCurrentRegion: Dispatch<SetStateAction<REGION>>,
  setCurrentRegionID: Dispatch<SetStateAction<string | null>>,
) => {
  const { over } = event;

  if (!over) {
    setCurrentRegion(null);
    setCurrentRegionID(null);
    return;
  }

  if (over.id.toString().startsWith("chat-")) {
    setCurrentRegionID(over.id.toString());
    const region = detectRegion(event);
    setCurrentRegion(region);
  } else {
    setCurrentRegion(null);
    setCurrentRegionID(null);
  }
};

export const handleDragOver = (
  event: DragOverEvent,
  setCurrentRegion: Dispatch<SetStateAction<REGION>>,
  setCurrentRegionID: Dispatch<SetStateAction<string | null>>,
) => {
  const { over } = event;

  if (!over) {
    setCurrentRegion(null);
    setCurrentRegionID(null);
    return;
  }

  if (over.id.toString().startsWith("chat-")) {
    setCurrentRegionID(over.id.toString());
  }
};

export const handleDragEnd = (
  event: DragEndEvent,
  context: DragEndContext,
  setters: DragEndSetters,
) => {
  const { active, over } = event;

  if (!over) {
    setters.setActiveId(null);
    setters.setCurrentRegion(null);
    return;
  }

  // Check if dropping on a chat window
  if (over.id.toString().startsWith("chat-")) {
    const targetRoomId = over.id.toString().replace("chat-", "");
    const room = getRoomById(active.id.toString(), context.allRooms);

    if (room && context.currentRegion) {
      //get current region
      const region = detectRegion(event);
      handleRoomSelect(
        context.userData,
        room,
        context.selectedRooms,
        setters.setAllMessages,
        setters.setSelectedRooms,
        setters.setActiveRoomIndex,
        region as string,
        targetRoomId as string,
      );
    }
    setters.setActiveId(null);
    setters.setCurrentRegion(null);
    return;
  }

  // Handle reordering in sidebar
  if (active.id !== over.id && over.id === "side-bar") {
    setters.setRooms((rooms) => {
      const oldIndex = rooms.findIndex((r) => r.id.toString() === active.id);
      const newIndex = rooms.findIndex((r) => r.id.toString() === over.id);
      return arrayMove(rooms, oldIndex, newIndex);
    });
  }

  setters.setActiveId(null);
  setters.setCurrentRegion(null);
};
