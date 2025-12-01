import React from "react";
import {
  ASPECT_RATIOS,
  PAPER_SIZES,
  type AspectRatioKey,
  type FrameState,
  type PaperSizeKey,
} from "../state/imageLayoutsEditorSlice";
import { Input } from "./ui/Input";
import { SliderInput } from "./SliderInput";

interface ImageLayoutFrameFormProps {
  frame: FrameState;
  onPaperSizeChange: (value: PaperSizeKey) => void;
  onPaperWidthChange: (value: number) => void;
  onPaperHeightChange: (value: number) => void;
  onDpiChange: (value: number) => void;
  onOrientationChange: (value: FrameState["orientation"]) => void;
  onUniformMarginsChange: (value: boolean) => void;
  onMarginAllChange: (value: number) => void;
  onMarginTopChange: (value: number) => void;
  onMarginRightChange: (value: number) => void;
  onMarginBottomChange: (value: number) => void;
  onMarginLeftChange: (value: number) => void;
  onAspectRatioChange: (value: AspectRatioKey) => void;
}

export const ImageLayoutFrameForm: React.FC<ImageLayoutFrameFormProps> = ({
  frame,
  onPaperSizeChange,
  onPaperWidthChange,
  onPaperHeightChange,
  onDpiChange,
  onOrientationChange,
  onUniformMarginsChange,
  onMarginAllChange,
  onMarginTopChange,
  onMarginRightChange,
  onMarginBottomChange,
  onMarginLeftChange,
  onAspectRatioChange,
}) => {
  return (
    <div className="space-y-6">
      {/* Page Setup */}
      <div className="space-y-4">
        <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
          Page Setup
        </h4>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Paper Size
          </label>
          <select
            value={frame.paperSize}
            onChange={(e) =>
              onPaperSizeChange(e.target.value as PaperSizeKey)
            }
            className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
          >
            {Object.keys(PAPER_SIZES).map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </div>

        {frame.paperSize === "Custom" && (
          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Width (inches)"
              type="number"
              value={frame.paperWidth}
              onChange={(e) =>
                onPaperWidthChange(parseFloat(e.target.value) || 8)
              }
              step="0.1"
              min="1"
            />
            <Input
              label="Height (inches)"
              type="number"
              value={frame.paperHeight}
              onChange={(e) =>
                onPaperHeightChange(parseFloat(e.target.value) || 10)
              }
              step="0.1"
              min="1"
            />
          </div>
        )}

        <SliderInput
          label="DPI"
          value={frame.dpi}
          onChange={onDpiChange}
          min={72}
          max={600}
          step={1}
        />

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Orientation
          </label>
          <div className="flex space-x-3">
            <button
              type="button"
              onClick={() => onOrientationChange("portrait")}
              className={`flex-1 py-2 px-4 rounded-md border-2 transition-all ${
                frame.orientation === "portrait"
                  ? "border-primary-600 bg-primary-50 text-primary-900"
                  : "border-gray-300 bg-white text-gray-700 hover:border-primary-400"
              }`}
            >
              Portrait ⬜
            </button>
            <button
              type="button"
              onClick={() => onOrientationChange("landscape")}
              className={`flex-1 py-2 px-4 rounded-md border-2 transition-all ${
                frame.orientation === "landscape"
                  ? "border-primary-600 bg-primary-50 text-primary-900"
                  : "border-gray-300 bg-white text-gray-700 hover:border-primary-400"
              }`}
            >
              Landscape ▭
            </button>
          </div>
        </div>
      </div>

      <hr className="border-gray-200" />

      {/* Margins */}
      <div className="space-y-4">
        <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
          Margins
        </h4>

        <div>
          <label className="flex items-center space-x-2 text-sm font-medium text-gray-700">
            <input
              type="checkbox"
              checked={frame.uniformMargins}
              onChange={(e) => onUniformMarginsChange(e.target.checked)}
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <span>Uniform margins</span>
          </label>
        </div>

        {frame.uniformMargins ? (
          <SliderInput
            label="All Margins"
            value={frame.marginAll}
            onChange={onMarginAllChange}
            min={0}
            max={2}
            step={0.05}
            unit="in"
          />
        ) : (
          <div className="grid grid-cols-2 gap-3">
            <SliderInput
              label="Top"
              value={frame.margins.top}
              onChange={onMarginTopChange}
              min={0}
              max={2}
              step={0.05}
              unit="in"
            />
            <SliderInput
              label="Right"
              value={frame.margins.right}
              onChange={onMarginRightChange}
              min={0}
              max={2}
              step={0.05}
              unit="in"
            />
            <SliderInput
              label="Bottom"
              value={frame.margins.bottom}
              onChange={onMarginBottomChange}
              min={0}
              max={2}
              step={0.05}
              unit="in"
            />
            <SliderInput
              label="Left"
              value={frame.margins.left}
              onChange={onMarginLeftChange}
              min={0}
              max={2}
              step={0.05}
              unit="in"
            />
          </div>
        )}
      </div>

      <hr className="border-gray-200" />

      {/* Frame & Crop */}
      <div className="space-y-4">
        <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
          Frame & Crop
        </h4>

        <p className="text-sm text-gray-600">
          Cropping is always cover-first: the image is cropped to the target aspect ratio so it fills the frame.
        </p>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Aspect Ratio
          </label>
          <select
            value={frame.aspectRatio}
            onChange={(e) => onAspectRatioChange(e.target.value as AspectRatioKey)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
          >
            {Object.keys(ASPECT_RATIOS).map((ratio) => (
              <option key={ratio} value={ratio}>
                {ratio}
              </option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
};
