import { Dispatch, SetStateAction } from "react";
import { MessageMap } from "./useRooms";
import { UserInfo } from "../../../bindings";
import { MessagePayload, MessageResponse, commands } from "../../../bindings";
import { Message } from "../../../db";

export const handleSendMessage =
  (
    roomId: string,
    userData: UserInfo,
    newMessage: NewMessageMap,
    setAllMessages: Dispatch<SetStateAction<MessageMap>>,
    setNewMessage: Dispatch<SetStateAction<NewMessageMap>>,
    saveChatData: (mgs: Message) => void,
  ) =>
  (e: React.FormEvent) => {
    e.preventDefault();

    if (!userData || !userData.id) {
      console.error("Current user data is null");
      return;
    }

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
      username: userData.username,
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
      username: userData.username,
      content: messageText,
      timeStamp: timestamp,
    };

    saveChatData(dbMessage);
  };

export type NewMessageMap = { [key: string]: string };
export const handleMessageChange =
  (roomId: string, setNewMessage: Dispatch<SetStateAction<NewMessageMap>>) =>
  (value: string) => {
    setNewMessage((prev) => ({ ...prev, [roomId]: value }));
  };

export const handleEditMessage = (
  messageId: string,
  roomId: string,
  value: string,
  setAllMessages: Dispatch<SetStateAction<MessageMap>>,
) => {
  setAllMessages((prev) => {
    return {
      ...prev,
      [roomId]: prev[roomId].map((msg) => {
        if (msg.id === messageId) {
          return { ...msg, content: value };
        }
        return msg;
      }),
    };
  });
};
