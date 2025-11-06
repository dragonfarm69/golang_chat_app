import { useState } from "react";
import "../../App.css";

function JoinRoomPage({ onJoin }: { onJoin: (roomId: string) => void}) {
    const [roomId, setRoomId] = useState("");
    const joinRoom = async () => {
        if (!roomId.trim()) {
            alert("Room ID is empty");
            return;
        }
        
        // try {
        //     const response = await fetch(`http://localhost:8080/join?hub=${roomId}`);
        //     if (response.ok) {
        //         onJoin(roomId);
        //     } else {
        //         const error = await response.text();
        //         alert(`ERROR: ${error}`);
        //     }
        // }
        // catch (error) {
        //     console.error("SOMETHING IS WRONG", error);
        // }
        onJoin(roomId)
    }

    return (
        <div className="app-container" style={{backgroundColor: "white"}}>
            <form className="room-form" onSubmit={(e) => {e.preventDefault(); joinRoom();}}> 
                <input
                    placeholder="input room id"
                    type="text" value={roomId} onChange={(e) => setRoomId(e.target.value)}
                    className="room-id-input">
                </input>
                <button className="join-room-button" type="submit">Join Room</button>
            </form>
        </div>
    )
}

export default JoinRoomPage