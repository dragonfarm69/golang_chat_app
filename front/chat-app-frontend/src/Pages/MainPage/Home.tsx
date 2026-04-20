import "../../App.css";
import { useState, useEffect, useRef } from "react";
import JoinRoomPopUp from "./JoinRoomPopUp";
import {
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  DragOverlay,
  closestCenter,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { RoomButton } from "../../Components/RoomButton";
import { DroppableZone, REGION } from "../../Components/DroppableZone";
import { useChatData } from "../../Context/DataContext";
import { useUser } from "../../Context/userContext";
import { Message } from "../../db";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import { commands, MessageResponse, RoomLitePayload } from "../../bindings";
import {
  handleDragEnd,
  handleDragMove,
  handleDragOver,
  handleDragStart,
} from "./Hooks/useDragDrop";
import {
  handleCloseRoom,
  handlePopUpSubmit,
  handleRoomSelect,
} from "./Hooks/useRooms";
import { ChatWindow } from "./ChatWindow";
import {
  handleMessageChange,
  handleSendMessage,
} from "./Hooks/useRoomMessages";
import { useNavigate } from "react-router-dom";
import { OptionPopUp } from "../Profile/OptionPopUp";

function HomePage() {
  const navigate = useNavigate();
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
  const [isProfileOpen, setIsProfilePopen] = useState(false);

  const [selectedRooms, setSelectedRooms] = useState<RoomLitePayload[]>([]);
  const [activeRoomIndex, setActiveRoomIndex] = useState(0); // for knowing which room is in focus

  const [currentRegion, setCurrentRegion] = useState<REGION>(null);
  const [currentRegionId, setCurrentRegionID] = useState<string | null>(null);

  const { updateMessage, saveChatData } = useChatData();

  const [hideSideBar, setHideSideBar] = useState(true);

  // roomid - user
  const [typingUser, setTypingUser] = useState<Record<string, string[]>>({});

  const handleJoinRoomPopUpOpen = () => {
    setIsJoinRoomPopupOpen(true);
  };

  const handleProfilePopUpOpen = () => {
    setIsProfilePopen(true);
  };

  const handleClosePopup = () => {
    setIsJoinRoomPopupOpen(false);
  };

  const handleCloseProfilePopup = () => {
    setIsProfilePopen(false);
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
        // console.log("message: ", msg.action);
        if (msg.action === "TYPING") {
          setTypingUser((prev) => {
            const currentUsers = prev[msg.room_id] || [];
            if (currentUsers.includes(msg.username)) {
              return prev;
            }

            return {
              ...prev,
              [msg.room_id]: [...currentUsers, msg.username],
            };
          });
          return;
        }
        if (msg.action === "STOP_TYPING") {
          setTypingUser((prev) => {
            const currentUsers = prev[msg.room_id] || [];

            return {
              ...prev,
              [msg.room_id]: currentUsers.filter(
                (user) => user !== msg.username,
              ),
            };
          });
          return;
        }

        const new_msg_data: Message = {
          id: msg.id,
          user_id: userData.id,
          username: userData.username,
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
      {isJoinRoomPopupOpen && (
        <JoinRoomPopUp
          onClose={handleClosePopup}
          onSubmit={(roomName, isCreateroom) =>
            handlePopUpSubmit(roomName, isCreateroom, userData!, setRooms)
          }
        />
      )}
      {isProfileOpen && <OptionPopUp onClose={handleCloseProfilePopup} />}
      <div className="app-container">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={(e) => handleDragStart(e, setActiveId)}
          onDragEnd={(e) =>
            handleDragEnd(
              e,
              {
                userData: userData!,
                allRooms: rooms,
                selectedRooms,
                currentRegion,
              },
              {
                setAllMessages,
                setCurrentRegion,
                setSelectedRooms,
                setActiveRoomIndex,
                setActiveId,
                setRooms,
              },
            )
          }
          onDragOver={(e) =>
            handleDragOver(e, setCurrentRegion, setCurrentRegionID)
          }
          onDragMove={(e) =>
            handleDragMove(e, setCurrentRegion, setCurrentRegionID)
          }
        >
          <div className="wrapper">
            <div
              className={`side-bar ${hideSideBar ? "hidden" : ""}`}
              onMouseEnter={() => setHideSideBar(false)}
              onMouseLeave={() => setHideSideBar(true)}
            >
              {!hideSideBar && (
                <DroppableZone id="side-bar" className="room-list">
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
                                  buttonSetAfterDelete.delete(
                                    room.id.toString(),
                                  );
                                  return buttonSetAfterDelete;
                                });
                              }, BUTTON_FLASHING_ANIMATION_TIMER);
                              return newSet;
                            });
                            return;
                          }
                          handleRoomSelect(
                            userData!,
                            room,
                            selectedRooms,
                            setAllMessages,
                            setSelectedRooms,
                            setActiveRoomIndex,
                          );
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
                  <button
                    onClick={handleJoinRoomPopUpOpen}
                    title="Join a new room"
                  >
                    +
                  </button>
                  <button
                    onClick={() => {
                      handleProfilePopUpOpen();
                    }}
                    title="Profile"
                  >
                    P
                  </button>
                </DroppableZone>
              )}
            </div>
            <div className="main-container">
              <div className={"multi-chat-container"}>
                {selectedRooms.slice(0, 3).map((room) => (
                  <ChatWindow
                    room={room}
                    currentRegionId={currentRegionId}
                    selectedRooms={selectedRooms}
                    currentRegion={currentRegion}
                    allMessages={allMessages}
                    newMessage={newMessage}
                    typingUser={typingUser[room.id.toString()]}
                    setAllMessages={setAllMessages}
                    onClose={() =>
                      handleCloseRoom(
                        room.id.toString(),
                        selectedRooms,
                        setSelectedRooms,
                        setActiveRoomIndex,
                        activeRoomIndex,
                      )
                    }
                    onMessageChange={handleMessageChange(
                      room.id,
                      setNewMessage,
                    )}
                    onSendMessage={handleSendMessage(
                      room.id,
                      userData!,
                      newMessage,
                      setAllMessages,
                      setNewMessage,
                      saveChatData,
                    )}
                  />
                ))}
              </div>

              {selectedRooms.length >= 4 && (
                <div className={"multi-chat-container"}>
                  {selectedRooms.slice(3, 6).map((room) => (
                    <ChatWindow
                      room={room}
                      currentRegionId={currentRegionId}
                      selectedRooms={selectedRooms}
                      currentRegion={currentRegion}
                      allMessages={allMessages}
                      typingUser={typingUser[room.id.toString()]}
                      newMessage={newMessage}
                      setAllMessages={setAllMessages}
                      onClose={() =>
                        handleCloseRoom(
                          room.id.toString(),
                          selectedRooms,
                          setSelectedRooms,
                          setActiveRoomIndex,
                          activeRoomIndex,
                        )
                      }
                      onMessageChange={handleMessageChange(
                        room.id,
                        setNewMessage,
                      )}
                      onSendMessage={handleSendMessage(
                        room.id,
                        userData!,
                        newMessage,
                        setAllMessages,
                        setNewMessage,
                        saveChatData,
                      )}
                    />
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
