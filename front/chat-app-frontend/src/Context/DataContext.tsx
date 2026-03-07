import { createContext, useContext, useEffect, useState } from "react";
import { ReactNode } from "react";
import { db, Message } from "../db";

interface ChatDataContextType {
    messages: Message[],
    saveChatData: (data: Message) => Promise<void>,
    deleteChatData: (id: string) => Promise<void>,
}

const chatDataContext = createContext<ChatDataContextType | undefined>(undefined);

export const ChatDataProvider = ({children}: {children: ReactNode}) => {
    const [messages, setMessages] = useState<Message[]>([]);
    

    async function saveChatData(data: Message) {
        try {
            const id = await db.messages.add(data)
            console.log("An item has been added to db: ", id)
        }
        catch (e) {
            console.log("An error has occured when trying to save data to DB: ", e)
        }
    }

    async function deleteChatData(id: string) {
        try {
            const status = await db.messages.delete(id)
            console.log("Status after trying to delete a data: ", status)
        }
        catch (e) {
            console.log("An error has occured when trying to delete data in DB: ", e)
        }
    }

    return (
        <chatDataContext.Provider value={{
            saveChatData,
            deleteChatData,
            messages
        }}>
            {children}
        </chatDataContext.Provider>
    )
}

export const useChatData = () => {
    const context = useContext(chatDataContext);
    if (!context) {
        throw new Error("useChatData must be used within a ChatDataProvider");
    }
    return context;
}