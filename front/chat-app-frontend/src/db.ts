import { Dexie, type EntityTable } from "dexie";

interface Message {
  id: string;
  room_id: string;
  user_id: string;
  username: string;
  content: string;
  timeStamp: string;
}

const db = new Dexie("MessageDB") as Dexie & {
  messages: EntityTable<Message, "id">; // primary key
};

db.version(1).stores({
  messages: "id, room_id, user_id, username, content, timeStamp",
});

export type { Message };
export { db };
