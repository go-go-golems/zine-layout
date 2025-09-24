import React, { useRef, useCallback } from 'react';
import { useAppDispatch } from '../../hooks/redux';
import { setImage } from '../../store/bookSpreadSlice';
import { loadImageFromFile } from '../../utils/bookSpreadUtils';
import { useUploadFileMutation } from '../../api';

export const ImageUploadSection: React.FC = () => {
  const dispatch = useAppDispatch();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploadFile, { isLoading: isUploading }] = useUploadFileMutation();

  const processFile = useCallback(async (file: File) => {
    if (!file.type.startsWith('image/')) return;
    
    try {
      // Upload to backend
      const uploadResult = await uploadFile(file).unwrap();
      
      // Also load for preview
      const img = await loadImageFromFile(file);
      
      dispatch(setImage({
        src: img.src,
        width: img.width,
        height: img.height,
        fileSize: file.size,
        fileName: file.name,
        uploadedPath: uploadResult.url, // Store backend path
      }));
    } catch (error) {
      console.error('Failed to process image:', error);
    }
  }, [dispatch, uploadFile]);

  const handleImageDrop = useCallback(async (e: React.DragEvent) => {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file) {
      await processFile(file);
    }
  }, [processFile]);

  const handleImageSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      await processFile(file);
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h3 className="text-lg font-semibold mb-4">📷 Image Upload</h3>

      <div
        className={`border-2 border-dashed border-gray-300 rounded-lg p-8 text-center cursor-pointer hover:border-blue-400 transition-colors ${
          isUploading ? 'opacity-50 pointer-events-none' : ''
        }`}
        onDrop={handleImageDrop}
        onDragOver={(e) => e.preventDefault()}
        onClick={() => !isUploading && fileInputRef.current?.click()}
      >
        <div className="text-4xl mb-2">{isUploading ? '⏳' : '📁'}</div>
        <p className="text-gray-600">
          {isUploading ? 'Uploading...' : 'Drop an image here or click to select'}
        </p>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleImageSelect}
          className="hidden"
          disabled={isUploading}
        />
      </div>
    </div>
  );
};
