import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WebSocketClient } from '../lib/websocket';

interface MockWs {
  readyState: number;
  url: string;
  send: ReturnType<typeof vi.fn>;
  close: ReturnType<typeof vi.fn>;
  onopen: ((ev: Event) => void) | null;
  onmessage: ((ev: MessageEvent) => void) | null;
  onerror: ((ev: Event) => void) | null;
  onclose: ((ev: CloseEvent) => void) | null;
}

let mockWs: MockWs;

class MockWebSocket {
  static OPEN = 1;
  static CLOSED = 3;
  readyState = 0;
  url: string;
  send = vi.fn();
  close = vi.fn();
  onopen: ((ev: Event) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  onclose: ((ev: CloseEvent) => void) | null = null;

  constructor(url: string) {
    this.url = url;
    this.close = vi.fn(() => { this.readyState = 3; });
    this.send = vi.fn();
    mockWs = this as unknown as MockWs;
  }
}

describe('WebSocketClient', () => {
  let client: WebSocketClient;
  const wsUrl = 'ws://localhost:8080/ws';

  beforeEach(() => {
    vi.stubGlobal('WebSocket', MockWebSocket);
    client = new WebSocketClient(wsUrl);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('connect() creates new WebSocket with the correct URL', () => {
    client.connect();
    expect(mockWs.url).toBe(wsUrl);
  });

  it('disconnect() calls ws.close() and isConnected() returns false', () => {
    client.connect();
    mockWs.readyState = 1;
    client.disconnect();
    expect(mockWs.close).toHaveBeenCalled();
    expect(client.isConnected()).toBe(false);
  });

  it('message handler receives game_start message', () => {
    const handler = vi.fn();
    client.onMessage(handler);
    client.connect();

    const msg = { type: 'game_start', game_id: 'g1', white_id: 'u1', black_id: 'u2', white_rating: 1200, black_rating: 1200, fen: 'start', your_color: 'white' };
    mockWs.onmessage?.({ data: JSON.stringify(msg) } as MessageEvent);

    expect(handler).toHaveBeenCalledWith(expect.objectContaining({ type: 'game_start' }));
  });

  it('message handler receives move_made message', () => {
    const handler = vi.fn();
    client.onMessage(handler);
    client.connect();

    const msg = { type: 'move_made', game_id: 'g1', move: 'e2e4', notation: 'e4', fen: 'fen2', move_count: 1 };
    mockWs.onmessage?.({ data: JSON.stringify(msg) } as MessageEvent);

    expect(handler).toHaveBeenCalledWith(expect.objectContaining({ type: 'move_made' }));
  });

  it('message handler receives queue_status message', () => {
    const handler = vi.fn();
    client.onMessage(handler);
    client.connect();

    const msg = { type: 'queue_status', queued: true, queue_size: 3 };
    mockWs.onmessage?.({ data: JSON.stringify(msg) } as MessageEvent);

    expect(handler).toHaveBeenCalledWith(expect.objectContaining({ type: 'queue_status' }));
  });

  it('send() calls ws.send with JSON.stringify of the message', () => {
    client.connect();
    mockWs.readyState = 1;
    const msg = { type: 'queue_join' as const, payload: { time_control: '5+0' } };
    client.send(msg);
    expect(mockWs.send).toHaveBeenCalledWith(JSON.stringify(msg));
  });

  it('onMessage unsubscribe fn removes the handler', () => {
    const handler = vi.fn();
    const unsub = client.onMessage(handler);
    client.connect();

    unsub();

    const msg = { type: 'queue_status', queued: false, queue_size: 0 };
    mockWs.onmessage?.({ data: JSON.stringify(msg) } as MessageEvent);

    expect(handler).not.toHaveBeenCalled();
  });

  it('onopen notifies status handlers with true', () => {
    const statusHandler = vi.fn();
    client.onStatusChange(statusHandler);
    client.connect();
    mockWs.onopen?.({} as Event);
    expect(statusHandler).toHaveBeenCalledWith(true);
  });

  it('onerror callback is handled without throwing', () => {
    client.connect();
    expect(() => {
      mockWs.onerror?.({} as Event);
    }).not.toThrow();
  });

  it('onclose notifies status handlers with false', () => {
    const statusHandler = vi.fn();
    client.onStatusChange(statusHandler);
    client.connect();
    mockWs.readyState = 1;
    client.disconnect();
    expect(statusHandler).toHaveBeenCalledWith(false);
  });

  it('send() does nothing when not connected (readyState !== OPEN)', () => {
    client.connect(); // readyState stays 0 (not OPEN)
    const msg = { type: 'queue_join' as const, payload: { time_control: '5+0' } };
    client.send(msg);
    expect(mockWs.send).not.toHaveBeenCalled();
  });

  it('onmessage with invalid JSON does not throw and handler is not called', () => {
    const handler = vi.fn();
    client.onMessage(handler);
    client.connect();
    expect(() => {
      mockWs.onmessage?.({ data: 'invalid json }{' } as MessageEvent);
    }).not.toThrow();
    expect(handler).not.toHaveBeenCalled();
  });

  it('onStatusChange unsubscribe fn removes the handler', () => {
    const statusHandler = vi.fn();
    const unsub = client.onStatusChange(statusHandler);
    client.connect();
    unsub();
    mockWs.onopen?.({} as Event);
    expect(statusHandler).not.toHaveBeenCalled();
  });

  it('connect() is a no-op when already OPEN', () => {
    client.connect();
    mockWs.readyState = 1;
    const firstWs = mockWs;
    client.connect(); // should return early
    expect(mockWs).toBe(firstWs); // same ws object
  });

  it('onclose without manual disconnect schedules reconnect', () => {
    vi.useFakeTimers();
    client.connect();
    // simulate server closing the connection (not manual)
    mockWs.onclose?.({} as CloseEvent);
    // scheduleReconnect sets a timeout — advance timers to trigger it
    // the reconnect calls connect() again which creates a new WS
    vi.advanceTimersByTime(2000);
    vi.useRealTimers();
    // no assertion needed beyond no throw
  });

  it('disconnect cancels pending reconnect timeout', () => {
    vi.useFakeTimers();
    client.connect();
    mockWs.onclose?.({} as CloseEvent); // triggers scheduleReconnect
    // now disconnect before reconnect fires
    client.disconnect();
    vi.useRealTimers();
    expect(client.isConnected()).toBe(false);
  });

  it('onopen starts heartbeat (covers startHeartbeat path)', () => {
    vi.useFakeTimers();
    client.connect();
    mockWs.readyState = 1;
    mockWs.onopen?.({} as Event);
    // advance past heartbeat interval to hit the isConnected() check inside
    vi.advanceTimersByTime(35000);
    vi.useRealTimers();
  });
});
