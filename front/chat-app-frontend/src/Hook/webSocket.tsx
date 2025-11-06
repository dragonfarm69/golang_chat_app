import { useEffect, useState } from "react"

function useWs(url: string) {
    const [messages, setMessages] = useState<string[]>([]);
    const [socket, setSocket] = useState<WebSocket | null>(null);
    const [connection, setConnection] = useState(false);

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
            setMessages((prev) => [...prev, event.data])
            console.log(event.data)
            console.log("sendinfg something")
        }

        ws.onerror = () => {
            console.log("FATAL ERROR")
        }

        setSocket(ws)

        return () => ws.close();
    }, [url]);

    const sendMessage = (msg: string) => {
        if (socket && connection) {
            socket.send(msg);
        }
    }

    return { connection, messages, sendMessage }
}

export default useWs