import { useRef, useEffect } from "react";
import { MessageResponse } from "../../bindings";
import { useUser } from "../../Context/userContext";
import { Virtuoso } from "react-virtuoso";

interface ChatAreaProps {
  selectedRoom: { id: string; name: string };
  messages: MessageResponse[];
  newMessage: string;
  onMessageChange: (value: string) => void;
  onSendMessage: (e: React.FormEvent) => void;
}

export function ChatArea({
  selectedRoom,
  messages,
  newMessage,
  onMessageChange,
  onSendMessage,
}: ChatAreaProps) {
  const chatLogRef = useRef<HTMLDivElement>(null);
  const { userData } = useUser();

  useEffect(() => {
    if (chatLogRef.current) {
      chatLogRef.current.scrollTop = chatLogRef.current.scrollHeight;
    }
  }, [messages, selectedRoom]);


  return (
    <div className="chat-area">
      <Virtuoso
        className={`chat-log`}
        totalCount={messages.length}
        data={messages}
        followOutput="smooth"

        //start from last index
        initialTopMostItemIndex={messages.length - 1}
        itemContent={(index, msg) => (
          <div className={`chat-message ${msg.owner_name === userData?.username ? "me" : "other"}`}>
            <div className="message-user">{msg.owner_name}</div>
            <div className="message-text">{msg.content}</div>
          </div>
        )}
      />
      <form className="message-form" onSubmit={onSendMessage}>
        <input
          type="text"
          className="message-input"
          placeholder={`Message #${selectedRoom.name.toLowerCase()}`}
          value={newMessage}
          onChange={(e) => onMessageChange(e.target.value)}
        />
        <button type="submit" className="send-button">
          Send
        </button>
        <button type="button" className="send-button">
          📎
        </button>
        <button type="button" className="send-button">
          🎤
        </button>
      </form>
    </div>
  );
}
