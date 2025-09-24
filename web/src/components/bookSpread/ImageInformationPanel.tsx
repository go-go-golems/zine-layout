import React from 'react';
import { useAppSelector } from '../../hooks/redux';
import { getCurrentDimensions } from '../../utils/bookSpreadUtils';

export const ImageInformationPanel: React.FC = () => {
  const { image, paperSize, paperWidthIn, paperHeightIn, orientation, isSpread, dpi, gutterMargin } = useAppSelector(
    (state) => state.bookSpread
  );

  if (!image) return null;

  const { width, height } = getCurrentDimensions(paperWidthIn, paperHeightIn, orientation, isSpread);
  const hasGutter = isSpread && gutterMargin > 0;
  const numImages = hasGutter ? 2 : 1;
  
  // For individual image dimensions when there's a gutter
  const individualWidth = hasGutter ? (width - gutterMargin) / 2 : width;
  const individualPixelWidth = individualWidth * dpi;
  const pixelHeight = height * dpi;
  const individualFileSize = Math.round((individualPixelWidth * pixelHeight * 3) / (1024 * 1024));
  
  // Total dimensions and file size
  const totalPixelWidth = width * dpi;
  const totalFileSize = Math.round((totalPixelWidth * pixelHeight * 3) / (1024 * 1024));

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
          <p>Layout: {width}" × {height}" ({isSpread ? 'spread' : 'single page'})</p>
          {hasGutter ? (
            <>
              <p>Export type: {numImages} split images (left & right halves)</p>
              <p>Each image: {individualWidth.toFixed(1)}" × {height}" ({individualPixelWidth} × {pixelHeight} pixels)</p>
              <p>Gutter margin: {gutterMargin}" (content lost in binding)</p>
              <p>Individual file size: ~{individualFileSize} MB each</p>
              <p>Combined download size: ~{individualFileSize * 2} MB</p>
            </>
          ) : (
            <>
              <p>Resolution: {totalPixelWidth} × {pixelHeight} pixels at {dpi} DPI</p>
              <p>Estimated file size: ~{totalFileSize} MB</p>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
