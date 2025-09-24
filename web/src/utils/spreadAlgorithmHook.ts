import { useEffect, useState } from 'react';
import { useAppSelector } from '../hooks/redux';
import { computePlacement, scaleResultForPreview, type Inputs } from './spreadAlgorithm';
import { CROP_RATIOS, PAPER_SIZES } from '../store/bookSpreadSlice';
import { useComputeSpreadMutation, type ComputeRequest, type SpreadSettings } from '../api';

export interface AlgorithmData {
  inputs: Inputs;
  result: any;
  scaledResult: any;
  previewScale: number;
  previewWidth: number;
  previewHeight: number;
  backendResult?: any; // Backend engine result
}

// Hook to get algorithm results using backend compute API
export const useSpreadAlgorithm = (): AlgorithmData | null => {
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

  const [computeSpread] = useComputeSpreadMutation();
  const [backendResult, setBackendResult] = useState<any>(null);

  if (!image) return null;

  const paperDims = PAPER_SIZES[paperSize];
  
  // Convert crop ratio to numeric value for backend
  let cropRatioValue: number | undefined;
  if (cropRatio && cropRatio !== 'original') {
    const ratio = CROP_RATIOS[cropRatio];
    if (ratio) {
      cropRatioValue = ratio;
    }
  }

  // Build backend request
  const backendRequest: ComputeRequest = {
    algorithm: 'sonnet',
    image_path: image.uploadedPath ? `/home/manuel/workspaces/2025-09-23/book-spread-generator/zine-layout/data${image.uploadedPath}` : undefined,
    meta: image.uploadedPath ? undefined : { width: image.width, height: image.height },
    name: 'preview',
    settings: {
      paper_width_in: paperDims.width,
      paper_height_in: paperDims.height,
      dpi: state.dpi,
      orientation: orientation,
      margin_top_in: margins.top,
      margin_right_in: margins.right,
      margin_bottom_in: margins.bottom,
      margin_left_in: margins.left,
      is_spread: isSpread,
      gutter_in: gutterMargin,
      crop_ratio: cropRatioValue,
      crop_to_fill: cropToFill,
      user_scale: imageScale,
      position_x: imagePosition.x,
      position_y: imagePosition.y,
      units: 'px',
      export: {
        format: 'png',
        quality: 90,
        background: 'transparent',
        out_dir: './out',
        filename_template: '{index:03d}-{name}-{panel}.{ext}',
      },
    },
  };

  // Call backend API when parameters change
  useEffect(() => {
    if (image.uploadedPath || (image.width && image.height)) {
      computeSpread(backendRequest)
        .unwrap()
        .then((result) => setBackendResult(result.result))
        .catch((error) => console.error('Backend compute failed:', error));
    }
  }, [
    image.uploadedPath, image.width, image.height,
    paperSize, orientation, isSpread, margins.top, margins.right, margins.bottom, margins.left,
    cropRatio, cropToFill, imageScale, imagePosition.x, imagePosition.y, gutterMargin, state.dpi
  ]);

  // Fallback to JavaScript implementation for legacy preview
  let cropRatioObj = null;
  if (cropRatio && cropRatio !== 'original') {
    const ratio = CROP_RATIOS[cropRatio];
    if (ratio) {
      if (cropRatio === '1:1') {
        cropRatioObj = { w: 1, h: 1 };
      } else if (cropRatio === '2:3') {
        cropRatioObj = { w: 2, h: 3 };
      } else if (cropRatio === '3:4') {
        cropRatioObj = { w: 3, h: 4 };
      } else if (cropRatio === '4:5') {
        cropRatioObj = { w: 4, h: 5 };
      } else if (cropRatio === '5:7') {
        cropRatioObj = { w: 5, h: 7 };
      } else if (cropRatio === '16:9') {
        cropRatioObj = { w: 16, h: 9 };
      }
    }
  }

  const inputs: Inputs = {
    srcW: image.width,
    srcH: image.height,
    paperWIn: paperDims.width,
    paperHIn: paperDims.height,
    orientation: orientation,
    marginTopIn: margins.top,
    marginRightIn: margins.right,
    marginBottomIn: margins.bottom,
    marginLeftIn: margins.left,
    dpi: state.dpi,
    isSpread: isSpread,
    gutterIn: gutterMargin,
    cropRatio: cropRatioObj,
    cropToFill: cropToFill,
    userScale: imageScale,
    imagePosition: { x: imagePosition.x, y: imagePosition.y },
    positionUnits: 'px'
  };

  const result = computePlacement(inputs);
  
  const actualWidth = isSpread ? 
    (paperDims.width * 2 * state.dpi) : 
    (paperDims.width * state.dpi);
  const previewWidth = Math.min(800, actualWidth * 0.2);
  const previewScale = previewWidth / actualWidth;
  const previewHeight = (paperDims.height * state.dpi) * previewScale;
  const scaledResult = scaleResultForPreview(result, previewScale);

  return {
    inputs,
    result,
    scaledResult,
    previewScale,
    previewWidth,
    previewHeight,
    backendResult
  };
};
