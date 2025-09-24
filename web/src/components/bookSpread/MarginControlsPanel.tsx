import React from 'react';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { setMargins } from '../../store/bookSpreadSlice';

export const MarginControlsPanel: React.FC = () => {
  const dispatch = useAppDispatch();
  const margins = useAppSelector((state) => state.bookSpread.margins);

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h3 className="text-lg font-semibold mb-4">📐 Margins (inches)</h3>

      <div className="grid grid-cols-2 gap-4">
        {Object.entries(margins).map(([key, value]) => (
          <div key={key}>
            <label className="block text-sm font-medium text-gray-700 mb-1 capitalize">
              {key}
            </label>
            <input
              type="number"
              step="0.1"
              min="0"
              max="2"
              value={value}
              onChange={(e) => dispatch(setMargins({ [key]: Number(e.target.value) }))}
              className="w-full p-2 border border-gray-300 rounded-md"
            />
          </div>
        ))}
      </div>
    </div>
  );
};
