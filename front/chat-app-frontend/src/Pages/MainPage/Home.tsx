import "../../App.css";
import { useState, useEffect } from "react";
import JoinRoomPopUp from "./JoinRoomPopUp";
import {
  DndContext,
  DragEndEvent,
  DragStartEvent,
  PointerSensor,
  useSensor,
  useSensors,
  DragOverlay,
  closestCenter,
  DragOverEvent,
  DragMoveEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  arrayMove,
} from "@dnd-kit/sortable";
import { RoomListBar } from "../../Components/RoomListBar";
import { DroppableZone, REGION } from "../../Components/DroppableZone";
// import { ChatArea } from "../../Components/ChatArea";
import { ChatArea } from "./ChatArea";

const initialRoomsData = [
  { id: 1, name: "General", icon: "G" },
  { id: 2, name: "Tech Talk", icon: "T" },
  { id: 3, name: "Project Phoenix", icon: "P" },
  { id: 4, name: "Random", icon: "R" },
];

const initialMessages: { [key: number]: any[] } = {
  1: [
    // General
    {
      id: 1,
      user: "Alice",
      text: "Hey everyone!",
      sender: "other",
      timestamp: "10:30 AM",
    },
    {
      id: 2,
      user: "You",
      text: "Hi Alice! How are you?",
      sender: "me",
      timestamp: "10:31 AM",
    },
    {
      id: 3,
      user: "Bob",
      text: "Welcome to the chat!",
      sender: "other",
      timestamp: "10:32 AM",
    },
  ],
  2: [
    // Tech Talk
    {
      id: 1,
      user: "Charlie",
      text: "Has anyone tried the new React 19 features?",
      sender: "other",
      timestamp: "11:00 AM",
    },
    {
      id: 2,
      user: "You",
      text: "Not yet, but I'm excited about the compiler.",
      sender: "me",
      timestamp: "11:01 AM",
    },
  ],
  3: [
    // Project Phoenix
    {
      id: 1,
      user: "Diana",
      text: "Weekly sync: What's everyone's status?",
      sender: "other",
      timestamp: "09:00 AM",
    },
    {
      id: 2,
      user: "Eve",
      text: "I've pushed the latest updates to the auth branch.",
      sender: "other",
      timestamp: "09:02 AM",
    },
    {
      id: 3,
      user: "You",
      text: "I'll be reviewing them this morning.",
      sender: "me",
      timestamp: "09:03 AM",
    },
  ],
  4: [
    // Random
    {
      id: 1,
      user: "Frank",
      text: "What's a good movie to watch this weekend?",
      sender: "other",
      timestamp: "03:45 PM",
    },
  ],
};

function HomePage() {
  const [activeId, setActiveId] = useState<string | null>(null);
  const [rooms, setRooms] = useState(initialRoomsData);
  const [allMessages, setAllMessages] = useState(initialMessages);
  const [newMessage, setNewMessage] = useState<{ [key: number]: string }>({});
  const [isJoinRoomPopupOpen, setIsJoinRoomPopupOpen] = useState(false);

  const [selectedRooms, setSelectedRooms] = useState([rooms[0]]);
  const [activeRoomIndex, setActiveRoomIndex] = useState(0); // for knowing which room is in focus

  const [currentRegion, setCurrentRegion] = useState<REGION>(null);
  const [currentRegionId, setCurrentRegionID] = useState<String | null>(null);

  const handleSendMessage = (roomId: number) => (e: React.FormEvent) => {
    e.preventDefault();
    const messageText = newMessage[roomId] || "";

    if (messageText.trim() === "") return;
    const message = {
      id: (allMessages[roomId]?.length || 0) + 1,
      user: "You",
      text: messageText,
      sender: "me" as const,
      timestamp: new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    };
    setAllMessages((prevMessages) => ({
      ...prevMessages,
      [roomId]: [...(prevMessages[roomId] || []), message],
    }));
    setNewMessage("");
  };

  const handleMessageChange = (roomId: number) => (value: string) => {
    setNewMessage((prev) => ({ ...prev, [roomId]: value }));
  };

  const handleJoinRoom = () => {
    setIsJoinRoomPopupOpen(true);
  };

  const handleClosePopup = () => {
    setIsJoinRoomPopupOpen(false);
  };

  const handleRoomSelect = (room: (typeof rooms)[0]) => {
    const roomIndex = selectedRooms.findIndex(
      (r) => r.id.toString() === room.id.toString(),
    );

    //not found
    if (roomIndex === -1) {
      setSelectedRooms((prev) => [...prev, room]);
      setActiveRoomIndex(selectedRooms.length);
    } else {
      setActiveRoomIndex(roomIndex);
    }
  };

  const handleCloseRoom = (roomsId: string) => {
    const roomIndex = selectedRooms.findIndex(
      (r) => r.id.toString() === roomsId,
    );

    setSelectedRooms((prev) => prev.filter((r) => r.id.toString() !== roomsId));

    if (activeRoomIndex >= selectedRooms.length - 1) {
      setActiveRoomIndex(Math.max(0, selectedRooms.length - 2));
    } else if (roomIndex <= activeRoomIndex) {
      setActiveRoomIndex(Math.max(0, activeRoomIndex - 1));
    }
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id.toString());
  };

  const getRoomById = (roomId: string) => {
    return rooms.find((r) => r.id.toString() === roomId);
  };

  const detectRegion = (event: DragEndEvent | DragOverEvent): REGION => {
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
      //Spawn window top
      return "top";
    }
      
    if (y > height * (1 - threshold)) {
      //Spawn window bottom
      return "bottom";
    } 

    if (x < width * threshold) {
      //Spawn window left

      return "left";
    } 
    
    if (x > width * (1 - threshold)) {
      //Spawn window right
      return "right";
    } 

    return null;
  };

  const handleDragMove = (event: DragMoveEvent) => {
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

  const handleDragOver = (event: DragOverEvent) => {
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

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over) {
      setActiveId(null);
      setCurrentRegion(null);
      return;
    }

    // Check if dropping on a chat window
    if (over.id.toString().startsWith("chat-")) {
      // const roomId = parseInt(over.id.toString().replace("chat-", ""));
      const room = getRoomById(active.id.toString());
      if (room) {
        handleRoomSelect(room);
      }
      setActiveId(null);
      setCurrentRegion(null);
      return;
    }

    // Handle reordering in sidebar
    if (active.id !== over.id && over.id === "side-bar") {
      setRooms((rooms) => {
        const oldIndex = rooms.findIndex((r) => r.id.toString() === active.id);
        const newIndex = rooms.findIndex((r) => r.id.toString() === over.id);
        return arrayMove(rooms, oldIndex, newIndex);
      });
    }

    setActiveId(null);
    setCurrentRegion(null);
  };

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
  );

  const activeRoom = rooms.find((r) => r.id.toString() === activeId);

  return (
    <>
      {isJoinRoomPopupOpen && <JoinRoomPopUp onClose={handleClosePopup} />}
      <div className="app-container">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
          onDragOver={handleDragOver}
          onDragMove={handleDragMove}
        >
          <div className="wrapper">
            <div className="side-bar">
              <DroppableZone id="side-bar" className="room-list">
                <button onClick={handleJoinRoom} title="Join a new room">
                  +
                </button>
                <SortableContext
                  items={rooms.map((r) => r.id.toString())}
                  strategy={verticalListSortingStrategy}
                >
                  {rooms.map((room) => (
                    <RoomListBar
                      id={room.id.toString()}
                      key={room.id}
                      className="room-icon"
                      title={room.name}
                      onClick={() => handleRoomSelect(room)}
                      style={{
                        backgroundColor: selectedRooms.some(
                          (r) => r.id === room.id,
                        )
                          ? "#5865F2"
                          : undefined,
                      }}
                    >
                      {room.icon}
                    </RoomListBar>
                  ))}
                </SortableContext>
              </DroppableZone>
            </div>
            <div className="main-container">
              <div className="multi-chat-container">
                {selectedRooms.map((room) => (
                  <DroppableZone
                    key={room.id}
                    id={`chat-${room.id}`}
                    className="chat-window"
                    hoveredRegion={
                      currentRegionId === `chat-${room.id}`
                        ? currentRegion
                        : null
                    }
                  >
                    <div className="chat-header">
                      <span className="chat-header-title"># {room.name}</span>
                      {selectedRooms.length > 1 && (
                        <button
                          className="close-chat-btn"
                          onClick={() => handleCloseRoom(room.id.toString())}
                        >
                          ×
                        </button>
                      )}
                    </div>
                    <ChatArea
                      selectedRoom={room}
                      messages={allMessages[room.id] || []}
                      newMessage={newMessage[room.id] || ""}
                      onMessageChange={handleMessageChange(room.id)}
                      onSendMessage={handleSendMessage(room.id)}
                    />
                  </DroppableZone>
                ))}
              </div>
            </div>
          </div>
          <DragOverlay>
            {activeId && activeRoom ? (
              <div
                className="room-icon"
                style={{
                  cursor: "grabbing",
                  transform: "scale(1.1) rotate(5deg)",
                  boxShadow: "0 20px 40px rgba(0,0,0,0.4)",
                  backgroundColor: selectedRooms.some(
                    (r) => r.id === activeRoom.id,
                  )
                    ? "#5865F2"
                    : undefined,
                }}
              >
                {activeRoom.icon}
              </div>
            ) : null}
          </DragOverlay>
        </DndContext>
      </div>
    </>
  );
}

export default HomePage;
