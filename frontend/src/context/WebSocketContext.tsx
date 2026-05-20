import { createContext, useContext, useEffect, useRef, useState, useCallback, type ReactNode } from 'react'
import { WebSocketClient } from '../lib/websocket'
import { useAuthContext } from './AuthContext'
import type { WsMessage, WsClientMessage } from '../types/websocket'

interface WebSocketContextValue {
  isConnected: boolean
  send: (msg: WsClientMessage) => void
  onMessage: (handler: (msg: WsMessage) => void) => () => void
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null)

function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.exp * 1000 < Date.now() + 30_000
  } catch {
    return true
  }
}

export function WebSocketProvider({ children, url }: { children: ReactNode; url: string }) {
  const clientRef = useRef<WebSocketClient | null>(null)
  const handlersRef = useRef<Set<(msg: WsMessage) => void>>(new Set())
  const [isConnected, setIsConnected] = useState(false)
  const { accessToken, refresh } = useAuthContext()

  useEffect(() => {
    if (!accessToken) {
      if (clientRef.current) {
        clientRef.current.disconnect()
        clientRef.current = null
        setIsConnected(false)
      }
      return
    }

    if (isTokenExpired(accessToken)) {
      refresh().catch(() => {})
      return
    }

    const wsUrl = `${url}?token=${encodeURIComponent(accessToken)}`
    const client = new WebSocketClient(wsUrl)
    clientRef.current = client

    const unsubStatus = client.onStatusChange(setIsConnected)
    const unsubMsg = client.onMessage((msg) => {
      handlersRef.current.forEach((h) => h(msg))
    })

    client.connect()

    return () => {
      unsubStatus()
      unsubMsg()
      client.disconnect()
      clientRef.current = null
    }
  }, [url, accessToken, refresh])

  const send = useCallback((msg: WsClientMessage) => {
    clientRef.current?.send(msg)
  }, [])

  const onMessage = useCallback((handler: (msg: WsMessage) => void) => {
    handlersRef.current.add(handler)
    return () => { handlersRef.current.delete(handler) }
  }, [])

  return (
    <WebSocketContext.Provider value={{ isConnected, send, onMessage }}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocketContext(): WebSocketContextValue {
  const ctx = useContext(WebSocketContext)
  if (!ctx) throw new Error('useWebSocketContext must be used within WebSocketProvider')
  return ctx
}
