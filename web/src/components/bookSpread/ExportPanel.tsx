import React from 'react';
import { useAppSelector } from '../../hooks/redux';
import { Button } from '../ui';
import { 
  getCurrentDimensions, 
  calculateImageDimensions, 
  calculateImageScale 
} from '../../utils/bookSpreadUtils';

export const ExportPanel: React.FC = () => {
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
    dpi,
    gutterMargin
  } = state;

  const exportImage = () => {
    if (!image) return;

    const { width, height } = getCurrentDimensions(paperSize, orientation, isSpread);
    const pixelWidth = width * dpi;
    const pixelHeight = height * dpi;

    const canvas = document.createElement('canvas');
    canvas.width = pixelWidth;
    canvas.height = pixelHeight;
    const ctx = canvas.getContext('2d');

    if (!ctx) return;

    // Fill with white background
    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, pixelWidth, pixelHeight);

    // Calculate content area
    const contentX = (margins.left * dpi);
    const contentY = (margins.top * dpi);
    const baseContentWidth = pixelWidth - (margins.left + margins.right) * dpi;
    const baseContentHeight = pixelHeight - (margins.top + margins.bottom) * dpi;

    // Account for gutter in spreads
    const hasGutter = isSpread && gutterMargin > 0;
    const gutterPixelWidth = hasGutter ? gutterMargin * dpi : 0;
    const effectiveContentWidth = hasGutter ? baseContentWidth - gutterPixelWidth : baseContentWidth;
    const leftPanelWidth = hasGutter ? effectiveContentWidth / 2 : baseContentWidth;
    const rightPanelWidth = hasGutter ? effectiveContentWidth / 2 : 0;
    const rightPanelX = hasGutter ? contentX + leftPanelWidth + gutterPixelWidth : 0;

    // Create mock image object for calculations
    const mockImage = {
      width: image.width,
      height: image.height
    } as HTMLImageElement;

    // Calculate image dimensions
    const { imgWidth, imgHeight, sourceX, sourceY } = calculateImageDimensions(
      mockImage,
      effectiveContentWidth,
      baseContentHeight,
      cropRatio,
      cropToFill,
      imagePosition
    );

    const { finalScale, displayWidth, displayHeight } = calculateImageScale(
      imgWidth,
      imgHeight,
      effectiveContentWidth,
      baseContentHeight,
      cropToFill,
      imageScale
    );

    const drawX = contentX + effectiveContentWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : (imagePosition.x * dpi / 50));
    const drawY = contentY + baseContentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : (imagePosition.y * dpi / 50));

    // Create image element for canvas drawing
    const imgElement = new Image();
    imgElement.onload = () => {
      // Clip to content area
      ctx.save();
      ctx.rect(contentX, contentY, effectiveContentWidth, baseContentHeight);
      ctx.clip();

      // Draw image with proper cropping
      if (cropRatio !== 'original' || cropToFill) {
        ctx.drawImage(imgElement, sourceX, sourceY, imgWidth, imgHeight, drawX, drawY, displayWidth, displayHeight);
      } else {
        ctx.drawImage(imgElement, drawX, drawY, displayWidth, displayHeight);
      }

      ctx.restore();

      // Download
      const link = document.createElement('a');
      link.download = `photobook-spread-${width}x${height}-${dpi}dpi.png`;
      link.href = canvas.toDataURL('image/png');
      link.click();
    };
    imgElement.src = image.src;
  };

  if (!image) return null;

  return (
    <Button onClick={exportImage}>
      💾 Export Image
    </Button>
  );
};
