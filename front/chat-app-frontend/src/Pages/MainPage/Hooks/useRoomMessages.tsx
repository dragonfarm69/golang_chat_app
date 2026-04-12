import { Dispatch, SetStateAction, useCallback } from "react";
import { MessageMap } from "./useRooms";
import { UserInfo } from "../../../bindings";
import { MessagePayload, MessageResponse, commands } from "../../../bindings";
import { Message } from "../../../db";
import { useChatData } from "../../../Context/DataContext";

export const useSendMessage = (
  roomId: string,
  userData: UserInfo,
  newMessage: NewMessageMap,
  setAllMessages: Dispatch<SetStateAction<MessageMap>>,
  setNewMessage: Dispatch<SetStateAction<NewMessageMap>>,
) => {
  const { saveChatData } = useChatData();

  //use call back to cache the function, but it is heavy so don't use it everywhere
  const handleSendMessage = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();

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
    },
    [roomId, userData, newMessage, setAllMessages, setNewMessage, saveChatData],
  );

  return handleSendMessage;
};

type NewMessageMap = { [key: string]: string };
export const handleMessageChange =
  (roomId: string, setNewMessage: Dispatch<SetStateAction<NewMessageMap>>) =>
  (value: string) => {
    setNewMessage((prev) => ({ ...prev, [roomId]: value }));
  };
