import { useEffect, useRef } from 'react'
import { useWebSocketContext } from '../context/WebSocketContext'
import type { WsMessage } from '../types/websocket'

export function useWebSocket() {
  return useWebSocketContext()
}

export function useWebSocketMessage(handler: (msg: WsMessage) => void) {
  const { onMessage } = useWebSocketContext()
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    return onMessage((msg) => handlerRef.current(msg))
  }, [onMessage])
}
