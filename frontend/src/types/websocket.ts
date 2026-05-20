/**
 * WebSocket Message Types
 * All server-to-client and client-to-server message shapes
 */

/**
 * Game start message from server
 */
export interface GameStartMessage {
  type: 'game_start';
  game_id: string;
  white_id: string;
  black_id: string;
  white_rating: number;
  black_rating: number;
  fen: string;
  your_color: 'white' | 'black';
}

/**
 * Move made message from server
 */
export interface MoveMadeMessage {
  type: 'move_made';
  game_id: string;
  move: string;
  notation: string;
  fen: string;
  move_count: number;
}

/**
 * Game end message from server
 */
export interface GameEndMessage {
  type: 'game_end';
  game_id: string;
  result: 'white' | 'black' | 'draw';
  reason: 'checkmate' | 'resignation' | 'stalemate' | 'draw_agreed' | 'timeout' | 'insufficient_material';
  white_rating_delta: number;
  black_rating_delta: number;
}

/**
 * Queue status message from server
 */
export interface QueueStatusMessage {
  type: 'queue_status';
  queued: boolean;
  position?: number;
  queue_size: number;
  estimate_wait_seconds?: number;
}

/**
 * Game state message from server
 */
export interface GameStateMessage {
  type: 'game_state';
  game_id: string;
  fen: string;
  last_move?: string;
  move_count: number;
  your_turn: boolean;
  white_rating: number;
  black_rating: number;
}

/**
 * Error message from server
 */
export interface ErrorMessage {
  type: 'error';
  message: string;
}

export interface PongMessage {
  type: 'pong';
}

/**
 * Union of all possible WebSocket messages from server
 */
export type WsMessage =
  | GameStartMessage
  | MoveMadeMessage
  | GameEndMessage
  | QueueStatusMessage
  | GameStateMessage
  | ErrorMessage
  | PongMessage;

/**
 * Client-side message types
 */
export interface JoinQueueClientMessage {
  type: 'queue_join';
  payload: { time_control: string };
}

export interface LeaveQueueClientMessage {
  type: 'leave_queue';
  payload: Record<string, unknown>;
}

export interface MoveClientMessage {
  type: 'move';
  payload: { game_id: string; from: string; to: string };
}

export interface ResignClientMessage {
  type: 'resign';
  payload: { game_id: string };
}

export interface OfferDrawClientMessage {
  type: 'offer_draw';
  payload: { game_id: string };
}

export interface DrawResponseClientMessage {
  type: 'draw_response';
  payload: { accepted: boolean };
}

export interface PingMessage {
  type: 'ping';
}

/**
 * Union of all possible WebSocket messages from client
 */
export type WsClientMessage =
  | JoinQueueClientMessage
  | LeaveQueueClientMessage
  | MoveClientMessage
  | ResignClientMessage
  | OfferDrawClientMessage
  | DrawResponseClientMessage
  | PingMessage;

/**
 * Type guard functions
 */

export function isGameStart(msg: WsMessage): msg is GameStartMessage {
  return msg.type === 'game_start';
}

export function isMoveMade(msg: WsMessage): msg is MoveMadeMessage {
  return msg.type === 'move_made';
}

export function isGameEnd(msg: WsMessage): msg is GameEndMessage {
  return msg.type === 'game_end';
}

export function isQueueStatus(msg: WsMessage): msg is QueueStatusMessage {
  return msg.type === 'queue_status';
}

export function isGameState(msg: WsMessage): msg is GameStateMessage {
  return msg.type === 'game_state';
}

export function isError(msg: WsMessage): msg is ErrorMessage {
  return msg.type === 'error';
}

export function isPong(msg: WsMessage): msg is PongMessage {
  return msg.type === 'pong';
}
