import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { apiClient } from '@/utils/api';
import { ThemeProvider } from '@/context/ThemeContext';
import { ProtectedRoute } from '@/components/Auth/ProtectedRoute';
import { OnboardingCheck } from '@/components/Auth/OnboardingCheck';
import { MainLayout } from '@/components/Layout/MainLayout';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { OnboardingPage } from '@/pages/OnboardingPage';
import { FeedPage } from '@/pages/FeedPage';
import { ProgramPage } from '@/pages/ProgramPage';
import { DMPage } from '@/pages/DMPage';
import { ToolsPage } from '@/pages/ToolsPage';
import { ChatPage } from '@/pages/ChatPage';
import './index.css';

const queryClient = new QueryClient();

function App() {
  useEffect(() => {
    apiClient.startOfflineQueueProcessor();
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <Router>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/onboarding" element={
              <ProtectedRoute>
                <OnboardingPage />
              </ProtectedRoute>
            } />
            <Route path="/chat" element={
              <ProtectedRoute>
                <ChatPage />
              </ProtectedRoute>
            } />
            <Route element={
              <ProtectedRoute>
                <OnboardingCheck>
                  <MainLayout />
                </OnboardingCheck>
              </ProtectedRoute>
            }>
              <Route path="/feed" element={<FeedPage />} />
              <Route path="/program" element={<ProgramPage />} />
              <Route path="/dm/*" element={<DMPage />} />
              <Route path="/tools" element={<ToolsPage />} />
              <Route path="/" element={<Navigate to="/feed" replace />} />
            </Route>
            <Route path="*" element={<Navigate to="/feed" replace />} />
          </Routes>
        </Router>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;
