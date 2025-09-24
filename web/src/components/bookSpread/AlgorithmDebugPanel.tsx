import React from 'react';
import { AlgorithmData } from '../../utils/spreadAlgorithmHook';

interface AlgorithmDebugPanelProps {
  algorithmData: AlgorithmData | null;
}

export const AlgorithmDebugPanel: React.FC<AlgorithmDebugPanelProps> = ({ algorithmData }) => {

  if (!algorithmData) {
    return (
      <div className="mt-4 p-4 bg-gray-100 rounded-lg">
        <h4 className="font-semibold text-sm">Algorithm Debug</h4>
        <p className="text-xs text-gray-500">No image loaded</p>
      </div>
    );
  }

  const debugText = JSON.stringify(algorithmData, null, 2);

  const copyToClipboard = () => {
    navigator.clipboard.writeText(debugText).then(() => {
      alert('Debug data copied to clipboard!');
    }).catch(err => {
      console.error('Failed to copy to clipboard:', err);
      alert('Failed to copy to clipboard');
    });
  };

  const openRenderedImage = () => {
    // Create a canvas and render the exact algorithm output
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const { result, inputs } = algorithmData;
    
    // Create image element to get the source
    const img = new Image();
    img.crossOrigin = 'anonymous';
    
    img.onload = () => {
      if (result.exportSingle) {
        // Single page render
        canvas.width = result.exportSingle.w;
        canvas.height = result.exportSingle.h;
        
        // Fill white background
        ctx.fillStyle = 'white';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        // Draw image using algorithm rectangles
        const src = result.srcRectGlobal;
        const dst = result.dstRectGlobal;
        ctx.drawImage(
          img,
          src.x, src.y, src.w, src.h,
          dst.x, dst.y, dst.w, dst.h
        );
      } else if (result.exportSpread) {
        // Spread render - for now just do left page
        canvas.width = result.exportSpread.left.w;
        canvas.height = result.exportSpread.left.h;
        
        ctx.fillStyle = 'white';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        const src = result.srcRectGlobal;
        const dst = result.dstRectLeft;
        if (dst) {
          ctx.drawImage(
            img,
            src.x, src.y, src.w, src.h,
            dst.x, dst.y, dst.w, dst.h
          );
        }
      }
      
      // Open in new tab
      canvas.toBlob((blob) => {
        if (blob) {
          const url = URL.createObjectURL(blob);
          window.open(url, '_blank');
        }
      });
    };
    
    // Get image source from the current state
    const imageData = inputs.srcW ? document.querySelector('img[alt*="Preview"]') as HTMLImageElement : null;
    if (imageData) {
      img.src = imageData.src;
    }
  };

  return (
    <div className="mt-4 p-4 bg-gray-100 rounded-lg">
      <div className="flex justify-between items-center mb-2">
        <h4 className="font-semibold text-sm">Algorithm Debug Data</h4>
        <div className="space-x-2">
          <button
            onClick={copyToClipboard}
            className="px-3 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition-colors"
          >
            📋 Copy Debug Data
          </button>
          <button
            onClick={openRenderedImage}
            className="px-3 py-1 bg-green-500 text-white text-xs rounded hover:bg-green-600 transition-colors"
          >
            🖼️ Open Render
          </button>
        </div>
      </div>
      <div className="text-xs text-gray-600 mb-2">
        Compare with Go CLI output - inputs, result, and scaled result for preview
      </div>
      <pre className="text-xs overflow-auto max-h-60 bg-white p-3 rounded border font-mono">
        {debugText}
      </pre>
    </div>
  );
};
