import { useState } from "react";
import "../../App.css";

interface JoinRoomPopUpProps {
  onClose: () => void;
  onSubmit: (roomName: string) => void;
}

function JoinRoomPopUp({ onClose, onSubmit }: JoinRoomPopUpProps) {
  const [formInputValue, setFormInputValue] = useState("");

  const joinRoom = async (value: string) => {
    try {
      console.log(value)
      // console.log(selectedMode)
      // // After successfully joining, you might want to close the popup
      onSubmit(formInputValue)
      onClose();
    } catch (error) {
      console.error("SOMETHING IS WRONG", error);
    }
  };

  return (
    <>
      <div className="popup-overlay" onClick={onClose}>
        <div
          className="join-room-popup"
          style={{ backgroundColor: "white" }}
          onClick={(e) => e.stopPropagation()}
        >
          <form
            className="room-form"
            onSubmit={(e) => {
              e.preventDefault();
              joinRoom(formInputValue);
            }}
          >
            <input
              placeholder={
                "Input invite code"
              }
              type="text"
              value={formInputValue}
              onChange={(e) => setFormInputValue(e.target.value)}
            ></input>
            <button className="join-room-button" type="submit">
              Join
            </button>
          </form>
          <button className="close-popup-button" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </>
  );
}

export default JoinRoomPopUp;
