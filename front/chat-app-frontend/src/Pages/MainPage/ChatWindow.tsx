import { RoomLitePayload } from "../../bindings";
import { DroppableZone, REGION } from "../../Components/DroppableZone";
import { MessageMap } from "./Hooks/useRooms";
import { ChatArea } from "./ChatArea";
import { NewMessageMap } from "./Hooks/useRoomMessages";
import { Dispatch, SetStateAction } from "react";

interface ChatWindowProps {
  room: RoomLitePayload;
  currentRegionId: string | null;
  selectedRooms: RoomLitePayload[];
  currentRegion: REGION;
  allMessages: MessageMap;
  newMessage: NewMessageMap;
  onClose: () => void;
  onMessageChange: (v: string) => void;
  onSendMessage: (e: React.FormEvent) => void;
  setAllMessages: Dispatch<SetStateAction<MessageMap>>;
}

export function ChatWindow({
  room,
  currentRegionId,
  currentRegion,
  selectedRooms,
  allMessages,
  newMessage,
  onClose,
  onMessageChange,
  onSendMessage,
  setAllMessages,
}: ChatWindowProps) {
  return (
    <DroppableZone
      key={room.id}
      id={`chat-${room.id}`}
      className="chat-window"
      hoveredRegion={
        currentRegionId === `chat-${room.id}` ? currentRegion : null
      }
    >
      <div className="chat-header">
        <span className="chat-header-title"># {room.name}</span>
        {selectedRooms.length > 1 && (
          <button className="close-chat-btn" onClick={onClose}>
            ×
          </button>
        )}
      </div>
      <ChatArea
        selectedRoom={{ id: room.id, name: room.name }}
        messages={allMessages[room.id] || []}
        newMessage={newMessage[room.id] || ""}
        onMessageChange={onMessageChange}
        onSendMessage={onSendMessage}
        setAllMessages={setAllMessages}
      />
    </DroppableZone>
  );
}
