import React, { useMemo, useState } from 'react';
import { useDeleteAssetMutation, useGetAssetsQuery, type Asset } from '../../api';
import { Button, Card, CardBody, CardHeader } from '../../components/ui';
import { ProjectAssetsPanel, type AssetSummary } from '../../components/ProjectAssetsPanel';

interface AssetsTabProps {
  projectId: string;
}

export const AssetsTab: React.FC<AssetsTabProps> = ({ projectId }) => {
  const assetsQuery = useGetAssetsQuery({ projectId }, { skip: !projectId });
  const [deleteAsset] = useDeleteAssetMutation();
  const [selectedAssetId, setSelectedAssetId] = useState<string | null>(null);

  const assets = useMemo<AssetSummary[]>(() => {
    if (!assetsQuery.data || !projectId) return [];
    return assetsQuery.data.map((asset) => {
      const base =
        asset.url ?? `/projects/${projectId}/images/${encodeURIComponent(asset.filename)}`;
      const bust = asset.uploaded_at
        ? new Date(asset.uploaded_at).getTime()
        : Date.now();
      return {
        id: asset.id,
        name: asset.filename,
        width: asset.width,
        height: asset.height,
        src: `${base}?t=${bust}`,
        uploadedPath: base,
      };
    });
  }, [assetsQuery.data, projectId]);

  const assetLookup = useMemo(() => {
    const map = new Map<string, AssetSummary>();
    for (const asset of assets) map.set(asset.id, asset);
    return map;
  }, [assets]);

  const selectedAsset = selectedAssetId ? assetLookup.get(selectedAssetId) : undefined;

  const handleDeleteAsset = async (assetId: string) => {
    if (!projectId) return;
    const asset = assetLookup.get(assetId);
    if (
      window.confirm(
        `Delete asset "${asset?.name ?? assetId}"? This will remove it from sequences as well.`
      )
    ) {
      await deleteAsset({ assetId, projectId }).unwrap();
      if (selectedAssetId === assetId) setSelectedAssetId(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header with actions */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">
            Assets ({assets.length} image{assets.length !== 1 ? 's' : ''})
          </h2>
          <p className="text-sm text-gray-500 mt-1">
            Upload and manage your project images
          </p>
        </div>
      </div>

      {/* Main Content */}
      <Card>
        <CardHeader>
          <h3 className="text-lg font-semibold text-gray-900">Image Gallery</h3>
        </CardHeader>
        <CardBody>
          <ProjectAssetsPanel
            projectId={projectId}
            assets={assets}
            selectedAssetId={selectedAssetId}
            onSelectAsset={(asset) => setSelectedAssetId(asset.id)}
            onAssetDragStart={(asset) => {
              // Drag functionality for sequences tab
            }}
            onAssetDragEnd={() => {
              // Drag functionality for sequences tab
            }}
          />
        </CardBody>
      </Card>

      {/* Selected Asset Details Panel */}
      {selectedAsset && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900">Selected Asset Details</h3>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setSelectedAssetId(null)}
              >
                ✕ Close
              </Button>
            </div>
          </CardHeader>
          <CardBody>
            <div className="grid md:grid-cols-2 gap-6">
              {/* Preview */}
              <div className="flex items-center justify-center bg-gray-50 rounded-lg p-4 min-h-[300px]">
                <img
                  src={selectedAsset.src}
                  alt={selectedAsset.name}
                  className="max-w-full max-h-[400px] object-contain"
                />
              </div>

              {/* Details */}
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium text-gray-700">Filename</label>
                  <p className="text-base text-gray-900 break-all">{selectedAsset.name}</p>
                </div>

                <div>
                  <label className="text-sm font-medium text-gray-700">Dimensions</label>
                  <p className="text-base text-gray-900">
                    {selectedAsset.width} × {selectedAsset.height} px
                  </p>
                  <p className="text-sm text-gray-500">
                    {((selectedAsset.width * selectedAsset.height) / 1000000).toFixed(1)} MP
                  </p>
                </div>

                <div>
                  <label className="text-sm font-medium text-gray-700">Aspect Ratio</label>
                  <p className="text-base text-gray-900">
                    {(selectedAsset.width / selectedAsset.height).toFixed(2)} : 1
                  </p>
                </div>

                <div className="pt-4 space-y-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    className="w-full"
                    onClick={() => window.open(selectedAsset.uploadedPath, '_blank')}
                  >
                    View Full Size
                  </Button>
                  <Button
                    variant="danger"
                    size="sm"
                    className="w-full"
                    onClick={() => handleDeleteAsset(selectedAsset.id)}
                  >
                    Delete Asset
                  </Button>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>
      )}
    </div>
  );
};

