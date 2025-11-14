import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { apiClient } from '@/utils/api';
import { ThemeProvider } from '@/context/ThemeContext';
import { DevModeProvider } from '@/context/DevModeContext';
import { DevModeToggle } from '@/components/DevMode/DevModeToggle';
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
import { AnalyticsPage } from '@/pages/AnalyticsPage';
import { ExerciseLibraryPage } from '@/pages/ExerciseLibraryPage';
import { WorkoutHistoryPage } from '@/pages/WorkoutHistoryPage';
import { CompPrepPage } from '@/pages/CompPrepPage';
import { CoachDirectoryPage } from '@/pages/CoachDirectoryPage';
import { CoachProfilePage } from '@/pages/CoachProfilePage';
import { RelationshipManagerPage } from '@/pages/RelationshipManagerPage';
import { AthleteProfilePage } from '@/pages/AthleteProfilePage';
import { PrivacySettingsPage } from '@/pages/PrivacySettingsPage';
import './index.css';

const queryClient = new QueryClient();

function App() {
  useEffect(() => {
    apiClient.startOfflineQueueProcessor();
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <DevModeProvider>
        <ThemeProvider>
          <DevModeToggle />
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
              <Route path="/analytics" element={<AnalyticsPage />} />
              <Route path="/exercises" element={<ExerciseLibraryPage />} />
              <Route path="/history" element={<WorkoutHistoryPage />} />
              <Route path="/comp-prep" element={<CompPrepPage />} />
              <Route path="/coaches" element={<CoachDirectoryPage />} />
              <Route path="/coaches/:coachId" element={<CoachProfilePage />} />
              <Route path="/relationships" element={<RelationshipManagerPage />} />
              <Route path="/athletes/:athleteId" element={<AthleteProfilePage />} />
              <Route path="/settings/privacy" element={<PrivacySettingsPage />} />
              <Route path="/dm/*" element={<DMPage />} />
              <Route path="/tools" element={<ToolsPage />} />
              <Route path="/" element={<Navigate to="/feed" replace />} />
            </Route>
            <Route path="*" element={<Navigate to="/feed" replace />} />
            </Routes>
          </Router>
        </ThemeProvider>
      </DevModeProvider>
    </QueryClientProvider>
  );
}

export default App;
