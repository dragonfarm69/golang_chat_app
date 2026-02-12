import { useRef, useEffect } from "react";
import { DroppableZone } from "../../Components/DroppableZone";

interface Message {
  id: number;
  user: string;
  text: string;
  sender: "me" | "other";
  timestamp: string;
}

interface ChatAreaProps {
  selectedRoom: { id: number; name: string; icon: string };
  messages: Message[];
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

  useEffect(() => {
    if (chatLogRef.current) {
      chatLogRef.current.scrollTop = chatLogRef.current.scrollHeight;
    }
  }, [messages, selectedRoom]);

  return (
    <div className="chat-area">
      <div className="chat-log" ref={chatLogRef}>
        {messages.map((msg) => (
          <div key={msg.id} className={`chat-message ${msg.sender}`}>
            <div className="message-user">{msg.user}</div>
            <div className="message-text">{msg.text}</div>
          </div>
        ))}
      </div>
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