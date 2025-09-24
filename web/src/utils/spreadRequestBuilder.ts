import { useMemo } from 'react';
import { useAppSelector } from '../hooks/redux';
import { CROP_RATIOS, type BookSpreadState } from '../store/bookSpreadSlice';
import { type ComputeRequest, type SpreadSettings } from '../api';

// Hook to build backend request parameters from Redux state
export const useSpreadRequest = (): ComputeRequest | null => {
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
    gutterMargin
  } = state;

  return useMemo(() => {
    if (!image?.uploadedPath) return null;

    // Convert crop ratio to numeric value for backend
    let cropRatioValue: number | undefined;
    if (cropRatio && cropRatio !== 'original') {
      const ratio = CROP_RATIOS[cropRatio];
      if (ratio) {
        cropRatioValue = ratio;
      }
    }

    return {
      image_path: `/home/manuel/workspaces/2025-09-23/book-spread-generator/zine-layout/data${image.uploadedPath}`,
      name: 'preview',
      settings: buildSpreadSettingsFromState(state, cropRatioValue),
    };
  }, [
    image?.uploadedPath, paperWidthIn, paperHeightIn, orientation, isSpread,
    margins.top, margins.right, margins.bottom, margins.left,
    cropRatio, cropToFill, imageScale, imagePosition.x, imagePosition.y,
    gutterMargin, state.dpi
  ]);
};

// Hook to build preview request with additional preview options
export const usePreviewRequest = (maxDimension = 600, panel = 'combined'): any => {
  const baseRequest = useSpreadRequest();
  
  return useMemo(() => {
    if (!baseRequest) return null;
    
    return {
      ...baseRequest,
      max_dimension: maxDimension,
      preview_format: 'jpg',
      panel: panel,
    };
  }, [baseRequest, maxDimension, panel]);
};

// Helper to suggest best crop ratio based on image aspect ratio
export const suggestCropRatio = (imageWidth: number, imageHeight: number): keyof typeof CROP_RATIOS => {
  const imageRatio = imageWidth / imageHeight;
  
  // Find the closest ratio
  let closestRatio: keyof typeof CROP_RATIOS = 'original';
  let closestDistance = Infinity;
  
  for (const [key, value] of Object.entries(CROP_RATIOS)) {
    if (value === null) continue; // Skip 'original'
    
    const distance = Math.abs(imageRatio - value);
    if (distance < closestDistance) {
      closestDistance = distance;
      closestRatio = key as keyof typeof CROP_RATIOS;
    }
  }
  
  return closestRatio;
};

export const buildSpreadSettingsFromState = (
  state: BookSpreadState,
  overrideCropRatio?: number
): SpreadSettings => {
  const { paperWidthIn, paperHeightIn, orientation, margins, isSpread, gutterMargin, cropRatio, cropToFill, imageScale, imagePosition, dpi } = state;

  let cropRatioValue: number | null = null;
  if (overrideCropRatio !== undefined) {
    cropRatioValue = overrideCropRatio ?? null;
  } else if (cropRatio !== 'original') {
    const value = CROP_RATIOS[cropRatio];
    cropRatioValue = value ?? null;
  }

  return {
    paper_width_in: paperWidthIn,
    paper_height_in: paperHeightIn,
    dpi,
    orientation,
    margin_top_in: margins.top,
    margin_right_in: margins.right,
    margin_bottom_in: margins.bottom,
    margin_left_in: margins.left,
    is_spread: isSpread,
    gutter_in: gutterMargin,
    crop_ratio: cropRatioValue === null ? undefined : cropRatioValue,
    crop_to_fill: cropToFill,
    user_scale: imageScale,
    position_x: imagePosition.x,
    position_y: imagePosition.y,
    units: 'px',
    export: {
      format: 'png',
      quality: 90,
      background: '#ffffff',
      out_dir: './out',
      filename_template: '{name}-{panel}.{ext}',
    },
  };
};

// Helper to categorize ratios
export const getRatioCategory = (ratio: keyof typeof CROP_RATIOS): 'vertical' | 'square' | 'horizontal' => {
  const value = CROP_RATIOS[ratio];
  if (value === null) return 'square'; // original
  if (value < 1) return 'vertical';
  if (value === 1) return 'square';
  return 'horizontal';
};
