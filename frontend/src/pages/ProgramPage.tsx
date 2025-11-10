import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '@/utils/api';
import { ProgramApprovalView } from '@/components/Program/ProgramApprovalView';
import { ProgramOverview } from '@/components/Program/ProgramOverview';

export const ProgramPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [activeProgram, setActiveProgram] = useState<any>(null);
  const [pendingProgram, setPendingProgram] = useState<any>(null);

  useEffect(() => {
    loadProgramData();
  }, []);

  const loadProgramData = async () => {
    try {
      setLoading(true);

      // Check for active program first
      const activeResponse = await apiClient.getActiveProgram();
      if (activeResponse.has_program && activeResponse.program) {
        setActiveProgram(activeResponse.program);
      }

      // Check for pending program
      const pendingResponse = await apiClient.getPendingProgram();
      if (pendingResponse.has_pending && pendingResponse.program) {
        setPendingProgram(pendingResponse.program);
      }

      // If neither exists, redirect to chat
      if (!activeResponse.has_program && !pendingResponse.has_pending) {
        navigate('/chat');
      }
    } catch (error) {
      console.error('Failed to load program data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleApproveProgram = async () => {
    if (!pendingProgram) return;

    try {
      await apiClient.approveProgram(pendingProgram.id);
      // Reload program data to show the approved program
      await loadProgramData();
    } catch (error) {
      console.error('Failed to approve program:', error);
      alert('Failed to approve program. Please try again.');
    }
  };

  const handleRejectProgram = async () => {
    if (!pendingProgram) return;

    try {
      await apiClient.rejectProgram(pendingProgram.id);
      // Redirect back to chat to create a new program
      navigate('/chat');
    } catch (error) {
      console.error('Failed to reject program:', error);
      alert('Failed to reject program. Please try again.');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading program...</p>
        </div>
      </div>
    );
  }

  // If there's a pending program, show the approval view
  if (pendingProgram) {
    return (
      <ProgramApprovalView
        pendingProgram={pendingProgram}
        currentProgram={activeProgram}
        onApprove={handleApproveProgram}
        onReject={handleRejectProgram}
      />
    );
  }

  // Otherwise show the active program overview
  if (activeProgram) {
    return <ProgramOverview program={activeProgram} onRefresh={loadProgramData} />;
  }

  // Fallback: no program found
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center max-w-md">
        <div className="text-6xl mb-4">📋</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">No Program Found</h2>
        <p className="text-gray-600 mb-6">
          You don't have an active training program yet. Let's create one with the AI coach!
        </p>
        <button
          onClick={() => navigate('/chat')}
          className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Create Program
        </button>
      </div>
    </div>
  );
};
