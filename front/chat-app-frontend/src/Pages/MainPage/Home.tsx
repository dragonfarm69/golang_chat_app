import "../../App.css";
import { useState, useRef, useEffect } from "react";

// Placeholder data for rooms and messages
const rooms = [
    { id: 1, name: "General", icon: "G" },
    { id: 2, name: "Random", icon: "R" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
    { id: 3, name: "Tech Talk", icon: "T" },
];

const initialMessages = [
    { id: 1, user: "Alice", text: "Hey everyone!", sender: "other" },
    { id: 2, user: "You", text: "Hi Alice! How are you?", sender: "me" },
    { id: 3, user: "Bob", text: "Welcome to the chat!", sender: "other" },
];

function HomePage() {
    const [messages, setMessages] = useState(initialMessages);
    const [newMessage, setNewMessage] = useState("");

    const chatLogRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (chatLogRef.current) {
            chatLogRef.current.scrollTop = chatLogRef.current.scrollHeight;
        }
    }, [messages]);

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault();
        if (newMessage.trim() === "") return;

        const message = {
            id: messages.length + 1,
            user: "You",
            text: newMessage,
            sender: "me" as const,
        };

        setMessages([...messages, message]);
        setNewMessage("");
    };

    return (
        <>
            <div className="app-container" style={{backgroundColor: "white"}}>
                <div className="wrapper">
                    <div className="side-bar">
                        <div className="room-list">
                            {rooms.map(room => (
                                <button key={room.id} className="room-icon" title={room.name}>
                                    {room.icon}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="main-container">
                        <div className="chat-area">
                            <div className="chat-log" ref={chatLogRef}>
                                {messages.map(msg => (
                                    <div key={msg.id} className={`chat-message ${msg.sender}`}>
                                        <div className="message-user">{msg.user}</div>
                                        <div className="message-text">{msg.text}</div>
                                    </div>
                                ))}
                            </div>
                            <form className="message-form" onSubmit={handleSendMessage}>
                                <input
                                    type="text"
                                    className="message-input"
                                    placeholder="Type a message..."
                                    value={newMessage}
                                    onChange={(e) => setNewMessage(e.target.value)}
                                />
                                <button type="submit" className="send-button">Send</button>
                                <button type="button" className="send-button">File</button>
                                <button type="button" className="send-button">Command</button>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}

export default HomePage;