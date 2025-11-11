import { useState } from "react";
import "../../App.css";
import { useUser } from "../../Context/userContext";
import UserInfo from "../../Modules/clientModule";

function JoinRoomPage({ onJoin }: { onJoin: (roomId: string, clientName: string) => void}) {
    const [roomId, setRoomId] = useState("");
    const [clientName, setClientName] = useState("");
    const {user, setUser} = useUser();
    
    const joinRoom = async () => {
        if (!roomId.trim()) {
            alert("Room ID is empty");
            return;
        }
        
        try {
            const response = await fetch(`http://localhost:8080/join?hub=${roomId}&client=${clientName}`);
            if (response.ok) {
                const userInformation: UserInfo = {
                    profileLink: "test",
                    name: clientName,
                }
                setUser(userInformation);
                onJoin(roomId, clientName);
            } else {
                const error = await response.text();
                // console.log(error)
                alert(`ERROR: ${error}`);
            }
        }
        catch (error) {
            console.error("SOMETHING IS WRONG", error);
        }
    }

    return (
        <div className="app-container" style={{backgroundColor: "white"}}>
            <form className="room-form" onSubmit={(e) => {e.preventDefault(); joinRoom();}}> 
                <input 
                    placeholder="Input name" 
                    type="text" 
                    value={clientName} 
                    onChange={(e) => setClientName(e.target.value)}>
                </input>
                <input
                    placeholder="Input room id"
                    type="text" value={roomId} onChange={(e) => setRoomId(e.target.value)}
                    className="room-id-input">
                </input>
                <button className="join-room-button" type="submit">Join Room</button>
            </form>
        </div>
    )
}

export default JoinRoomPage