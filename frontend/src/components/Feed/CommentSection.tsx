import React, { useState, useEffect } from 'react';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';
import { Comment } from '@/types';

interface CommentSectionProps {
  postId: string;
}

export const CommentSection: React.FC<CommentSectionProps> = ({ postId }) => {
  const { user } = useAuthStore();
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentText, setCommentText] = useState('');
  const [replyingTo, setReplyingTo] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadComments();
  }, [postId]);

  const loadComments = async () => {
    try {
      const response = await apiClient.getPostComments(postId);
      setComments(response.comments || []);
    } catch (err) {
      console.error('Failed to load comments', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmitComment = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!user || !commentText.trim()) return;

    setIsSubmitting(true);

    try {
      await apiClient.submitComment(user.id, postId, commentText, replyingTo || undefined);

      const tempComment: Comment = {
        id: crypto.randomUUID(),
        comment_id: crypto.randomUUID(),
        post_id: postId,
        user_id: user.id,
        parent_comment_id: replyingTo || undefined,
        comment_text: commentText,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      setComments([...comments, tempComment]);
      setCommentText('');
      setReplyingTo(null);
      console.info('Comment submitted', { post_id: postId });
    } catch (err: any) {
      console.error('Failed to submit comment', err);
      if (err.queued) {
        setCommentText('');
        setReplyingTo(null);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const buildCommentTree = (comments: Comment[]): Comment[] => {
    return comments.filter(c => !c.parent_comment_id);
  };

  const getReplies = (commentId: string): Comment[] => {
    return comments.filter(c => c.parent_comment_id === commentId);
  };

  const renderComment = (comment: Comment, depth = 0) => {
    const replies = getReplies(comment.comment_id);

    return (
      <div key={comment.comment_id} className={depth > 0 ? 'ml-8 mt-2' : 'mt-3'}>
        <div className="bg-gray-50 rounded-lg p-3">
          <div className="flex items-start justify-between mb-1">
            <span className="text-sm font-semibold text-gray-900">
              {comment.user_id === user?.id ? 'You' : 'User'}
            </span>
            <span className="text-xs text-gray-500">
              {new Date(comment.created_at).toLocaleString()}
            </span>
          </div>
          <p className="text-sm text-gray-800">{comment.comment_text}</p>
          <button
            onClick={() => setReplyingTo(comment.comment_id)}
            className="text-xs text-blue-600 hover:text-blue-700 mt-2"
          >
            Reply
          </button>
        </div>
        {replies.map(reply => renderComment(reply, depth + 1))}
      </div>
    );
  };

  const rootComments = buildCommentTree(comments);

  return (
    <div className="mt-4">
      <h4 className="text-sm font-semibold text-gray-900 mb-3">
        Comments ({comments.length})
      </h4>

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading comments...</p>
      ) : (
        <>
          {rootComments.map(comment => renderComment(comment))}

          {comments.length === 0 && (
            <p className="text-sm text-gray-500 text-center py-4">
              No comments yet. Be the first to comment!
            </p>
          )}
        </>
      )}

      <form onSubmit={handleSubmitComment} className="mt-4">
        {replyingTo && (
          <div className="mb-2 flex items-center justify-between bg-blue-50 px-3 py-2 rounded">
            <span className="text-xs text-blue-700">Replying to comment</span>
            <button
              type="button"
              onClick={() => setReplyingTo(null)}
              className="text-xs text-blue-600 hover:text-blue-800"
            >
              Cancel
            </button>
          </div>
        )}
        <div className="flex gap-2">
          <input
            type="text"
            value={commentText}
            onChange={(e) => setCommentText(e.target.value)}
            placeholder="Add a comment..."
            className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-md"
          />
          <button
            type="submit"
            disabled={isSubmitting || !commentText.trim()}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            Post
          </button>
        </div>
      </form>
    </div>
  );
};
