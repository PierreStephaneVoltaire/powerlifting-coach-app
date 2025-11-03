import React from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import './index.css';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div className="min-h-screen bg-gray-100">
          <h1 className="text-3xl font-bold text-center py-8">
            Powerlifting Coach
          </h1>
        </div>
      </Router>
    </QueryClientProvider>
  );
}

export default App;
