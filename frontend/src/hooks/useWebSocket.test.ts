import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WebSocketClient } from '../lib/websocket';

interface MockWs {
  readyState: number;
  send: ReturnType<typeof vi.fn>;
  close: ReturnType<typeof vi.fn>;
  onopen: ((ev: Event) => void) | null;
  onmessage: ((ev: MessageEvent) => void) | null;
  onerror: ((ev: Event) => void) | null;
  onclose: ((ev: CloseEvent) => void) | null;
}

let mockWs: MockWs;
const MockWebSocket = vi.fn().mockImplementation(() => {
  mockWs = {
    readyState: 0,
    send: vi.fn(),
    close: vi.fn().mockImplementation(function (this: MockWs) {
      this.readyState = 3;
    }),
    onopen: null,
    onmessage: null,
    onerror: null,
    onclose: null,
  };
  return mockWs;
});
(MockWebSocket as unknown as { OPEN: number }).OPEN = 1;
(MockWebSocket as unknown as { CLOSED: number }).CLOSED = 3;

describe('WebSocketClient', () => {
  let client: WebSocketClient;
  const wsUrl = 'ws://localhost:8080/ws';

  beforeEach(() => {
    vi.stubGlobal('WebSocket', MockWebSocket);
    client = new WebSocketClient(wsUrl);
    MockWebSocket.mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('connect() creates new WebSocket with the correct URL', () => {
    client.connect();
    expect(MockWebSocket).toHaveBeenCalledWith(wsUrl);
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
});
