/**
 * Application Constants
 */

import type { City } from '../types/user';

/**
 * Kazakhstan cities
 */
export const CITIES: City[] = ['Almaty', 'Astana', 'Shymkent'];

/**
 * Time control options
 */
export const TIME_CONTROLS = [
  { label: 'Bullet (1+0)', value: 'bullet_1_0' },
  { label: 'Blitz (3+2)', value: 'blitz_3_2' },
  { label: 'Blitz (5+0)', value: 'blitz_5_0' },
  { label: 'Rapid (10+0)', value: 'rapid_10_0' },
  { label: 'Classical (30+0)', value: 'classical_30_0' },
] as const;

/**
 * WebSocket reconnection settings (milliseconds)
 */
export const WS_RECONNECT_DELAY_MS = 1000;
export const WS_MAX_RECONNECT_DELAY_MS = 30000;

/**
 * WebSocket heartbeat interval (milliseconds)
 */
export const WS_HEARTBEAT_INTERVAL_MS = 30000;

/**
 * Stockfish analysis debounce (milliseconds)
 */
export const STOCKFISH_DEBOUNCE_MS = 100;

/**
 * Stockfish default analysis depth
 */
export const STOCKFISH_DEFAULT_DEPTH = 20;

/**
 * Default pagination settings
 */
export const DEFAULT_PAGE_SIZE = 20;
