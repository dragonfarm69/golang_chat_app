import { useEffect, useRef, useState } from "react"
import messagePayload from "../Payload/messagePayload";

function useWs<T>(url: string, onMessage: (msg: T) => void) {
    const [messages, setMessages] = useState<string[]>([]);
    const [socket, setSocket] = useState<WebSocket | null>(null);
    const [connection, setConnection] = useState(false);
    const onMessageRef = useRef(onMessage);
    onMessageRef.current = onMessage;

    useEffect(() => {
        const ws = new WebSocket(url);

        ws.onopen = () => {
            console.log("connceted");
            setConnection(true);
        }

        ws.onclose = () => {
            console.log("disconnected");
            setConnection(false);
        }

        ws.onmessage = (event) => {
            console.log("event: ", event)
            const messages = event.data.split('\n');
            messages.forEach((msg: string) => {
                if (msg.trim()) {
                    console.log("message: ", msg)
                    try {
                        const parsedData = JSON.parse(msg);
                        onMessageRef.current(parsedData);
                    } catch (e) {
                        console.error("Failed to parse JSON: ", e)
                        // onMessageRef.current(msg);
                    }
                }
            });
            console.log(event.data)
        }

        ws.onerror = () => {
            console.log("FATAL ERROR")
        }

        setSocket(ws)

        return () => ws.close();
    }, [url]);

    const sendMessage = (msg: messagePayload) => {
        if (socket && connection) {
            socket.send(JSON.stringify(msg));
        }
    }

    return { connection, messages, sendMessage }
}

export default useWs