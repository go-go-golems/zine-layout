import React from 'react';
import { useAppSelector } from '../../hooks/redux';
import { PAPER_SIZES } from '../../store/bookSpreadSlice';
import { AlgorithmData } from '../../utils/spreadAlgorithmHook';

const DEBUG = true;

interface ImagePreviewProps {
  image: { src: string; width: number; height: number };
  srcRect: { x: number; y: number; w: number; h: number };
  dstRect: { x: number; y: number; w: number; h: number };
  containerWidth: number;
  containerHeight: number;
  containerX: number;
  containerY: number;
  side?: 'left' | 'right' | 'single';
}

const ImagePreview: React.FC<ImagePreviewProps> = ({
  image,
  srcRect,
  dstRect,
  containerWidth,
  containerHeight,
  containerX,
  containerY,
  side = 'single'
}) => {
  const canvasRef = React.useRef<HTMLCanvasElement>(null);
  
  React.useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    // Set canvas size to container size
    canvas.width = containerWidth;
    canvas.height = containerHeight;
    
    // Fill white background
    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    
    // Load and draw image
    const img = new Image();
    img.onload = () => {
      // Draw the source rectangle to the destination rectangle
      ctx.drawImage(
        img,
        srcRect.x, srcRect.y, srcRect.w, srcRect.h,  // source crop
        dstRect.x, dstRect.y, dstRect.w, dstRect.h   // destination
      );
    };
    img.src = image.src;
  }, [image.src, srcRect, dstRect, containerWidth, containerHeight]);
  
  // Calculate scale for debugging
  const scaleX = dstRect.w / srcRect.w;
  const scaleY = dstRect.h / srcRect.h;

  return (
    <div 
      className="absolute bg-white border border-gray-300"
      style={{
        left: containerX,
        top: containerY,
        width: containerWidth,
        height: containerHeight
      }}
    >
      <canvas
        ref={canvasRef}
        className="w-full h-full"
        style={{ display: 'block' }}
      />
      
      {/* Debug overlay showing image metrics */}
      <div className="absolute top-1 left-1 bg-black bg-opacity-70 text-white text-xs px-2 py-1 rounded">
        <div>{side}</div>
        <div>container: {Math.round(containerWidth)}×{Math.round(containerHeight)}</div>
        <div>dst: {Math.round(dstRect.x)},{Math.round(dstRect.y)} {Math.round(dstRect.w)}×{Math.round(dstRect.h)}</div>
        <div>scale: {scaleX.toFixed(3)}</div>
        <div>crop: {Math.round(srcRect.x)},{Math.round(srcRect.y)} {Math.round(srcRect.w)}×{Math.round(srcRect.h)}</div>
      </div>
    </div>
  );
};

interface PreviewCanvasProps {
  algorithmData: AlgorithmData;
}

export const PreviewCanvasNew: React.FC<PreviewCanvasProps> = ({ algorithmData }) => {
  const { scaledResult, previewWidth, previewHeight } = algorithmData;
  const state = useAppSelector((state) => state.bookSpread);
  const { image, isSpread, orientation } = state;
  const paperDims = PAPER_SIZES[state.paperSize];

  if (!image) {
    return (
      <div className="text-center text-gray-500 py-12">
        <div className="text-6xl mb-4">🖼️</div>
        <p>Upload an image to start designing your photobook spread</p>
      </div>
    );
  }

  if (DEBUG) {
    console.log('[PreviewCanvasNew] Debug info:', algorithmData);
  }

  const renderSinglePage = () => (
    <ImagePreview
      image={image!}
      srcRect={scaledResult.srcRectGlobal}
      dstRect={scaledResult.dstRectGlobal}
      containerWidth={scaledResult.contentRect.w}
      containerHeight={scaledResult.contentRect.h}
      containerX={0}
      containerY={0}
      side="single"
    />
  );

  const renderSpread = () => {
    if (!scaledResult.leftPanel || !scaledResult.rightPanel || !scaledResult.dstRectLeft || !scaledResult.dstRectRight || !scaledResult.pageW) {
      return null;
    }

    return (
      <>
        {/* Left page */}
        <ImagePreview
          image={image!}
          srcRect={scaledResult.srcRectGlobal}
          dstRect={scaledResult.dstRectLeft}
          containerWidth={scaledResult.leftPanel.w}
          containerHeight={scaledResult.leftPanel.h}
          containerX={scaledResult.leftPanel.x}
          containerY={scaledResult.leftPanel.y}
          side="left"
        />

        {/* Right page */}
        <ImagePreview
          image={image!}
          srcRect={scaledResult.srcRectGlobal}
          dstRect={scaledResult.dstRectRight}
          containerWidth={scaledResult.rightPanel.w}
          containerHeight={scaledResult.rightPanel.h}
          containerX={scaledResult.rightPanel.x}
          containerY={scaledResult.rightPanel.y}
          side="right"
        />
      </>
    );
  };

  return (
    <div className="border-2 border-gray-300 relative bg-gray-100"
      style={{ width: scaledResult.contentRect.w, height: scaledResult.contentRect.h }}>
      
      {/* Content area guide */}
      <div className="absolute border border-red-200"
        style={{
          left: scaledResult.contentRect.x,
          top: scaledResult.contentRect.y,
          width: scaledResult.contentRect.w,
          height: scaledResult.contentRect.h
        }} />

      {/* Gutter guide for spreads */}
      {isSpread && scaledResult.gutterPx > 0 && scaledResult.pageW && (
        <>
          <div className="absolute border-2 border-blue-300 bg-blue-100 bg-opacity-30"
            style={{
              left: scaledResult.contentRect.x + scaledResult.pageW - scaledResult.gutterPx / 2,
              top: scaledResult.contentRect.y,
              width: scaledResult.gutterPx,
              height: scaledResult.contentRect.h
            }} />
          <div className="absolute text-blue-600 text-xs font-semibold pointer-events-none"
            style={{
              left: scaledResult.contentRect.x + scaledResult.pageW - 20,
              top: scaledResult.contentRect.y + scaledResult.contentRect.h / 2 - 8
            }}>
            GUTTER
          </div>
        </>
      )}

      {/* Image rendering */}
      {isSpread ? renderSpread() : renderSinglePage()}

      {/* Paper info overlay */}
      <div className="absolute top-2 right-2 bg-black bg-opacity-50 text-white text-xs px-2 py-1 rounded">
        {paperDims.width}"×{paperDims.height}" {orientation} {isSpread ? 'spread' : 'single'}
      </div>

    </div>
  );
};
