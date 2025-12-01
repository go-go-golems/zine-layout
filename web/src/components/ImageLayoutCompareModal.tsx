import React from "react";
import { Button } from "./ui";
import { ImageLayoutViewportResult } from "../api";

interface Props {
  open: boolean;
  onClose: () => void;
  previewResult?: ImageLayoutViewportResult | null;
  previewAssetUrl?: string;
  previewPlacement?: {
    targetLeft: number;
    targetTop: number;
    targetWidth: number;
    targetHeight: number;
    imgWidth: number;
    imgHeight: number;
    offsetX: number;
    offsetY: number;
  } | null;
  previewCanvas?: { w: number; h: number } | null;
  previewScale: number;
  renderUrl?: string | null;
}

export const ImageLayoutCompareModal: React.FC<Props> = ({
  open,
  onClose,
  previewResult,
  previewAssetUrl,
  previewPlacement,
  previewCanvas,
  previewScale,
  renderUrl,
}) => {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-lg shadow-2xl max-w-6xl w-full p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Compare Preview vs Render</h3>
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
        </div>
        <div className="grid md:grid-cols-2 gap-4">
          <div className="border rounded-lg p-3 bg-gray-50">
            <p className="text-sm font-medium mb-2">Preview (geometry)</p>
            {previewResult && previewCanvas && previewPlacement ? (
              <div
                className="relative bg-white border border-gray-300"
                style={{
                  width: `${previewCanvas.w * previewScale}px`,
                  height: `${previewCanvas.h * previewScale}px`,
                }}
              >
                <div
                  className="absolute border-2 border-primary-500/80 bg-primary-100/60 overflow-hidden shadow-inner"
                  style={{
                    left: `${previewPlacement.targetLeft}px`,
                    top: `${previewPlacement.targetTop}px`,
                    width: `${previewPlacement.targetWidth}px`,
                    height: `${previewPlacement.targetHeight}px`,
                  }}
                >
                  {previewAssetUrl && (
                    <img
                      src={previewAssetUrl}
                      alt="Preview asset"
                      className="pointer-events-none select-none"
                      style={{
                        position: "absolute",
                        width: `${previewPlacement.imgWidth}px`,
                        height: `${previewPlacement.imgHeight}px`,
                        left: `${previewPlacement.offsetX}px`,
                        top: `${previewPlacement.offsetY}px`,
                      }}
                    />
                  )}
                </div>
                <div className="absolute inset-0 border border-dashed border-gray-300 pointer-events-none" />
              </div>
            ) : (
              <div className="text-sm text-gray-500">Preview not available.</div>
            )}
          </div>
          <div className="border rounded-lg p-3 bg-gray-50">
            <p className="text-sm font-medium mb-2">Render (backend)</p>
            {renderUrl ? (
              <img
                src={renderUrl}
                alt="Rendered"
                className="w-full h-auto border border-gray-300 rounded"
              />
            ) : (
              <div className="text-sm text-gray-500">Render not available.</div>
            )}
          </div>
        </div>
        <p className="text-xs text-gray-500">
          Compare alignment, scale, and clipping. If mismatched, check DPI, margins, and crop.
        </p>
      </div>
    </div>
  );
};
