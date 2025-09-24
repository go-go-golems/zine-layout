import React from 'react';
import { useAppSelector } from '../../hooks/redux';
import { getCurrentDimensions } from '../../utils/bookSpreadUtils';

export const ImageInformationPanel: React.FC = () => {
  const { image, paperSize, orientation, isSpread, dpi } = useAppSelector(
    (state) => state.bookSpread
  );

  if (!image) return null;

  const { width, height } = getCurrentDimensions(paperSize, orientation, isSpread);
  const pixelWidth = width * dpi;
  const pixelHeight = height * dpi;
  const fileSize = Math.round((pixelWidth * pixelHeight * 3) / (1024 * 1024));

  return (
    <div className="space-y-4">
      {/* Image Information */}
      <div className="bg-green-50 p-4 rounded-lg">
        <h4 className="font-medium text-green-900 mb-2">🖼️ Image Information</h4>
        <div className="text-sm text-green-800 space-y-1">
          <p>File: {image.fileName || 'Unknown'}</p>
          <p>Original size: {image.width} × {image.height} pixels</p>
          <p>File size: {image.fileSize ? `${(image.fileSize / (1024 * 1024)).toFixed(2)} MB` : 'Unknown'}</p>
          <p>Aspect ratio: {(image.width / image.height).toFixed(2)}:1</p>
        </div>
      </div>

      {/* Export Info */}
      <div className="bg-blue-50 p-4 rounded-lg">
        <h4 className="font-medium text-blue-900 mb-2">📋 Export Information</h4>
        <div className="text-sm text-blue-800 space-y-1">
          <p>Dimensions: {width}" × {height}" ({isSpread ? 'spread' : 'single page'})</p>
          <p>Resolution: {pixelWidth} × {pixelHeight} pixels at {dpi} DPI</p>
          <p>Estimated file size: ~{fileSize} MB</p>
        </div>
      </div>
    </div>
  );
};
