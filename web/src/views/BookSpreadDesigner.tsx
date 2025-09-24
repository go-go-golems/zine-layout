import React from 'react';
import { ImageUploadSection } from '../components/bookSpread/ImageUploadSection';
import { PaperSettingsPanel } from '../components/bookSpread/PaperSettingsPanel';
import { MarginControlsPanel } from '../components/bookSpread/MarginControlsPanel';
import { ImageControlsPanel } from '../components/bookSpread/ImageControlsPanel';
import { ImageInformationPanel } from '../components/bookSpread/ImageInformationPanel';
import { ExportPanel } from '../components/bookSpread/ExportPanel';
import { AlgorithmDebugPanel } from '../components/bookSpread/AlgorithmDebugPanel';
import { useAppSelector } from '../hooks/redux';
import { useGetPreviewSpreadQuery } from '../api';
import { usePreviewRequest } from '../utils/spreadRequestBuilder';

export const BookSpreadDesigner: React.FC = () => {
  const image = useAppSelector((state) => state.bookSpread.image);
  const isSpread = useAppSelector((state) => state.bookSpread.isSpread);
  
  // Get preview requests for different panels
  const leftRequest = usePreviewRequest(600, 'left');
  const rightRequest = usePreviewRequest(600, 'right');
  const singleRequest = usePreviewRequest(600, 'single');
  
  // Get preview images from backend
  const { data: leftPreviewUrl, isLoading: leftLoading, error: leftError } = useGetPreviewSpreadQuery(
    leftRequest!,
    { skip: !leftRequest || !isSpread }
  );
  
  const { data: rightPreviewUrl, isLoading: rightLoading, error: rightError } = useGetPreviewSpreadQuery(
    rightRequest!,
    { skip: !rightRequest || !isSpread }
  );
  
  const { data: singlePreviewUrl, isLoading: singleLoading, error: singleError } = useGetPreviewSpreadQuery(
    singleRequest!,
    { skip: !singleRequest || isSpread }
  );
  
  const isLoading = isSpread ? (leftLoading || rightLoading) : singleLoading;
  const error = isSpread ? (leftError || rightError) : singleError;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">📖 Photobook Spread Designer</h1>
        
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Controls Panel */}
          <div className="space-y-6">
            <ImageUploadSection />
            <PaperSettingsPanel />
            <MarginControlsPanel />
          </div>

          {/* Preview Panel */}
          <div className="lg:col-span-2">
            <div className="bg-white p-6 rounded-lg shadow">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold">🖼️ Preview</h3>
                <ExportPanel />
              </div>

              <div className="space-y-6">
                {/* Image Controls */}
                <ImageControlsPanel />

                {/* Preview */}
                <div className="flex justify-center">
                  {image ? (
                    <div className="text-center">
                      {isLoading && (
                        <div className="text-gray-500 py-12">
                          <div className="text-4xl mb-4">⏳</div>
                          <p>Generating preview...</p>
                        </div>
                      )}
                      {error && (
                        <div className="text-red-500 py-12">
                          <div className="text-4xl mb-4">❌</div>
                          <p>Error generating preview</p>
                        </div>
                      )}
                      {!isLoading && !error && (
                        <div className="preview-container">
                          {isSpread ? (
                            // Show both panels side by side for spreads
                            <div className="flex gap-2 items-start">
                              <div className="text-center">
                                <div className="text-xs text-gray-500 mb-1">Left Page</div>
                                {leftPreviewUrl ? (
                                  <img 
                                    src={leftPreviewUrl} 
                                    alt="Left Page Preview" 
                                    className="border border-gray-300 rounded shadow-lg"
                                    style={{ maxHeight: '500px' }}
                                  />
                                ) : (
                                  <div className="w-32 h-40 bg-gray-200 border border-gray-300 rounded flex items-center justify-center text-gray-400">
                                    ⏳
                                  </div>
                                )}
                              </div>
                              <div className="text-center">
                                <div className="text-xs text-gray-500 mb-1">Right Page</div>
                                {rightPreviewUrl ? (
                                  <img 
                                    src={rightPreviewUrl} 
                                    alt="Right Page Preview" 
                                    className="border border-gray-300 rounded shadow-lg"
                                    style={{ maxHeight: '500px' }}
                                  />
                                ) : (
                                  <div className="w-32 h-40 bg-gray-200 border border-gray-300 rounded flex items-center justify-center text-gray-400">
                                    ⏳
                                  </div>
                                )}
                              </div>
                            </div>
                          ) : (
                            // Show single page
                            singlePreviewUrl && (
                              <img 
                                src={singlePreviewUrl} 
                                alt="Single Page Preview" 
                                className="max-w-full h-auto border border-gray-300 rounded shadow-lg"
                                style={{ maxHeight: '600px' }}
                              />
                            )
                          )}
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="text-center text-gray-500 py-12">
                      <div className="text-6xl mb-4">🖼️</div>
                      <p>Upload an image to start designing your photobook spread</p>
                    </div>
                  )}
                </div>

                {/* Image Information */}
                <ImageInformationPanel />

                {/* Debug Panel */}
                <AlgorithmDebugPanel />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
