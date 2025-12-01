import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { SequenceList } from './components/SequenceList';
import { SequenceEditor } from './components/SequenceEditor';

interface SequencingPageProps {
  projectId?: string;
}

export const SequencingPage: React.FC<SequencingPageProps> = ({ projectId: propProjectId }) => {
  const { projectId: routeProjectId } = useParams<{ projectId: string }>();
  const projectId = propProjectId ?? routeProjectId;
  const [selectedSequenceId, setSelectedSequenceId] = useState<string | null>(null);

  if (!projectId) {
    return (
      <div className="p-8">
        <div className="text-red-600">Error: Project ID is required</div>
      </div>
    );
  }

  return (
    <div className="flex min-h-[600px] bg-gray-50">
      {/* Left sidebar: Sequence list */}
      <div className="w-80 border-r border-gray-200 bg-white overflow-y-auto">
        <SequenceList
          projectId={projectId}
          selectedSequenceId={selectedSequenceId}
          onSelectSequence={setSelectedSequenceId}
        />
      </div>

      {/* Main content: Sequence editor */}
      <div className="flex-1 overflow-y-auto min-h-[600px]">
        {selectedSequenceId ? (
          <SequenceEditor
            projectId={projectId}
            sequenceId={selectedSequenceId}
          />
        ) : (
          <div className="flex items-center justify-center h-full min-h-[600px]">
            <div className="text-center text-gray-500">
              <p className="text-lg font-medium">No sequence selected</p>
              <p className="text-sm mt-2">Select a sequence from the list or create a new one</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

