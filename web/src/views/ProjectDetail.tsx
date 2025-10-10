import React, { useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  useAddImageSequenceItemMutation,
  useCreateImageSequenceMutation,
  useDeleteAssetMutation,
  useDeleteImageSequenceItemMutation,
  useDeleteImageSequenceMutation,
  useGetAssetsQuery,
  useGetImageSequenceDetailQuery,
  useGetImageSequencesQuery,
  useGetProjectsQuery,
} from '../api';
import { ProjectAssetsPanel, type AssetSummary } from '../components/bookSpread/ProjectAssetsPanel';
import { Button, Card, CardBody, CardHeader, Input } from '../components/ui';

const formatDateTime = (iso?: string) => {
  if (!iso) return '—';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '—';
  return date.toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
};

export const ProjectDetail: React.FC = () => {
  const { id = '' } = useParams();
  const { data: projects } = useGetProjectsQuery();
  const project = projects?.find((p) => p.id === id);

  const assetsQuery = useGetAssetsQuery({ projectId: id }, { skip: !id });
  const [selectedAssetId, setSelectedAssetId] = useState<string | null>(null);

  const assets = useMemo<AssetSummary[]>(() => {
    if (!assetsQuery.data || !id) return [];
    return assetsQuery.data.map((asset) => {
      const base =
        asset.url ?? `/projects/${id}/images/${encodeURIComponent(asset.filename)}`;
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
  }, [assetsQuery.data, id]);

  const assetLookup = useMemo(() => {
    const map = new Map<string, AssetSummary>();
    for (const asset of assets) map.set(asset.id, asset);
    return map;
  }, [assets]);

  const sequencesQuery = useGetImageSequencesQuery(
    { projectId: id },
    { skip: !id }
  );
  const [selectedSequenceId, setSelectedSequenceId] = useState<string | null>(null);

  useEffect(() => {
    if (!selectedSequenceId && sequencesQuery.data?.length) {
      setSelectedSequenceId(sequencesQuery.data[0]!.id);
    }
  }, [sequencesQuery.data, selectedSequenceId]);

  const sequenceDetailQuery = useGetImageSequenceDetailQuery(
    { sequenceId: selectedSequenceId ?? '' },
    { skip: !selectedSequenceId }
  );

  const [createSequence, createState] = useCreateImageSequenceMutation();
  const [addSequenceItem, addItemState] = useAddImageSequenceItemMutation();
  const [deleteSequence] = useDeleteImageSequenceMutation();
  const [deleteSequenceItem] = useDeleteImageSequenceItemMutation();
  const [deleteAsset] = useDeleteAssetMutation();

  const [newSequenceName, setNewSequenceName] = useState('');
  const [newSequenceDescription, setNewSequenceDescription] = useState('');

  const handleCreateSequence = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!id || !newSequenceName.trim()) return;
    const sequence = await createSequence({
      projectId: id,
      name: newSequenceName.trim(),
      description: newSequenceDescription.trim() || undefined,
    }).unwrap();
    setNewSequenceName('');
    setNewSequenceDescription('');
    setSelectedSequenceId(sequence.id);
  };

  const handleAppendSelectedAsset = async () => {
    if (!selectedSequenceId || !selectedAssetId) return;
    await addSequenceItem({
      sequenceId: selectedSequenceId,
      assetId: selectedAssetId,
    }).unwrap();
  };

  const handleAppendGap = async () => {
    if (!selectedSequenceId) return;
    await addSequenceItem({ sequenceId: selectedSequenceId }).unwrap();
  };

  const handleDeleteSequence = async (sequenceId: string) => {
    if (!id) return;
    const seq = sequencesQuery.data?.find((s) => s.id === sequenceId);
    if (
      window.confirm(
        `Delete sequence "${seq?.name ?? sequenceId}"? This cannot be undone.`
      )
    ) {
      await deleteSequence({ sequenceId, projectId: id }).unwrap();
      if (sequenceId === selectedSequenceId) {
        setSelectedSequenceId(null);
      }
    }
  };

  const handleDeleteItem = async (position: number) => {
    if (!selectedSequenceId) return;
    await deleteSequenceItem({ sequenceId: selectedSequenceId, position }).unwrap();
  };

  const handleDeleteAsset = async (assetId: string) => {
    if (!id) return;
    const asset = assetLookup.get(assetId);
    if (
      window.confirm(
        `Delete asset "${asset?.name ?? assetId}"? This will remove it from sequences as well.`
      )
    ) {
      await deleteAsset({ assetId, projectId: id }).unwrap();
      if (selectedAssetId === assetId) setSelectedAssetId(null);
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <nav className="flex items-center space-x-2 text-sm text-gray-600 mb-2">
          <Link to="/projects" className="hover:text-gray-900">
            Projects
          </Link>
          <span>/</span>
          <span className="text-gray-900 font-medium">{project?.name ?? id}</span>
        </nav>
        <h1 className="text-3xl font-bold text-gray-900">
          {project?.name ?? 'Project'}
        </h1>
        <p className="text-sm text-gray-500">
          Created {project ? formatDateTime(project.created_at) : '…'} · Updated{' '}
          {project ? formatDateTime(project.updated_at) : '…'}
        </p>
        {project?.description && (
          <p className="mt-4 max-w-2xl text-gray-700">{project.description}</p>
        )}
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <div className="xl:col-span-1 space-y-6">
          <Card>
            <CardHeader>
              <h2 className="text-lg font-semibold text-gray-900">Images</h2>
            </CardHeader>
            <CardBody>
              <ProjectAssetsPanel
                projectId={id}
                assets={assets}
                selectedAssetId={selectedAssetId}
                onSelectAsset={(asset) => setSelectedAssetId(asset.id)}
              />
            </CardBody>
          </Card>

          {selectedAssetId && (
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">
                  Selected Asset
                </h3>
              </CardHeader>
              <CardBody className="space-y-3">
                <div className="text-sm text-gray-600">
                  <div className="font-medium">
                    {assetLookup.get(selectedAssetId)?.name}
                  </div>
                  <div>
                    {assetLookup.get(selectedAssetId)?.width} ×{' '}
                    {assetLookup.get(selectedAssetId)?.height} px
                  </div>
                </div>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    const asset = assetLookup.get(selectedAssetId);
                    if (asset) handleDeleteAsset(asset.id);
                  }}
                >
                  Delete Asset
                </Button>
              </CardBody>
            </Card>
          )}
        </div>

        <div className="xl:col-span-2 space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900">
                  Image Sequences
                </h2>
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => setSelectedSequenceId(null)}
                >
                  + New Sequence
                </Button>
              </div>
            </CardHeader>
            <CardBody className="space-y-6">
              {!selectedSequenceId && (
                <form onSubmit={handleCreateSequence} className="space-y-4">
                  <Input
                    label="Sequence name"
                    value={newSequenceName}
                    onChange={(e) => setNewSequenceName(e.target.value)}
                    placeholder="Contact Sheet"
                    required
                  />
                  <Input
                    label="Description"
                    value={newSequenceDescription}
                    onChange={(e) => setNewSequenceDescription(e.target.value)}
                    placeholder="Optional description"
                  />
                  <div className="flex justify-end space-x-3">
                    <Button
                      variant="secondary"
                      type="button"
                      onClick={() => {
                        setSelectedSequenceId(
                          sequencesQuery.data?.[0]?.id ?? null
                        );
                      }}
                    >
                      Cancel
                    </Button>
                    <Button type="submit" isLoading={createState.isLoading}>
                      Create Sequence
                    </Button>
                  </div>
                </form>
              )}

              {Boolean(sequencesQuery.data?.length) && (
                <div className="grid md:grid-cols-2 gap-4">
                  {sequencesQuery.data?.map((sequence) => (
                    <button
                      key={sequence.id}
                      type="button"
                      onClick={() => setSelectedSequenceId(sequence.id)}
                      className={`border rounded-lg p-4 text-left transition-colors ${
                        sequence.id === selectedSequenceId
                          ? 'border-primary-500 ring-2 ring-primary-200 bg-primary-50'
                          : 'border-gray-200 hover:border-primary-300'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <h3 className="font-semibold text-gray-900 truncate">
                          {sequence.name}
                        </h3>
                        <button
                          type="button"
                          className="text-gray-400 hover:text-red-600"
                          onClick={(event) => {
                            event.stopPropagation();
                            handleDeleteSequence(sequence.id);
                          }}
                        >
                          ✕
                        </button>
                      </div>
                      {sequence.description && (
                        <p className="text-sm text-gray-600 line-clamp-2 mb-2">
                          {sequence.description}
                        </p>
                      )}
                      <p className="text-xs text-gray-500">
                        Updated {formatDateTime(sequence.updated_at)}
                      </p>
                    </button>
                  ))}
                </div>
              )}

              {selectedSequenceId && (
                <div className="border rounded-lg p-4 space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-semibold text-gray-900">
                        Sequence Items
                      </h3>
                      <p className="text-sm text-gray-500">
                        Append the selected asset or insert gap placeholders.
                      </p>
                    </div>
                    <div className="flex space-x-2">
                      <Button
                        size="sm"
                        disabled={!selectedAssetId}
                        isLoading={addItemState.isLoading}
                        onClick={handleAppendSelectedAsset}
                      >
                        Add Selected Asset
                      </Button>
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={handleAppendGap}
                        disabled={addItemState.isLoading}
                      >
                        Insert Gap
                      </Button>
                    </div>
                  </div>

                  <div className="space-y-2">
                    {sequenceDetailQuery.data?.items.length ? (
                      sequenceDetailQuery.data.items.map((item) => (
                        <div
                          key={item.position}
                          className="flex items-center justify-between border rounded-md px-3 py-2 bg-white"
                        >
                          <div>
                            <div className="text-sm font-medium text-gray-800">
                              Position {item.position + 1}
                            </div>
                            <div className="text-xs text-gray-500">
                              {item.is_gap
                                ? 'Gap'
                                : assetLookup.get(item.asset_id ?? '')?.name ??
                                  item.asset_id ??
                                  'Unknown asset'}
                            </div>
                          </div>
                          <button
                            type="button"
                            className="text-gray-400 hover:text-red-600"
                            onClick={() => handleDeleteItem(item.position)}
                          >
                            Remove
                          </button>
                        </div>
                      ))
                    ) : (
                      <p className="text-sm text-gray-500">
                        This sequence is empty. Select an asset to add it or add a
                        gap.
                      </p>
                    )}
                  </div>
                </div>
              )}
            </CardBody>
          </Card>
        </div>
      </div>
    </div>
  );
};
