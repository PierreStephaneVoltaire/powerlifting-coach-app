import React from 'react';

interface ExerciseDetailProps {
  exercise: any;
  onClose: () => void;
}

export const ExerciseDetail: React.FC<ExerciseDetailProps> = ({ exercise, onClose }) => {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex justify-between items-start">
            <div>
              <h2 className="text-2xl font-bold">{exercise.name}</h2>
              <div className="flex gap-2 mt-2">
                <span className={`px-2 py-1 rounded text-xs font-medium ${
                  exercise.lift_type === 'squat' ? 'bg-red-100 text-red-800' :
                  exercise.lift_type === 'bench' ? 'bg-blue-100 text-blue-800' :
                  exercise.lift_type === 'deadlift' ? 'bg-green-100 text-green-800' :
                  'bg-purple-100 text-purple-800'
                }`}>
                  {exercise.lift_type}
                </span>
                {exercise.difficulty && (
                  <span className={`px-2 py-1 rounded text-xs font-medium ${
                    exercise.difficulty === 'beginner' ? 'bg-green-100 text-green-800' :
                    exercise.difficulty === 'intermediate' ? 'bg-yellow-100 text-yellow-800' :
                    'bg-red-100 text-red-800'
                  }`}>
                    {exercise.difficulty}
                  </span>
                )}
                {exercise.is_custom && (
                  <span className="px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded">
                    Custom
                  </span>
                )}
              </div>
            </div>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Description */}
          {exercise.description && (
            <div>
              <h3 className="font-semibold mb-2">Description</h3>
              <p className="text-gray-700">{exercise.description}</p>
            </div>
          )}

          {/* Demo Video */}
          {exercise.demo_video_url && (
            <div>
              <h3 className="font-semibold mb-2">Demo Video</h3>
              <div className="aspect-video bg-gray-100 rounded-lg overflow-hidden">
                {exercise.demo_video_url.includes('youtube.com') || exercise.demo_video_url.includes('youtu.be') ? (
                  <iframe
                    src={exercise.demo_video_url.replace('watch?v=', 'embed/')}
                    className="w-full h-full"
                    allowFullScreen
                  />
                ) : (
                  <div className="flex items-center justify-center h-full">
                    <a
                      href={exercise.demo_video_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:text-blue-800"
                    >
                      View Demo Video →
                    </a>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Muscles Worked */}
          {((exercise.primary_muscles && exercise.primary_muscles.length > 0) ||
            (exercise.secondary_muscles && exercise.secondary_muscles.length > 0)) && (
            <div>
              <h3 className="font-semibold mb-2">Muscles Worked</h3>
              <div className="grid grid-cols-2 gap-4">
                {exercise.primary_muscles && exercise.primary_muscles.length > 0 && (
                  <div>
                    <div className="text-sm font-medium text-gray-600 mb-2">Primary</div>
                    <div className="flex flex-wrap gap-1">
                      {exercise.primary_muscles.map((muscle: string, idx: number) => (
                        <span key={idx} className="px-2 py-1 bg-blue-100 text-blue-800 text-sm rounded">
                          {muscle}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {exercise.secondary_muscles && exercise.secondary_muscles.length > 0 && (
                  <div>
                    <div className="text-sm font-medium text-gray-600 mb-2">Secondary</div>
                    <div className="flex flex-wrap gap-1">
                      {exercise.secondary_muscles.map((muscle: string, idx: number) => (
                        <span key={idx} className="px-2 py-1 bg-gray-100 text-gray-800 text-sm rounded">
                          {muscle}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Equipment */}
          {exercise.equipment_needed && exercise.equipment_needed.length > 0 && (
            <div>
              <h3 className="font-semibold mb-2">Equipment Needed</h3>
              <div className="flex flex-wrap gap-2">
                {exercise.equipment_needed.map((equipment: string, idx: number) => (
                  <span key={idx} className="px-3 py-1 bg-gray-100 text-gray-800 rounded">
                    {equipment.replace(/_/g, ' ')}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Instructions */}
          {exercise.instructions && (
            <div>
              <h3 className="font-semibold mb-2">Instructions</h3>
              <p className="text-gray-700 whitespace-pre-wrap">{exercise.instructions}</p>
            </div>
          )}

          {/* Form Cues */}
          {exercise.form_cues && exercise.form_cues.length > 0 && (
            <div>
              <h3 className="font-semibold mb-2">Form Cues</h3>
              <ul className="space-y-2">
                {exercise.form_cues.map((cue: string, idx: number) => (
                  <li key={idx} className="flex items-start gap-2">
                    <svg className="w-5 h-5 text-green-600 mt-0.5 shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    <span className="text-gray-700">{cue}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
          <button
            onClick={onClose}
            className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
