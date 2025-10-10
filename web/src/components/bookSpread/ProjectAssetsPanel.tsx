import React, { useRef } from 'react';
import { useUploadAssetsMutation } from '../../api';

export interface AssetSummary {
  id: string;
  name: string;
  width: number;
  height: number;
  src: string;
  uploadedPath: string;
}

interface ProjectAssetsPanelProps {
  projectId: string | null;
  assets: AssetSummary[];
  selectedAssetId: string | null;
  onSelectAsset: (asset: AssetSummary) => void;
}

export const ProjectAssetsPanel: React.FC<ProjectAssetsPanelProps> = ({
  projectId,
  assets,
  selectedAssetId,
  onSelectAsset,
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploadAssets, { isLoading: isUploading }] = useUploadAssetsMutation();

  const handleFiles = async (files: FileList | File[]) => {
    if (!projectId) return;
    const fileArray = Array.from(files);
    if (fileArray.length === 0) return;

    try {
      const uploaded = await uploadAssets({ projectId, files: fileArray }).unwrap();
      if (uploaded.length) {
        const latest = uploaded[uploaded.length - 1];
        const summary: AssetSummary = {
          id: latest.id,
          name: latest.filename,
          width: latest.width,
          height: latest.height,
          src:
            (latest.url && `${latest.url}?t=${Date.now()}`) ||
            `/projects/${projectId}/images/${latest.filename}?t=${Date.now()}`,
          uploadedPath: latest.url ?? `/projects/${projectId}/images/${latest.filename}`,
        };
        onSelectAsset(summary);
      }
    } catch (error) {
      console.error('Failed to upload images', error);
    }
  };

  const handleDrop: React.DragEventHandler<HTMLDivElement> = async (event) => {
    event.preventDefault();
    if (!projectId) return;
    const files = event.dataTransfer.files;
    await handleFiles(files);
  };

  const handleSelect = () => {
    if (!projectId || isUploading) return;
    fileInputRef.current?.click();
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow space-y-4">
      <div>
        <h3 className="text-lg font-semibold">📁 Project Assets</h3>
        <p className="text-sm text-gray-500">Upload images or pick an existing asset to drive previews.</p>
      </div>

      <div
        className={`border-2 border-dashed rounded-lg p-6 text-center transition-colors cursor-pointer ${
          projectId ? 'border-gray-300 hover:border-primary-400' : 'border-gray-200 cursor-not-allowed'
        } ${isUploading ? 'opacity-60 pointer-events-none' : ''}`}
        onDragOver={(event) => event.preventDefault()}
        onDrop={handleDrop}
        onClick={handleSelect}
      >
        <div className="text-3xl mb-2">{isUploading ? '⏳' : '⬆️'}</div>
        <p className="text-sm text-gray-600">
          {projectId ? (isUploading ? 'Uploading…' : 'Drop images here or click to select') : 'Select a project to enable uploads'}
        </p>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/png"
          multiple
          hidden
          onChange={(event) => {
            if (event.target.files) {
              void handleFiles(event.target.files);
              event.target.value = '';
            }
          }}
          disabled={!projectId || isUploading}
        />
      </div>

      <div>
        <h4 className="text-sm font-semibold text-gray-700 mb-2">Project library</h4>
        {projectId ? (
          assets.length ? (
            <div className="grid grid-cols-2 gap-3">
              {assets.map((asset) => (
                <button
                  key={asset.id}
                  type="button"
                  onClick={() => onSelectAsset(asset)}
                  className={`border rounded-lg overflow-hidden flex flex-col items-center p-2 text-sm transition-colors ${
                    asset.id === selectedAssetId ? 'border-primary-500 ring-2 ring-primary-200' : 'border-gray-200 hover:border-primary-300'
                  }`}
                >
                  <div className="w-full h-24 bg-gray-100 flex items-center justify-center overflow-hidden mb-2">
                    <img
                      src={asset.src}
                      alt={asset.name}
                      className="object-contain max-h-full"
                    />
                  </div>
                  <div className="text-gray-800 truncate w-full text-left">{asset.name}</div>
                  <div className="text-xs text-gray-500 w-full text-left">
                    {asset.width} × {asset.height} px
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500">No images uploaded yet.</p>
          )
        ) : (
          <p className="text-sm text-gray-400">Choose a project to browse its assets.</p>
        )}
      </div>
    </div>
  );
};
