import {
  useRef,
  useEffect,
  useCallback,
  Dispatch,
  SetStateAction,
  useState,
} from "react";
import { commands, MessageResponse } from "../../bindings";
import { useUser } from "../../Context/userContext";
import { Virtuoso } from "react-virtuoso";
import { MessageMap } from "./Hooks/useRooms";
import { MessagePayload } from "../../bindings";

interface ChatAreaProps {
  selectedRoom: { id: string; name: string };
  messages: MessageResponse[];
  newMessage: string;
  typingUser: string[];
  onMessageChange: (value: string) => void;
  onSendMessage: (e: React.FormEvent) => void;
  setAllMessages: Dispatch<SetStateAction<MessageMap>>;
}

export function ChatArea({
  selectedRoom,
  messages,
  typingUser,
  newMessage,
  onMessageChange,
  onSendMessage,
  setAllMessages,
}: ChatAreaProps) {
  const [firstItemIdex, setFirstItemIndex] = useState(1_000_000);
  const isLoadingMore = useRef(true);
  const chatLogRef = useRef<HTMLDivElement>(null);
  const { userData } = useUser();

  const [editingMessageId, setEditingMessageId] = useState<string>("");
  const [editValue, setEditValue] = useState<string>("");

  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isTypingRef = useRef(false);

  const [showOptions, setShowOptions] = useState<string | null>(null);

  useEffect(() => {
    if (chatLogRef.current) {
      chatLogRef.current.scrollTop = chatLogRef.current.scrollHeight;
    }
  }, [messages, selectedRoom]);

  useEffect(() => {
    setFirstItemIndex(1_000_000);
    isLoadingMore.current = true; // block loadMore until fresh messages arrive
  }, [selectedRoom.id]);

  useEffect(() => {
    if (messages.length > 0) {
      isLoadingMore.current = false; // messages loaded, allow loadMore on scroll
    }
  }, [messages]);

  const triggerEventTyping = () => {
    if (!userData || !userData.id) {
      console.error("Current user data is null");
      return;
    }

    const timestamp = new Date().toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });

    const tempId = crypto.randomUUID();
    if (!isTypingRef.current) {
      isTypingRef.current = true;

      const messagePayload: MessagePayload = {
        id: tempId,
        user_id: userData.id,
        username: userData.username,
        room_id: selectedRoom.id,
        content: "",
        timeStamp: timestamp,
        action: "TYPING",
      };

      commands.sendMessage(messagePayload);
    }

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    //create timer
    typingTimeoutRef.current = setTimeout(() => {
      isTypingRef.current = false;
      const stopTypingMessagePayload: MessagePayload = {
        id: tempId,
        user_id: userData.id,
        username: userData.username,
        room_id: selectedRoom.id,
        content: "",
        timeStamp: timestamp,
        action: "STOP_TYPING",
      };
      commands.sendMessage(stopTypingMessagePayload);
    }, 2000);
  };

  const triggerEventEditing = (msgId: string, value: string) => {
    if (!userData || !userData.id) {
      console.error("Current user data is null");
      return;
    }

    const timestamp = new Date().toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });

    const messagePayload: MessagePayload = {
      id: msgId,
      user_id: userData.id,
      username: userData.username,
      room_id: selectedRoom.id,
      content: value,
      timeStamp: timestamp,
      action: "EDIT",
    };

    console.log("Trigger edit event: ", messagePayload);

    commands.sendMessage(messagePayload);
    setEditValue(""); //reset edit value
  };

  const loadMore = useCallback(async () => {
    if (isLoadingMore.current || messages.length === 0) return;
    isLoadingMore.current = true;
    console.log("Reached old messages, fetching more ....");
    //load more data
    const result = await commands.fetchRoomMessages(
      selectedRoom.id,
      messages[0].id,
    );

    if (result.status === "ok") {
      if (result.data != null) {
        const reversedData = result.data.reverse();
        setFirstItemIndex((prev) => prev - reversedData.length);
        setAllMessages((prev) => ({
          ...prev,
          [selectedRoom.id]: [...reversedData, ...prev[selectedRoom.id]],
        }));
      }
    } else {
      console.error("Error fetching rooms: ", result.error);
    }

    isLoadingMore.current = false;
  }, [messages, selectedRoom.id, setAllMessages]);

  // const handleOptionClick = (msg: MessageResponse) => {
  //   console.log("Option clicked for message: ", msg);
  // };

  return (
    <div className="chat-area">
      <Virtuoso
        className={`chat-log`}
        totalCount={messages.length}
        data={messages}
        followOutput="smooth"
        startReached={loadMore}
        firstItemIndex={firstItemIdex}
        //start from last index
        initialTopMostItemIndex={messages.length - 1}
        onMouseDown={() => {
          setShowOptions(null);
        }}
        itemContent={(index, msg) => (
          <div
            style={{ position: "relative" }}
            onMouseLeave={() => {
              setShowOptions(null);
            }}
          >
            {showOptions === msg.id && (
              <div
                className="options-popup"
                onMouseDown={(e) => e.stopPropagation()}
                style={{
                  position: "absolute",
                  bottom: "calc(55% + 15px)",
                  right: "1px",
                  padding: "6px 10px",
                  display: "flex",
                  gap: 4,
                  zIndex: 10,
                }}
              >
                <button
                  onClick={() => {
                    setEditingMessageId(msg.id);
                    setEditValue(msg.content);
                    console.log("editing: ", editingMessageId);
                  }}
                >
                  Edit
                </button>
                <button>Delete</button>
              </div>
            )}
            <div
              className={`chat-message ${msg.owner_name === userData?.username ? "me" : "other"}`}
            >
              <div className="message-content">
                <div className="message-user">{msg.owner_name}</div>
                {editingMessageId && editingMessageId === msg.id ? (
                  <>
                    <input
                      type="text"
                      value={editValue}
                      onChange={(e) => {
                        setEditValue(e.target.value);
                      }}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") {
                          e.preventDefault();
                          triggerEventEditing(msg.id, editValue);
                          setEditingMessageId("");
                        } else if (e.key === "Escape") {
                          e.preventDefault();
                          setEditValue("");
                          setEditingMessageId("");
                        }
                      }}
                    />
                    <span>
                      Press <kbd className="key-style">Esc</kbd> to exit. Press{" "}
                      <kbd className="key-style">Enter</kbd> to save.
                    </span>
                  </>
                ) : (
                  <div className="message-text">{msg.content}</div>
                )}
              </div>
              <div
                className="message-option-button"
                onMouseEnter={() => {
                  setShowOptions(msg.id);
                }}
              >
                :::
              </div>
            </div>
          </div>
        )}
      />
      {typingUser && typingUser.length > 0 && (
        <div className="typing-indicator">
          <span>{typingUser.join(", ") + " is typing "} </span>
        </div>
      )}
      <form className="message-form" onSubmit={onSendMessage}>
        <input
          type="text"
          className="message-input"
          placeholder={`Message #${selectedRoom.name.toLowerCase()}`}
          value={newMessage}
          onChange={(e) => {
            triggerEventTyping();
            onMessageChange(e.target.value);
          }}
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
