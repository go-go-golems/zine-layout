import React from 'react';
import { useAppSelector } from '../../hooks/redux';
import { useComputeSpreadMutation } from '../../api';
import { useSpreadRequest } from '../../utils/spreadRequestBuilder';

export const AlgorithmDebugPanel: React.FC = () => {
  const image = useAppSelector((state) => state.bookSpread.image);
  const algorithm = useAppSelector((state) => state.bookSpread.algorithm);
  const spreadRequest = useSpreadRequest();
  const [computeSpread, { data: computeResult, isLoading }] = useComputeSpreadMutation();

  if (!image) {
    return (
      <div className="mt-4 p-4 bg-gray-100 rounded-lg">
        <h4 className="font-semibold text-sm">Algorithm Debug</h4>
        <p className="text-xs text-gray-500">No image loaded</p>
      </div>
    );
  }

  const handleComputeDebug = () => {
    if (spreadRequest) {
      computeSpread(spreadRequest);
    }
  };

  const debugData = {
    algorithm,
    computeResult: computeResult?.result,
    request: spreadRequest,
  };

  const debugText = JSON.stringify(debugData, null, 2);

  const copyToClipboard = () => {
    navigator.clipboard.writeText(debugText).then(() => {
      alert('Debug data copied to clipboard!');
    }).catch(err => {
      console.error('Failed to copy to clipboard:', err);
      alert('Failed to copy to clipboard');
    });
  };

  return (
    <div className="mt-4 p-4 bg-gray-100 rounded-lg">
      <div className="flex justify-between items-center mb-2">
        <h4 className="font-semibold text-sm">Algorithm Debug ({algorithm})</h4>
        <div className="space-x-2">
          <button
            onClick={handleComputeDebug}
            disabled={isLoading || !spreadRequest}
            className="px-3 py-1 bg-green-500 text-white text-xs rounded hover:bg-green-600 transition-colors disabled:opacity-50"
          >
            {isLoading ? '⏳' : '🔄'} Compute
          </button>
          <button
            onClick={copyToClipboard}
            className="px-3 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition-colors"
          >
            📋 Copy Debug Data
          </button>
        </div>
      </div>
      <div className="text-xs text-gray-600 mb-2">
        Backend {algorithm} algorithm result and request parameters
      </div>
      <pre className="text-xs overflow-auto max-h-60 bg-white p-3 rounded border font-mono">
        {debugText}
      </pre>
    </div>
  );
};
