import React, { useMemo, useRef, useState } from 'react';
import {
  useDeleteImageMutation,
  useGetImagesQuery,
  useReorderImagesMutation,
  useUploadImagesMutation,
} from '../api';
import { ImgCell } from './ImgCell';
import { Button } from './ui';

export const ImageTray: React.FC<{ id: string }> = ({ id }) => {
  const { data, isLoading, refetch } = useGetImagesQuery({ id });
  const [uploadImages, { isLoading: isUploading }] = useUploadImagesMutation();
  const [deleteImage] = useDeleteImageMutation();
  const [reorderImages, { isLoading: isReordering }] = useReorderImagesMutation();
  const fileRef = useRef<HTMLInputElement>(null);
  const [order, setOrder] = useState<string[] | null>(null);
  const [dragIndex, setDragIndex] = useState<number | null>(null);
  const [isDragOverDropzone, setIsDragOverDropzone] = useState(false);

  const currentOrder = order ?? data?.order ?? [];
  const imagesById = useMemo(() => {
    const m = new Map<string, { id: string; name: string; width: number; height: number }>();
    if (data?.images) {
      for (const im of data.images) {
        m.set(im.id, im);
      }
    }
    return m;
  }, [data]);
  const orderedImages = currentOrder.map((i) => imagesById.get(i)).filter(Boolean) as Array<{
    id: string;
    name: string;
    width: number;
    height: number;
  }>;

  const onUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    const files = fileRef.current?.files;
    if (!files || files.length === 0) return;
    await uploadImages({ id, files }).unwrap();
    if (fileRef.current) fileRef.current.value = '';
    setOrder(null);
    refetch();
  };

  const onDelete = async (imageId: string) => {
    await deleteImage({ id, imageId }).unwrap();
    setOrder(null);
    refetch();
  };

  const move = (idx: number, dir: -1 | 1) => {
    const next = [...currentOrder];
    const j = idx + dir;
    if (j < 0 || j >= next.length) return;
    [next[idx], next[j]] = [next[j], next[idx]];
    setOrder(next);
  };

  const saveOrder = async () => {
    if (!order) return;
    await reorderImages({ id, order }).unwrap();
    setOrder(null);
    refetch();
  };

  const onItemDragStart = (idx: number, imageId: string) => (e: React.DragEvent) => {
    setDragIndex(idx);
    e.dataTransfer.effectAllowed = 'move';
    try {
      e.dataTransfer.setData('application/x-zine-image-id', imageId);
      e.dataTransfer.setData('text/plain', imageId);
    } catch {
      // ignore
    }
  };
  const onItemDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  };
  const onItemDrop = (targetIndex: number) => (e: React.DragEvent) => {
    e.preventDefault();
    if (dragIndex === null || dragIndex === targetIndex) return;
    const next = [...currentOrder];
    const [moved] = next.splice(dragIndex, 1);
    next.splice(targetIndex, 0, moved);
    setOrder(next);
    setDragIndex(null);
  };

  const onDropzoneDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOverDropzone(true);
  };
  const onDropzoneDragLeave = () => setIsDragOverDropzone(false);
  const onDropzoneDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOverDropzone(false);
    const files = Array.from(e.dataTransfer.files || []);
    if (files.length === 0) return;
    const pngs = files.filter(
      (f) => f.type === 'image/png' || f.name.toLowerCase().endsWith('.png'),
    );
    if (pngs.length === 0) return;
    await uploadImages({ id, files: pngs }).unwrap();
    setOrder(null);
    refetch();
  };

  return (
    <div className="space-y-6">
      {/* Upload Section */}
      <div>
        <div className="mb-4">
          <form onSubmit={onUpload} className="flex items-center gap-3">
            <input 
              ref={fileRef} 
              type="file" 
              accept="image/png" 
              multiple 
              className="file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-primary-50 file:text-primary-700 hover:file:bg-primary-100"
            />
            <Button type="submit" isLoading={isUploading} size="sm">
              Upload Images
            </Button>
          </form>
        </div>
        
        {/* Drop Zone */}
        <div
          onDragOver={onDropzoneDragOver}
          onDragLeave={onDropzoneDragLeave}
          onDrop={onDropzoneDrop}
          className={`drop-zone ${isDragOverDropzone ? 'drop-zone-active' : ''}`}
        >
          <svg className="w-8 h-8 mx-auto mb-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          <p className="text-sm text-gray-600">Drop PNG files here to upload</p>
          <p className="text-xs text-gray-400 mt-1">or click above to browse</p>
        </div>
      </div>

      {/* Images Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center py-8">
          <div className="w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full animate-spin mr-2"></div>
          <span className="text-gray-600">Loading images...</span>
        </div>
      ) : orderedImages.length === 0 ? (
        <div className="text-center py-8">
          <svg className="w-12 h-12 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <p className="text-gray-500 mb-2">No images uploaded yet</p>
          <p className="text-sm text-gray-400">Upload PNG images to get started</p>
        </div>
      ) : (
        <div>
          {/* Images List */}
          <div className="space-y-3">
            {orderedImages.map((im, i) => (
              <div
                key={im.id}
                className="flex items-center space-x-4 p-3 bg-white border border-gray-200 rounded-lg hover:border-gray-300 transition-colors duration-200"
              >
                {/* Order Number */}
                <div className="flex-shrink-0">
                  <span className="inline-flex items-center justify-center w-8 h-8 bg-primary-100 text-primary-700 text-sm font-medium rounded-full">
                    {i + 1}
                  </span>
                </div>

                {/* Image Preview */}
                <div
                  draggable
                  onDragStart={onItemDragStart(i, im.id)}
                  onDragOver={onItemDragOver}
                  onDrop={onItemDrop(i)}
                  className="flex-shrink-0 cursor-move hover:bg-gray-50 p-2 rounded-lg transition-colors duration-200"
                  title="Drag to reorder"
                >
                  <ImgCell
                    src={`/api/projects/${id}/images/${encodeURIComponent(im.id)}`}
                    alt={im.name}
                    w={im.width}
                    h={im.height}
                  />
                </div>

                {/* Image Info */}
                <div className="flex-grow min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">{im.name}</p>
                  <p className="text-xs text-gray-500">
                    {im.width} × {im.height} pixels
                  </p>
                </div>

                {/* Actions */}
                <div className="flex items-center space-x-2">
                  <button
                    type="button"
                    onClick={() => move(i, -1)}
                    disabled={i === 0}
                    className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Move up"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 14l5-5 5 5" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    onClick={() => move(i, +1)}
                    disabled={i === orderedImages.length - 1}
                    className="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Move down"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 10l-5 5-5-5" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    onClick={() => onDelete(im.id)}
                    className="p-1 text-gray-400 hover:text-red-600 transition-colors duration-200"
                    title="Delete image"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            ))}
          </div>

          {/* Actions */}
          <div className="flex items-center justify-between mt-6 pt-4 border-t border-gray-200">
            <div className="flex space-x-3">
              <Button variant="secondary" size="sm" onClick={() => refetch()}>
                Refresh
              </Button>
              {order && (
                <Button 
                  size="sm" 
                  onClick={saveOrder} 
                  isLoading={isReordering}
                >
                  Save Order
                </Button>
              )}
            </div>
            <p className="text-sm text-gray-500">
              {orderedImages.length} image{orderedImages.length !== 1 ? 's' : ''}
            </p>
          </div>
        </div>
      )}
    </div>
  );
};
