import React from 'react';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { 
  setCropRatio, 
  setCropToFill, 
  setImageScale, 
  setImagePosition,
  CROP_RATIOS 
} from '../../store/bookSpreadSlice';
import { getCurrentDimensions, shouldEnablePositionControl } from '../../utils/bookSpreadUtils';

export const ImageControlsPanel: React.FC = () => {
  const dispatch = useAppDispatch();
  const { 
    image, 
    cropRatio, 
    cropToFill, 
    imageScale, 
    imagePosition,
    paperSize,
    orientation,
    isSpread,
    margins,
    gutterMargin
  } = useAppSelector((state) => state.bookSpread);

  if (!image) return null;

  // Calculate current dimensions and content area
  const { width, height } = getCurrentDimensions(paperSize, orientation, isSpread);
  const totalMarginWidth = margins.left + margins.right;
  const totalMarginHeight = margins.top + margins.bottom;
  const baseContentWidth = width - totalMarginWidth;
  const baseContentHeight = height - totalMarginHeight;

  // Account for gutter in spread calculations
  const effectiveContentWidth = (isSpread && gutterMargin > 0) 
    ? baseContentWidth - gutterMargin 
    : baseContentWidth;

  // Create mock image object for shouldEnablePositionControl
  const mockImage = image ? {
    width: image.width,
    height: image.height
  } as HTMLImageElement : null;

  const { enableX, enableY } = shouldEnablePositionControl(
    mockImage,
    cropToFill,
    effectiveContentWidth,
    baseContentHeight
  );

  return (
    <div className="bg-gray-50 p-4 rounded-lg">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Crop Ratio</label>
          <select
            value={cropRatio}
            onChange={(e) => dispatch(setCropRatio(e.target.value as keyof typeof CROP_RATIOS))}
            className="w-full p-2 border border-gray-300 rounded-md text-sm"
          >
            {Object.keys(CROP_RATIOS).map(ratio => (
              <option key={ratio} value={ratio}>{ratio}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Fill Mode</label>
          <div className="flex items-center h-10">
            <input
              type="checkbox"
              id="cropToFill"
              checked={cropToFill}
              onChange={(e) => dispatch(setCropToFill(e.target.checked))}
              className="mr-2"
            />
            <label htmlFor="cropToFill" className="text-sm font-medium text-gray-700">
              Crop to fill
            </label>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Scale ({Math.round(imageScale * 100)}%)
          </label>
          <input
            type="range"
            min="0.1"
            max="3"
            step="0.01"
            value={imageScale}
            onChange={(e) => dispatch(setImageScale(Number(e.target.value)))}
            className="w-full"
            disabled={cropToFill}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Position {cropToFill && '(crop adjust)'}
          </label>
          <div className="flex gap-2">
            <input
              type="range"
              min="-100"
              max="100"
              value={imagePosition.x}
              onChange={(e) => dispatch(setImagePosition({ x: Number(e.target.value) }))}
              className={`w-full ${!enableX ? 'opacity-30' : ''}`}
              title="Horizontal"
              disabled={!enableX && cropToFill}
            />
            <input
              type="range"
              min="-100"
              max="100"
              value={imagePosition.y}
              onChange={(e) => dispatch(setImagePosition({ y: Number(e.target.value) }))}
              className={`w-full ${!enableY ? 'opacity-30' : ''}`}
              title="Vertical"
              disabled={!enableY && cropToFill}
            />
          </div>
        </div>
      </div>
    </div>
  );
};
