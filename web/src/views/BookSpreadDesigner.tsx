import React, { CSSProperties } from 'react';
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
import { getCurrentDimensions } from '../utils/bookSpreadUtils';

const MAX_EDGE_SINGLE = 520;
const MAX_EDGE_PANEL = 420;

const computeFrameDimensions = (widthIn: number, heightIn: number, maxEdge: number): CSSProperties => {
  if (widthIn <= 0 || heightIn <= 0) {
    return { width: maxEdge, height: maxEdge * 1.2 };
  }
  if (widthIn >= heightIn) {
    const frameWidth = maxEdge;
    const frameHeight = Math.max(maxEdge * (heightIn / widthIn), maxEdge * 0.55);
    return { width: frameWidth, height: frameHeight };
  }
  const frameHeight = maxEdge;
  const frameWidth = Math.max(maxEdge * (widthIn / heightIn), maxEdge * 0.55);
  return { width: frameWidth, height: frameHeight };
};

const frameClassName = 'relative bg-white border-8 border-gray-200 rounded-3xl shadow-lg flex items-center justify-center overflow-hidden transition-all duration-300';

const renderFramedPreview = (label: string, url: string | undefined | null, frameStyle: CSSProperties) => (
  <div className="text-center">
    <div className="text-xs text-gray-500 mb-2">{label}</div>
    <div className={frameClassName} style={frameStyle}>
      {url ? (
        <img
          src={url}
          alt={`${label} preview`}
          className="w-full h-full object-contain"
        />
      ) : (
        <div className="flex flex-col items-center justify-center text-gray-400 gap-2">
          <span className="text-3xl">⏳</span>
          <span className="text-xs">Rendering…</span>
        </div>
      )}
    </div>
  </div>
);

export const BookSpreadDesigner: React.FC = () => {
  const { image, isSpread, paperSize, orientation } = useAppSelector((state) => state.bookSpread);
  
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

  const pageDimensions = getCurrentDimensions(paperSize, orientation, false);
  const singleFrameStyle = computeFrameDimensions(pageDimensions.width, pageDimensions.height, MAX_EDGE_SINGLE);
  const panelFrameStyle = computeFrameDimensions(pageDimensions.width, pageDimensions.height, MAX_EDGE_PANEL);

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
                            <div className="flex flex-wrap gap-6 items-start justify-center">
                              {renderFramedPreview('Left Page', leftPreviewUrl, panelFrameStyle)}
                              {renderFramedPreview('Right Page', rightPreviewUrl, panelFrameStyle)}
                            </div>
                          ) : (
                            renderFramedPreview('Single Page', singlePreviewUrl, singleFrameStyle)
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
