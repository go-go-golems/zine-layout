import { useAppSelector } from '../hooks/redux';
import { computePlacement, scaleResultForPreview, type Inputs } from './spreadAlgorithm';
import { CROP_RATIOS, PAPER_SIZES } from '../store/bookSpreadSlice';

export interface AlgorithmData {
  inputs: Inputs;
  result: any;
  scaledResult: any;
  previewScale: number;
  previewWidth: number;
  previewHeight: number;
}

// Hook to get algorithm results for debugging
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

  if (!image) return null;

  const paperDims = PAPER_SIZES[paperSize];
  
  // Convert crop ratio
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
  const previewWidth = Math.min(800, actualWidth * 0.2); // Larger preview
  const previewScale = previewWidth / actualWidth;
  const previewHeight = (paperDims.height * state.dpi) * previewScale;
  const scaledResult = scaleResultForPreview(result, previewScale);

  return {
    inputs,
    result,
    scaledResult,
    previewScale,
    previewWidth,
    previewHeight
  };
};
