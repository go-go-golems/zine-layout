import React from "react";
import type { Asset, ImageLayoutComputation } from "../api";
import type { PreviewState } from "../state/imageLayoutsEditorSlice";
import { Button } from "./ui/Button";

interface PreviewPlacement {
  targetLeft: number;
  targetTop: number;
  targetWidth: number;
  targetHeight: number;
  imgWidth: number;
  imgHeight: number;
  offsetX: number;
  offsetY: number;
}

interface ImageLayoutPreviewPanelProps {
  projectId: string;
  assets: Asset[];
  previewAssetId: string;
  onPreviewAssetChange: (id: string) => void;
  previewState: PreviewState;
  previewResult: ImageLayoutComputation | null;
  previewCanvasScale: number;
  previewPlacement: PreviewPlacement | null;
  onRender: () => void;
  onOpenCompare: () => void;
}

export const ImageLayoutPreviewPanel: React.FC<ImageLayoutPreviewPanelProps> = ({
  projectId,
  assets,
  previewAssetId,
  onPreviewAssetChange,
  previewState,
  previewResult,
  previewCanvasScale,
  previewPlacement,
  onRender,
  onOpenCompare,
}) => {
  const previewAsset = assets.find((a) => a.id === previewAssetId);
  const previewCanvas = previewResult?.result?.canvas_rect;
  const previewTarget = previewResult?.result?.target_rect;
  const previewSource = previewResult?.result?.source_rect;

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Preview Asset
        </label>
        <select
          value={previewAssetId}
          onChange={(e) => onPreviewAssetChange(e.target.value)}
          className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
        >
          <option value="">Select an asset to preview...</option>
          {assets.map((asset) => (
            <option key={asset.id} value={asset.id}>
              {asset.filename} ({asset.width}×{asset.height})
            </option>
          ))}
        </select>
      </div>

      <div className="border-2 border-gray-300 rounded-lg p-6 bg-gray-50 min-h-[500px] flex flex-col">
        {previewAsset ? (
          <div className="w-full space-y-3">
            <div className="flex items-center justify-between text-xs text-gray-500">
              <span>Backend preview updates as you edit.</span>
              {previewState.isLoading && (
                <span className="text-primary-600 animate-pulse">
                  Computing...
                </span>
              )}
            </div>

            {previewState.error && (
              <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {previewState.error}
              </div>
            )}
            {previewState.renderError && (
              <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {previewState.renderError}
              </div>
            )}

            <div className="flex flex-1 items-center justify-center">
              {previewResult &&
              previewCanvas &&
              previewTarget &&
              previewSource &&
              previewPlacement ? (
                <div
                  className="relative bg-white border-2 border-gray-400 shadow-lg"
                  style={{
                    width: `${previewCanvas.w * previewCanvasScale}px`,
                    height: `${previewCanvas.h * previewCanvasScale}px`,
                  }}
                >
                  <div className="absolute inset-0 bg-gray-50" />
                  <div
                    className="absolute border-2 border-primary-500/80 bg-primary-100/60 overflow-hidden shadow-inner"
                    style={{
                      left: `${previewPlacement.targetLeft}px`,
                      top: `${previewPlacement.targetTop}px`,
                      width: `${previewPlacement.targetWidth}px`,
                      height: `${previewPlacement.targetHeight}px`,
                    }}
                  >
                    <img
                      src={
                        previewAsset.url ??
                        `/projects/${projectId}/images/${previewAsset.filename}`
                      }
                      alt={previewAsset.filename}
                      className="pointer-events-none select-none"
                      style={{
                        position: "absolute",
                        width: `${previewPlacement.imgWidth}px`,
                        height: `${previewPlacement.imgHeight}px`,
                        left: `${previewPlacement.offsetX}px`,
                        top: `${previewPlacement.offsetY}px`,
                      }}
                    />
                    <div className="absolute inset-0 border border-white/70 pointer-events-none" />
                  </div>
                  <div className="absolute inset-0 border border-dashed border-gray-300 pointer-events-none" />
                </div>
              ) : (
                <div className="text-center text-gray-500 text-sm">
                  {previewState.isLoading
                    ? "Computing preview..."
                    : "Adjust settings to compute preview geometry."}
                </div>
              )}
            </div>

            <div className="space-y-1 text-sm text-gray-600 text-center">
              <p className="font-medium">{previewAsset.filename}</p>
              {previewResult && previewResult.result && (
                <div className="grid grid-cols-2 gap-1 text-xs text-gray-600">
                  <span>
                    Canvas: {Math.round(previewResult.result.canvas_rect.w)} ×{" "}
                    {Math.round(previewResult.result.canvas_rect.h)} px
                  </span>
                  <span>
                    Target: {Math.round(previewResult.result.target_rect.w)} ×{" "}
                    {Math.round(previewResult.result.target_rect.h)} px
                  </span>
                  <span>
                    Source: {Math.round(previewResult.result.source_rect.w)} ×{" "}
                    {Math.round(previewResult.result.source_rect.h)} px
                  </span>
                  <span>
                    Scale: {previewResult.result.scale.toFixed(3)}× (
                    {previewResult.result.mode})
                  </span>
                </div>
              )}
            </div>
            <div className="flex items-center justify-between pt-2">
              <div className="space-x-2">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={onRender}
                  disabled={previewState.isRenderLoading}
                >
                  {previewState.isRenderLoading ? "Rendering..." : "Render backend image"}
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={onOpenCompare}
                  disabled={!previewState.renderUrl || !previewResult}
                >
                  Open Compare
                </Button>
              </div>
              {previewState.renderUrl && (
                <div className="flex items-center space-x-2 text-xs text-gray-600">
                  <span>Rendered:</span>
                  <img
                    src={previewState.renderUrl}
                    alt="Rendered thumbnail"
                    className="h-12 w-12 object-contain border border-gray-200 rounded"
                  />
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="flex flex-1 flex-col items-center justify-center text-center text-gray-400">
            <div className="text-6xl mb-4">📐</div>
            <p className="text-sm">Select an asset to preview the template</p>
          </div>
        )}
      </div>
    </div>
  );
};
