// import { useCallback, useEffect, useRef, useState } from "react";
// // import { invoke } from "@tauri-apps/api/core";
// import "../../App.css";
// import reactImg from '../../assets/react.svg'
// import { useUser } from "../../Context/userContext";
// import LogData from "../../Modules/messageModule";
// import useWs from "../../Hook/webSocket";
// import messagePayload from "../../Payload/messagePayload";

// function useHover() {
//     const [isHovered, setIsHovered] = useState(false);
//     const ref = useRef<HTMLDivElement | null>(null);

//     useEffect(() => {
//         const node = ref.current;
//         if (!node) return;

//         const handleEnter = () => setIsHovered(true)
//         const handleExit = () => setIsHovered(false)

//         node.addEventListener("mouseenter", handleEnter)
//         node.addEventListener("mouseleave", handleExit)

//         return () => {
//             node.removeEventListener("mouseenter", handleEnter);
//             node.removeEventListener("mouseleave", handleExit);
//         };
//     }, []);

//     return { ref, isHovered }
// }

// function ChatPage({ roomId, clientName }: { roomId: string, clientName: string }) {
//     const {user} = useUser()
//     const { ref, isHovered } = useHover();
//     const [logs, setLogs] = useState<LogData[]>([]);
//     const [hubs, setHubs] = useState<string[]>([]);
//     const [message, setMessage] = useState("");

//     interface wsMessage {
//         username: string;
//         content: string;
//     }

//     const { connection, messages, sendMessage } = useWs<wsMessage>(`ws://localhost:8080/ws?hub=${roomId}&username=${clientName}`, (msg) => {
//         console.log("test: ", msg)
//         try {
//             // const message: wsMessage = JSON.parse(msg)
//             addLog(msg.username, "server", msg.content)
//         } catch (err) {
//             console.error("Failed to parse WS message: ", err)
//         }
//     });


//     const LogItem = ({ data }: { data: LogData }) => {
//         // console.log(data)
//         return (
//             <>
//                 {data.username === "!server" ? (
//                     <div className="log-item" style={{ display: "flex", margin: "5px", gap: "10px" }}>
//                         <div>
//                             <div className="username">{data.username}</div>
//                         </div>
//                         <div style={{ display: "block" }}>
//                             <div className="content">{data.message}</div>
//                             <div className="timestamp">{data.timestamp.toLocaleTimeString()}</div>
//                         </div>
//                     </div>
//                 ) : (   
//                     <div className="log-item" style={{ display: "flex", margin: "5px", gap: "10px" }}>
//                         {/* <div className="profile-picture">{data.profile}</div> */}
//                         {/* <img className="profile-img" src={reactImg}></img> */}
//                         {/* <div style={{background-image: {reactImg}}}></div> */}
//                         <div>
//                             <div className="profile-img" style={{ backgroundImage: `url(${reactImg})` }}></div>
//                         </div>
//                         <div style={{ display: "block" }}>
//                             <div className="username">{data.username}</div>
//                             <div className="content">{data.message}</div>
//                             <div className="timestamp">{data.timestamp.toLocaleTimeString()}</div>
//                         </div>
//                     </div>
//                 )}
//             </>
//         );
//     }

//     const addLog = useCallback((username: string, profile: string, message: string) => {
//         if (user?.name) {
//             const newLog: LogData = {
//                 username: username,
//                 profile,
//                 id: Date.now().toString(),
//                 message,
//                 timestamp: new Date(),
//             };
//             setLogs(prev => [...prev, newLog]);
//         }
//     }, []);

//     return (
//         <div className="app-container">
//             {/* Main Content Area */}
//             <div className="main-content">
//                 {/* Log Area */}
//                 <div className="log-area">
//                     <div className="room-info">
//                         <h1>{roomId}</h1>
//                     </div>
//                     {logs.map(log => (
//                         <LogItem key={log.id} data={log} />
//                     ))}
//                 </div>

//                 {/* Hubs List */}
//                 <div ref={ref}
//                     className={`${isHovered ? "hubs-list" : "hidden"}`}
//                     style={!isHovered ? { display: "flex", alignItems: "center", justifyContent: "center" } : undefined}>
//                     {isHovered ?
//                         <>
//                             <h3 className="hubs-title">Hubs</h3>
//                             <div className="hubs-items">
//                                 {hubs.map((hub, i) => (
//                                     <div key={i} className="hub-item">
//                                         {hub}
//                                     </div>
//                                 ))}
//                             </div>
//                         </>
//                         : <>
//                             <h1>{`<`}</h1>
//                         </>
//                     }
//                 </div>
//             </div>

//             {/* Form */}
//             <form className="form-container" onSubmit={(e) => { e.preventDefault() }}>
//                 <button type="button" onClick={() => {
//                     if (message.trim() && user?.name) {
//                         const msg: messagePayload = {
//                             username: user.name,
//                             hubId: roomId,
//                             content: message.toString(),
//                         }
//                         sendMessage(msg);
//                         setMessage('');
//                     }
//                 }}>Send</button>
//                 <input
//                     type="text"
//                     value={message}
//                     onChange={(e) => setMessage(e.target.value)}
//                     className="message-input"
//                     autoFocus
//                     onKeyDown={(e) => {
//                         if (e.key == "Enter") {
//                             e.preventDefault();
//                             if (message.trim() && user?.name) {
//                                 const msg: messagePayload = {
//                                     username: user.name,
//                                     hubId: roomId,
//                                     content: message.toString(),
//                                 }
//                                 sendMessage(msg);
//                                 setMessage('');
//                             }
//                         }
//                     }}
//                 />
//                 <button type="button" >Create new hub</button>
//                 <button type="button" >Refresh list</button>
//                 <button type="button" >Disconnect</button>
//             </form>
//         </div>
//     );
// }

// export default ChatPage;