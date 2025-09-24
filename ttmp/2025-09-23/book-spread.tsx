import React, { useState, useRef, useCallback } from 'react';

const PhotobookSpreadDesigner = () => {
  const [image, setImage] = useState(null);
  const [paperSize, setPaperSize] = useState('8x10');
  const [isSpread, setIsSpread] = useState(false);
  const [margins, setMargins] = useState({ top: 0.5, right: 0.5, bottom: 0.5, left: 0.5 });
  const [orientation, setOrientation] = useState('portrait');
  const [cropRatio, setCropRatio] = useState('original');
  const [cropToFill, setCropToFill] = useState(false);
  const [imageScale, setImageScale] = useState(1);
  const [imagePosition, setImagePosition] = useState({ x: 0, y: 0 });
  const [dpi, setDpi] = useState(300);
  const canvasRef = useRef(null);
  const fileInputRef = useRef(null);

  // Paper size definitions (in inches)
  const paperSizes = {
    '4x6': { width: 4, height: 6 },
    '5x7': { width: 5, height: 7 },
    '8x10': { width: 8, height: 10 },
    '11x14': { width: 11, height: 14 },
    '12x12': { width: 12, height: 12 },
    '8.5x11': { width: 8.5, height: 11 },
    'A4': { width: 8.27, height: 11.69 }
  };

  // Crop ratio definitions
  const cropRatios = {
    'original': null,
    '1:1': 1,
    '2:3': 2 / 3,
    '3:4': 3 / 4,
    '4:5': 4 / 5,
    '5:7': 5 / 7,
    '16:9': 16 / 9
  };

  const getCurrentDimensions = () => {
    const base = paperSizes[paperSize];
    let width = base.width;
    let height = base.height;

    // Swap dimensions for landscape
    if (orientation === 'landscape') {
      [width, height] = [height, width];
    }

    // Apply spread multiplier
    if (isSpread) {
      width = width * 2;
    }

    return { width, height };
  };

  const handleImageDrop = useCallback((e) => {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file && file.type.startsWith('image/')) {
      loadImage(file);
    }
  }, []);

  const handleImageSelect = (e) => {
    const file = e.target.files[0];
    if (file) {
      loadImage(file);
    }
  };

  const loadImage = (file) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        setImage(img);
        setImageScale(1);
        setImagePosition({ x: 0, y: 0 });
        // Store file info for display
        img.fileSize = file.size;
        img.fileName = file.name;
      };
      img.src = e.target.result;
    };
    reader.readAsDataURL(file);
  };

  const renderPreview = () => {
    if (!image) return null;

    const { width, height } = getCurrentDimensions();
    const previewWidth = Math.min(600, width * 50); // Scale for preview
    const previewHeight = (height / width) * previewWidth;

    // Calculate content area after margins
    const contentWidth = previewWidth * (1 - (margins.left + margins.right) / width);
    const contentHeight = previewHeight * (1 - (margins.top + margins.bottom) / height);
    const contentX = previewWidth * (margins.left / width);
    const contentY = previewHeight * (margins.top / height);

    // Calculate gutter split for spreads
    const hasGutter = isSpread && gutterMargin > 0;
    const gutterWidth = hasGutter ? previewWidth * (gutterMargin / width) : 0;
    const leftPanelWidth = hasGutter ? (contentWidth - gutterWidth) / 2 : contentWidth;
    const rightPanelWidth = hasGutter ? (contentWidth - gutterWidth) / 2 : 0;
    const rightPanelX = hasGutter ? contentX + leftPanelWidth + gutterWidth : 0;

    // Calculate image dimensions based on crop ratio and crop-to-fill mode
    let imgWidth = image.width;
    let imgHeight = image.height;
    let sourceX = 0;
    let sourceY = 0;
    let cropAdjustmentX = 0;
    let cropAdjustmentY = 0;

    if (cropRatio !== 'original' && cropRatios[cropRatio]) {
      const targetRatio = cropRatios[cropRatio];
      const currentRatio = imgWidth / imgHeight;

      if (currentRatio > targetRatio) {
        // Image is wider than target ratio - crop width
        const newWidth = imgHeight * targetRatio;
        sourceX = (imgWidth - newWidth) / 2;
        imgWidth = newWidth;
      } else {
        // Image is taller than target ratio - crop height
        const newHeight = imgWidth / targetRatio;
        sourceY = (imgHeight - newHeight) / 2;
        imgHeight = newHeight;
      }
    } else if (cropToFill) {
      // When crop to fill is enabled, crop the image to match content area ratio
      const contentRatio = contentWidth / contentHeight;
      const currentRatio = imgWidth / imgHeight;

      if (currentRatio > contentRatio) {
        // Image is wider than content area - crop width, allow horizontal adjustment
        const newWidth = imgHeight * contentRatio;
        const maxCropAdjustment = (imgWidth - newWidth) / 2;
        cropAdjustmentX = (imagePosition.x / 100) * maxCropAdjustment;
        sourceX = (imgWidth - newWidth) / 2 + cropAdjustmentX;
        imgWidth = newWidth;
      } else {
        // Image is taller than content area - crop height, allow vertical adjustment
        const newHeight = imgWidth / contentRatio;
        const maxCropAdjustment = (imgHeight - newHeight) / 2;
        cropAdjustmentY = (imagePosition.y / 100) * maxCropAdjustment;
        sourceY = (imgHeight - newHeight) / 2 + cropAdjustmentY;
        imgHeight = newHeight;
      }
    }

    // Scale image to fit content area
    const scaleX = contentWidth / imgWidth;
    const scaleY = contentHeight / imgHeight;
    let baseScale, finalScale, displayWidth, displayHeight;

    if (cropToFill) {
      // Scale to fill - use the larger scale to fill the entire content area
      baseScale = Math.max(scaleX, scaleY);
      finalScale = baseScale; // Ignore user scale when crop to fill is enabled
      displayWidth = imgWidth * finalScale;
      displayHeight = imgHeight * finalScale;
    } else {
      // Scale to fit - use the smaller scale to fit within content area
      baseScale = Math.min(scaleX, scaleY);
      finalScale = baseScale * imageScale;
      displayWidth = imgWidth * finalScale;
      displayHeight = imgHeight * finalScale;
    }

    return (
      <div className="border-2 border-gray-300 relative bg-white"
        style={{ width: previewWidth, height: previewHeight }}>
        {/* Margin guides */}
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute border border-red-200"
            style={{
              left: contentX,
              top: contentY,
              width: contentWidth,
              height: contentHeight
            }} />
          {/* Gutter guide for spreads */}
          {hasGutter && (
            <>
              <div className="absolute border-2 border-blue-300 bg-blue-100 bg-opacity-30"
                style={{
                  left: contentX + leftPanelWidth,
                  top: contentY,
                  width: gutterWidth,
                  height: contentHeight
                }} />
              <div className="absolute text-blue-600 text-xs font-semibold pointer-events-none"
                style={{
                  left: contentX + leftPanelWidth + gutterWidth / 2 - 20,
                  top: contentY + contentHeight / 2 - 8
                }}>
                GUTTER
              </div>
            </>
          )}
        </div>

        {/* Image - Left panel or full width */}
        <div className="absolute overflow-hidden"
          style={{
            left: contentX,
            top: contentY,
            width: hasGutter ? leftPanelWidth : contentWidth,
            height: contentHeight
          }}>
          <img
            src={image.src}
            className="absolute"
            style={{
              width: displayWidth,
              height: displayHeight,
              left: hasGutter ?
                (leftPanelWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x)) :
                (contentWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x)),
              top: contentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : imagePosition.y),
              objectFit: 'cover'
            }}
            alt="Preview"
          />
        </div>

        {/* Image - Right panel for gutter spreads */}
        {hasGutter && (
          <div className="absolute overflow-hidden"
            style={{
              left: rightPanelX,
              top: contentY,
              width: rightPanelWidth,
              height: contentHeight
            }}>
            <img
              src={image.src}
              className="absolute"
              style={{
                width: displayWidth,
                height: displayHeight,
                left: rightPanelWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : imagePosition.x) - (leftPanelWidth + gutterWidth),
                top: contentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : imagePosition.y),
                objectFit: 'cover'
              }}
              alt="Preview"
            />
          </div>
        )}

        {/* Paper info overlay */}
        <div className="absolute top-2 right-2 bg-black bg-opacity-50 text-white text-xs px-2 py-1 rounded">
          {width}"×{height}" {orientation} {isSpread ? 'spread' : 'single'}
        </div>
      </div>
    );
  };

  const exportImage = () => {
    if (!image) return;

    const { width, height } = getCurrentDimensions();
    const pixelWidth = width * dpi;
    const pixelHeight = height * dpi;

    const canvas = document.createElement('canvas');
    canvas.width = pixelWidth;
    canvas.height = pixelHeight;
    const ctx = canvas.getContext('2d');

    // Fill with white background
    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, pixelWidth, pixelHeight);

    // Calculate content area
    const contentX = (margins.left * dpi);
    const contentY = (margins.top * dpi);
    const baseContentWidth = pixelWidth - (margins.left + margins.right) * dpi;
    const baseContentHeight = pixelHeight - (margins.top + margins.bottom) * dpi;

    // Account for gutter in spreads
    const hasGutter = isSpread && gutterMargin > 0;
    const gutterPixelWidth = hasGutter ? gutterMargin * dpi : 0;
    const effectiveContentWidth = hasGutter ? baseContentWidth - gutterPixelWidth : baseContentWidth;
    const leftPanelWidth = hasGutter ? effectiveContentWidth / 2 : baseContentWidth;
    const rightPanelWidth = hasGutter ? effectiveContentWidth / 2 : 0;
    const rightPanelX = hasGutter ? contentX + leftPanelWidth + gutterPixelWidth : 0;

    // Calculate image dimensions
    let imgWidth = image.width;
    let imgHeight = image.height;
    let sourceX = 0;
    let sourceY = 0;
    let cropAdjustmentX = 0;
    let cropAdjustmentY = 0;

    if (cropRatio !== 'original' && cropRatios[cropRatio]) {
      const targetRatio = cropRatios[cropRatio];
      const currentRatio = imgWidth / imgHeight;

      if (currentRatio > targetRatio) {
        // Image is wider than target ratio - crop width
        const newWidth = imgHeight * targetRatio;
        sourceX = (imgWidth - newWidth) / 2;
        imgWidth = newWidth;
      } else {
        // Image is taller than target ratio - crop height
        const newHeight = imgWidth / targetRatio;
        sourceY = (imgHeight - newHeight) / 2;
        imgHeight = newHeight;
      }
    } else if (cropToFill) {
      // When crop to fill is enabled, crop the image to match content area ratio
      const contentRatio = contentWidth / contentHeight;
      const currentRatio = imgWidth / imgHeight;

      if (currentRatio > contentRatio) {
        // Image is wider than content area - crop width, allow horizontal adjustment
        const newWidth = imgHeight * contentRatio;
        const maxCropAdjustment = (imgWidth - newWidth) / 2;
        cropAdjustmentX = (imagePosition.x / 100) * maxCropAdjustment;
        sourceX = (imgWidth - newWidth) / 2 + cropAdjustmentX;
        imgWidth = newWidth;
      } else {
        // Image is taller than content area - crop height, allow vertical adjustment
        const newHeight = imgWidth / contentRatio;
        const maxCropAdjustment = (imgHeight - newHeight) / 2;
        cropAdjustmentY = (imagePosition.y / 100) * maxCropAdjustment;
        sourceY = (imgHeight - newHeight) / 2 + cropAdjustmentY;
        imgHeight = newHeight;
      }
    }

    const scaleX = contentWidth / imgWidth;
    const scaleY = contentHeight / imgHeight;
    let baseScale, finalScale, displayWidth, displayHeight;

    if (cropToFill) {
      // Scale to fill - image should exactly fill the content area
      baseScale = 1; // Since we've already cropped to the right ratio
      finalScale = Math.max(scaleX, scaleY); // Ensure it fills completely
      displayWidth = imgWidth * finalScale;
      displayHeight = imgHeight * finalScale;
    } else {
      // Scale to fit - use the smaller scale to fit within content area
      baseScale = Math.min(scaleX, scaleY);
      finalScale = baseScale * imageScale;
      displayWidth = imgWidth * finalScale;
      displayHeight = imgHeight * finalScale;
    }

    const drawX = contentX + contentWidth / 2 - displayWidth / 2 + (cropToFill ? 0 : (imagePosition.x * dpi / 50));
    const drawY = contentY + contentHeight / 2 - displayHeight / 2 + (cropToFill ? 0 : (imagePosition.y * dpi / 50));

    // Clip to content area
    ctx.save();
    ctx.rect(contentX, contentY, contentWidth, contentHeight);
    ctx.clip();

    // Draw image with proper cropping
    if (cropRatio !== 'original' && cropRatios[cropRatio]) {
      ctx.drawImage(image, sourceX, sourceY, imgWidth, imgHeight, drawX, drawY, displayWidth, displayHeight);
    } else if (cropToFill) {
      ctx.drawImage(image, sourceX, sourceY, imgWidth, imgHeight, drawX, drawY, displayWidth, displayHeight);
    } else {
      ctx.drawImage(image, drawX, drawY, displayWidth, displayHeight);
    }

    ctx.restore();

    // Download
    const link = document.createElement('a');
    link.download = `photobook-spread-${width}x${height}-${dpi}dpi.png`;
    link.href = canvas.toDataURL('image/png');
    link.click();
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">📖 Photobook Spread Designer</h1>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Controls Panel */}
          <div className="space-y-6">
            {/* Image Upload */}
            <div className="bg-white p-6 rounded-lg shadow">
              <h3 className="text-lg font-semibold mb-4">📷 Image Upload</h3>

              <div
                className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center cursor-pointer hover:border-blue-400 transition-colors"
                onDrop={handleImageDrop}
                onDragOver={(e) => e.preventDefault()}
                onClick={() => fileInputRef.current?.click()}
              >
                <div className="text-4xl mb-2">📁</div>
                <p className="text-gray-600">Drop an image here or click to select</p>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleImageSelect}
                  className="hidden"
                />
              </div>
            </div>

            {/* Paper Settings */}
            <div className="bg-white p-6 rounded-lg shadow">
              <h3 className="text-lg font-semibold mb-4">📏 Paper Settings</h3>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Paper Size</label>
                  <select
                    value={paperSize}
                    onChange={(e) => setPaperSize(e.target.value)}
                    className="w-full p-2 border border-gray-300 rounded-md"
                  >
                    {Object.entries(paperSizes).map(([key, size]) => (
                      <option key={key} value={key}>
                        {key} ({size.width}" × {size.height}")
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Orientation</label>
                  <select
                    value={orientation}
                    onChange={(e) => setOrientation(e.target.value)}
                    className="w-full p-2 border border-gray-300 rounded-md"
                  >
                    <option value="portrait">Portrait</option>
                    <option value="landscape">Landscape</option>
                  </select>
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="spread"
                    checked={isSpread}
                    onChange={(e) => setIsSpread(e.target.checked)}
                    className="mr-2"
                  />
                  <label htmlFor="spread" className="text-sm font-medium text-gray-700">
                    Double-page spread
                  </label>
                </div>

                {isSpread && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Gutter/Spine Margin (inches)
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="2"
                      value={gutterMargin}
                      onChange={(e) => setGutterMargin(Number(e.target.value))}
                      className="w-full p-2 border border-gray-300 rounded-md"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Space in the center where binding occurs
                    </p>
                  </div>
                )}

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Print DPI</label>
                  <select
                    value={dpi}
                    onChange={(e) => setDpi(Number(e.target.value))}
                    className="w-full p-2 border border-gray-300 rounded-md"
                  >
                    <option value={150}>150 DPI</option>
                    <option value={300}>300 DPI</option>
                    <option value={600}>600 DPI</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Margins */}
            <div className="bg-white p-6 rounded-lg shadow">
              <h3 className="text-lg font-semibold mb-4">📐 Margins (inches)</h3>

              <div className="grid grid-cols-2 gap-4">
                {Object.entries(margins).map(([key, value]) => (
                  <div key={key}>
                    <label className="block text-sm font-medium text-gray-700 mb-1 capitalize">
                      {key}
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="2"
                      value={value}
                      onChange={(e) => setMargins(prev => ({ ...prev, [key]: Number(e.target.value) }))}
                      className="w-full p-2 border border-gray-300 rounded-md"
                    />
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Preview Panel */}
          <div className="lg:col-span-2">
            <div className="bg-white p-6 rounded-lg shadow">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold">🖼️ Preview</h3>
                {image && (
                  <button
                    onClick={exportImage}
                    className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
                  >
                    💾 Export Image
                  </button>
                )}
              </div>

              {image ? (
                <div className="space-y-6">
                  {/* Image Controls */}
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Crop Ratio</label>
                        <select
                          value={cropRatio}
                          onChange={(e) => setCropRatio(e.target.value)}
                          className="w-full p-2 border border-gray-300 rounded-md text-sm"
                        >
                          {Object.keys(cropRatios).map(ratio => (
                            <option key={ratio} value={ratio}>{ratio}</option>
                          ))}
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Fill Mode</label>
                        <div className="flex items-center h-10">
                          <input
                            type="checkbox"
                            id="cropToFill"
                            checked={cropToFill}
                            onChange={(e) => setCropToFill(e.target.checked)}
                            className="mr-2"
                          />
                          <label htmlFor="cropToFill" className="text-sm font-medium text-gray-700">
                            Crop to fill
                          </label>
                        </div>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Scale ({Math.round(imageScale * 100)}%)
                        </label>
                        <input
                          type="range"
                          min="0.1"
                          max="3"
                          step="0.01"
                          value={imageScale}
                          onChange={(e) => setImageScale(Number(e.target.value))}
                          className="w-full"
                          disabled={cropToFill}
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Position {cropToFill && '(crop adjust)'}
                        </label>
                        {(() => {
                          if (!image || !cropToFill) {
                            // Normal mode - both sliders enabled
                            return (
                              <div className="flex gap-2">
                                <input
                                  type="range"
                                  min="-100"
                                  max="100"
                                  value={imagePosition.x}
                                  onChange={(e) => setImagePosition(prev => ({ ...prev, x: Number(e.target.value) }))}
                                  className="w-full"
                                  title="Horizontal"
                                />
                                <input
                                  type="range"
                                  min="-100"
                                  max="100"
                                  value={imagePosition.y}
                                  onChange={(e) => setImagePosition(prev => ({ ...prev, y: Number(e.target.value) }))}
                                  className="w-full"
                                  title="Vertical"
                                />
                              </div>
                            );
                          }

                          // Crop to fill mode - determine which axis to enable
                          const { width, height } = getCurrentDimensions();
                          const totalMarginWidth = margins.left + margins.right;
                          const totalMarginHeight = margins.top + margins.bottom;
                          const baseContentWidth = width - totalMarginWidth;
                          const baseContentHeight = height - totalMarginHeight;

                          // Account for gutter in spread calculations
                          const effectiveContentWidth = (isSpread && gutterMargin > 0) ?
                            baseContentWidth - gutterMargin : baseContentWidth;

                          const contentRatio = effectiveContentWidth / baseContentHeight;
                          const imageRatio = image.width / image.height;

                          if (imageRatio > contentRatio) {
                            // Image is wider - we cropped horizontally, allow horizontal adjustment
                            return (
                              <div className="flex gap-2">
                                <input
                                  type="range"
                                  min="-100"
                                  max="100"
                                  value={imagePosition.x}
                                  onChange={(e) => setImagePosition(prev => ({ ...prev, x: Number(e.target.value) }))}
                                  className="w-full"
                                  title="Horizontal crop adjustment"
                                />
                                <input
                                  type="range"
                                  min="0"
                                  max="0"
                                  value="0"
                                  className="w-full opacity-30"
                                  title="Vertical (disabled)"
                                  disabled
                                />
                              </div>
                            );
                          } else {
                            // Image is taller - we cropped vertically, allow vertical adjustment
                            return (
                              <div className="flex gap-2">
                                <input
                                  type="range"
                                  min="0"
                                  max="0"
                                  value="0"
                                  className="w-full opacity-30"
                                  title="Horizontal (disabled)"
                                  disabled
                                />
                                <input
                                  type="range"
                                  min="-100"
                                  max="100"
                                  value={imagePosition.y}
                                  onChange={(e) => setImagePosition(prev => ({ ...prev, y: Number(e.target.value) }))}
                                  className="w-full"
                                  title="Vertical crop adjustment"
                                />
                              </div>
                            );
                          }
                        })()}
                      </div>
                    </div>
                  </div>

                  {/* Preview */}
                  <div className="flex justify-center">
                    {renderPreview()}
                  </div>

                  {/* Image Information */}
                  <div className="bg-green-50 p-4 rounded-lg">
                    <h4 className="font-medium text-green-900 mb-2">🖼️ Image Information</h4>
                    <div className="text-sm text-green-800 space-y-1">
                      <p>File: {image.fileName || 'Unknown'}</p>
                      <p>Original size: {image.width} × {image.height} pixels</p>
                      <p>File size: {image.fileSize ? `${(image.fileSize / (1024 * 1024)).toFixed(2)} MB` : 'Unknown'}</p>
                      <p>Aspect ratio: {(image.width / image.height).toFixed(2)}:1</p>
                    </div>
                  </div>

                  {/* Export Info */}
                  <div className="bg-blue-50 p-4 rounded-lg">
                    <h4 className="font-medium text-blue-900 mb-2">📋 Export Information</h4>
                    <div className="text-sm text-blue-800 space-y-1">
                      {(() => {
                        const { width, height } = getCurrentDimensions();
                        const pixelWidth = width * dpi;
                        const pixelHeight = height * dpi;
                        const fileSize = Math.round((pixelWidth * pixelHeight * 3) / (1024 * 1024));
                        return (
                          <>
                            <p>Dimensions: {width}" × {height}" ({isSpread ? 'spread' : 'single page'})</p>
                            <p>Resolution: {pixelWidth} × {pixelHeight} pixels at {dpi} DPI</p>
                            <p>Estimated file size: ~{fileSize} MB</p>
                          </>
                        );
                      })()}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-center text-gray-500 py-12">
                  <div className="text-6xl mb-4">🖼️</div>
                  <p>Upload an image to start designing your photobook spread</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PhotobookSpreadDesigner;
