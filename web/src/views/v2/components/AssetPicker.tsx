import React, { useState } from 'react';
import { useGetAssetsQuery, type Asset } from '../../../api';
import { Button, Card, CardBody, CardHeader } from '../../../components/ui';

interface AssetPickerProps {
  projectId: string;
  isOpen: boolean;
  onClose: () => void;
  onSelect: (assetIds: string[]) => void;
  multiple?: boolean;
}

export const AssetPicker: React.FC<AssetPickerProps> = ({
  projectId,
  isOpen,
  onClose,
  onSelect,
  multiple = true,
}) => {
  const { data: assets = [], isLoading } = useGetAssetsQuery({ projectId });
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  if (!isOpen) return null;

  const handleToggleSelection = (assetId: string) => {
    if (multiple) {
      setSelectedIds((prev) => {
        const next = new Set(prev);
        if (next.has(assetId)) {
          next.delete(assetId);
        } else {
          next.add(assetId);
        }
        return next;
      });
    } else {
      setSelectedIds(new Set([assetId]));
    }
  };

  const handleConfirm = () => {
    if (selectedIds.size > 0) {
      onSelect(Array.from(selectedIds));
      setSelectedIds(new Set());
      onClose();
    }
  };

  const handleCancel = () => {
    setSelectedIds(new Set());
    onClose();
  };

  const getImageUrl = (asset: Asset): string => {
    return asset.url ?? `/projects/${projectId}/images/${encodeURIComponent(asset.filename)}`;
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50" onClick={handleCancel}>
      <Card className="w-full max-w-4xl max-h-[90vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold">Select Images</h2>
            <button
              onClick={handleCancel}
              className="text-gray-400 hover:text-gray-600 text-2xl leading-none"
            >
              ×
            </button>
          </div>
        </CardHeader>
        <CardBody className="flex-1 overflow-y-auto">
          {isLoading ? (
            <div className="text-center py-8 text-gray-500">Loading assets...</div>
          ) : assets.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <p>No assets available</p>
              <p className="text-sm mt-2">Upload images in the Assets tab first</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              {assets.map((asset) => {
                const isSelected = selectedIds.has(asset.id);
                return (
                  <button
                    key={asset.id}
                    type="button"
                    onClick={() => handleToggleSelection(asset.id)}
                    className={`relative border-2 rounded-lg overflow-hidden transition-all ${
                      isSelected
                        ? 'border-primary-500 ring-2 ring-primary-200'
                        : 'border-gray-200 hover:border-primary-300'
                    }`}
                  >
                    <div className="aspect-square bg-gray-100 flex items-center justify-center overflow-hidden">
                      <img
                        src={`${getImageUrl(asset)}?t=${new Date(asset.uploaded_at).getTime()}`}
                        alt={asset.filename}
                        className="w-full h-full object-cover"
                      />
                    </div>
                    {isSelected && (
                      <div className="absolute top-2 right-2 bg-primary-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm">
                        ✓
                      </div>
                    )}
                    <div className="p-2 text-left">
                      <div className="text-xs text-gray-800 truncate">{asset.filename}</div>
                      <div className="text-xs text-gray-500">
                        {asset.width} × {asset.height}
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </CardBody>
        <div className="p-4 border-t border-gray-200 flex items-center justify-between">
          <div className="text-sm text-gray-600">
            {selectedIds.size > 0
              ? `${selectedIds.size} image${selectedIds.size === 1 ? '' : 's'} selected`
              : 'Select images to add'}
          </div>
          <div className="flex gap-2">
            <Button onClick={handleCancel} variant="secondary" size="sm">
              Cancel
            </Button>
            <Button
              onClick={handleConfirm}
              disabled={selectedIds.size === 0}
              size="sm"
            >
              Add {selectedIds.size > 0 ? `(${selectedIds.size})` : ''}
            </Button>
          </div>
        </div>
      </Card>
    </div>
  );
};

