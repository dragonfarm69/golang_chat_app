import { useEffect, useRef, useState } from "react";
// import { invoke } from "@tauri-apps/api/core";
import "../../App.css";
import reactImg from '../../assets/react.svg'
import LogData from "../../Modules/messageModule";
import useWs from "../../Hook/webSocket";

function useHover() {
    const [isHovered, setIsHovered] = useState(false);
    const ref = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        const node = ref.current;
        if (!node) return;

        const handleEnter = () => setIsHovered(true)
        const handleExit = () => setIsHovered(false)

        node.addEventListener("mouseenter", handleEnter)
        node.addEventListener("mouseleave", handleExit)

        return () => {
            node.removeEventListener("mouseenter", handleEnter);
            node.removeEventListener("mouseleave", handleExit);
        };
    }, []);

    return { ref, isHovered }
}

function ChatPage({roomId}: {roomId: string}) {
    const { ref, isHovered } = useHover();
    const [logs, setLogs] = useState<LogData[]>([]);
    const [hubs, setHubs] = useState<string[]>([]);
    const [message, setMessage] = useState("");

    const { connection, messages, sendMessage } = useWs(`ws://localhost:8080/ws?hub=${roomId}`)


    const LogItem = ({ data }: { data: LogData }) => {
        return (
            <div className="log-item" style={{ display: "flex", margin: "5px", gap: "10px" }}>
                {/* <div className="profile-picture">{data.profile}</div> */}
                {/* <img className="profile-img" src={reactImg}></img> */}
                {/* <div style={{background-image: {reactImg}}}></div> */}
                <div className="profile-img" style={{ backgroundImage: `url(${reactImg})` }}></div>
                <div style={{ display: "block" }}>
                    <div className="content">{data.message}</div>
                    <div className="timestamp">{data.timestamp.toLocaleTimeString()}</div>
                </div>
            </div>
        );
    }

    const addLog = (profile: string, message: string) => {
        const newLog: LogData = {
            profile,
            id: Date.now().toString(),
            message,
            timestamp: new Date(),
        };
        setLogs(prev => [...prev, newLog]);
    }

    useEffect(() => {
        setLogs([
            { profile: 'profileImg', id: '1', message: 'Hello', timestamp: new Date() },
            { profile: 'profileImg', id: '2', message: 'Test', timestamp: new Date() },
        ]);
    }, []);


    return (
        <div className="app-container">
            {/* Main Content Area */}
            <div className="main-content">
                {/* Log Area */}
                <div className="log-area">
                    <div className="room-info">
                        <h1>roomName</h1>
                    </div>
                    {logs.map(log => (
                        <LogItem key={log.id} data={log} />
                    ))}
                </div>

                {/* Hubs List */}
                <div ref={ref}
                    className={`${isHovered ? "hubs-list" : "hidden"}`}
                    style={!isHovered ? { display: "flex", alignItems: "center", justifyContent: "center" } : undefined}>
                    {isHovered ?
                        <>
                            <h3 className="hubs-title">Hubs</h3>
                            <div className="hubs-items">
                                {hubs.map((hub, i) => (
                                    <div key={i} className="hub-item">
                                        {hub}
                                    </div>
                                ))}
                            </div>
                        </>
                        : <>
                            <h1>{`<`}</h1>
                        </>
                    }
                </div>
            </div>

            {/* Form */}
            <form className="form-container" onSubmit={(e) => {e.preventDefault()}}>
                <button type="button" onClick={() => {
                    if (message.trim()) {
                        sendMessage(message);
                        addLog('profileImg', message)
                        setMessage('');
                    }
                }}>Send</button>
                <input
                    type="text"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    className="message-input"
                    autoFocus
                    onKeyDown={(e) => {
                        if (e.key == "Enter") {
                            e.preventDefault();
                            if (message.trim()) {
                                sendMessage(message);
                                addLog('profileImg', message)
                                setMessage('');
                            }
                        }
                    }}
                />
                <button type="button" >Create new hub</button>
                <button type="button" >Refresh list</button>
                <button type="button" >Disconnect</button>
            </form>
        </div>
    );
}

export default ChatPage;