import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { apiClient } from '@/utils/api';

import { generateUUID } from '@/utils/uuid';
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
        client_generated_id: generateUUID(),
        user_id: response.user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          user_id: response.user.id,
          session_id: response.tokens.access_token.substring(0, 10),
        },
      };

      await apiClient.submitEvent(event);

      if (response.user.user_type === 'athlete') {
        try {
          await apiClient.getUserSettings();
          navigate('/feed');
        } catch (settingsError: any) {
          if (settingsError.response?.status === 404) {
            navigate('/onboarding');
          } else {
            navigate('/feed');
          }
        }
      } else {
        navigate('/feed');
      }
    } catch (err: any) {
      console.error('Login error:', err);
      console.error('Response data:', err.response?.data);
      console.error('Response status:', err.response?.status);

      if (err.response) {
        const errorMessage = err.response.data?.error || err.response.data?.message;
        if (errorMessage) {
          setError(errorMessage);
        } else {
          setError(`Login failed (${err.response.status}). Please try again.`);
        }
      } else if (err.request) {
        setError('Cannot connect to server. Please check your connection.');
      } else {
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
