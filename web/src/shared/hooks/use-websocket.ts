import { useEffect } from "react";
import { useWebSocketContext } from "@/shared/contexts/websocket-context";

type MessageHandler = (data: any) => void;

export function useWebSocket(onMessage: MessageHandler) {
  const { subscribe } = useWebSocketContext();

  useEffect(() => {
    // The subscribe function returns an unsubscribe function
    const unsubscribe = subscribe(onMessage);
    return () => unsubscribe();
  }, [subscribe, onMessage]);
}
