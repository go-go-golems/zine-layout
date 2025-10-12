import React from 'react';

interface AnchorGridProps {
  value?: string;
  onChange: (anchor: string) => void;
  className?: string;
}

const ANCHOR_POSITIONS = [
  { value: 'top-left', label: 'TL', row: 0, col: 0 },
  { value: 'top-center', label: 'TC', row: 0, col: 1 },
  { value: 'top-right', label: 'TR', row: 0, col: 2 },
  { value: 'middle-left', label: 'ML', row: 1, col: 0 },
  { value: 'middle-center', label: 'MC', row: 1, col: 1 },
  { value: 'middle-right', label: 'MR', row: 1, col: 2 },
  { value: 'bottom-left', label: 'BL', row: 2, col: 0 },
  { value: 'bottom-center', label: 'BC', row: 2, col: 1 },
  { value: 'bottom-right', label: 'BR', row: 2, col: 2 },
];

export const AnchorGrid: React.FC<AnchorGridProps> = ({
  value = 'middle-center',
  onChange,
  className = '',
}) => {
  return (
    <div className={className}>
      <label className="block text-sm font-medium text-gray-700 mb-2">Anchor Point</label>
      <div className="grid grid-cols-3 gap-2 w-fit">
        {ANCHOR_POSITIONS.map((anchor) => {
          const isSelected = value === anchor.value;
          return (
            <button
              key={anchor.value}
              type="button"
              onClick={() => onChange(anchor.value)}
              className={`w-12 h-12 flex items-center justify-center text-xs font-medium border-2 rounded transition-all ${
                isSelected
                  ? 'border-primary-600 bg-primary-100 text-primary-900 shadow-md'
                  : 'border-gray-300 bg-white text-gray-600 hover:border-primary-400 hover:bg-gray-50'
              }`}
              title={anchor.value.replace('-', ' ')}
            >
              {anchor.label}
            </button>
          );
        })}
      </div>
      <p className="text-xs text-gray-500 mt-2">
        Select where the image should be anchored within the canvas
      </p>
    </div>
  );
};

