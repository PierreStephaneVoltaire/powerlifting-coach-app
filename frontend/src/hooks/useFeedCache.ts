import { openDB, DBSchema, IDBPDatabase } from 'idb';
import { FeedPost } from '@/types';

interface FeedCacheDB extends DBSchema {
  feed: {
    key: string;
    value: {
      posts: FeedPost[];
      timestamp: number;
    };
  };
}

const DB_NAME = 'powerlifting_feed_cache';
const DB_VERSION = 1;
const STORE_NAME = 'feed';
const CACHE_KEY = 'main_feed';
const CACHE_TTL_MS = 1000 * 60 * 30;

export const useFeedCache = () => {
  let db: IDBPDatabase<FeedCacheDB> | null = null;

  const initDB = async () => {
    if (db) return db;

    db = await openDB<FeedCacheDB>(DB_NAME, DB_VERSION, {
      upgrade(database) {
        if (!database.objectStoreNames.contains(STORE_NAME)) {
          database.createObjectStore(STORE_NAME);
        }
      },
    });

    return db;
  };

  const cacheFeed = async (posts: FeedPost[]) => {
    const database = await initDB();
    await database.put(STORE_NAME, {
      posts,
      timestamp: Date.now(),
    }, CACHE_KEY);
    console.info('Feed cached', { count: posts.length });
  };

  const getCachedFeed = async (): Promise<FeedPost[]> => {
    const database = await initDB();
    const cached = await database.get(STORE_NAME, CACHE_KEY);

    if (!cached) {
      console.info('No cached feed found');
      return [];
    }

    const age = Date.now() - cached.timestamp;
    console.info('Retrieved cached feed', {
      count: cached.posts.length,
      age_minutes: Math.round(age / 1000 / 60),
    });

    return cached.posts;
  };

  const isFeedStale = async (): Promise<boolean> => {
    const database = await initDB();
    const cached = await database.get(STORE_NAME, CACHE_KEY);

    if (!cached) return true;

    const age = Date.now() - cached.timestamp;
    return age > CACHE_TTL_MS;
  };

  const clearCache = async () => {
    const database = await initDB();
    await database.delete(STORE_NAME, CACHE_KEY);
    console.info('Feed cache cleared');
  };

  return {
    cacheFeed,
    getCachedFeed,
    isFeedStale,
    clearCache,
  };
};
