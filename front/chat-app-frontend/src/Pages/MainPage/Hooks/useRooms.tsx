import { Dispatch, SetStateAction } from "react";
import {
  commands,
  RoomLitePayload,
  UserInfo,
  MessagePayload,
  MessageResponse,
} from "../../../bindings";

export const ROOMS_CAP = 6; // the amount of rooms can be opened at the same tim
export type MessageMap = { [key: string]: MessageResponse[] };

export const handleRoomSelect = async (
  userData: UserInfo,
  room: RoomLitePayload,
  selectedRooms: RoomLitePayload[],
  setAllMessages: Dispatch<SetStateAction<MessageMap>>,
  setSelectedRooms: Dispatch<SetStateAction<RoomLitePayload[]>>,
  setActiveRoomIndex: Dispatch<SetStateAction<number>>,
  location?: string,
  droppedOnRoomWithId?: string,
) => {
  if (!userData || !userData.id) {
    console.error("Current user data is null");
    return;
  }
  //prevent user from opening more room
  if (selectedRooms.length >= ROOMS_CAP) {
    return;
  }

  const roomIndex = selectedRooms.findIndex((r) => r.id.toString() === room.id);
  if (roomIndex !== -1) {
    setActiveRoomIndex(roomIndex);
    return;
  }

  let newActiveIndex = selectedRooms.length;
  let dropTargetIndex = -1;

  if (location && droppedOnRoomWithId) {
    dropTargetIndex = selectedRooms.findIndex(
      (r) => r.id.toString() === droppedOnRoomWithId,
    );

    if (dropTargetIndex !== -1) {
      if (location === "left") {
        newActiveIndex = dropTargetIndex;
      } else if (location === "right") {
        newActiveIndex = dropTargetIndex + 1;
      }
    }
  }

  setSelectedRooms((prev) => {
    const newRooms = [...prev];

    const currentDropIndex =
      location && droppedOnRoomWithId
        ? newRooms.findIndex(
            (r) => r.id.toString() === droppedOnRoomWithId.toString(),
          )
        : -1;

    switch (location) {
      case "left":
        newRooms.splice(currentDropIndex, 0, room);
        break;
      case "right":
        newRooms.splice(currentDropIndex + 1, 0, room);
        break;
      default:
        newRooms.push(room);
        break;
    }
    return newRooms;
  });

  setActiveRoomIndex(newActiveIndex);

  const tempId = crypto.randomUUID();
  const messagePayload: MessagePayload = {
    id: tempId,
    user_id: userData.id,
    room_id: room.id,
    content: "",
    timeStamp: "",
    action: "JOIN",
  };
  commands.sendMessage(messagePayload);

  const result = await commands.fetchRoomMessages(room.id, "");
  if (result.status === "ok") {
    if (result.data != null) {
      const reversedData = result.data.reverse();
      setAllMessages((prev) => ({ ...prev, [room.id]: reversedData }));
    }
  } else {
    console.error("Error fetching rooms: ", result.error);
  }
};

export const handleCloseRoom = (
  roomsId: string,
  selectedRooms: RoomLitePayload[],
  setSelectedRooms: Dispatch<SetStateAction<RoomLitePayload[]>>,
  setActiveRoomIndex: Dispatch<SetStateAction<number>>,
  activeRoomIndex: number,
) => {
  const roomIndex = selectedRooms.findIndex((r) => r.id.toString() === roomsId);

  setSelectedRooms((prev) => prev.filter((r) => r.id.toString() !== roomsId));

  if (activeRoomIndex >= selectedRooms.length - 1) {
    setActiveRoomIndex(Math.max(0, selectedRooms.length - 2));
  } else if (roomIndex <= activeRoomIndex) {
    setActiveRoomIndex(Math.max(0, activeRoomIndex - 1));
  }
};

export const handlePopUpSubmit = async (
  roomName: string,
  isCreateRoom: boolean,
  userData: UserInfo,
  setRooms: Dispatch<SetStateAction<RoomLitePayload[]>>,
) => {
  if (!userData || !userData.id) {
    console.error("Current user data is null");
    return;
  }

  if (isCreateRoom) {
    const stats = await commands.createRoom(userData?.id, roomName);
    if (stats.status === "ok") {
      if (stats.data === true) {
        const roomList = await commands.fetchRoomsList(userData.id);
        if (roomList.status === "ok") {
          if (roomList.data != null) {
            setRooms(roomList.data);
          }
        } else {
          console.error("Error fetching rooms: ", roomList.error);
        }
      }
    }
  } else {
    const stats = await commands.joinRoom(userData?.id, roomName);
    if (stats.status === "ok") {
      if (stats.data === true) {
        const roomList = await commands.fetchRoomsList(userData.id);
        if (roomList.status === "ok") {
          if (roomList.data != null) {
            setRooms(roomList.data);
          }
        } else {
          console.error("Error fetching rooms: ", roomList.error);
        }
      }
    }
  }
};
