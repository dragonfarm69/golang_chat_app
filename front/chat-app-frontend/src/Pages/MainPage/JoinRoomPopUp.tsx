import { useState } from "react";
import "../../App.css";

function JoinRoomPopUp() {
    const [roomId, setRoomId] = useState("");
    const joinRoom = async (id: string) => {        
        try {
            console.log(id)
        }
        catch (error) {
            console.error("SOMETHING IS WRONG", error);
        }
    }

    return (
        <>
            <div className="join-room-popup" style={{backgroundColor: "white"}}>
                <form className="room-form" onSubmit={(e) => {e.preventDefault(); joinRoom(roomId)}}> 
                    <input 
                        placeholder="Input first name" 
                        type="text" 
                        value={roomId} 
                        onChange={(e) => setRoomId(e.target.value)}>
                    </input>
                    <button className="join-room-button" type="submit">Join</button>
                </form>
            </div>
        </>
    )
}

export default JoinRoomPopUp