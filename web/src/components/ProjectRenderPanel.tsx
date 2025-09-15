import React from 'react';
import { useGetRendersQuery, useRenderProjectMutation } from '../api';
import { Button } from './ui';

export const ProjectRenderPanel: React.FC<{ id: string }> = ({ id }) => {
  const { data, refetch, isFetching } = useGetRendersQuery({ id });
  const [renderProject, { isLoading }] = useRenderProjectMutation();
  const [test, setTest] = React.useState(false);
  const [testBW, setTestBW] = React.useState(false);
  const [testDimensions, setTestDimensions] = React.useState('600px,800px');

  const onRender = async () => {
    await renderProject({ id, test, test_bw: testBW, test_dimensions: testDimensions }).unwrap();
    refetch();
  };

  return (
    <div className="space-y-4">
      {/* Render Options */}
      <div className="space-y-3">
        <div className="flex items-center space-x-4">
          <label className="flex items-center text-sm">
            <input 
              type="checkbox" 
              checked={test} 
              onChange={(e) => setTest(e.target.checked)}
              className="mr-2 h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
            />
            Use test images
          </label>
          <label className="flex items-center text-sm">
            <input 
              type="checkbox" 
              checked={testBW} 
              onChange={(e) => setTestBW(e.target.checked)}
              className="mr-2 h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
            />
            Black & White
          </label>
        </div>
        
        {test && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Test Dimensions
            </label>
            <input
              value={testDimensions}
              onChange={(e) => setTestDimensions(e.target.value)}
              placeholder="600px,800px"
              className="input"
            />
          </div>
        )}
        
        <div className="flex space-x-2">
          <Button onClick={onRender} isLoading={isLoading} className="flex-1">
            Generate Render
          </Button>
          <Button variant="secondary" onClick={() => refetch()} isLoading={isFetching}>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </Button>
        </div>
      </div>

      {/* Renders List */}
      {data?.renders?.length ? (
        <div className="space-y-4">
          <h4 className="text-sm font-medium text-gray-900">Recent Renders</h4>
          <div className="space-y-3">
            {data.renders.map((render) => (
              <div key={render.id} className="bg-white border border-gray-200 rounded-lg p-4">
                <div className="flex items-center justify-between mb-3">
                  <div>
                    <span className="text-sm font-medium text-gray-900">{render.id}</span>
                  </div>
                  <a 
                    href={`/api/projects/${id}/renders/${render.id}/download.zip`}
                    className="inline-flex items-center px-3 py-1 text-xs font-medium text-blue-700 bg-blue-50 rounded-full hover:bg-blue-100 transition-colors duration-200"
                  >
                    <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3M3 17V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
                    </svg>
                    Download ZIP
                  </a>
                </div>
                
                {/* Preview Thumbnails */}
                {render.files && render.files.length > 0 && (
                  <div className="flex space-x-2 overflow-x-auto">
                    {render.files.map((file, index) => (
                      <div key={file} className="flex-shrink-0">
                        <a
                          href={`/api/projects/${id}/renders/${render.id}/files/${encodeURIComponent(file)}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="block border border-gray-200 rounded-lg overflow-hidden hover:border-blue-300 transition-colors duration-200"
                        >
                          <img
                            src={`/api/projects/${id}/renders/${render.id}/files/${encodeURIComponent(file)}`}
                            alt={`Render ${render.id} - Page ${index + 1}`}
                            className="w-20 h-20 object-contain bg-gray-50"
                          />
                        </a>
                      </div>
                    ))}
                  </div>
                )}
                
                {(!render.files || render.files.length === 0) && (
                  <div className="text-center py-4">
                    <svg className="w-8 h-8 text-gray-300 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    <p className="text-sm text-gray-500">No preview available</p>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div className="text-center py-8">
          <svg className="w-12 h-12 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <p className="text-gray-500 mb-2">No renders yet</p>
          <p className="text-sm text-gray-400">Generate your first render to see the output</p>
        </div>
      )}
    </div>
  );
};
