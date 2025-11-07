import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import { apiClient } from '@/utils/api';

export const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuthStore();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [userType, setUserType] = useState<'athlete' | 'coach'>('athlete');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const response = await apiClient.register(email, password, name, userType);

      // Validate response structure
      if (!response || !response.tokens || !response.user) {
        console.error('Invalid response structure:', response);
        setError('Registration succeeded but received invalid response format. Please try logging in.');
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

      navigate('/onboarding');
    } catch (err: any) {
      console.error('Registration error:', err);
      console.error('Response data:', err.response?.data);
      console.error('Response status:', err.response?.status);

      // Handle different error scenarios
      if (err.response) {
        // Server responded with error status
        const errorMessage = err.response.data?.error || err.response.data?.message;
        if (errorMessage) {
          setError(errorMessage);
        } else {
          setError(`Registration failed (${err.response.status}). Please try again.`);
        }
      } else if (err.request) {
        // Request made but no response
        setError('Cannot connect to server. Please check your connection.');
      } else {
        // Something else went wrong
        setError(err.message || 'Registration failed. Please try again.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center">
      <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
        <h1 className="text-3xl font-bold text-center mb-6">Create Account</h1>

        <form onSubmit={handleRegister} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full border rounded p-2"
              required
              disabled={isLoading}
            />
          </div>

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
              minLength={8}
              disabled={isLoading}
            />
            <p className="text-xs text-gray-500 mt-1">Minimum 8 characters</p>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">I am a...</label>
            <div className="flex gap-4">
              <label className="flex items-center">
                <input
                  type="radio"
                  value="athlete"
                  checked={userType === 'athlete'}
                  onChange={(e) => setUserType(e.target.value as 'athlete')}
                  className="mr-2"
                  disabled={isLoading}
                />
                Athlete
              </label>
              <label className="flex items-center">
                <input
                  type="radio"
                  value="coach"
                  checked={userType === 'coach'}
                  onChange={(e) => setUserType(e.target.value as 'coach')}
                  className="mr-2"
                  disabled={isLoading}
                />
                Coach
              </label>
            </div>
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
            {isLoading ? 'Creating account...' : 'Create Account'}
          </button>
        </form>

        <p className="text-center mt-4 text-sm text-gray-600">
          Already have an account?{' '}
          <Link to="/login" className="text-blue-500 hover:text-blue-600">
            Login here
          </Link>
        </p>
      </div>
    </div>
  );
};
