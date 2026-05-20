/**
 * Formatting Utilities
 */

import type { Color, GameResult } from '../types/game';

/**
 * Format ELO rating with thousand separator
 */
export function formatElo(rating: number): string {
  return new Intl.NumberFormat('en-US').format(Math.round(rating));
}

/**
 * Format date string to readable format
 */
export function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

/**
 * Format short date without time
 */
export function formatDateShort(dateStr: string): string {
  const date = new Date(dateStr);
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date);
}

/**
 * Format game result from player's perspective
 */
export function formatResult(result: GameResult | null, myColor: Color): string {
  if (result === null) return 'Ongoing';
  if (result === 'draw') return 'Draw';
  if (result === myColor) return 'Win';
  return 'Loss';
}

/**
 * Format rating delta with sign
 */
export function formatRatingDelta(delta: number): string {
  if (delta > 0) return `+${delta}`;
  if (delta < 0) return `${delta}`;
  return '0';
}

/**
 * Format time in seconds to MM:SS
 */
export function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

/**
 * Format centipawn score to display format
 */
export function formatCentipawns(centipawns: number): string {
  const pawns = Math.abs(centipawns / 100);
  const sign = centipawns >= 0 ? '+' : '-';
  return `${sign}${pawns.toFixed(2)}`;
}
