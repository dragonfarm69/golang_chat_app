import "../../App.css";
import { useState, useEffect, useRef } from "react";
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
import { RoomButton } from "../../Components/RoomButton";
import { DroppableZone, REGION } from "../../Components/DroppableZone";
// import { ChatArea } from "../../Components/ChatArea";
import { ChatArea } from "./ChatArea";
import { useChatData } from "../../Context/DataContext";
import { useUser } from "../../Context/userContext";
import { Message } from "../../db";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import {
  commands,
  MessagePayload,
  MessageResponse,
  RoomLitePayload,
} from "../../bindings";
import { Store } from "@tauri-apps/plugin-store";

function HomePage() {
  const { userData } = useUser();
  const isConnected = useRef(false);
  const BUTTON_FLASHING_ANIMATION_TIMER = 2000;
  const ROOMS_CAP = 6; // the amount of rooms can be opened at the same time
  const [buttonFlashing, setIsButtonFlashing] = useState<Set<String>>(
    new Set(),
  ); // to track which button should be flashing at the moment

  const [activeId, setActiveId] = useState<string | null>(null);
  const [rooms, setRooms] = useState<RoomLitePayload[]>([]);
  const [allMessages, setAllMessages] = useState<{
    [key: string]: MessageResponse[];
  }>({});
  const [newMessage, setNewMessage] = useState<{ [key: string]: string }>({});
  const [isJoinRoomPopupOpen, setIsJoinRoomPopupOpen] = useState(false);

  const [selectedRooms, setSelectedRooms] = useState<RoomLitePayload[]>([]);
  const [activeRoomIndex, setActiveRoomIndex] = useState(0); // for knowing which room is in focus

  const [currentRegion, setCurrentRegion] = useState<REGION>(null);
  const [currentRegionId, setCurrentRegionID] = useState<String | null>(null);

  const { saveChatData, deleteChatData, updateMessage } = useChatData();

  const handleSendMessage = (roomId: string) => (e: React.FormEvent) => {
    if (!userData || !userData.id) {
      console.error("Current user data is null");
      return;
    }

    e.preventDefault();
    const messageText = newMessage[roomId] || "";

    const timestamp = new Date().toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });

    const tempId = crypto.randomUUID();

    if (messageText.trim() === "") return;

    const messagePayload: MessagePayload = {
      id: tempId,
      user_id: userData.id,
      room_id: roomId,
      content: messageText,
      timeStamp: timestamp,
      action: "SEND",
    };

    const messageResponse: MessageResponse = {
      id: tempId,
      owner_name: userData.username || "me",
      room_id: roomId,
      content: messageText,
      timeStamp: timestamp,
    };

    setAllMessages((prevMessages) => ({
      ...prevMessages,
      [roomId]: [...(prevMessages[roomId] || []), messageResponse],
    }));
    setNewMessage((prev) => ({ ...prev, [roomId]: "" }));

    commands.sendMessage(messagePayload);

    //save to indexDB
    const dbMessage: Message = {
      id: tempId,
      room_id: roomId.toString(),
      user_id: userData.id,
      content: messageText,
      timeStamp: timestamp,
    };

    saveChatData(dbMessage);
  };

  const handleMessageChange = (roomId: string) => (value: string) => {
    setNewMessage((prev) => ({ ...prev, [roomId]: value }));
  };

  const handleJoinRoom = async (roomName: string) => {
    if (!userData || !userData.id) {
      console.error("Current user data is null");
      return;
    }

    const stats = await commands.joinRoom(userData?.id, roomName)
    if(stats.status === "ok") {
      if (stats.data === true) {
          const roomList = await commands.fetchRoomsList(userData.id);
          if (roomList.status === "ok") {
            if (roomList.data != null) {
              setRooms((prev) => [...prev, ...roomList.data]);
            }
          } else {
            console.error("Error fetching rooms: ", roomList.error);
        }
      }
    }
  }

  const handleJoinRoomPopUpOpen = () => {
    setIsJoinRoomPopupOpen(true);
  };

  const handleClosePopup = () => {
    setIsJoinRoomPopupOpen(false);
  };

  useEffect(() => {
    if (!userData || !userData.id) {
      return;
    }

    if (!isConnected.current) {
      invoke("establish_ws", {
        userId: userData?.id,
      })
        .then(() => {
          isConnected.current = true;
        })
        .catch(console.error);

      //fetch room list
      async function fetchRoomList() {
        if (userData && userData?.id !== null) {
          const roomList = await commands.fetchRoomsList(userData.id);
          if (roomList.status === "ok") {
            if (roomList.data != null) {
              setRooms((prev) => [...prev, ...roomList.data]);
            }
          } else {
            console.error("Error fetching rooms: ", roomList.error);
          }
        } else {
          console.error("Error fetching rooms user data is null");
        }
      }

      fetchRoomList();

      // Object Prototype
      // Listen to WS messages
      const unlisten = listen("ws-message", (event) => {
        const msg = JSON.parse(event.payload as string);
        const new_msg_data: Message = {
          id: msg.id,
          user_id: userData.id,
          room_id: msg.room_id,
          content: msg.content,
          timeStamp: msg.timeStamp,
        };

        updateMessage(msg.original_id, new_msg_data);
      });
      return () => {
        unlisten.then((stop) => stop());
      };
    }
  }, [userData?.id]);

  const handleRoomSelect = async (
    room: (typeof rooms)[0],
    location?: string,
    droppedOnRoomWithId?: string,
  ) => {
    if (!userData || !userData.id) {
      console.error("Current user data is null");
      return;
    }

    //prevent user from opening more room
    if (selectedRooms.length === ROOMS_CAP) {
      return;
    }

    const roomIndex = selectedRooms.findIndex(
      (r) => r.id.toString() === room.id,
    );

    //not found
    if (roomIndex === -1) {
      const tempId = crypto.randomUUID();
      const messagePayload: MessagePayload = {
        id: tempId,
        user_id: userData.id,
        room_id: room.id,
        content: "",
        timeStamp: "",
        action: "JOIN",
      };
      commands.sendMessage(messagePayload);

      const result = await commands.fetchRoomMessages(room.id);
      if (result.status === "ok") {
        if (result.data != null) {
          setAllMessages((prev) => ({ ...prev, [room.id]: result.data }));
        }
      } else {
        console.error("Error fetching rooms: ", result.error);
      }
      // console.log("room_messages: ", result);

      if (location) {
        //get the needed index to add the chat room to rooms
        const dropIndex = selectedRooms.findIndex(
          (r) => r.id.toString() === droppedOnRoomWithId,
        );
        // console.log(selectedRooms);
        // console.log(
        //   "dropped on index: ",
        //   rooms.findIndex((r) => r.id.toString() === droppedOnRoomWithId),
        // );

        switch (location) {
          case "top":
            break;
          case "bottom":
            break;
          case "left":
            setSelectedRooms((prev) => {
              const newRooms = [...prev];
              newRooms.splice(dropIndex, 0, room);
              return newRooms;
            });
            break;
          case "right":
            if (dropIndex < selectedRooms.length) {
              setSelectedRooms((prev) => {
                const newRooms = [...prev];
                newRooms.splice(dropIndex + 1, 0, room);
                return newRooms;
              });
            } else {
              setSelectedRooms((prev) => [...prev, room]);
            }
            break;
          default:
            setSelectedRooms((prev) => [...prev, room]);
            break;
        }
      } else {
        setSelectedRooms((prev) => [...prev, room]);
      }
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
      return "top";
    }

    if (y > height * (1 - threshold)) {
      return "bottom";
    }

    if (x < width * threshold) {
      return "left";
    }

    if (x > width * (1 - threshold)) {
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
      const targetRoomId = over.id.toString().replace("chat-", "");
      const room = getRoomById(active.id.toString());

      if (room && currentRegion) {
        //get current region
        const region = detectRegion(event);
        // console.log("current region: ", region)
        // console.log("dropped on: ", targetRoomId)
        handleRoomSelect(room, region as string, targetRoomId as string);
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
      {isJoinRoomPopupOpen && <JoinRoomPopUp onClose={handleClosePopup} onSubmit={handleJoinRoom} />}
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
                <button onClick={handleJoinRoomPopUpOpen} title="Join a new room">
                  +
                </button>
                <SortableContext
                  items={rooms.map((r) => r.id.toString())}
                  strategy={verticalListSortingStrategy}
                >
                  {rooms.map((room) => (
                    <RoomButton
                      id={room.id.toString()}
                      key={room.id}
                      title={room.name}
                      flashing={buttonFlashing.has(room.id.toString())}
                      onClick={() => {
                        if (selectedRooms.length === ROOMS_CAP) {
                          //get the button id and set state
                          setIsButtonFlashing((prev) => {
                            const newSet = new Set(prev);
                            newSet.add(room.id.toString());

                            //set time out to reset the button state
                            setTimeout(() => {
                              setIsButtonFlashing((prev) => {
                                const buttonSetAfterDelete = new Set(prev);
                                buttonSetAfterDelete.delete(room.id.toString());
                                return buttonSetAfterDelete;
                              });
                            }, BUTTON_FLASHING_ANIMATION_TIMER);
                            return newSet;
                          });
                          return;
                        }
                        handleRoomSelect(room);
                      }}
                      style={{
                        backgroundColor: selectedRooms.some(
                          (r) => r.id === room.id,
                        )
                          ? "#5865F2"
                          : undefined,
                      }}
                    >
                      {"r"}
                    </RoomButton>
                  ))}
                </SortableContext>
              </DroppableZone>
            </div>
            <div className="main-container">
              <div className={"multi-chat-container"}>
                {selectedRooms.slice(0, 3).map((room) => (
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
                      selectedRoom={{ id: room.id, name: room.name }}
                      messages={allMessages[room.id] || []}
                      newMessage={newMessage[room.id] || ""}
                      onMessageChange={handleMessageChange(room.id)}
                      onSendMessage={handleSendMessage(room.id)}
                    />
                  </DroppableZone>
                ))}
              </div>

              {selectedRooms.length >= 4 && (
                <div className={"multi-chat-container"}>
                  {selectedRooms.slice(3, 6).map((room) => (
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
              )}
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
                {"R"}
              </div>
            ) : null}
          </DragOverlay>
        </DndContext>
      </div>
    </>
  );
}

export default HomePage;
