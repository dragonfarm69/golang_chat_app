import { useState } from "react";
import "../../App.css";

interface JoinRoomPopUpProps {
  onClose: () => void;
  onSubmit: (roomName: string, isCreateRoom: boolean) => void;
}

function JoinRoomPopUp({ onClose, onSubmit }: JoinRoomPopUpProps) {
  const [formInputValue, setFormInputValue] = useState("");
  const [formCreateRoomState, setFormCreateRoomState] = useState(false);

  const handleSubmit = () => {
    try {
      if (formCreateRoomState) {
        onSubmit(formInputValue, true);
      } else {
        onSubmit(formInputValue, false);
      }
      onClose();
    } catch (error) {
      console.error("SOMETHING IS WRONG", error);
    }
  };

  return (
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
            handleSubmit();
          }}
        >
          <input
            placeholder={
              formCreateRoomState ? "Input room name" : "Input invite code"
            }
            type="text"
            value={formInputValue}
            onChange={(e) => setFormInputValue(e.target.value)}
          ></input>
          <button className="join-room-button" type="submit">
            {formCreateRoomState ? "Create" : "Join"}
          </button>
        </form>
        <button
          className="join-popup-button"
          onClick={() => {
            setFormCreateRoomState(!formCreateRoomState);
          }}
        >
          {formCreateRoomState ? "Join A Room" : "Create New Room"}
        </button>
        <button className="close-popup-button" onClick={onClose}>
          Close
        </button>
      </div>
    </div>
  );
}

export default JoinRoomPopUp;
