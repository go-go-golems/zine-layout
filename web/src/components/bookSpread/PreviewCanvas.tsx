import React from 'react';
import { useAppSelector } from '../../hooks/redux';
import { 
  getCurrentDimensions, 
  calculateImageDimensions, 
  calculateImageScale 
} from '../../utils/bookSpreadUtils';
import { CROP_RATIOS } from '../../store/bookSpreadSlice';

const DEBUG = true;

interface SinglePreviewProps {
  image: { src: string; width: number; height: number };
  dimensions: { width: number; height: number };
  previewDimensions: { width: number; height: number };
  contentArea: { x: number; y: number; width: number; height: number };
  imageSettings: {
    cropRatio: keyof typeof CROP_RATIOS;
    cropToFill: boolean;
    imageScale: number;
    imagePosition: { x: number; y: number };
  };
  side?: 'left' | 'right';
  panelWidth?: number;
  offsetX?: number;
}

const SinglePreview: React.FC<SinglePreviewProps> = ({
  image,
  dimensions,
  previewDimensions,
  contentArea,
  imageSettings,
  side,
  panelWidth,
  offsetX = 0
}) => {
  const { cropRatio, cropToFill, imageScale, imagePosition } = imageSettings;

  // Create mock image object for calculations
  const mockImage = {
    width: image.width,
    height: image.height,
    src: image.src
  } as HTMLImageElement;

  // Calculate image dimensions based on full content area (for proper scaling)
  const { imgWidth, imgHeight, sourceX, sourceY } = calculateImageDimensions(
    mockImage,
    contentArea.width,
    contentArea.height,
    cropRatio,
    cropToFill,
    imagePosition
  );

  // Scale image to fit the full content area
  const { displayWidth, displayHeight } = calculateImageScale(
    imgWidth,
    imgHeight,
    contentArea.width,
    contentArea.height,
    cropToFill,
    imageScale
  );

  // Calculate image position within this panel
  let imageX, imageY;
  
  if (side) {
    // For spread panels: position image to show correct portion
    // The image is scaled for the full spread, but each panel shows only its half
    const fullSpreadCenterX = contentArea.width / 2;
    const baseImageX = fullSpreadCenterX - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x);
    
    if (side === 'left') {
      // Left panel: show left portion of image, positioned normally
      imageX = baseImageX;
    } else {
      // Right panel: show right portion by shifting image left by the offset amount
      imageX = baseImageX - offsetX;
    }
    
    imageY = contentArea.height / 2 - displayHeight / 2 + (cropToFill ? 0 : imagePosition.y);
  } else {
    // Single panel: normal centering
    imageX = contentArea.width / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x);
    imageY = contentArea.height / 2 - displayHeight / 2 + (cropToFill ? 0 : imagePosition.y);
  }

  if (DEBUG && side) {
    console.log(`[SinglePreview ${side}] Debug info:`, {
      side,
      panelWidth,
      offsetX,
      contentArea,
      image: { imgWidth, imgHeight, sourceX, sourceY },
      display: { displayWidth, displayHeight },
      positioning: { 
        fullSpreadCenterX: side ? contentArea.width / 2 : 'N/A',
        baseImageX: side ? (contentArea.width / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x)) : 'N/A',
        finalImageX: imageX, 
        finalImageY: imageY 
      },
      calculations: {
        imageExceedsPanel: displayWidth > (panelWidth || contentArea.width),
        imagePanelRatio: displayWidth / (panelWidth || contentArea.width)
      },
      imageSettings
    });
  }

  return (
    <div className="absolute overflow-hidden"
      style={{
        left: contentArea.x,
        top: contentArea.y,
        width: panelWidth || contentArea.width,
        height: contentArea.height
      }}>
      <img
        src={image.src}
        className="absolute"
        style={{
          width: displayWidth,
          height: displayHeight,
          left: imageX,
          top: imageY,
          objectFit: 'cover'
        }}
        alt={`Preview ${side || 'Single'}`}
      />
    </div>
  );
};

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

  // Prepare image settings for SinglePreview components
  const imageSettings = {
    cropRatio,
    cropToFill,
    imageScale,
    imagePosition
  };

  const previewDimensions = { width: previewWidth, height: previewHeight };
  const dimensions = { width, height };

  if (DEBUG) {
    console.log('[PreviewCanvas] Debug info:', {
      paperSize, orientation, isSpread,
      dimensions: { width, height },
      preview: { previewWidth, previewHeight },
      content: { contentX, contentY, contentWidth, contentHeight },
      gutter: { hasGutter, gutterWidth, gutterMargin },
      panels: { leftPanelWidth, rightPanelWidth, rightPanelX },
      controls: { cropRatio, cropToFill, imageScale, imagePosition },
      margins
    });
  }

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

      {/* Image rendering */}
      {!isSpread || !hasGutter ? (
        /* Single page or spread without gutter */
        <SinglePreview
          image={image}
          dimensions={dimensions}
          previewDimensions={previewDimensions}
          contentArea={{
            x: contentX,
            y: contentY,
            width: contentWidth,
            height: contentHeight
          }}
          imageSettings={imageSettings}
        />
      ) : (
        /* Double-page spread with gutter - two separate preview canvases */
        <>
          {/* Left panel preview */}
          <SinglePreview
            image={image}
            dimensions={dimensions}
            previewDimensions={previewDimensions}
            contentArea={{
              x: contentX, // Absolute position for left panel
              y: contentY,
              width: contentWidth, // Full spread width for calculations
              height: contentHeight
            }}
            imageSettings={imageSettings}
            side="left"
            panelWidth={leftPanelWidth}
            offsetX={0}
          />

          {/* Right panel preview */}
          <SinglePreview
            image={image}
            dimensions={dimensions}
            previewDimensions={previewDimensions}
            contentArea={{
              x: rightPanelX, // Absolute position for right panel
              y: contentY,
              width: contentWidth, // Full spread width for calculations
              height: contentHeight
            }}
            imageSettings={imageSettings}
            side="right"
            panelWidth={rightPanelWidth}
            offsetX={leftPanelWidth + gutterWidth}
          />
        </>
      )}

      {/* Paper info overlay */}
      <div className="absolute top-2 right-2 bg-black bg-opacity-50 text-white text-xs px-2 py-1 rounded">
        {width}"×{height}" {orientation} {isSpread ? 'spread' : 'single'}
      </div>
    </div>
  );
};
