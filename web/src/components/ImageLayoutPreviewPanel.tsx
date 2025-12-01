import React, { useEffect, useRef } from "react";
import type { Asset, ImageLayoutComputation, ImageLayoutRequest } from "../api";
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
  layout: ImageLayoutRequest;
  previewPayload: {
    layout: ImageLayoutRequest;
    asset_id?: string;
    asset: { id: string; filename: string; width: number; height: number } | null;
  };
  previewState: PreviewState;
  previewResult: ImageLayoutComputation | null;
  previewCanvasScale: number;
  previewPlacement: PreviewPlacement | null;
  canvasRect: { w: number; h: number; x: number; y: number } | null;
  margins?: { top: number; right: number; bottom: number; left: number };
  dpi?: number;
  onRender: () => void;
  renderUrl: string | null;
  renderError: string | null;
  isRenderLoading: boolean;
  onOpenCompare: () => void;
}

export const ImageLayoutPreviewPanel: React.FC<ImageLayoutPreviewPanelProps> = ({
  projectId,
  assets,
  previewAssetId,
  onPreviewAssetChange,
  layout,
  previewPayload,
  previewState,
  previewResult,
  previewCanvasScale,
  previewPlacement,
  canvasRect,
  margins,
  dpi,
  onRender,
  renderUrl,
  renderError,
  isRenderLoading,
  onOpenCompare,
}) => {
  const previewAsset = assets.find((a) => a.id === previewAssetId);
  const previewTarget = previewResult?.result?.target_rect;
  const previewSource = previewResult?.result?.source_rect;
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [showAsset, setShowAsset] = React.useState(true);

  useEffect(() => {
    if (!previewAsset || !canvasRect || !previewTarget || !previewSource) return;
    const canvas = canvasRef.current;
    if (!canvas) return;
    const scale = previewCanvasScale;
    const width = canvasRect.w * scale;
    const height = canvasRect.h * scale;
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    ctx.clearRect(0, 0, width, height);
    ctx.fillStyle = "#f8fafc";
    ctx.fillRect(0, 0, width, height);

    if (margins && dpi && canvasRect) {
      const { top, right, bottom, left } = margins;
      const mTop = top * dpi * scale;
      const mRight = right * dpi * scale;
      const mBottom = bottom * dpi * scale;
      const mLeft = left * dpi * scale;
      ctx.fillStyle = "rgba(255, 182, 193, 0.35)";
      ctx.fillRect(0, 0, width, mTop);
      ctx.fillRect(0, 0, mLeft, height);
      ctx.fillRect(0, height - mBottom, width, mBottom);
      ctx.fillRect(width - mRight, 0, mRight, height);
    }

    if (showAsset) {
      const img = new Image();
      img.crossOrigin = "anonymous";
      img.onload = () => {
        ctx.drawImage(
          img,
          previewSource.x,
          previewSource.y,
          previewSource.w,
          previewSource.h,
          previewTarget.x * scale,
          previewTarget.y * scale,
          previewTarget.w * scale,
          previewTarget.h * scale,
        );
        ctx.strokeStyle = "rgba(59,130,246,0.8)";
        ctx.lineWidth = 2;
        ctx.strokeRect(
          previewTarget.x * scale,
          previewTarget.y * scale,
          previewTarget.w * scale,
          previewTarget.h * scale,
        );
        ctx.strokeStyle = "rgba(148,163,184,0.8)";
        ctx.setLineDash([4, 4]);
        ctx.strokeRect(0, 0, width, height);
        ctx.setLineDash([]);
      };
      img.src =
        previewAsset.url ?? `/projects/${projectId}/images/${previewAsset.filename}`;
    } else {
      ctx.fillStyle = "rgba(52, 211, 153, 0.25)";
      ctx.fillRect(
        previewTarget.x * scale,
        previewTarget.y * scale,
        previewTarget.w * scale,
        previewTarget.h * scale,
      );
      ctx.strokeStyle = "rgba(52, 211, 153, 0.5)";
      ctx.lineWidth = 2;
      ctx.strokeRect(
        previewTarget.x * scale,
        previewTarget.y * scale,
        previewTarget.w * scale,
        previewTarget.h * scale,
      );
      ctx.strokeStyle = "rgba(148,163,184,0.8)";
      ctx.setLineDash([4, 4]);
      ctx.strokeRect(0, 0, width, height);
      ctx.setLineDash([]);
    }
  }, [
    previewAsset,
    canvasRect,
    previewTarget,
    previewSource,
    previewCanvasScale,
    projectId,
    showAsset,
    margins,
    dpi,
  ]);

  return (
    <div className="space-y-6">
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
              {renderError && (
                <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                  {renderError}
                </div>
              )}

              <div className="flex flex-1 items-center justify-center">
                {previewResult && canvasRect ? (
                  <canvas
                    ref={canvasRef}
                    className="bg-white border-2 border-gray-400 shadow-lg"
                    style={{
                      width: `${canvasRect.w * previewCanvasScale}px`,
                      height: `${canvasRect.h * previewCanvasScale}px`,
                    }}
                  />
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
                    disabled={isRenderLoading}
                  >
                    {isRenderLoading ? "Rendering..." : "Render backend image"}
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={onOpenCompare}
                    disabled={!renderUrl || !previewResult}
                  >
                    Open Compare
                  </Button>
                </div>
                {renderUrl && (
                  <div className="flex items-center space-x-2 text-xs text-gray-600">
                    <span>Rendered:</span>
                    <img
                      src={renderUrl}
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

      <div className="space-y-2 text-xs text-gray-700">
        <details className="rounded-md border border-gray-200 bg-white px-3 py-2">
          <summary className="cursor-pointer font-semibold text-gray-800">Preview request payload</summary>
          <pre className="mt-2 whitespace-pre-wrap break-words text-[11px] text-gray-700">
            {JSON.stringify(previewPayload, null, 2)}
          </pre>
        </details>

        <div className="flex items-center space-x-3">
          <label className="flex items-center space-x-2 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={showAsset}
              onChange={(e) => setShowAsset(e.target.checked)}
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <span>Show asset in preview</span>
          </label>
          <span className="text-xs text-gray-500">
            Toggle to view just margins (pink) and target (green).
          </span>
        </div>

        {previewResult && (
          <details className="rounded-md border border-gray-200 bg-white px-3 py-2">
            <summary className="cursor-pointer font-semibold text-gray-800">Backend response</summary>
            <pre className="mt-2 whitespace-pre-wrap break-words text-[11px] text-gray-700">
              {JSON.stringify(
                {
                  result: previewResult.result,
                  trace: previewResult.trace ?? null,
                },
                null,
                2,
              )}
            </pre>
          </details>
        )}
      </div>
    </div>
  );
};
