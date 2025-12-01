import React from 'react';
import type { ImageSequenceItem, Asset } from '../../../api';

interface SequenceItemProps {
  item: ImageSequenceItem;
  asset?: Asset;
  projectId: string;
  index: number;
  onDelete?: () => void;
}

export const SequenceItem: React.FC<SequenceItemProps> = ({
  item,
  asset,
  projectId,
  index,
  onDelete,
}) => {
  if (item.is_gap) {
    return (
      <div className="aspect-square border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="text-gray-400 text-2xl mb-1">⏸</div>
          <div className="text-xs text-gray-500">Gap</div>
        </div>
      </div>
    );
  }

  if (!asset) {
    return (
      <div className="aspect-square border border-gray-200 rounded-lg flex items-center justify-center bg-gray-100">
        <div className="text-xs text-gray-500">Missing asset</div>
      </div>
    );
  }

  const imageUrl = asset.url ?? `/projects/${projectId}/images/${encodeURIComponent(asset.filename)}`;
  const imageUrlWithCache = `${imageUrl}?t=${new Date(asset.uploaded_at).getTime()}`;

  return (
    <div className="aspect-square border border-gray-200 rounded-lg overflow-hidden bg-gray-100 relative group">
      <img
        src={imageUrlWithCache}
        alt={asset.filename}
        className="w-full h-full object-cover"
      />
      <div className="absolute top-2 left-2 bg-black bg-opacity-50 text-white text-xs px-2 py-1 rounded">
        {index + 1}
      </div>
      {onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="absolute top-2 right-2 bg-red-500 hover:bg-red-600 text-white rounded-full w-6 h-6 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity text-xs"
          title="Delete item"
        >
          ×
        </button>
      )}
      <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-10 transition-opacity flex items-center justify-center">
        <div className="opacity-0 group-hover:opacity-100 transition-opacity text-white text-xs">
          Drag to reorder
        </div>
      </div>
    </div>
  );
};

