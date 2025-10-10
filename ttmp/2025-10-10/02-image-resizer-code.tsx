import React, { useState, useRef, useEffect } from 'react';

export default function ImageResizer() {
  const [image, setImage] = useState(null);
  const [mode, setMode] = useState('page'); // 'page', 'crop', 'fit', 'spread'
  const canvasRef = useRef(null);
  const leftPageCanvasRef = useRef(null);
  const rightPageCanvasRef = useRef(null);
  const combinedSpreadCanvasRef = useRef(null);
  const imageRef = useRef(null);

  // Mode A: Page + margins
  const [pageWidth, setPageWidth] = useState(1920);
  const [pageHeight, setPageHeight] = useState(1080);
  const [marginLeft, setMarginLeft] = useState(100);
  const [marginRight, setMarginRight] = useState(100);
  const [marginTop, setMarginTop] = useState(100);
  const [marginBottom, setMarginBottom] = useState(100);
  const [marginLinked, setMarginLinked] = useState(true);
  const [pageCropFill, setPageCropFill] = useState(false);
  const [pageCentering, setPageCentering] = useState(false);
  const [pageAnchorX, setPageAnchorX] = useState(0.5); // 0=left, 0.5=center, 1=right
  const [pageAnchorY, setPageAnchorY] = useState(0.5); // 0=top, 0.5=middle, 1=bottom
  
  // Mode B: Fixed crop
  const [cropAspect, setCropAspect] = useState('1:1');
  const [cropWidth, setCropWidth] = useState(1000);
  const [cropHeight, setCropHeight] = useState(1000);
  const [cropFitMode, setCropFitMode] = useState('cover');
  const [cropAnchorX, setCropAnchorX] = useState(0.5);
  const [cropAnchorY, setCropAnchorY] = useState(0.5);
  
  // Mode C: Fit to dimension
  const [fitMode, setFitMode] = useState('width'); // 'width' or 'height'
  const [fitWidth, setFitWidth] = useState(1920);
  const [fitHeight, setFitHeight] = useState(1080);
  const [fitCropFill, setFitCropFill] = useState(false);
  const [fitAnchorX, setFitAnchorX] = useState(0.5);
  const [fitAnchorY, setFitAnchorY] = useState(0.5);
  
  // Mode D: Spread
  const [spreadWidth, setSpreadWidth] = useState(3840); // Full spread width
  const [spreadHeight, setSpreadHeight] = useState(1920);
  const [gutterWidth, setGutterWidth] = useState(100);
  const [gutterPosition, setGutterPosition] = useState('center'); // 'center', 'left', 'right'
  const [gutterOverlap, setGutterOverlap] = useState(50); // How much each page overlaps into gutter
  const [spreadCropFill, setSpreadCropFill] = useState(true);
  const [spreadAnchorX, setSpreadAnchorX] = useState(0.5);
  const [spreadAnchorY, setSpreadAnchorY] = useState(0.5);
  
  // Common
  const [zoom, setZoom] = useState(1.0);
  const [dragX, setDragX] = useState(0);
  const [dragY, setDragY] = useState(0);
  const [dragLimits, setDragLimits] = useState({ minX: -2000, maxX: 2000, minY: -2000, maxY: 2000, canDragX: true, canDragY: true });

  const handleImageUpload = (e) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        const img = new Image();
        img.onload = () => {
          imageRef.current = img;
          setImage(img);
          setDragX(0);
          setDragY(0);
          setZoom(1.0);
        };
        img.src = event.target.result;
      };
      reader.readAsDataURL(file);
    }
  };

  const updateCropDimensions = (aspect) => {
    if (aspect === 'custom') return;
    const [w, h] = aspect.split(':').map(Number);
    const ratio = w / h;
    setCropHeight(Math.round(cropWidth / ratio));
  };

  useEffect(() => {
    if (cropAspect !== 'custom') {
      updateCropDimensions(cropAspect);
    }
  }, [cropAspect, cropWidth]);

  useEffect(() => {
    if (marginLinked) {
      setMarginRight(marginLeft);
      setMarginTop(marginLeft);
      setMarginBottom(marginLeft);
    }
  }, [marginLeft, marginLinked]);

  const baseScale = (fitMode, Wi, Hi, Wt, Ht) => {
    if (fitMode === 'cover') return Math.max(Wt / Wi, Ht / Hi);
    if (fitMode === 'contain') return Math.min(Wt / Wi, Ht / Hi);
    if (fitMode === 'fitWidth') return Wt / Wi;
    if (fitMode === 'fitHeight') return Ht / Hi;
    return 1;
  };

  const anchorToTranslation = (ax, ay, s, Wi, Hi, Wt, Ht) => {
    const ux = (2 * ax - 1) * (s * Wi - Wt) / 2;
    const uy = (2 * ay - 1) * (s * Hi - Ht) / 2;
    return { ux, uy };
  };

  const clampCover = (ux, uy, s, Wi, Hi, Wt, Ht) => {
    const xMin = -(s * Wi - Wt) / 2;
    const xMax = +(s * Wi - Wt) / 2;
    const yMin = -(s * Hi - Ht) / 2;
    const yMax = +(s * Hi - Ht) / 2;
    
    // Only clamp if there's actually room to move (image is larger than target)
    // If image dimension equals or is smaller than target, center it (force to 0)
    const uxClamped = (s * Wi > Wt) ? Math.min(xMax, Math.max(xMin, ux)) : 0;
    const uyClamped = (s * Hi > Ht) ? Math.min(yMax, Math.max(yMin, uy)) : 0;
    
    return {
      ux: uxClamped,
      uy: uyClamped,
    };
  };

  const renderSpreadPages = (s, ux, uy, Wt, Ht, gutterX) => {
    if (!image) return;

    const Wi = image.width;
    const Hi = image.height;

    // Calculate page dimensions
    let leftPageWidth, rightPageWidth;
    
    if (gutterPosition === 'center') {
      leftPageWidth = gutterX + gutterOverlap;
      rightPageWidth = gutterX + gutterOverlap;
    } else if (gutterPosition === 'left') {
      leftPageWidth = gutterWidth + gutterOverlap;
      rightPageWidth = Wt - gutterWidth + gutterOverlap;
    } else {
      leftPageWidth = Wt - gutterWidth + gutterOverlap;
      rightPageWidth = gutterWidth + gutterOverlap;
    }

    // Create a temporary canvas for the full spread
    const tempCanvas = document.createElement('canvas');
    tempCanvas.width = Wt;
    tempCanvas.height = Ht;
    const tempCtx = tempCanvas.getContext('2d');

    // Draw full spread image
    tempCtx.fillStyle = '#ffffff';
    tempCtx.fillRect(0, 0, Wt, Ht);
    tempCtx.save();
    tempCtx.translate(Wt / 2, Ht / 2);
    tempCtx.translate(ux, uy);
    tempCtx.scale(s, s);
    tempCtx.drawImage(image, -Wi / 2, -Hi / 2);
    tempCtx.restore();

    // Render left page
    const leftCanvas = leftPageCanvasRef.current;
    if (leftCanvas) {
      leftCanvas.width = leftPageWidth;
      leftCanvas.height = Ht;
      const leftCtx = leftCanvas.getContext('2d');
      
      const leftSourceX = gutterX - leftPageWidth;
      leftCtx.fillStyle = '#ffffff';
      leftCtx.fillRect(0, 0, leftPageWidth, Ht);
      leftCtx.drawImage(
        tempCanvas,
        Math.max(0, leftSourceX), 0, Math.min(leftPageWidth, Wt), Ht,
        0, 0, leftPageWidth, Ht
      );
      
      // Draw gutter line on left page (at right edge where it overlaps)
      const gutterLineX = leftPageWidth - gutterOverlap;
      leftCtx.strokeStyle = 'rgba(255, 0, 0, 0.6)';
      leftCtx.lineWidth = 2;
      leftCtx.setLineDash([10, 5]);
      leftCtx.beginPath();
      leftCtx.moveTo(gutterLineX, 0);
      leftCtx.lineTo(gutterLineX, Ht);
      leftCtx.stroke();
      leftCtx.setLineDash([]);
      
      // Draw border
      leftCtx.strokeStyle = '#333';
      leftCtx.lineWidth = 2;
      leftCtx.strokeRect(0, 0, leftPageWidth, Ht);
    }

    // Render right page
    const rightCanvas = rightPageCanvasRef.current;
    if (rightCanvas) {
      rightCanvas.width = rightPageWidth;
      rightCanvas.height = Ht;
      const rightCtx = rightCanvas.getContext('2d');
      
      const rightSourceX = gutterX - gutterOverlap;
      rightCtx.fillStyle = '#ffffff';
      rightCtx.fillRect(0, 0, rightPageWidth, Ht);
      rightCtx.drawImage(
        tempCanvas,
        Math.min(rightSourceX, Wt - rightPageWidth), 0, Math.min(rightPageWidth, Wt), Ht,
        0, 0, rightPageWidth, Ht
      );
      
      // Draw gutter line on right page (at left edge where it overlaps)
      rightCtx.strokeStyle = 'rgba(255, 0, 0, 0.6)';
      rightCtx.lineWidth = 2;
      rightCtx.setLineDash([10, 5]);
      rightCtx.beginPath();
      rightCtx.moveTo(gutterOverlap, 0);
      rightCtx.lineTo(gutterOverlap, Ht);
      rightCtx.stroke();
      rightCtx.setLineDash([]);
      
      // Draw border
      rightCtx.strokeStyle = '#333';
      rightCtx.lineWidth = 2;
      rightCtx.strokeRect(0, 0, rightPageWidth, Ht);
    }

    // Render combined spread (both pages together)
    const combinedCanvas = combinedSpreadCanvasRef.current;
    if (combinedCanvas && leftCanvas && rightCanvas) {
      const gap = 4; // Small gap to show the binding/gutter
      combinedCanvas.width = leftPageWidth + rightPageWidth + gap;
      combinedCanvas.height = Ht;
      const combinedCtx = combinedCanvas.getContext('2d');
      
      // Fill background (gap will show through)
      combinedCtx.fillStyle = '#333';
      combinedCtx.fillRect(0, 0, combinedCanvas.width, combinedCanvas.height);
      
      // Draw left page
      combinedCtx.drawImage(leftCanvas, 0, 0);
      
      // Draw right page
      combinedCtx.drawImage(rightCanvas, leftPageWidth + gap, 0);
    }
  };

  const renderCanvas = () => {
    const canvas = canvasRef.current;
    if (!canvas || !image) return;

    const ctx = canvas.getContext('2d');
    const Wi = image.width;
    const Hi = image.height;

    let Wt, Ht, s, ux, uy, ct, outputW, outputH;

    if (mode === 'page') {
      // Mode A: Page + margins
      outputW = pageWidth;
      outputH = pageHeight;
      canvas.width = outputW;
      canvas.height = outputH;

      const Wu = pageWidth - (marginLeft + marginRight);
      const Ht = pageHeight - (marginTop + marginBottom);
      Wt = Wu;

      const fitMode = pageCropFill ? 'cover' : 'contain';
      s = baseScale(fitMode, Wi, Hi, Wt, Ht) * zoom;

      if (pageCentering) {
        // Use anchor-based positioning
        const anchor = anchorToTranslation(pageAnchorX, pageAnchorY, s, Wi, Hi, Wt, Ht);
        ux = anchor.ux + dragX;
        uy = anchor.uy + dragY;
      } else {
        // No centering - just use drag (image centered by default at 0,0)
        ux = dragX;
        uy = dragY;
      }

      if (pageCropFill) {
        const clamped = clampCover(ux, uy, s, Wi, Hi, Wt, Ht);
        ux = clamped.ux;
        uy = clamped.uy;
      }

      ct = { x: marginLeft + Wu / 2, y: marginTop + Ht / 2 };

      // Fill background
      ctx.fillStyle = '#f0f0f0';
      ctx.fillRect(0, 0, outputW, outputH);

      // Draw image at reduced opacity first (shows what's cut off)
      ctx.save();
      ctx.globalAlpha = 0.3;
      ctx.translate(ct.x, ct.y);
      ctx.translate(ux, uy);
      ctx.scale(s, s);
      ctx.drawImage(image, -Wi / 2, -Hi / 2);
      ctx.restore();

      // Clip to visible area and draw image at full opacity
      ctx.save();
      ctx.beginPath();
      ctx.rect(marginLeft, marginTop, Wu, Ht);
      ctx.clip();
      ctx.translate(ct.x, ct.y);
      ctx.translate(ux, uy);
      ctx.scale(s, s);
      ctx.drawImage(image, -Wi / 2, -Hi / 2);
      ctx.restore();

      // Draw content area outline
      ctx.strokeStyle = '#0066cc';
      ctx.lineWidth = 2;
      ctx.strokeRect(marginLeft, marginTop, Wu, Ht);
    } else if (mode === 'crop') {
      // Mode B: Fixed crop
      Wt = cropWidth;
      Ht = cropHeight;
      outputW = Wt;
      outputH = Ht;
      canvas.width = outputW;
      canvas.height = outputH;

      s = baseScale(cropFitMode, Wi, Hi, Wt, Ht) * zoom;

      const anchor = anchorToTranslation(cropAnchorX, cropAnchorY, s, Wi, Hi, Wt, Ht);
      ux = anchor.ux + dragX;
      uy = anchor.uy + dragY;

      if (cropFitMode === 'cover') {
        const clamped = clampCover(ux, uy, s, Wi, Hi, Wt, Ht);
        ux = clamped.ux;
        uy = clamped.uy;
      }

      ct = { x: Wt / 2, y: Ht / 2 };

      ctx.fillStyle = '#ffffff';
      ctx.fillRect(0, 0, outputW, outputH);

    } else if (mode === 'fit') {
      // Mode C: Fit to width/height
      if (!fitCropFill) {
        // No crop - output is scaled image
        if (fitMode === 'width') {
          s = fitWidth / Wi;
          outputW = fitWidth;
          outputH = Math.floor(s * Hi);
        } else {
          s = fitHeight / Hi;
          outputH = fitHeight;
          outputW = Math.floor(s * Wi);
        }
        s *= zoom;
        Wt = outputW;
        Ht = outputH;
        canvas.width = outputW;
        canvas.height = outputH;

        ux = dragX;
        uy = dragY;
        ct = { x: outputW / 2, y: outputH / 2 };

        ctx.fillStyle = '#ffffff';
        ctx.fillRect(0, 0, outputW, outputH);

      } else {
        // Crop/Fill on - force both dimensions
        Wt = fitWidth;
        Ht = fitHeight;
        outputW = Wt;
        outputH = Ht;
        canvas.width = outputW;
        canvas.height = outputH;

        s = baseScale('cover', Wi, Hi, Wt, Ht) * zoom;

        const anchor = anchorToTranslation(fitAnchorX, fitAnchorY, s, Wi, Hi, Wt, Ht);
        ux = anchor.ux + dragX;
        uy = anchor.uy + dragY;

        const clamped = clampCover(ux, uy, s, Wi, Hi, Wt, Ht);
        ux = clamped.ux;
        uy = clamped.uy;

        ct = { x: Wt / 2, y: Ht / 2 };

        ctx.fillStyle = '#ffffff';
        ctx.fillRect(0, 0, outputW, outputH);
      }
    } else if (mode === 'spread') {
      // Mode D: Spread with gutter
      Wt = spreadWidth;
      Ht = spreadHeight;
      outputW = Wt;
      outputH = Ht;
      canvas.width = outputW;
      canvas.height = outputH;
      
      const fitMode = spreadCropFill ? 'cover' : 'contain';
      s = baseScale(fitMode, Wi, Hi, Wt, Ht) * zoom;

      // Calculate drag limits
      const xRange = (s * Wi - Wt) / 2;
      const yRange = (s * Hi - Ht) / 2;
      
      if (spreadCropFill) {
        newDragLimits = {
          minX: s * Wi > Wt ? -xRange : 0,
          maxX: s * Wi > Wt ? xRange : 0,
          minY: s * Hi > Ht ? -yRange : 0,
          maxY: s * Hi > Ht ? yRange : 0,
          canDragX: s * Wi > Wt,
          canDragY: s * Hi > Ht
        };
      } else {
        newDragLimits = {
          minX: -Math.max(xRange, 500),
          maxX: Math.max(xRange, 500),
          minY: -Math.max(yRange, 500),
          maxY: Math.max(yRange, 500),
          canDragX: true,
          canDragY: true
        };
      }

      const anchor = anchorToTranslation(spreadAnchorX, spreadAnchorY, s, Wi, Hi, Wt, Ht);
      ux = anchor.ux + dragX;
      uy = anchor.uy + dragY;

      if (spreadCropFill) {
        const clamped = clampCover(ux, uy, s, Wi, Hi, Wt, Ht);
        ux = clamped.ux;
        uy = clamped.uy;
      }

      // Draw full spread in main canvas
      ctx.fillStyle = '#ffffff';
      ctx.fillRect(0, 0, Wt, Ht);
      ctx.save();
      ctx.translate(Wt / 2, Ht / 2);
      ctx.translate(ux, uy);
      ctx.scale(s, s);
      ctx.drawImage(image, -Wi / 2, -Hi / 2);
      ctx.restore();

      // Calculate gutter position
      let gutterX;
      
      if (gutterPosition === 'center') {
        gutterX = Wt / 2;
      } else if (gutterPosition === 'left') {
        gutterX = gutterWidth;
      } else { // right
        gutterX = Wt - gutterWidth;
      }

      // Draw gutter visualization
      ctx.fillStyle = 'rgba(255, 0, 0, 0.15)';
      ctx.fillRect(gutterX - gutterWidth / 2, 0, gutterWidth, Ht);
      ctx.strokeStyle = 'rgba(255, 0, 0, 0.8)';
      ctx.lineWidth = 2;
      ctx.setLineDash([10, 5]);
      ctx.beginPath();
      ctx.moveTo(gutterX, 0);
      ctx.lineTo(gutterX, Ht);
      ctx.stroke();
      ctx.setLineDash([]);

      // Add label
      ctx.fillStyle = 'rgba(255, 0, 0, 0.8)';
      ctx.font = 'bold 14px sans-serif';
      ctx.fillText('GUTTER', gutterX - 30, 30);

      ct = { x: Wt / 2, y: Ht / 2 };

      // Render split pages
      renderSpreadPages(s, ux, uy, Wt, Ht, gutterX);
    }

    // Draw image (for non-page modes, page mode handles its own drawing)
    if (mode !== 'page') {
      ctx.save();
      ctx.translate(ct.x, ct.y);
      ctx.translate(ux, uy);
      ctx.scale(s, s);
      ctx.drawImage(image, -Wi / 2, -Hi / 2);
      ctx.restore();
    }
  };

  useEffect(() => {
    renderCanvas();
  }, [image, mode, pageWidth, pageHeight, marginLeft, marginRight, marginTop, marginBottom,
      pageCropFill, pageCentering, pageAnchorX, pageAnchorY, cropWidth, cropHeight, cropFitMode,
      cropAnchorX, cropAnchorY, fitMode, fitWidth, fitHeight, fitCropFill,
      fitAnchorX, fitAnchorY, spreadWidth, spreadHeight, gutterWidth, gutterPosition,
      gutterOverlap, spreadCropFill, spreadAnchorX, spreadAnchorY, zoom, dragX, dragY]);

  const downloadImage = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const link = document.createElement('a');
    link.download = `resized-${mode}.png`;
    link.href = canvas.toDataURL();
    link.click();
  };

  const downloadSpreadPages = (page) => {
    if (!image || mode !== 'spread') return;

    const Wi = image.width;
    const Hi = image.height;
    const Wt = spreadWidth;
    const Ht = spreadHeight;

    const fitMode = spreadCropFill ? 'cover' : 'contain';
    const s = baseScale(fitMode, Wi, Hi, Wt, Ht) * zoom;

    const anchor = anchorToTranslation(spreadAnchorX, spreadAnchorY, s, Wi, Hi, Wt, Ht);
    let ux = anchor.ux + dragX;
    let uy = anchor.uy + dragY;

    if (spreadCropFill) {
      const clamped = clampCover(ux, uy, s, Wi, Hi, Wt, Ht);
      ux = clamped.ux;
      uy = clamped.uy;
    }

    // Calculate gutter and page dimensions
    let gutterX, leftPageWidth, rightPageWidth;
    
    if (gutterPosition === 'center') {
      gutterX = Wt / 2;
      leftPageWidth = gutterX + gutterOverlap;
      rightPageWidth = gutterX + gutterOverlap;
    } else if (gutterPosition === 'left') {
      gutterX = gutterWidth;
      leftPageWidth = gutterWidth + gutterOverlap;
      rightPageWidth = Wt - gutterWidth + gutterOverlap;
    } else {
      gutterX = Wt - gutterWidth;
      leftPageWidth = Wt - gutterWidth + gutterOverlap;
      rightPageWidth = gutterWidth + gutterOverlap;
    }

    // Create full spread
    const tempCanvas = document.createElement('canvas');
    tempCanvas.width = Wt;
    tempCanvas.height = Ht;
    const tempCtx = tempCanvas.getContext('2d');

    tempCtx.fillStyle = '#ffffff';
    tempCtx.fillRect(0, 0, Wt, Ht);
    tempCtx.save();
    tempCtx.translate(Wt / 2, Ht / 2);
    tempCtx.translate(ux, uy);
    tempCtx.scale(s, s);
    tempCtx.drawImage(image, -Wi / 2, -Hi / 2);
    tempCtx.restore();

    // Create output canvas for single page or combined
    const outputCanvas = document.createElement('canvas');
    const outputCtx = outputCanvas.getContext('2d');

    if (page === 'combined') {
      // Download the combined spread preview
      const combinedCanvas = combinedSpreadCanvasRef.current;
      if (combinedCanvas) {
        const link = document.createElement('a');
        link.download = 'spread-combined.png';
        link.href = combinedCanvas.toDataURL();
        link.click();
      }
      return;
    } else if (page === 'left') {
      outputCanvas.width = leftPageWidth;
      outputCanvas.height = Ht;
      const sourceX = gutterX - leftPageWidth;
      outputCtx.drawImage(
        tempCanvas,
        Math.max(0, sourceX), 0, Math.min(leftPageWidth, Wt), Ht,
        0, 0, leftPageWidth, Ht
      );
    } else if (page === 'right') {
      outputCanvas.width = rightPageWidth;
      outputCanvas.height = Ht;
      const sourceX = gutterX - gutterOverlap;
      outputCtx.drawImage(
        tempCanvas,
        Math.min(sourceX, Wt - rightPageWidth), 0, Math.min(rightPageWidth, Wt), Ht,
        0, 0, rightPageWidth, Ht
      );
    } else {
      // Both pages (full uncut spread)
      outputCanvas.width = Wt;
      outputCanvas.height = Ht;
      outputCtx.drawImage(tempCanvas, 0, 0);
    }

    const link = document.createElement('a');
    link.download = `spread-${page}.png`;
    link.href = outputCanvas.toDataURL();
    link.click();
  };

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold mb-8 text-gray-900">Image Resizer</h1>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Controls */}
          <div className="lg:col-span-1 space-y-6">
            {/* Upload */}
            <div className="bg-white p-6 rounded-lg shadow">
              <h2 className="text-xl font-semibold mb-4">📤 Upload Image</h2>
              <input
                type="file"
                accept="image/*"
                onChange={handleImageUpload}
                className="w-full text-sm"
              />
            </div>

            {/* Mode Selection */}
            <div className="bg-white p-6 rounded-lg shadow">
              <h2 className="text-xl font-semibold mb-4">🎯 Target Size Mode</h2>
              <div className="space-y-2">
                <button
                  onClick={() => setMode('page')}
                  className={`w-full p-3 rounded text-left ${
                    mode === 'page' ? 'bg-blue-500 text-white' : 'bg-gray-100'
                  }`}
                >
                  A) Page + Margins
                </button>
                <button
                  onClick={() => setMode('crop')}
                  className={`w-full p-3 rounded text-left ${
                    mode === 'crop' ? 'bg-blue-500 text-white' : 'bg-gray-100'
                  }`}
                >
                  B) Fixed Crop Format
                </button>
                <button
                  onClick={() => setMode('fit')}
                  className={`w-full p-3 rounded text-left ${
                    mode === 'fit' ? 'bg-blue-500 text-white' : 'bg-gray-100'
                  }`}
                >
                  C) Fit to Width/Height
                </button>
                <button
                  onClick={() => setMode('spread')}
                  className={`w-full p-3 rounded text-left ${
                    mode === 'spread' ? 'bg-blue-500 text-white' : 'bg-gray-100'
                  }`}
                >
                  D) Spread with Gutter
                </button>
              </div>
            </div>

            {/* Mode-specific controls */}
            {mode === 'page' && (
              <div className="bg-white p-6 rounded-lg shadow space-y-4">
                <h2 className="text-xl font-semibold">📄 Page + Margins</h2>
                
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-sm font-medium mb-1">Page Width (px)</label>
                    <input
                      type="number"
                      value={pageWidth}
                      onChange={(e) => setPageWidth(Number(e.target.value))}
                      className="w-full p-2 border rounded"
                    />
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium mb-1">Page Height (px)</label>
                    <input
                      type="number"
                      value={pageHeight}
                      onChange={(e) => setPageHeight(Number(e.target.value))}
                      className="w-full p-2 border rounded"
                    />
                  </div>
                </div>

                <button
                  onClick={() => {
                    const temp = pageWidth;
                    setPageWidth(pageHeight);
                    setPageHeight(temp);
                  }}
                  className="w-full p-2 bg-gray-100 rounded hover:bg-gray-200 text-sm font-medium"
                >
                  🔄 Swap Portrait/Landscape
                </button>

                <div className="border-t pt-4">
                  <label className="flex items-center gap-2 mb-2">
                    <input
                      type="checkbox"
                      checked={marginLinked}
                      onChange={(e) => setMarginLinked(e.target.checked)}
                    />
                    <span className="text-sm font-medium">🔗 Link margins</span>
                  </label>
                  
                  <label className="block text-sm font-medium mb-1">
                    {marginLinked ? 'All Margins (px)' : 'Left Margin (px)'}
                  </label>
                  <input
                    type="number"
                    value={marginLeft}
                    onChange={(e) => setMarginLeft(Number(e.target.value))}
                    className="w-full p-2 border rounded"
                  />
                </div>

                {!marginLinked && (
                  <>
                    <div>
                      <label className="block text-sm font-medium mb-1">Right Margin (px)</label>
                      <input
                        type="number"
                        value={marginRight}
                        onChange={(e) => setMarginRight(Number(e.target.value))}
                        className="w-full p-2 border rounded"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1">Top Margin (px)</label>
                      <input
                        type="number"
                        value={marginTop}
                        onChange={(e) => setMarginTop(Number(e.target.value))}
                        className="w-full p-2 border rounded"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1">Bottom Margin (px)</label>
                      <input
                        type="number"
                        value={marginBottom}
                        onChange={(e) => setMarginBottom(Number(e.target.value))}
                        className="w-full p-2 border rounded"
                      />
                    </div>
                  </>
                )}

                {/* Visible area display */}
                <div className="bg-blue-50 border-2 border-blue-200 rounded p-3">
                  <div className="text-sm font-bold text-blue-900 mb-1">📐 Visible Area (Content)</div>
                  <div className="text-lg font-mono text-blue-700">
                    {pageWidth - (marginLeft + marginRight)} × {pageHeight - (marginTop + marginBottom)} px
                  </div>
                </div>

                <div className="border-t pt-4">
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={pageCropFill}
                      onChange={(e) => setPageCropFill(e.target.checked)}
                    />
                    <span className="text-sm font-medium">✂️ Crop/Fill (Cover mode)</span>
                  </label>
                </div>

                <div>
                  <label className="flex items-center gap-2 mb-2">
                    <input
                      type="checkbox"
                      checked={pageCentering}
                      onChange={(e) => setPageCentering(e.target.checked)}
                    />
                    <span className="text-sm font-medium">🎯 Use anchor centering</span>
                  </label>
                </div>

                {pageCentering && (
                  <>
                    <div>
                      <label className="block text-sm font-medium mb-2">Horizontal Anchor</label>
                      <div className="flex gap-2">
                        <button
                          onClick={() => setPageAnchorX(0)}
                          className={`flex-1 p-2 rounded text-sm ${pageAnchorX === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ⬅️ Left
                        </button>
                        <button
                          onClick={() => setPageAnchorX(0.5)}
                          className={`flex-1 p-2 rounded text-sm ${pageAnchorX === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ↔️ Center
                        </button>
                        <button
                          onClick={() => setPageAnchorX(1)}
                          className={`flex-1 p-2 rounded text-sm ${pageAnchorX === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ➡️ Right
                        </button>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium mb-2">Vertical Anchor</label>
                      <div className="flex gap-2">
                        <button
                          onClick={() => setPageAnchorY(0)}
                          className={`flex-1 p-2 rounded text-sm ${pageAnchorY === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ⬆️ Top
                        </button>
                        <button
                          onClick={() => setPageAnchorY(0.5)}
                          className={`flex-1 p-2 rounded text-sm ${pageAnchorY === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ↕️ Middle
                        </button>
                        <button
                          onClick={() => setPageAnchorY(1)}
                          className={`flex-1 p-2 rounded text-sm ${pageAnchorY === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ⬇️ Bottom
                        </button>
                      </div>
                    </div>
                  </>
                )}

                {!pageCentering && (
                  <div className="text-sm text-gray-600 italic bg-gray-50 p-3 rounded">
                    💡 Image will be centered in visible area. Use drag to position manually.
                  </div>
                )}
              </div>
            )}

            {mode === 'spread' && (
              <div className="bg-white p-6 rounded-lg shadow space-y-4">
                <h2 className="text-xl font-semibold">📖 Spread with Gutter</h2>
                
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-sm font-medium mb-1">Spread Width (px)</label>
                    <input
                      type="number"
                      value={spreadWidth}
                      onChange={(e) => setSpreadWidth(Number(e.target.value))}
                      className="w-full p-2 border rounded"
                    />
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium mb-1">Spread Height (px)</label>
                    <input
                      type="number"
                      value={spreadHeight}
                      onChange={(e) => setSpreadHeight(Number(e.target.value))}
                      className="w-full p-2 border rounded"
                    />
                  </div>
                </div>

                <button
                  onClick={() => {
                    const temp = spreadWidth;
                    setSpreadWidth(spreadHeight);
                    setSpreadHeight(temp);
                  }}
                  className="w-full p-2 bg-gray-100 rounded hover:bg-gray-200 text-sm font-medium"
                >
                  🔄 Swap Portrait/Landscape
                </button>

                <div className="border-t pt-4">
                  <label className="block text-sm font-medium mb-1">Gutter Width (px)</label>
                  <input
                    type="number"
                    value={gutterWidth}
                    onChange={(e) => setGutterWidth(Number(e.target.value))}
                    className="w-full p-2 border rounded"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Gutter Position</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setGutterPosition('left')}
                      className={`flex-1 p-2 rounded text-sm ${gutterPosition === 'left' ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      Left
                    </button>
                    <button
                      onClick={() => setGutterPosition('center')}
                      className={`flex-1 p-2 rounded text-sm ${gutterPosition === 'center' ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      Center
                    </button>
                    <button
                      onClick={() => setGutterPosition('right')}
                      className={`flex-1 p-2 rounded text-sm ${gutterPosition === 'right' ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      Right
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Overlap into Gutter (px)</label>
                  <input
                    type="number"
                    value={gutterOverlap}
                    onChange={(e) => setGutterOverlap(Number(e.target.value))}
                    className="w-full p-2 border rounded"
                  />
                  <p className="text-xs text-gray-600 mt-1">
                    How much each page extends into the gutter area
                  </p>
                </div>

                {/* Info about page sizes */}
                <div className="bg-blue-50 border-2 border-blue-200 rounded p-3 space-y-1">
                  <div className="text-sm font-bold text-blue-900">📐 Output Pages</div>
                  {gutterPosition === 'center' && (
                    <>
                      <div className="text-sm text-blue-700">
                        <strong>Left Page:</strong> {Math.round((spreadWidth / 2) + gutterOverlap)} × {spreadHeight} px
                      </div>
                      <div className="text-sm text-blue-700">
                        <strong>Right Page:</strong> {Math.round((spreadWidth / 2) + gutterOverlap)} × {spreadHeight} px
                      </div>
                    </>
                  )}
                  {gutterPosition === 'left' && (
                    <>
                      <div className="text-sm text-blue-700">
                        <strong>Left Page:</strong> {gutterWidth + gutterOverlap} × {spreadHeight} px
                      </div>
                      <div className="text-sm text-blue-700">
                        <strong>Right Page:</strong> {spreadWidth - gutterWidth + gutterOverlap} × {spreadHeight} px
                      </div>
                    </>
                  )}
                  {gutterPosition === 'right' && (
                    <>
                      <div className="text-sm text-blue-700">
                        <strong>Left Page:</strong> {spreadWidth - gutterWidth + gutterOverlap} × {spreadHeight} px
                      </div>
                      <div className="text-sm text-blue-700">
                        <strong>Right Page:</strong> {gutterWidth + gutterOverlap} × {spreadHeight} px
                      </div>
                    </>
                  )}
                </div>

                <div className="border-t pt-4">
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={spreadCropFill}
                      onChange={(e) => setSpreadCropFill(e.target.checked)}
                    />
                    <span className="text-sm font-medium">✂️ Crop/Fill (Cover mode)</span>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Horizontal Anchor</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setSpreadAnchorX(0)}
                      className={`flex-1 p-2 rounded text-sm ${spreadAnchorX === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ⬅️ Left
                    </button>
                    <button
                      onClick={() => setSpreadAnchorX(0.5)}
                      className={`flex-1 p-2 rounded text-sm ${spreadAnchorX === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ↔️ Center
                    </button>
                    <button
                      onClick={() => setSpreadAnchorX(1)}
                      className={`flex-1 p-2 rounded text-sm ${spreadAnchorX === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ➡️ Right
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Vertical Anchor</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setSpreadAnchorY(0)}
                      className={`flex-1 p-2 rounded text-sm ${spreadAnchorY === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ⬆️ Top
                    </button>
                    <button
                      onClick={() => setSpreadAnchorY(0.5)}
                      className={`flex-1 p-2 rounded text-sm ${spreadAnchorY === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ↕️ Middle
                    </button>
                    <button
                      onClick={() => setSpreadAnchorY(1)}
                      className={`flex-1 p-2 rounded text-sm ${spreadAnchorY === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ⬇️ Bottom
                    </button>
                  </div>
                </div>
              </div>
            )}

            {mode === 'crop' && (
              <div className="bg-white p-6 rounded-lg shadow space-y-4">
                <h2 className="text-xl font-semibold">✂️ Fixed Crop</h2>
                
                <div>
                  <label className="block text-sm font-medium mb-1">Aspect Ratio</label>
                  <select
                    value={cropAspect}
                    onChange={(e) => setCropAspect(e.target.value)}
                    className="w-full p-2 border rounded"
                  >
                    <option value="1:1">1:1 (Square)</option>
                    <option value="3:2">3:2 (Landscape)</option>
                    <option value="2:3">2:3 (Portrait)</option>
                    <option value="4:3">4:3 (Landscape)</option>
                    <option value="3:4">3:4 (Portrait)</option>
                    <option value="5:4">5:4 (Landscape)</option>
                    <option value="4:5">4:5 (Portrait)</option>
                    <option value="16:9">16:9 (Landscape)</option>
                    <option value="9:16">9:16 (Portrait)</option>
                    <option value="custom">Custom</option>
                  </select>
                </div>

                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-sm font-medium mb-1">Crop Width (px)</label>
                    <input
                      type="number"
                      value={cropWidth}
                      onChange={(e) => setCropWidth(Number(e.target.value))}
                      className="w-full p-2 border rounded"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1">Crop Height (px)</label>
                    <input
                      type="number"
                      value={cropHeight}
                      onChange={(e) => setCropHeight(Number(e.target.value))}
                      disabled={cropAspect !== 'custom'}
                      className="w-full p-2 border rounded disabled:bg-gray-100"
                    />
                  </div>
                </div>

                <button
                  onClick={() => {
                    const temp = cropWidth;
                    setCropWidth(cropHeight);
                    setCropHeight(temp);
                    // Also swap the aspect ratio if it's not square or custom
                    if (cropAspect === '3:2') setCropAspect('2:3');
                    else if (cropAspect === '2:3') setCropAspect('3:2');
                    else if (cropAspect === '4:3') setCropAspect('3:4');
                    else if (cropAspect === '3:4') setCropAspect('4:3');
                    else if (cropAspect === '5:4') setCropAspect('4:5');
                    else if (cropAspect === '4:5') setCropAspect('5:4');
                    else if (cropAspect === '16:9') setCropAspect('9:16');
                    else if (cropAspect === '9:16') setCropAspect('16:9');
                  }}
                  className="w-full p-2 bg-gray-100 rounded hover:bg-gray-200 text-sm font-medium"
                >
                  🔄 Swap Portrait/Landscape
                </button>

                <div>
                  <label className="block text-sm font-medium mb-2">Fit Mode</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setCropFitMode('cover')}
                      className={`flex-1 p-2 rounded ${cropFitMode === 'cover' ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      Cover
                    </button>
                    <button
                      onClick={() => setCropFitMode('contain')}
                      className={`flex-1 p-2 rounded ${cropFitMode === 'contain' ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      Contain
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Horizontal Anchor</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setCropAnchorX(0)}
                      className={`flex-1 p-2 rounded ${cropAnchorX === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ⬅️ Left
                    </button>
                    <button
                      onClick={() => setCropAnchorX(0.5)}
                      className={`flex-1 p-2 rounded ${cropAnchorX === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ↔️ Center
                    </button>
                    <button
                      onClick={() => setCropAnchorX(1)}
                      className={`flex-1 p-2 rounded ${cropAnchorX === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ➡️ Right
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Vertical Anchor</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setCropAnchorY(0)}
                      className={`flex-1 p-2 rounded ${cropAnchorY === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ⬆️ Top
                    </button>
                    <button
                      onClick={() => setCropAnchorY(0.5)}
                      className={`flex-1 p-2 rounded ${cropAnchorY === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ↕️ Middle
                    </button>
                    <button
                      onClick={() => setCropAnchorY(1)}
                      className={`flex-1 p-2 rounded ${cropAnchorY === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      ⬇️ Bottom
                    </button>
                  </div>
                </div>
              </div>
            )}

            {mode === 'fit' && (
              <div className="bg-white p-6 rounded-lg shadow space-y-4">
                <h2 className="text-xl font-semibold">📏 Fit to Dimension</h2>
                
                <div>
                  <label className="block text-sm font-medium mb-2">Target</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setFitMode('width')}
                      className={`flex-1 p-2 rounded ${fitMode === 'width' ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      Width
                    </button>
                    <button
                      onClick={() => setFitMode('height')}
                      className={`flex-1 p-2 rounded ${fitMode === 'height' ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                    >
                      Height
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Target Width (px)</label>
                  <input
                    type="number"
                    value={fitWidth}
                    onChange={(e) => setFitWidth(Number(e.target.value))}
                    className="w-full p-2 border rounded"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Target Height (px)</label>
                  <input
                    type="number"
                    value={fitHeight}
                    onChange={(e) => setFitHeight(Number(e.target.value))}
                    className="w-full p-2 border rounded"
                  />
                </div>

                <div>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={fitCropFill}
                      onChange={(e) => setFitCropFill(e.target.checked)}
                    />
                    <span className="text-sm font-medium">✂️ Crop/Fill (force both dimensions)</span>
                  </label>
                </div>

                {fitCropFill && (
                  <>
                    <div>
                      <label className="block text-sm font-medium mb-2">Horizontal Anchor</label>
                      <div className="flex gap-2">
                        <button
                          onClick={() => setFitAnchorX(0)}
                          className={`flex-1 p-2 rounded ${fitAnchorX === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ⬅️ Left
                        </button>
                        <button
                          onClick={() => setFitAnchorX(0.5)}
                          className={`flex-1 p-2 rounded ${fitAnchorX === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ↔️ Center
                        </button>
                        <button
                          onClick={() => setFitAnchorX(1)}
                          className={`flex-1 p-2 rounded ${fitAnchorX === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ➡️ Right
                        </button>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium mb-2">Vertical Anchor</label>
                      <div className="flex gap-2">
                        <button
                          onClick={() => setFitAnchorY(0)}
                          className={`flex-1 p-2 rounded ${fitAnchorY === 0 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ⬆️ Top
                        </button>
                        <button
                          onClick={() => setFitAnchorY(0.5)}
                          className={`flex-1 p-2 rounded ${fitAnchorY === 0.5 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ↕️ Middle
                        </button>
                        <button
                          onClick={() => setFitAnchorY(1)}
                          className={`flex-1 p-2 rounded ${fitAnchorY === 1 ? 'bg-blue-500 text-white' : 'bg-gray-100'}`}
                        >
                          ⬇️ Bottom
                        </button>
                      </div>
                    </div>
                  </>
                )}
              </div>
            )}

            {/* Common controls */}
            <div className="bg-white p-6 rounded-lg shadow space-y-4">
              <h2 className="text-xl font-semibold">🎛️ Adjustments</h2>
              <div className="text-xs text-gray-600 mb-3 bg-gray-50 p-2 rounded">
                These controls move the <strong>image</strong> within the visible area
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">Zoom: {zoom.toFixed(2)}x</label>
                <input
                  type="range"
                  min="0.1"
                  max="3"
                  step="0.1"
                  value={zoom}
                  onChange={(e) => setZoom(Number(e.target.value))}
                  className="w-full"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">
                  Horizontal Position {dragLimits.canDragX ? `(${Math.round(dragX)}px)` : '(centered)'}
                </label>
                {dragLimits.canDragX && (
                  <>
                    <div className="flex gap-1 mb-2">
                      <button
                        onClick={() => setDragX(dragLimits.minX)}
                        className="flex-1 p-1 text-xs bg-gray-100 rounded hover:bg-gray-200"
                      >
                        ⬅️ Left
                      </button>
                      <button
                        onClick={() => setDragX(0)}
                        className="flex-1 p-1 text-xs bg-gray-100 rounded hover:bg-gray-200"
                      >
                        ↔️ Center
                      </button>
                      <button
                        onClick={() => setDragX(dragLimits.maxX)}
                        className="flex-1 p-1 text-xs bg-gray-100 rounded hover:bg-gray-200"
                      >
                        ➡️ Right
                      </button>
                    </div>
                    <input
                      type="range"
                      min={dragLimits.minX}
                      max={dragLimits.maxX}
                      step="1"
                      value={dragX}
                      onChange={(e) => setDragX(Number(e.target.value))}
                      className="w-full"
                    />
                  </>
                )}
                {!dragLimits.canDragX && (
                  <div className="text-sm text-gray-500 italic bg-gray-100 p-2 rounded">
                    Image fits horizontally - no dragging needed
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">
                  Vertical Position {dragLimits.canDragY ? `(${Math.round(dragY)}px)` : '(centered)'}
                </label>
                {dragLimits.canDragY && (
                  <>
                    <div className="flex gap-1 mb-2">
                      <button
                        onClick={() => setDragY(dragLimits.minY)}
                        className="flex-1 p-1 text-xs bg-gray-100 rounded hover:bg-gray-200"
                      >
                        ⬆️ Top
                      </button>
                      <button
                        onClick={() => setDragY(0)}
                        className="flex-1 p-1 text-xs bg-gray-100 rounded hover:bg-gray-200"
                      >
                        ↕️ Middle
                      </button>
                      <button
                        onClick={() => setDragY(dragLimits.maxY)}
                        className="flex-1 p-1 text-xs bg-gray-100 rounded hover:bg-gray-200"
                      >
                        ⬇️ Bottom
                      </button>
                    </div>
                    <input
                      type="range"
                      min={dragLimits.minY}
                      max={dragLimits.maxY}
                      step="1"
                      value={dragY}
                      onChange={(e) => setDragY(Number(e.target.value))}
                      className="w-full"
                    />
                  </>
                )}
                {!dragLimits.canDragY && (
                  <div className="text-sm text-gray-500 italic bg-gray-100 p-2 rounded">
                    Image fits vertically - no dragging needed
                  </div>
                )}
              </div>

              <button
                onClick={() => {
                  setZoom(1);
                  setDragX(0);
                  setDragY(0);
                }}
                className="w-full p-2 bg-gray-200 rounded hover:bg-gray-300"
              >
                🔄 Reset All
              </button>
            </div>

            {/* Export */}
            <div className="bg-white p-6 rounded-lg shadow">
              <h2 className="text-xl font-semibold mb-4">💾 Export</h2>
              {mode === 'spread' ? (
                <div className="space-y-2">
                  <button
                    onClick={() => downloadSpreadPages('combined')}
                    disabled={!image}
                    className="w-full p-3 bg-green-600 text-white rounded hover:bg-green-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-semibold"
                  >
                    📖 Download Combined Spread
                  </button>
                  <button
                    onClick={() => downloadSpreadPages('left')}
                    disabled={!image}
                    className="w-full p-3 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-semibold"
                  >
                    Download Left Page
                  </button>
                  <button
                    onClick={() => downloadSpreadPages('right')}
                    disabled={!image}
                    className="w-full p-3 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-semibold"
                  >
                    Download Right Page
                  </button>
                  <button
                    onClick={() => downloadSpreadPages('full')}
                    disabled={!image}
                    className="w-full p-3 bg-gray-600 text-white rounded hover:bg-gray-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-semibold text-sm"
                  >
                    Download Full Spread (uncut)
                  </button>
                </div>
              ) : (
                <button
                  onClick={downloadImage}
                  disabled={!image}
                  className="w-full p-3 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-semibold"
                >
                  Download Image
                </button>
              )}
            </div>
          </div>

          {/* Preview */}
          <div className="lg:col-span-2">
            <div className="bg-white p-6 rounded-lg shadow">
              <h2 className="text-xl font-semibold mb-4">🖼️ Preview</h2>
              <div className="border-2 border-dashed border-gray-300 rounded-lg p-4 bg-gray-50 overflow-auto max-h-[800px]">
                {!image ? (
                  <div className="text-center text-gray-500 py-20">
                    Upload an image to get started
                  </div>
                ) : (
                  <>
                    <div className="mb-4">
                      <div className="text-sm font-medium text-gray-700 mb-2">
                        {mode === 'spread' ? 'Full Spread' : 'Output'}
                      </div>
                      <canvas
                        ref={canvasRef}
                        className="mx-auto shadow-lg"
                        style={{ maxWidth: '100%', height: 'auto' }}
                      />
                    </div>
                    
                    {mode === 'spread' && (
                      <div className="space-y-4 mt-6">
                        <div className="border-t pt-4">
                          <div className="text-sm font-medium text-gray-700 mb-2">📖 Combined Spread (as printed)</div>
                          <canvas
                            ref={combinedSpreadCanvasRef}
                            className="mx-auto shadow-lg border-2 border-gray-400"
                            style={{ maxWidth: '100%', height: 'auto' }}
                          />
                          <p className="text-xs text-gray-600 text-center mt-2">
                            Dark gap represents the binding
                          </p>
                        </div>
                        
                        <div className="border-t pt-4">
                          <div className="text-sm font-medium text-gray-700 mb-2">Left Page</div>
                          <canvas
                            ref={leftPageCanvasRef}
                            className="mx-auto shadow-lg border-2 border-gray-300"
                            style={{ maxWidth: '100%', height: 'auto' }}
                          />
                          <p className="text-xs text-gray-600 text-center mt-2">
                            Red dashed line shows gutter overlap area
                          </p>
                        </div>
                        
                        <div className="border-t pt-4">
                          <div className="text-sm font-medium text-gray-700 mb-2">Right Page</div>
                          <canvas
                            ref={rightPageCanvasRef}
                            className="mx-auto shadow-lg border-2 border-gray-300"
                            style={{ maxWidth: '100%', height: 'auto' }}
                          />
                          <p className="text-xs text-gray-600 text-center mt-2">
                            Red dashed line shows gutter overlap area
                          </p>
                        </div>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}