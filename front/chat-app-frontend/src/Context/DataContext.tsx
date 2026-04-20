import { createContext, useContext, useEffect, useState } from "react";
import { ReactNode } from "react";
import { db, Message } from "../db";

interface ChatDataContextType {
  messages: Message[];
  saveChatData: (data: Message) => Promise<void>;
  deleteChatData: (id: string) => Promise<void>;
  clearData: () => Promise<void>;
  updateMessage: (
    origional_message_id: string,
    new_data: Message,
  ) => Promise<void>;
}

const chatDataContext = createContext<ChatDataContextType | undefined>(
  undefined,
);

export const ChatDataProvider = ({ children }: { children: ReactNode }) => {
  const [messages, setMessages] = useState<Message[]>([]);

  async function updateMessage(
    origional_message_id: string,
    new_data: Message,
  ) {
    try {
      // if updating the id, we delete and create new one
      if (origional_message_id != new_data.id) {
        await deleteChatData(origional_message_id).then(async () => {
          await saveChatData(new_data);
        });
      } else {
        const message_to_update = await db.messages
          .where("id")
          .equals(origional_message_id)
          .first();
        if (message_to_update) {
          await db.messages.update(message_to_update.id, {
            ...new_data,
          });
        }
      }
    } catch (e) {
      console.log("An error has occured when trying to update data in DB: ", e);
    }
  }

  async function saveChatData(data: Message) {
    try {
      const id = await db.messages.add(data);
    } catch (e) {
      console.log("An error has occured when trying to save data to DB: ", e);
    }
  }

  async function deleteChatData(id: string) {
    try {
      const exists = await db.messages.get(id);
      if (!exists) {
        console.warn(
          `Attempted to delete ID ${id}, but it doesn't exist in DB.`,
        );
        return;
      }

      await db.messages.delete(id);
      setMessages((prev) => prev.filter((msg) => msg.id !== id));
    } catch (e) {
      console.error("IndexedDB Delete Error:", e);
    }
  }

  async function clearData() {
    try {
      await db.messages.clear();
      setMessages([]);
    } catch (e) {
      console.error("Error while trying to clear data: ", e);
    }
  }

  return (
    <chatDataContext.Provider
      value={{
        saveChatData,
        deleteChatData,
        updateMessage,
        clearData,
        messages,
      }}
    >
      {children}
    </chatDataContext.Provider>
  );
};

export const useChatData = () => {
  const context = useContext(chatDataContext);
  if (!context) {
    throw new Error("useChatData must be used within a ChatDataProvider");
  }
  return context;
};
