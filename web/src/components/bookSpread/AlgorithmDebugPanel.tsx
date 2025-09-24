import React, { useEffect } from 'react';
import { useAppSelector } from '../../hooks/redux';
import { useBuildYamlMutation, useComputeSpreadMutation } from '../../api';
import { useSpreadRequest } from '../../utils/spreadRequestBuilder';

const formatError = (error: unknown): string => {
  if (!error) return '';
  if (typeof error === 'string') return error;
  if (typeof error === 'object') {
    const maybeObj = error as Record<string, unknown>;
    const status = maybeObj['status'];
    const data = maybeObj['data'];
    if (typeof status !== 'undefined' && typeof data !== 'undefined') {
      return `${String(status)}: ${String(data)}`;
    }
    if (typeof maybeObj['message'] === 'string') {
      return String(maybeObj['message']);
    }
  }
  try {
    return JSON.stringify(error);
  } catch {
    return String(error);
  }
};

export const AlgorithmDebugPanel: React.FC = () => {
  const image = useAppSelector((state) => state.bookSpread.image);
  const spreadRequest = useSpreadRequest();
  const [computeSpread, { data: computeResult, isLoading }] = useComputeSpreadMutation();
  const [buildYaml, { data: yamlText, isLoading: yamlLoading, error: yamlError }] = useBuildYamlMutation();

  useEffect(() => {
    if (!spreadRequest) return;
    buildYaml(spreadRequest).unwrap().catch(() => {
      /* error handled via yamlError */
    });
  }, [spreadRequest, buildYaml]);

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
    computeResult: computeResult?.result,
    trace: computeResult?.trace,
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

  const copyYaml = () => {
    if (!yamlText) return;
    navigator.clipboard.writeText(yamlText).then(() => {
      alert('YAML copied to clipboard!');
    }).catch(err => {
      console.error('Failed to copy YAML:', err);
      alert('Failed to copy YAML');
    });
  };

  return (
    <div className="mt-4 p-4 bg-gray-100 rounded-lg">
      <div className="flex justify-between items-center mb-2">
        <h4 className="font-semibold text-sm">Algorithm Debug</h4>
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
        Backend simple algorithm result and request parameters
      </div>
      <pre className="text-xs overflow-auto max-h-60 bg-white p-3 rounded border font-mono">
        {debugText}
      </pre>
      <div className="mt-4">
        <div className="flex justify-between items-center mb-2">
          <h5 className="font-semibold text-xs">Sonnet YAML</h5>
          <div className="space-x-2">
            <span className="text-[10px] text-gray-500">
              {yamlLoading ? 'Generating…' : yamlError ? 'Failed to generate YAML' : 'Copy to reproduce via CLI'}
            </span>
            <button
              onClick={copyYaml}
              disabled={!yamlText || yamlLoading}
              className="px-3 py-1 bg-purple-500 text-white text-xs rounded hover:bg-purple-600 transition-colors disabled:opacity-50"
            >
              📋 Copy YAML
            </button>
          </div>
        </div>
        <pre className="text-xs overflow-auto max-h-60 bg-white p-3 rounded border font-mono">
          {yamlText || (yamlError ? formatError(yamlError) : 'YAML will appear here once generated.')}
        </pre>
      </div>
    </div>
  );
};
