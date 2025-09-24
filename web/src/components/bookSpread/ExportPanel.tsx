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
    paperWidthIn,
    paperHeightIn,
    orientation,
    isSpread,
    margins,
    cropRatio,
    cropToFill,
    imageScale,
    imagePosition,
    dpi,
    gutterMargin,
  } = state;

  const exportImage = () => {
    if (!image) return;

    const { width, height } = getCurrentDimensions(paperWidthIn, paperHeightIn, orientation, isSpread);
    const hasGutter = isSpread && gutterMargin > 0;

    if (hasGutter) {
      // Export two separate images
      exportTwoSeparateImages(width, height);
    } else {
      // Export single image (normal behavior)
      exportSingleImage(width, height);
    }
  };

  const exportSingleImage = (width: number, height: number) => {
    if (!image || !image.src) return;
    const imageSrc = image.src;
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
    const contentWidth = pixelWidth - (margins.left + margins.right) * dpi;
    const contentHeight = pixelHeight - (margins.top + margins.bottom) * dpi;

    // Create mock image object for calculations
    const mockImage = {
      width: image.width,
      height: image.height
    } as HTMLImageElement;

    // Calculate image dimensions
    const { imgWidth, imgHeight, sourceX, sourceY } = calculateImageDimensions(
      mockImage,
      contentWidth,
      contentHeight,
      cropRatio,
      cropToFill,
      imagePosition
    );

    const { displayWidth, displayHeight } = calculateImageScale(
      imgWidth,
      imgHeight,
      contentWidth,
      contentHeight,
      cropToFill,
      imageScale
    );

    const imgElement = new Image();
    imgElement.onload = () => {
      ctx.save();
      ctx.rect(contentX, contentY, contentWidth, contentHeight);
      ctx.clip();

      const drawX = contentX + contentWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : (imagePosition.x * dpi / 50));
      const drawY = contentY + contentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : (imagePosition.y * dpi / 50));

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
    imgElement.src = imageSrc;
  };

  const exportTwoSeparateImages = (width: number, height: number) => {
    if (!image || !image.src) return;
    const imageSrc = image.src;
    // Calculate dimensions
    const fullPixelWidth = width * dpi;
    const pixelHeight = height * dpi;
    const individualWidth = (width - gutterMargin) / 2;
    const individualPixelWidth = individualWidth * dpi;
    
    // Calculate full content dimensions for image sizing
    const fullContentX = (margins.left * dpi);
    const fullContentY = (margins.top * dpi);
    const fullContentWidth = fullPixelWidth - (margins.left + margins.right) * dpi;
    const fullContentHeight = pixelHeight - (margins.top + margins.bottom) * dpi;
    
    // Calculate panel dimensions
    const gutterPixelWidth = gutterMargin * dpi;
    const leftPanelPixelWidth = (fullContentWidth - gutterPixelWidth) / 2;
    const rightPanelPixelWidth = leftPanelPixelWidth;
    
    // Individual image content area
    const contentX = (margins.left * dpi);
    const contentY = (margins.top * dpi);
    const individualContentWidth = individualPixelWidth - (margins.left + margins.right) * dpi;
    const individualContentHeight = pixelHeight - (margins.top + margins.bottom) * dpi;

    // Create mock image object for calculations (using full content area)
    const mockImage = {
      width: image.width,
      height: image.height
    } as HTMLImageElement;

    // Calculate image dimensions based on full content area (like in preview)
    const { imgWidth, imgHeight, sourceX, sourceY } = calculateImageDimensions(
      mockImage,
      fullContentWidth,
      fullContentHeight,
      cropRatio,
      cropToFill,
      imagePosition
    );

    const { displayWidth, displayHeight } = calculateImageScale(
      imgWidth,
      imgHeight,
      fullContentWidth,
      fullContentHeight,
      cropToFill,
      imageScale
    );

    console.log('[Export Split Images] Debug info:', {
      originalDimensions: { width, height },
      individualWidth, individualPixelWidth, pixelHeight,
      fullContentArea: { fullContentWidth, fullContentHeight },
      panelWidths: { leftPanelPixelWidth, rightPanelPixelWidth, gutterPixelWidth },
      imageCalculation: { imgWidth, imgHeight, sourceX, sourceY },
      display: { displayWidth, displayHeight },
      settings: { cropRatio, cropToFill, imageScale, imagePosition }
    });

    const createLeftImage = () => {
      const canvas = document.createElement('canvas');
      canvas.width = individualPixelWidth;
      canvas.height = pixelHeight;
      const ctx = canvas.getContext('2d');

      if (!ctx) return;

      // Fill with white background
      ctx.fillStyle = 'white';
      ctx.fillRect(0, 0, individualPixelWidth, pixelHeight);

      const imgElement = new Image();
      imgElement.onload = () => {
        ctx.save();
        ctx.rect(contentX, contentY, individualContentWidth, individualContentHeight);
        ctx.clip();

        // Position image as if it spans the full content area, but centered in individual canvas
        const drawX = contentX + fullContentWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : (imagePosition.x * dpi / 50));
        const drawY = contentY + fullContentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : (imagePosition.y * dpi / 50));

        if (cropRatio !== 'original' || cropToFill) {
          ctx.drawImage(imgElement, sourceX, sourceY, imgWidth, imgHeight, drawX, drawY, displayWidth, displayHeight);
        } else {
          ctx.drawImage(imgElement, drawX, drawY, displayWidth, displayHeight);
        }

        ctx.restore();

        // Download
        const link = document.createElement('a');
        link.download = `photobook-left-${individualWidth.toFixed(1)}x${height}-${dpi}dpi.png`;
        link.href = canvas.toDataURL('image/png');
        link.click();
      };
      imgElement.src = imageSrc;
    };

    const createRightImage = () => {
      const canvas = document.createElement('canvas');
      canvas.width = individualPixelWidth;
      canvas.height = pixelHeight;
      const ctx = canvas.getContext('2d');

      if (!ctx) return;

      // Fill with white background
      ctx.fillStyle = 'white';
      ctx.fillRect(0, 0, individualPixelWidth, pixelHeight);

      const imgElement = new Image();
      imgElement.onload = () => {
        ctx.save();
        ctx.rect(contentX, contentY, individualContentWidth, individualContentHeight);
        ctx.clip();

        // Position image to show right portion - shift left by (leftPanel + gutter)
        const baseDrawX = contentX + fullContentWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : (imagePosition.x * dpi / 50));
        const drawX = baseDrawX - (leftPanelPixelWidth + gutterPixelWidth);
        const drawY = contentY + fullContentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : (imagePosition.y * dpi / 50));

        if (cropRatio !== 'original' || cropToFill) {
          ctx.drawImage(imgElement, sourceX, sourceY, imgWidth, imgHeight, drawX, drawY, displayWidth, displayHeight);
        } else {
          ctx.drawImage(imgElement, drawX, drawY, displayWidth, displayHeight);
        }

        ctx.restore();

        // Download
        const link = document.createElement('a');
        link.download = `photobook-right-${individualWidth.toFixed(1)}x${height}-${dpi}dpi.png`;
        link.href = canvas.toDataURL('image/png');
        link.click();
      };
      imgElement.src = imageSrc;
    };

    // Create both images with slight delay
    setTimeout(createLeftImage, 100);
    setTimeout(createRightImage, 500);
  };

  if (!image) return null;

  return (
    <Button onClick={exportImage}>
      💾 Export Image
    </Button>
  );
};
