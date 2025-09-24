import React from 'react';
import { useAppSelector } from '../../hooks/redux';
import { 
  getCurrentDimensions, 
  calculateImageDimensions, 
  calculateImageScale 
} from '../../utils/bookSpreadUtils';

export const PreviewCanvas: React.FC = () => {
  const state = useAppSelector((state) => state.bookSpread);
  const { 
    image, 
    paperSize, 
    orientation, 
    isSpread, 
    margins, 
    cropRatio, 
    cropToFill, 
    imageScale, 
    imagePosition,
    gutterMargin
  } = state;

  if (!image) {
    return (
      <div className="text-center text-gray-500 py-12">
        <div className="text-6xl mb-4">🖼️</div>
        <p>Upload an image to start designing your photobook spread</p>
      </div>
    );
  }

  const { width, height } = getCurrentDimensions(paperSize, orientation, isSpread);
  const previewWidth = Math.min(600, width * 50); // Scale for preview
  const previewHeight = (height / width) * previewWidth;

  // Calculate content area after margins
  const contentWidth = previewWidth * (1 - (margins.left + margins.right) / width);
  const contentHeight = previewHeight * (1 - (margins.top + margins.bottom) / height);
  const contentX = previewWidth * (margins.left / width);
  const contentY = previewHeight * (margins.top / height);

  // Calculate gutter split for spreads
  const hasGutter = isSpread && gutterMargin > 0;
  const gutterWidth = hasGutter ? previewWidth * (gutterMargin / width) : 0;
  const leftPanelWidth = hasGutter ? (contentWidth - gutterWidth) / 2 : contentWidth;
  const rightPanelWidth = hasGutter ? (contentWidth - gutterWidth) / 2 : 0;
  const rightPanelX = hasGutter ? contentX + leftPanelWidth + gutterWidth : 0;

  // Create mock image object for calculations
  const mockImage = {
    width: image.width,
    height: image.height,
    src: image.src
  } as HTMLImageElement;

  // Calculate image dimensions based on crop ratio and crop-to-fill mode
  const { imgWidth, imgHeight, sourceX, sourceY } = calculateImageDimensions(
    mockImage,
    contentWidth,
    contentHeight,
    cropRatio,
    cropToFill,
    imagePosition
  );

  // Scale image to fit content area
  const { finalScale, displayWidth, displayHeight } = calculateImageScale(
    imgWidth,
    imgHeight,
    contentWidth,
    contentHeight,
    cropToFill,
    imageScale
  );

  return (
    <div className="border-2 border-gray-300 relative bg-white"
      style={{ width: previewWidth, height: previewHeight }}>
      {/* Margin guides */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute border border-red-200"
          style={{
            left: contentX,
            top: contentY,
            width: contentWidth,
            height: contentHeight
          }} />
        {/* Gutter guide for spreads */}
        {hasGutter && (
          <>
            <div className="absolute border-2 border-blue-300 bg-blue-100 bg-opacity-30"
              style={{
                left: contentX + leftPanelWidth,
                top: contentY,
                width: gutterWidth,
                height: contentHeight
              }} />
            <div className="absolute text-blue-600 text-xs font-semibold pointer-events-none"
              style={{
                left: contentX + leftPanelWidth + gutterWidth / 2 - 20,
                top: contentY + contentHeight / 2 - 8
              }}>
              GUTTER
            </div>
          </>
        )}
      </div>

      {/* Image - Left panel or full width */}
      <div className="absolute overflow-hidden"
        style={{
          left: contentX,
          top: contentY,
          width: hasGutter ? leftPanelWidth : contentWidth,
          height: contentHeight
        }}>
        <img
          src={image.src}
          className="absolute"
          style={{
            width: displayWidth,
            height: displayHeight,
            left: hasGutter ?
              (leftPanelWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x)) :
              (contentWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x)),
            top: contentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : imagePosition.y),
            objectFit: 'cover'
          }}
          alt="Preview"
        />
      </div>

      {/* Image - Right panel for gutter spreads */}
      {hasGutter && (
        <div className="absolute overflow-hidden"
          style={{
            left: rightPanelX,
            top: contentY,
            width: rightPanelWidth,
            height: contentHeight
          }}>
          <img
            src={image.src}
            className="absolute"
            style={{
              width: displayWidth,
              height: displayHeight,
              left: rightPanelWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x) - (leftPanelWidth + gutterWidth),
              top: contentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : imagePosition.y),
              objectFit: 'cover'
            }}
            alt="Preview"
          />
        </div>
      )}

      {/* Paper info overlay */}
      <div className="absolute top-2 right-2 bg-black bg-opacity-50 text-white text-xs px-2 py-1 rounded">
        {width}"×{height}" {orientation} {isSpread ? 'spread' : 'single'}
      </div>
    </div>
  );
};
