import React from 'react';

interface SliderInputProps {
  label: string;
  value: number;
  onChange: (value: number) => void;
  min: number;
  max: number;
  step?: number;
  unit?: string;
  className?: string;
}

export const SliderInput: React.FC<SliderInputProps> = ({
  label,
  value,
  onChange,
  min,
  max,
  step = 0.1,
  unit = '',
  className = '',
}) => {
  return (
    <div className={className}>
      <div className="flex items-center justify-between mb-2">
        <label className="text-sm font-medium text-gray-700">{label}</label>
        <div className="flex items-center space-x-2">
          <input
            type="number"
            value={value}
            onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
            min={min}
            max={max}
            step={step}
            className="w-20 px-2 py-1 text-sm border border-gray-300 rounded focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
          />
          {unit && <span className="text-sm text-gray-500">{unit}</span>}
        </div>
      </div>
      <div className="flex items-center space-x-3">
        <span className="text-xs text-gray-500">{min}</span>
        <input
          type="range"
          value={value}
          onChange={(e) => onChange(parseFloat(e.target.value))}
          min={min}
          max={max}
          step={step}
          className="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-primary-600"
        />
        <span className="text-xs text-gray-500">{max}</span>
      </div>
    </div>
  );
};

