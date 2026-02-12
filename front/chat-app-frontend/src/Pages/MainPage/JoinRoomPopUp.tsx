import { useState } from "react";
import "../../App.css";

interface JoinRoomPopUpProps {
    onClose: () => void;
}

function JoinRoomPopUp({ onClose }: JoinRoomPopUpProps) {
    const [formInputValue, setFormInputValue] = useState("");
    const [selectedMode, setSelectedMode] = useState("IDMode");

    const joinRoom = async (value: string) => {        
        try {
            console.log(value)
            console.log(selectedMode)
            // After successfully joining, you might want to close the popup
            onClose();
        }
        catch (error) {
            console.error("SOMETHING IS WRONG", error);
        }
    }

    const handleModeChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        setSelectedMode(event.target.value)
    }

    return (
        <>
            <div className="popup-overlay" onClick={onClose}>
                <div className="join-room-popup" style={{backgroundColor: "white"}} onClick={(e) => e.stopPropagation()}>
                    <select value={selectedMode} onChange={handleModeChange}>
                        <option value="IDMode">Join with ID</option>
                        <option value="URLMode">Join with URL</option>
                    </select>
                    <form className="room-form" onSubmit={(e) => {e.preventDefault(); joinRoom(formInputValue)}}> 
                        <input 
                            placeholder={selectedMode === "IDMode" ? "Input room ID" : "Input room URL"}
                            type="text" 
                            value={formInputValue} 
                            onChange={(e) => setFormInputValue(e.target.value)}>
                        </input>
                        <button className="join-room-button" type="submit">Join</button>
                    </form>
                    <button className="close-popup-button" onClick={onClose}>Close</button>
                </div>
            </div>
        </>
    )
}

export default JoinRoomPopUp