import React, { useMemo, useState } from 'react';
import {
  useCreateLaidOutImageMutation,
  useDeleteLaidOutImageMutation,
  useGetAssetsQuery,
  useGetImageLayoutTemplatesQuery,
  useGetLaidOutImageQuery,
  useGetLaidOutImagesQuery,
  usePreviewLaidOutImageQuery,
  useUpdateLaidOutImageMutation,
  type Asset,
  type ImageLayoutTemplate,
  type LaidOutImage,
} from '../api';
import { Button, Card, CardBody, CardHeader, Input } from '../components/ui';

interface LaidOutImageViewerProps {
  projectId: string;
}

const jsonString = (value: unknown) => {
  if (!value) return '';
  try {
    return JSON.stringify(value, null, 2);
  } catch (err) {
    return '';
  }
};

const parseJSON = (value: string) => {
  if (!value.trim()) return undefined;
  const parsed = JSON.parse(value);
  if (typeof parsed !== 'object' || Array.isArray(parsed) || parsed === null) {
    throw new Error('Overrides JSON must be an object');
  }
  return parsed as Record<string, unknown>;
};

export const LaidOutImageViewer: React.FC<LaidOutImageViewerProps> = ({ projectId }) => {
  const assetsQuery = useGetAssetsQuery({ projectId }, { skip: !projectId });
  const templatesQuery = useGetImageLayoutTemplatesQuery({ projectId }, { skip: !projectId });
  const laidOutImagesQuery = useGetLaidOutImagesQuery({ projectId }, { skip: !projectId });

  const [selectedImageId, setSelectedImageId] = useState<string | null>(null);
  const selectedImageQuery = useGetLaidOutImageQuery({ id: selectedImageId ?? '' }, { skip: !selectedImageId });
  const previewQuery = usePreviewLaidOutImageQuery({ id: selectedImageId ?? '' }, { skip: !selectedImageId });

  const [createImage, createState] = useCreateLaidOutImageMutation();
  const [updateImage, updateState] = useUpdateLaidOutImageMutation();
  const [deleteImage, deleteState] = useDeleteLaidOutImageMutation();

  const [createAsset, setCreateAsset] = useState('');
  const [createTemplate, setCreateTemplate] = useState('');
  const [createOverrides, setCreateOverrides] = useState('');

  const [editTemplate, setEditTemplate] = useState('');
  const [editOverrides, setEditOverrides] = useState('');

  const assets = useMemo(() => assetsQuery.data ?? [], [assetsQuery.data]);
  const templates = useMemo(() => templatesQuery.data ?? [], [templatesQuery.data]);
  const laidOutImages = useMemo(() => laidOutImagesQuery.data ?? [], [laidOutImagesQuery.data]);

  const assetLookup = useMemo(() => {
    const map = new Map<string, Asset>();
    for (const asset of assets) map.set(asset.id, asset);
    return map;
  }, [assets]);

  const templateLookup = useMemo(() => {
    const map = new Map<string, ImageLayoutTemplate>();
    for (const tpl of templates) map.set(tpl.id, tpl);
    return map;
  }, [templates]);

  const handleCreate = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!createAsset || !createTemplate) {
      alert('Choose both asset and template');
      return;
    }
    try {
      const overrides = createOverrides.trim() ? parseJSON(createOverrides) : undefined;
      const created = await createImage({
        projectId,
        assetId: createAsset,
        templateId: createTemplate,
        overrides,
      }).unwrap();
      setSelectedImageId(created.id);
      setEditTemplate(created.template_id);
      setEditOverrides(jsonString(created.overrides));
      setCreateOverrides('');
    } catch (err) {
      alert((err as Error).message);
    }
  };

  const handleUpdate = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!selectedImageId) return;
    try {
      const overrides = editOverrides.trim() ? parseJSON(editOverrides) : undefined;
      const updated = await updateImage({
        id: selectedImageId,
        templateId: editTemplate || undefined,
        overrides,
      }).unwrap();
      setEditTemplate(updated.template_id);
      setEditOverrides(jsonString(updated.overrides));
    } catch (err) {
      alert((err as Error).message);
    }
  };

  const handleDelete = async (image: LaidOutImage) => {
    if (!window.confirm('Delete this laid-out image?')) return;
    await deleteImage({ id: image.id, projectId: projectId }).unwrap();
    if (selectedImageId === image.id) {
      setSelectedImageId(null);
      setEditTemplate('');
      setEditOverrides('');
    }
  };

  const handleSelectImage = (image: LaidOutImage) => {
    setSelectedImageId(image.id);
    setEditTemplate(image.template_id);
    setEditOverrides(jsonString(image.overrides));
  };

  return (
    <Card className="mt-8">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Laid-Out Images</h2>
            <p className="text-sm text-gray-500">
              Apply templates to assets and inspect the computed placement results.
            </p>
          </div>
        </div>
      </CardHeader>
      <CardBody className="space-y-8">
        <form onSubmit={handleCreate} className="grid gap-4 md:grid-cols-4 p-4 border border-gray-200 rounded-lg">
          <div className="md:col-span-1">
            <label className="block text-sm font-medium text-gray-700">Asset</label>
            <select
              value={createAsset}
              onChange={(e) => setCreateAsset(e.target.value)}
              className="w-full border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              <option value="">Select asset…</option>
              {assets.map((asset) => (
                <option key={asset.id} value={asset.id}>
                  {asset.filename}
                </option>
              ))}
            </select>
          </div>
          <div className="md:col-span-1">
            <label className="block text-sm font-medium text-gray-700">Template</label>
            <select
              value={createTemplate}
              onChange={(e) => setCreateTemplate(e.target.value)}
              className="w-full border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              <option value="">Select template…</option>
              {templates.map((tpl) => (
                <option key={tpl.id} value={tpl.id}>
                  {tpl.name} ({tpl.scope})
                </option>
              ))}
            </select>
          </div>
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700">Overrides (JSON)</label>
            <textarea
              value={createOverrides}
              onChange={(e) => setCreateOverrides(e.target.value)}
              rows={4}
              className="w-full font-mono text-sm border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="Optional partial settings overrides"
            />
          </div>
          <div className="md:col-span-4 flex justify-end">
            <Button type="submit" disabled={createState.isLoading || !createAsset || !createTemplate}>
              {createState.isLoading ? 'Creating…' : 'Create Laid-Out Image'}
            </Button>
          </div>
        </form>

        <div className="grid gap-4 md:grid-cols-3">
          {laidOutImages.map((image) => {
            const asset = assetLookup.get(image.asset_id);
            const template = templateLookup.get(image.template_id);
            return (
              <div
                key={image.id}
                className={`cursor-pointer`}
                onClick={() => handleSelectImage(image)}
              >
                <Card className={`border ${selectedImageId === image.id ? 'border-primary-500' : 'border-gray-200'}`}>
                  <CardBody className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-semibold text-gray-800">{asset?.filename ?? image.asset_id}</p>
                      <p className="text-xs text-gray-500">Template: {template?.name ?? image.template_id}</p>
                    </div>
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(image);
                      }}
                      disabled={deleteState.isLoading}
                    >
                      Delete
                    </Button>
                  </div>
                  <p className="text-xs text-gray-500">
                    Updated {new Date(image.updated_at).toLocaleString()}
                  </p>
                  </CardBody>
                </Card>
              </div>
            );
          })}
          {!laidOutImagesQuery.isLoading && laidOutImages.length === 0 && (
            <div className="col-span-full text-sm text-gray-500">No laid-out images yet.</div>
          )}
        </div>

        {selectedImageId && selectedImageQuery.data && (
          <div className="grid gap-6 md:grid-cols-2">
            <form onSubmit={handleUpdate} className="space-y-4 p-4 border border-primary-200 rounded-lg">
              <h3 className="text-lg font-semibold text-gray-900">Update Selection</h3>
              <div>
                <label className="block text-sm font-medium text-gray-700">Template</label>
                <select
                  value={editTemplate}
                  onChange={(e) => setEditTemplate(e.target.value)}
                  className="w-full border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
                >
                  <option value="">(unchanged)</option>
                  {templates.map((tpl) => (
                    <option key={tpl.id} value={tpl.id}>
                      {tpl.name} ({tpl.scope})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Overrides (JSON)</label>
                <textarea
                  value={editOverrides}
                  onChange={(e) => setEditOverrides(e.target.value)}
                  rows={8}
                  className="w-full font-mono text-sm border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  placeholder="Optional partial overrides"
                />
              </div>
              <div className="flex justify-end space-x-3">
                <Button type="submit" disabled={updateState.isLoading}>
                  {updateState.isLoading ? 'Saving…' : 'Save Changes'}
                </Button>
              </div>
            </form>

            <div className="space-y-4 p-4 border border-gray-200 rounded-lg">
              <h3 className="text-lg font-semibold text-gray-900">Computed Placement</h3>
              <div className="text-sm text-gray-500">
                <p><strong>Asset:</strong> {assetLookup.get(selectedImageQuery.data.asset_id)?.filename ?? selectedImageQuery.data.asset_id}</p>
                <p><strong>Template:</strong> {templateLookup.get(selectedImageQuery.data.template_id)?.name ?? selectedImageQuery.data.template_id}</p>
              </div>
              <div>
                <h4 className="text-sm font-medium text-gray-700 mb-1">Preview Payload</h4>
                {previewQuery.isFetching && <p className="text-xs text-gray-500">Loading preview…</p>}
                {previewQuery.data && (
                  <pre className="text-xs bg-gray-900 text-green-200 rounded-md p-3 overflow-x-auto">
                    {jsonString(previewQuery.data)}
                  </pre>
                )}
              </div>
            </div>
          </div>
        )}
      </CardBody>
    </Card>
  );
};
