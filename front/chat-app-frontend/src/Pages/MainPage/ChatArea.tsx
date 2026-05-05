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
import { Virtuoso, VirtuosoHandle } from "react-virtuoso";
import { MessageMap } from "./Hooks/useRooms";
import { MessagePayload } from "../../bindings";
import { ImageCard } from "../../Components/CustomImageComponent";
import { message, open } from "@tauri-apps/plugin-dialog";
import { convertFileSrc } from "@tauri-apps/api/core";
import { stat } from "@tauri-apps/plugin-fs";
import { FileMetaData } from "../../bindings";

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
  const mediaBaseURL = "http://localhost:8082";
  const virtuosoRef = useRef<VirtuosoHandle>(null);
  const [firstItemIdex, setFirstItemIndex] = useState(1_000_000);
  const isLoadingMore = useRef(true);
  const chatLogRef = useRef<HTMLDivElement>(null);
  const { userData } = useUser();

  const [editingMessageId, setEditingMessageId] = useState<string>("");
  const [editValue, setEditValue] = useState<string>("");

  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isTypingRef = useRef(false);

  const [showOptions, setShowOptions] = useState<string | null>(null);
  const [files, setFiles] = useState<string[]>([]);
  const [filePath, setFilePath] = useState<string[]>([]);
  const [fileMetadatas, setFileMetadatas] = useState<FileMetaData[]>([]);

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

  useEffect(() => {
    if (messages.length === 0) return;
    virtuosoRef.current?.scrollToIndex({
      index: messages.length - 1,
      behavior: "smooth",
    });
  }, [messages.length]);

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
    commands.editMessage(selectedRoom.id, msgId, value);
    setEditValue(""); //reset edit value
  };

  const triggerEventDelete = (msgId: string) => {
    commands.deleteMessage(selectedRoom.id, msgId);
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

  const handleRemoveMedia = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
    setFileMetadatas((prev) => prev.filter((_, i) => i !== index));

    console.log("after: ", files);
  };

  const handleRemoveAll = () => {
    setFiles([]);
    setFileMetadatas([]);
  };

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
                  }}
                >
                  Edit
                </button>
                <button
                  onClick={() => {
                    triggerEventDelete(msg.id);
                  }}
                >
                  Delete
                </button>
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
                        // 1. Guard against OS key-repeat (holding the key down)
                        if (e.repeat) return;

                        // 2. Guard against IME composition double-fires (e.g., UniKey / EVKey)
                        if (e.nativeEvent.isComposing || e.keyCode === 229)
                          return;
                        if (e.key === "Enter") {
                          e.preventDefault();
                          e.stopPropagation();
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
                ) : msg.message_type === "text" ? (
                  <div className="message-text">{msg.content}</div>
                ) : (
                  <img
                    src={`${mediaBaseURL}/media/insecure/rs:auto:800:0/plain/s3://chat-media/${msg.content}`}
                    alt="media"
                  />
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
      {files.length >= 1 && (
        <div className="media-preview-box">
          <div className="media-preview-ultilities">
            <button>Hide all</button>
            <button onClick={handleRemoveAll}>X</button>
          </div>
          <ul className="media-preview-content">
            {files.map((url, index) => (
              <li key={index}>
                <ImageCard
                  image={url}
                  handleRemove={() => handleRemoveMedia(index)}
                />
              </li>
            ))}
          </ul>
        </div>
      )}
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
        <button
          type="button"
          className="send-button"
          onClick={async () => {
            if (files.length >= 1) {
              if (!userData) {
                console.error("user info not found");
                return;
              }
              try {
                const presignedURLs = await commands.sendMedia(
                  fileMetadatas,
                  selectedRoom.id,
                  userData?.id,
                );

                console.log("got the urls: ", presignedURLs);
                if (presignedURLs.status === "error") {
                  console.error(
                    "Error when trying to upload file: ",
                    presignedURLs.error,
                  );
                  return;
                }

                await Promise.all(
                  presignedURLs.data.map((url, i) => {
                    console.log("sending file with url: ", url);
                    console.log("The file sent: ", filePath[i]);
                    commands.uploadFile(url, filePath[i]);
                  }),
                ).then(() => {
                  setFileMetadatas([]);
                  setFiles([]);
                  setFilePath([]);
                });
              } catch (err) {
                console.error("Error when trying to upload file: ", err);
              }
            }
          }}
        >
          Send
        </button>
        <button
          type="button"
          className="send-button"
          onClick={async () => {
            const file = await open({
              multiple: false,
              directory: false,
              filters: [{ name: "Images", extensions: ["png", "jpg", "webp"] }],
            });

            if (file) {
              try {
                const assetUrl = convertFileSrc(file);
                const file_data = file.split("/");
                const file_name = file_data[file_data.length - 1];
                const ext = file_name.split(".").pop()?.toLowerCase() ?? ""; //TODO: Handle weird case -> .txt.exam/.png.jsp
                const file_type = ((): string => {
                  switch (ext) {
                    // Images
                    case "png":
                      return "image/png";
                    case "jpg":
                    case "jpeg":
                      return "image/jpeg";
                    case "gif":
                      return "image/gif";
                    case "webp":
                      return "image/webp";
                    // Videos
                    case "mp4":
                      return "video/mp4";
                    case "webm":
                      return "video/webm";
                    case "mov":
                      return "video/quicktime";
                    default:
                      return `application/octet-stream`;
                  }
                })();

                setFilePath((prev) => [...prev, file]);
                await stat(file).then((data) => {
                  setFileMetadatas((prev) => [
                    ...prev,
                    {
                      file_name: file_name,
                      file_size: data.size.toString(),
                      file_type: file_type,
                    },
                  ]);
                });
                setFiles((prev) => [...prev, assetUrl]);
              } catch (error) {
                console.error("error when fetching file: ", error);
              }
            }
          }}
        >
          📎
        </button>
        <button type="button" className="send-button">
          🎤
        </button>
      </form>
    </div>
  );
}
