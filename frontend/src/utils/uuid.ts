/**
 * Generate a UUID v4 compatible with both secure and non-secure contexts
 * Falls back to a custom implementation when crypto.randomUUID is not available
 */
export function generateUUID(): string {
  // Check if crypto.randomUUID is available (HTTPS or Node.js)
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }

  // Fallback implementation for non-secure contexts (HTTP)
  // This uses Math.random() which is less secure but works everywhere
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
