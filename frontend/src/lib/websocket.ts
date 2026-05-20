/**
 * WebSocket Client
 * Handles connection lifecycle, message dispatch, and auto-reconnection
 */

import type { WsMessage, WsClientMessage } from '../types/websocket';
import {
  WS_RECONNECT_DELAY_MS,
  WS_MAX_RECONNECT_DELAY_MS,
  WS_HEARTBEAT_INTERVAL_MS,
} from '../utils/constants';

export type MessageHandler = (message: WsMessage) => void;
export type StatusChangeHandler = (connected: boolean) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private messageHandlers: MessageHandler[] = [];
  private statusHandlers: StatusChangeHandler[] = [];
  private reconnectDelay = WS_RECONNECT_DELAY_MS;
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private heartbeatInterval: ReturnType<typeof setInterval> | null = null;
  private isManuallyDisconnected = false;

  constructor(url: string) {
    this.url = url;
  }

  /**
   * Connect to the WebSocket server
   */
  public connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.isManuallyDisconnected = false;

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('[WebSocket] Connected');
        this.reconnectDelay = WS_RECONNECT_DELAY_MS;
        this.notifyStatusChange(true);
        this.startHeartbeat();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WsMessage = JSON.parse(event.data);
          this.messageHandlers.forEach((handler) => handler(message));
        } catch (error) {
          console.error('[WebSocket] Failed to parse message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
      };

      this.ws.onclose = () => {
        console.log('[WebSocket] Disconnected');
        this.stopHeartbeat();
        this.notifyStatusChange(false);

        if (!this.isManuallyDisconnected) {
          this.scheduleReconnect();
        }
      };
    } catch (error) {
      console.error('[WebSocket] Failed to connect:', error);
      this.scheduleReconnect();
    }
  }

  /**
   * Disconnect from the WebSocket server
   */
  public disconnect(): void {
    this.isManuallyDisconnected = true;
    this.stopHeartbeat();

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.notifyStatusChange(false);
  }

  /**
   * Send a message to the server
   */
  public send(message: WsClientMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(message));
      } catch (error) {
        console.error('[WebSocket] Failed to send message:', error);
      }
    } else {
      console.warn('[WebSocket] Not connected, cannot send message');
    }
  }

  /**
   * Register a message handler
   */
  public onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.push(handler);
    // Return unsubscribe function
    return () => {
      this.messageHandlers = this.messageHandlers.filter((h) => h !== handler);
    };
  }

  /**
   * Register a status change handler
   */
  public onStatusChange(handler: StatusChangeHandler): () => void {
    this.statusHandlers.push(handler);
    // Return unsubscribe function
    return () => {
      this.statusHandlers = this.statusHandlers.filter((h) => h !== handler);
    };
  }

  /**
   * Check if connected
   */
  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    if (this.isManuallyDisconnected) {
      return;
    }

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }

    console.log(
      `[WebSocket] Reconnecting in ${this.reconnectDelay}ms...`
    );

    this.reconnectTimeout = setTimeout(() => {
      this.reconnectTimeout = null;
      this.reconnectDelay = Math.min(
        this.reconnectDelay * 2,
        WS_MAX_RECONNECT_DELAY_MS
      );
      this.connect();
    }, this.reconnectDelay);
  }

  /**
   * Start heartbeat ping
   */
  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected()) {
        this.send({ type: 'ping' });
      }
    }, WS_HEARTBEAT_INTERVAL_MS);
  }

  /**
   * Stop heartbeat ping
   */
  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  /**
   * Notify all status change handlers
   */
  private notifyStatusChange(connected: boolean): void {
    this.statusHandlers.forEach((handler) => handler(connected));
  }
}
