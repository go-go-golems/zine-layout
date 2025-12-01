import React from "react";
import { SliderInput } from "./SliderInput";
import { AnchorGrid } from "./AnchorGrid";
import { Input } from "./ui";

type Strategy = "auto" | "anchor" | "focus" | "manual";

interface Props {
  cropStrategy: Strategy;
  setCropStrategy: (value: Strategy) => void;
  anchorPreset: string;
  setAnchorPreset: (value: string) => void;
  panX: number;
  setPanX: (value: number) => void;
  panY: number;
  setPanY: (value: number) => void;
  zoom: number;
  setZoom: (value: number) => void;
  userScale: number;
  setUserScale: (value: number) => void;
  offsetX: number;
  setOffsetX: (value: number) => void;
  offsetY: number;
  setOffsetY: (value: number) => void;
  clampToCanvas: boolean;
  setClampToCanvas: (value: boolean) => void;
}

export const ImageLayoutCropControls: React.FC<Props> = ({
  cropStrategy,
  setCropStrategy,
  anchorPreset,
  setAnchorPreset,
  panX,
  setPanX,
  panY,
  setPanY,
  zoom,
  setZoom,
  userScale,
  setUserScale,
  offsetX,
  setOffsetX,
  offsetY,
  setOffsetY,
  clampToCanvas,
  setClampToCanvas,
}) => {
  return (
    <div className="space-y-4">
      <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
        Crop & Placement
      </h4>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Crop Strategy
        </label>
        <select
          value={cropStrategy}
          onChange={(e) =>
            setCropStrategy(
              e.target.value as "auto" | "anchor" | "focus" | "manual",
            )
          }
          className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
        >
          <option value="auto">Auto center</option>
          <option value="anchor">Anchor preset</option>
          <option value="manual">Manual pan</option>
        </select>
      </div>

      {cropStrategy === "anchor" && (
        <AnchorGrid value={anchorPreset} onChange={setAnchorPreset} />
      )}

      {cropStrategy === "manual" && (
        <div className="grid grid-cols-2 gap-3">
          <SliderInput
            label="Pan X"
            value={panX}
            onChange={setPanX}
            min={-1}
            max={1}
            step={0.01}
            unit="norm"
          />
          <SliderInput
            label="Pan Y"
            value={panY}
            onChange={setPanY}
            min={-1}
            max={1}
            step={0.01}
            unit="norm"
          />
        </div>
      )}

      <SliderInput
        label="Zoom"
        value={zoom}
        onChange={setZoom}
        min={0.5}
        max={2}
        step={0.01}
        unit="×"
      />

      <SliderInput
        label="User Scale"
        value={userScale}
        onChange={setUserScale}
        min={0.5}
        max={2}
        step={0.05}
        unit="×"
      />

      <div className="grid grid-cols-2 gap-3">
        <Input
          label="Offset X (px)"
          type="number"
          value={offsetX}
          onChange={(e) => setOffsetX(parseFloat(e.target.value) || 0)}
        />
        <Input
          label="Offset Y (px)"
          type="number"
          value={offsetY}
          onChange={(e) => setOffsetY(parseFloat(e.target.value) || 0)}
        />
      </div>

      <label className="flex items-center space-x-2 text-sm text-gray-700">
        <input
          type="checkbox"
          className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
          checked={clampToCanvas}
          onChange={(e) => setClampToCanvas(e.target.checked)}
        />
        <span>Clamp to canvas</span>
      </label>
    </div>
  );
};
