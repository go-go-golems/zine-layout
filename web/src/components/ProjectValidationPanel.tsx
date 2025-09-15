import type React from 'react';
import { useLazyValidateProjectQuery } from '../api';
import { Button } from './ui';

export const ProjectValidationPanel: React.FC<{ id: string }> = ({ id }) => {
  const [trigger, { data, isFetching, isUninitialized }] = useLazyValidateProjectQuery();

  const onRun = () => {
    trigger({ id });
  };

  return (
    <div className="space-y-4">
      {/* Trigger Button */}
      <Button 
        onClick={onRun} 
        isLoading={isFetching}
        className="w-full"
      >
        Run Validation
      </Button>

      {/* Results */}
      {!isUninitialized && data && (
        <div className="space-y-3">
          {/* Status Badge */}
          <div className="flex items-center justify-center">
            {data.ok ? (
              <div className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                <svg className="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                Validation Passed
              </div>
            ) : (
              <div className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800">
                <svg className="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Issues Found
              </div>
            )}
          </div>

          {/* Issues List */}
          {data.issues?.length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <h4 className="text-sm font-medium text-red-800 mb-2">Issues to resolve:</h4>
              <ul className="space-y-1">
                {data.issues.map((issue, index) => (
                  <li key={index} className="text-sm text-red-700 flex items-start">
                    <svg className="w-4 h-4 mr-2 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                    {issue}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Project Details */}
          {data.details && (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-3">
              <h4 className="text-sm font-medium text-gray-800 mb-2">Project Details:</h4>
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <span className="text-gray-500">Images:</span>
                  <span className="ml-1 font-medium">{data.details.count}</span>
                </div>
                <div>
                  <span className="text-gray-500">Dimensions:</span>
                  <span className="ml-1 font-medium">{data.details.width}×{data.details.height}</span>
                </div>
                <div>
                  <span className="text-gray-500">Grid:</span>
                  <span className="ml-1 font-medium">{data.details.rows}×{data.details.columns}</span>
                </div>
                <div>
                  <span className="text-gray-500">Pages:</span>
                  <span className="ml-1 font-medium">{data.details.pages}</span>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* No validation run yet */}
      {isUninitialized && (
        <div className="text-center py-4">
          <svg className="w-8 h-8 text-gray-300 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm text-gray-500">Click to validate your project</p>
        </div>
      )}
    </div>
  );
};
