import { openDB, DBSchema, IDBPDatabase } from 'idb';

import { generateUUID } from '@/utils/uuid';
interface QueuedEvent {
  id: string;
  event: any;
  timestamp: number;
  retryCount: number;
  nextRetryAt: number;
}

interface OfflineDB extends DBSchema {
  events: {
    key: string;
    value: QueuedEvent;
    indexes: { 'by-nextRetry': number };
  };
}

const DB_NAME = 'powerlifting_offline_queue';
const DB_VERSION = 1;
const STORE_NAME = 'events';
const MAX_RETRIES = 5;
const INITIAL_BACKOFF_MS = 1000;

class OfflineQueue {
  private db: IDBPDatabase<OfflineDB> | null = null;
  private processingQueue = false;

  async init() {
    if (this.db) return;

    this.db = await openDB<OfflineDB>(DB_NAME, DB_VERSION, {
      upgrade(db) {
        if (!db.objectStoreNames.contains(STORE_NAME)) {
          const store = db.createObjectStore(STORE_NAME, { keyPath: 'id' });
          store.createIndex('by-nextRetry', 'nextRetryAt');
        }
      },
    });
  }

  private calculateBackoff(retryCount: number): number {
    const backoffMs = Math.min(
      INITIAL_BACKOFF_MS * Math.pow(2, retryCount),
      60000
    );
    const jitter = Math.random() * backoffMs * 0.1;
    return backoffMs + jitter;
  }

  async enqueue(event: any): Promise<string> {
    await this.init();
    if (!this.db) throw new Error('Failed to init DB');

    const id = event.client_generated_id || generateUUID();
    const queuedEvent: QueuedEvent = {
      id,
      event,
      timestamp: Date.now(),
      retryCount: 0,
      nextRetryAt: Date.now(),
    };

    await this.db.put(STORE_NAME, queuedEvent);
    console.info('Event enqueued for offline submission', { id, event_type: event.event_type });

    this.processQueue();

    return id;
  }

  async processQueue(): Promise<void> {
    if (this.processingQueue) return;

    this.processingQueue = true;

    try {
      await this.init();
      if (!this.db) return;

      const tx = this.db.transaction(STORE_NAME, 'readonly');
      const index = tx.store.index('by-nextRetry');
      const events = await index.getAll(IDBKeyRange.upperBound(Date.now()));

      for (const queuedEvent of events) {
        await this.processEvent(queuedEvent);
      }
    } catch (error) {
      console.error('Error processing queue', error);
    } finally {
      this.processingQueue = false;
    }
  }

  private async processEvent(queuedEvent: QueuedEvent): Promise<void> {
    if (!this.db) return;

    try {
      console.log(`Processing queued event: ${queuedEvent.event.event_type}`, {
        id: queuedEvent.id,
        retry: queuedEvent.retryCount,
      });

      const { apiClient } = await import('./api');
      await apiClient.submitEvent(queuedEvent.event);

      await this.db.delete(STORE_NAME, queuedEvent.id);
      console.log(`Event successfully submitted: ${queuedEvent.event.event_type}`, {
        id: queuedEvent.id,
      });
    } catch (error: any) {
      console.error(`Failed to submit queued event: ${queuedEvent.event.event_type}`, {
        id: queuedEvent.id,
        error: error.message,
        retry: queuedEvent.retryCount,
      });

      if (queuedEvent.retryCount >= MAX_RETRIES) {
        console.error('Max retries exceeded, removing event from queue', {
          id: queuedEvent.id,
        });
        await this.db.delete(STORE_NAME, queuedEvent.id);
        return;
      }

      const updatedEvent: QueuedEvent = {
        ...queuedEvent,
        retryCount: queuedEvent.retryCount + 1,
        nextRetryAt: Date.now() + this.calculateBackoff(queuedEvent.retryCount),
      };

      await this.db.put(STORE_NAME, updatedEvent);
    }
  }

  async getPendingCount(): Promise<number> {
    await this.init();
    if (!this.db) return 0;

    const count = await this.db.count(STORE_NAME);
    return count;
  }

  async clearQueue(): Promise<void> {
    await this.init();
    if (!this.db) return;

    await this.db.clear(STORE_NAME);
    console.info('Offline queue cleared');
  }

  startAutoProcess(intervalMs = 30000) {
    setInterval(() => {
      this.processQueue();
    }, intervalMs);

    window.addEventListener('online', () => {
      console.info('Network online, processing queue');
      this.processQueue();
    });
  }
}

export const offlineQueue = new OfflineQueue();
