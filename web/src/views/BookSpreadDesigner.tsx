import React from 'react';
import { ImageUploadSection } from '../components/bookSpread/ImageUploadSection';
import { PaperSettingsPanel } from '../components/bookSpread/PaperSettingsPanel';
import { MarginControlsPanel } from '../components/bookSpread/MarginControlsPanel';
import { ImageControlsPanel } from '../components/bookSpread/ImageControlsPanel';
import { PreviewCanvas } from '../components/bookSpread/PreviewCanvas';
import { ImageInformationPanel } from '../components/bookSpread/ImageInformationPanel';
import { ExportPanel } from '../components/bookSpread/ExportPanel';

export const BookSpreadDesigner: React.FC = () => {
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
                  <PreviewCanvas />
                </div>

                {/* Image Information */}
                <ImageInformationPanel />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
