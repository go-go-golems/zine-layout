import { PAPER_SIZES, CROP_RATIOS, BookSpreadState } from '../store/bookSpreadSlice';

export interface Dimensions {
  width: number;
  height: number;
}

export const getCurrentDimensions = (
  paperSize: keyof typeof PAPER_SIZES,
  orientation: 'portrait' | 'landscape',
  isSpread: boolean
): Dimensions => {
  const base = PAPER_SIZES[paperSize];
  let width = base.width;
  let height = base.height;

  // Swap dimensions for landscape
  if (orientation === 'landscape') {
    [width, height] = [height, width];
  }

  // Apply spread multiplier
  if (isSpread) {
    width = width * 2;
  }

  return { width, height };
};

export const loadImageFromFile = (file: File): Promise<HTMLImageElement> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        resolve(img);
      };
      img.onerror = reject;
      img.src = e.target?.result as string;
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
};

export const calculateImageDimensions = (
  image: HTMLImageElement,
  contentWidth: number,
  contentHeight: number,
  cropRatio: keyof typeof CROP_RATIOS,
  cropToFill: boolean,
  imagePosition: { x: number; y: number }
) => {
  let imgWidth = image.width;
  let imgHeight = image.height;
  let sourceX = 0;
  let sourceY = 0;
  let cropAdjustmentX = 0;
  let cropAdjustmentY = 0;

  if (cropRatio !== 'original' && CROP_RATIOS[cropRatio]) {
    const targetRatio = CROP_RATIOS[cropRatio]!;
    const currentRatio = imgWidth / imgHeight;

    if (currentRatio > targetRatio) {
      // Image is wider than target ratio - crop width
      const newWidth = imgHeight * targetRatio;
      sourceX = (imgWidth - newWidth) / 2;
      imgWidth = newWidth;
    } else {
      // Image is taller than target ratio - crop height
      const newHeight = imgWidth / targetRatio;
      sourceY = (imgHeight - newHeight) / 2;
      imgHeight = newHeight;
    }
  } else if (cropToFill) {
    // When crop to fill is enabled, crop the image to match content area ratio
    const contentRatio = contentWidth / contentHeight;
    const currentRatio = imgWidth / imgHeight;

    if (currentRatio > contentRatio) {
      // Image is wider than content area - crop width, allow horizontal adjustment
      const newWidth = imgHeight * contentRatio;
      const maxCropAdjustment = (imgWidth - newWidth) / 2;
      cropAdjustmentX = (imagePosition.x / 100) * maxCropAdjustment;
      sourceX = (imgWidth - newWidth) / 2 + cropAdjustmentX;
      imgWidth = newWidth;
    } else {
      // Image is taller than content area - crop height, allow vertical adjustment
      const newHeight = imgWidth / contentRatio;
      const maxCropAdjustment = (imgHeight - newHeight) / 2;
      cropAdjustmentY = (imagePosition.y / 100) * maxCropAdjustment;
      sourceY = (imgHeight - newHeight) / 2 + cropAdjustmentY;
      imgHeight = newHeight;
    }
  }

  return {
    imgWidth,
    imgHeight,
    sourceX,
    sourceY,
    cropAdjustmentX,
    cropAdjustmentY,
  };
};

export const calculateImageScale = (
  imgWidth: number,
  imgHeight: number,
  contentWidth: number,
  contentHeight: number,
  cropToFill: boolean,
  imageScale: number
) => {
  const scaleX = contentWidth / imgWidth;
  const scaleY = contentHeight / imgHeight;
  let baseScale: number;
  let finalScale: number;

  if (cropToFill) {
    // Scale to fill - use the larger scale to fill the entire content area
    baseScale = Math.max(scaleX, scaleY);
    finalScale = baseScale; // Ignore user scale when crop to fill is enabled
  } else {
    // Scale to fit - use the smaller scale to fit within content area
    baseScale = Math.min(scaleX, scaleY);
    finalScale = baseScale * imageScale;
  }

  return {
    baseScale,
    finalScale,
    displayWidth: imgWidth * finalScale,
    displayHeight: imgHeight * finalScale,
  };
};

export const shouldEnablePositionControl = (
  image: HTMLImageElement | null,
  cropToFill: boolean,
  contentWidth: number,
  contentHeight: number
): { enableX: boolean; enableY: boolean } => {
  if (!image || !cropToFill) {
    return { enableX: true, enableY: true };
  }

  const contentRatio = contentWidth / contentHeight;
  const imageRatio = image.width / image.height;

  if (imageRatio > contentRatio) {
    // Image is wider - we cropped horizontally, allow horizontal adjustment
    return { enableX: true, enableY: false };
  } else {
    // Image is taller - we cropped vertically, allow vertical adjustment
    return { enableX: false, enableY: true };
  }
};
