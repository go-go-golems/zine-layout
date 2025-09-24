import React, { useRef, useCallback } from 'react';
import { useAppDispatch } from '../../hooks/redux';
import { setImage } from '../../store/bookSpreadSlice';
import { loadImageFromFile } from '../../utils/bookSpreadUtils';

export const ImageUploadSection: React.FC = () => {
  const dispatch = useAppDispatch();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImageDrop = useCallback(async (e: React.DragEvent) => {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file && file.type.startsWith('image/')) {
      try {
        const img = await loadImageFromFile(file);
        dispatch(setImage({
          src: img.src,
          width: img.width,
          height: img.height,
          fileSize: file.size,
          fileName: file.name,
        }));
      } catch (error) {
        console.error('Failed to load image:', error);
      }
    }
  }, [dispatch]);

  const handleImageSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      try {
        const img = await loadImageFromFile(file);
        dispatch(setImage({
          src: img.src,
          width: img.width,
          height: img.height,
          fileSize: file.size,
          fileName: file.name,
        }));
      } catch (error) {
        console.error('Failed to load image:', error);
      }
    }
  };

  return (
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
  );
};
