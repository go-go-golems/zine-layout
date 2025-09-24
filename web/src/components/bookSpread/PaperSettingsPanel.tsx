import React from 'react';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { 
  setPaperSize, 
  setOrientation, 
  setIsSpread, 
  setDpi, 
  setGutterMargin,
  PAPER_SIZES 
} from '../../store/bookSpreadSlice';

export const PaperSettingsPanel: React.FC = () => {
  const dispatch = useAppDispatch();
  const { paperSize, orientation, isSpread, dpi, gutterMargin } = useAppSelector(
    (state) => state.bookSpread
  );

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h3 className="text-lg font-semibold mb-4">📏 Paper Settings</h3>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Paper Size</label>
          <select
            value={paperSize}
            onChange={(e) => dispatch(setPaperSize(e.target.value as keyof typeof PAPER_SIZES))}
            className="w-full p-2 border border-gray-300 rounded-md"
          >
            {Object.entries(PAPER_SIZES).map(([key, size]) => (
              <option key={key} value={key}>
                {key} ({size.width}" × {size.height}")
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Orientation</label>
          <select
            value={orientation}
            onChange={(e) => dispatch(setOrientation(e.target.value as 'portrait' | 'landscape'))}
            className="w-full p-2 border border-gray-300 rounded-md"
          >
            <option value="portrait">Portrait</option>
            <option value="landscape">Landscape</option>
          </select>
        </div>

        <div className="flex items-center">
          <input
            type="checkbox"
            id="spread"
            checked={isSpread}
            onChange={(e) => dispatch(setIsSpread(e.target.checked))}
            className="mr-2"
          />
          <label htmlFor="spread" className="text-sm font-medium text-gray-700">
            Double-page spread
          </label>
        </div>

        {isSpread && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Gutter/Spine Margin (inches)
            </label>
            <input
              type="number"
              step="0.1"
              min="0"
              max="2"
              value={gutterMargin}
              onChange={(e) => dispatch(setGutterMargin(Number(e.target.value)))}
              className="w-full p-2 border border-gray-300 rounded-md"
            />
            <p className="text-xs text-gray-500 mt-1">
              Space in the center where binding occurs
            </p>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Print DPI</label>
          <select
            value={dpi}
            onChange={(e) => dispatch(setDpi(Number(e.target.value)))}
            className="w-full p-2 border border-gray-300 rounded-md"
          >
            <option value={150}>150 DPI</option>
            <option value={300}>300 DPI</option>
            <option value={600}>600 DPI</option>
          </select>
        </div>
      </div>
    </div>
  );
};
