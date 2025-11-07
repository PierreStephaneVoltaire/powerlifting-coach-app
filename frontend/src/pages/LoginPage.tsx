import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { apiClient } from '@/utils/api';

export const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuthStore();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const response = await apiClient.login(email, password);

      // Validate response structure
      if (!response || !response.tokens || !response.user) {
        console.error('Invalid response structure:', response);
        setError('Login succeeded but received invalid response format. Please try again.');
        setIsLoading(false);
        return;
      }

      login(response.tokens, response);

      const event = {
        schema_version: '1.0.0',
        event_type: 'auth.user.logged_in',
        client_generated_id: crypto.randomUUID(),
        user_id: response.user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          user_id: response.user.id,
          session_id: response.tokens.access_token.substring(0, 10),
        },
      };

      await apiClient.submitEvent(event);

      // Check if user has completed onboarding by trying to fetch settings
      try {
        await apiClient.getUserSettings();
        // Settings exist, user has completed onboarding
        navigate('/feed');
      } catch (settingsErr: any) {
        // Settings don't exist (404), user needs onboarding
        if (settingsErr.response?.status === 404) {
          navigate('/onboarding');
        } else {
          // Other error, default to feed
          navigate('/feed');
        }
      }
    } catch (err: any) {
      console.error('Login error:', err);
      console.error('Response data:', err.response?.data);
      console.error('Response status:', err.response?.status);

      // Handle different error scenarios
      if (err.response) {
        // Server responded with error status
        const errorMessage = err.response.data?.error || err.response.data?.message;
        if (errorMessage) {
          setError(errorMessage);
        } else {
          setError(`Login failed (${err.response.status}). Please try again.`);
        }
      } else if (err.request) {
        // Request made but no response
        setError('Cannot connect to server. Please check your connection.');
      } else {
        // Something else went wrong
        setError(err.message || 'Login failed. Please try again.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center">
      <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
        <h1 className="text-3xl font-bold text-center mb-6">Powerlifting Coach</h1>

        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full border rounded p-2"
              required
              disabled={isLoading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full border rounded p-2"
              required
              disabled={isLoading}
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 disabled:bg-gray-300"
          >
            {isLoading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <p className="text-center mt-4 text-sm text-gray-600">
          Don't have an account?{' '}
          <Link to="/register" className="text-blue-500 hover:text-blue-600">
            Sign up here
          </Link>
        </p>
      </div>
    </div>
  );
};
