/**
 * Stockfish WASM Worker Bridge
 * Communicates with Stockfish engine running in Web Worker
 */

export interface StockfishEvaluation {
  fen: string;
  depth: number;
  score: number;
  mate: number | null;
  pv: string[];
}

export type StockfishEventHandler = (evaluation: StockfishEvaluation) => void;

export class StockfishWorker {
  private worker: Worker | null = null;
  private evalHandlers: StockfishEventHandler[] = [];
  private isInitialized = false;
  private initPromise: Promise<void> | null = null;

  /**
   * Initialize the Stockfish worker
   */
  public async init(): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = new Promise((resolve, reject) => {
      try {
        // Import the worker module
        this.worker = new Worker(
          new URL('./stockfish-worker-script.js', import.meta.url),
          { type: 'module' }
        );

        this.worker.onmessage = (event) => {
          const data = event.data;

          if (data.type === 'ready') {
            this.isInitialized = true;
            resolve();
          } else if (data.type === 'evaluation') {
            this.evalHandlers.forEach((handler) => handler(data.evaluation));
          } else if (data.type === 'error') {
            console.error('[Stockfish] Worker error:', data.error);
          }
        };

        this.worker.onerror = (error) => {
          console.error('[Stockfish] Worker initialization error:', error);
          reject(error);
        };

        // Send initialization command
        this.worker.postMessage({ type: 'init' });
      } catch (error) {
        reject(error);
      }
    });

    return this.initPromise;
  }

  /**
   * Analyze a position
   */
  public analyze(fen: string, depth: number = 20): void {
    if (!this.worker || !this.isInitialized) {
      console.warn('[Stockfish] Worker not initialized');
      return;
    }

    this.worker.postMessage({
      type: 'analyze',
      fen,
      depth,
    });
  }

  /**
   * Register an evaluation handler
   */
  public onEval(handler: StockfishEventHandler): () => void {
    this.evalHandlers.push(handler);
    // Return unsubscribe function
    return () => {
      this.evalHandlers = this.evalHandlers.filter((h) => h !== handler);
    };
  }

  /**
   * Stop current analysis
   */
  public stop(): void {
    if (!this.worker || !this.isInitialized) {
      return;
    }

    this.worker.postMessage({ type: 'stop' });
  }

  /**
   * Terminate the worker
   */
  public terminate(): void {
    if (this.worker) {
      this.worker.terminate();
      this.worker = null;
    }
    this.isInitialized = false;
    this.initPromise = null;
    this.evalHandlers = [];
  }

  /**
   * Check if worker is initialized
   */
  public isReady(): boolean {
    return this.isInitialized;
  }
}

/**
 * Create a singleton instance
 */
let instance: StockfishWorker | null = null;

export function getStockfishInstance(): StockfishWorker {
  if (!instance) {
    instance = new StockfishWorker();
  }
  return instance;
}

export function resetStockfishInstance(): void {
  if (instance) {
    instance.terminate();
  }
  instance = null;
}
