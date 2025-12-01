import React, { useState } from 'react';
import { SequencesTab } from './SequencesTab';
import { SequencingPage } from '../v2/SequencingPage';
import { Button } from '../../components/ui';

interface SequencesTabWrapperProps {
  projectId: string;
}

export const SequencesTabWrapper: React.FC<SequencesTabWrapperProps> = ({ projectId }) => {
  const [useOldUI, setUseOldUI] = useState(false);

  return (
    <div className="space-y-4">
      {/* Toggle between old and new UI */}
      <div className="flex items-center justify-between bg-gray-50 p-3 rounded-lg border border-gray-200">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-gray-700">UI Version:</span>
          <span className={`text-sm px-2 py-1 rounded ${useOldUI ? 'bg-yellow-100 text-yellow-800' : 'bg-green-100 text-green-800'}`}>
            {useOldUI ? 'Legacy' : 'New (v2)'}
          </span>
        </div>
        <Button
          onClick={() => setUseOldUI(!useOldUI)}
          variant="secondary"
          size="sm"
        >
          Switch to {useOldUI ? 'New' : 'Legacy'} UI
        </Button>
      </div>

      {/* Render appropriate UI */}
      {useOldUI ? (
        <SequencesTab projectId={projectId} />
      ) : (
        <div className="border border-gray-200 rounded-lg overflow-hidden -m-6">
          <SequencingPage projectId={projectId} />
        </div>
      )}
    </div>
  );
};

