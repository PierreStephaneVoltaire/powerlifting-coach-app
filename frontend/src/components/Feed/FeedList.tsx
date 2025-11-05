import React, { useState, useEffect } from 'react';
import { apiClient } from '@/utils/api';
import { useFeedCache } from '@/hooks/useFeedCache';
import { useAuthStore } from '@/store/authStore';
import { FeedPost } from '@/types';
import { CommentSection } from './CommentSection';

export const FeedList: React.FC = () => {
  const { user } = useAuthStore();
  const [posts, setPosts] = useState<FeedPost[]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [hasMore, setHasMore] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isStale, setIsStale] = useState(false);
  const [expandedPost, setExpandedPost] = useState<string | null>(null);
  const [likedPosts, setLikedPosts] = useState<Set<string>>(new Set());

  const { getCachedFeed, cacheFeed } = useFeedCache();

  const loadFeed = async (nextCursor?: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await apiClient.getFeed(20, nextCursor);

      const newPosts = nextCursor ? [...posts, ...response.posts] : response.posts;
      setPosts(newPosts);
      setCursor(response.next_cursor);
      setHasMore(!!response.next_cursor);
      setIsStale(false);

      await cacheFeed(newPosts);
    } catch (err: any) {
      console.error('Failed to load feed', err);

      const isNetworkError = !err.response || err.code === 'ECONNABORTED' || err.code === 'ERR_NETWORK';

      if (isNetworkError && !nextCursor) {
        const cachedPosts = await getCachedFeed();
        if (cachedPosts.length > 0) {
          setPosts(cachedPosts);
          setIsStale(true);
          setError('Showing cached feed. Network unavailable.');
        } else {
          setError('Unable to load feed. Please check your connection.');
        }
      } else {
        setError('Failed to load more posts.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadFeed();
  }, []);

  const handleLoadMore = () => {
    if (!isLoading && hasMore && cursor) {
      loadFeed(cursor);
    }
  };

  const handleLike = async (postId: string) => {
    if (!user) return;

    const isLiked = likedPosts.has(postId);
    const newLikedPosts = new Set(likedPosts);

    if (isLiked) {
      newLikedPosts.delete(postId);
    } else {
      newLikedPosts.add(postId);
    }

    setLikedPosts(newLikedPosts);

    const updatedPosts = posts.map(p =>
      p.post_id === postId
        ? { ...p, likes_count: (p.likes_count || 0) + (isLiked ? -1 : 1) }
        : p
    );
    setPosts(updatedPosts);

    try {
      await apiClient.submitLike(user.id, 'post', postId, isLiked ? 'unlike' : 'like');
      console.info('Like toggled', { post_id: postId, action: isLiked ? 'unlike' : 'like' });
    } catch (err: any) {
      console.error('Failed to submit like', err);
      if (!err.queued) {
        setLikedPosts(likedPosts);
        setPosts(posts);
      }
    }
  };

  return (
    <div className="max-w-2xl mx-auto p-4">
      {isStale && (
        <div className="mb-4 p-3 bg-yellow-50 border border-yellow-300 rounded-lg">
          <p className="text-sm text-yellow-800">
            Data may be out of date. Network connection unavailable.
          </p>
        </div>
      )}

      {error && !isStale && (
        <div className="mb-4 p-3 bg-red-50 border border-red-300 rounded-lg">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      <div className="space-y-4">
        {posts.map((post) => (
          <div
            key={post.post_id}
            className="bg-white rounded-lg shadow p-4"
          >
            <div className="flex items-center mb-2">
              <div className="flex-1">
                <h3 className="font-semibold text-gray-900">{post.user_name || 'Anonymous'}</h3>
                <p className="text-xs text-gray-500">
                  {new Date(post.created_at).toLocaleString()}
                </p>
              </div>
            </div>

            {post.media_url && (
              <div className="mb-3">
                <video
                  src={post.media_url}
                  controls
                  className="w-full rounded-lg"
                  poster={post.thumbnail_url}
                />
              </div>
            )}

            <div className="space-y-1 text-sm">
              {post.movement_label && (
                <p className="text-gray-700">
                  <span className="font-medium">Movement:</span> {post.movement_label}
                </p>
              )}
              {post.weight && (
                <p className="text-gray-700">
                  <span className="font-medium">Weight:</span> {post.weight}
                </p>
              )}
              {post.rpe && (
                <p className="text-gray-700">
                  <span className="font-medium">RPE:</span> {post.rpe}
                </p>
              )}
            </div>

            {post.comment_text && (
              <p className="mt-2 text-gray-800">{post.comment_text}</p>
            )}

            <div className="mt-3 flex items-center gap-4 text-sm text-gray-600">
              <button
                onClick={() => handleLike(post.post_id)}
                className={`flex items-center gap-1 transition-colors ${
                  likedPosts.has(post.post_id)
                    ? 'text-red-600 hover:text-red-700'
                    : 'hover:text-blue-600'
                }`}
              >
                <svg
                  className="w-5 h-5"
                  fill={likedPosts.has(post.post_id) ? 'currentColor' : 'none'}
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
                  />
                </svg>
                <span>{post.likes_count || 0}</span>
              </button>
              <button
                onClick={() => setExpandedPost(expandedPost === post.post_id ? null : post.post_id)}
                className="hover:text-blue-600 flex items-center gap-1"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                  />
                </svg>
                <span>{post.comments_count || 0}</span>
              </button>
            </div>

            {expandedPost === post.post_id && (
              <CommentSection postId={post.post_id} />
            )}
          </div>
        ))}
      </div>

      {hasMore && (
        <div className="mt-6 text-center">
          <button
            onClick={handleLoadMore}
            disabled={isLoading}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Loading...' : 'Load More'}
          </button>
        </div>
      )}

      {!isLoading && posts.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <p>No posts yet. Start uploading your lifts!</p>
        </div>
      )}
    </div>
  );
};
