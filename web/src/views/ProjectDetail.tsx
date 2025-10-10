import React, { useEffect, useMemo, useRef, useState } from 'react';
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
  useReorderImageSequenceItemsMutation,
} from '../api';
import {
  ProjectAssetsPanel,
  type AssetSummary,
} from '../components/bookSpread/ProjectAssetsPanel';
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

const mapItemsToPayload = (
  items: Array<{ asset_id?: string; is_gap: boolean }>
) =>
  items.map((item) => ({
    assetId: item.is_gap ? undefined : item.asset_id,
    isGap: item.is_gap,
  }));

export const ProjectDetail: React.FC = () => {
  const { id = '' } = useParams();
  const { data: projects } = useGetProjectsQuery();
  const project = projects?.find((p) => p.id === id);

  const assetsQuery = useGetAssetsQuery({ projectId: id }, { skip: !id });
  const [selectedAssetId, setSelectedAssetId] = useState<string | null>(null);
  const [dragAssetId, setDragAssetId] = useState<string | null>(null);
  const [dragSourceIndex, setDragSourceIndex] = useState<number | null>(null);

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

  const sortedItems = useMemo(() => {
    if (!sequenceDetailQuery.data?.items) return [];
    return [...sequenceDetailQuery.data.items].sort(
      (a, b) => a.position - b.position
    );
  }, [sequenceDetailQuery.data?.items]);

  const [slideIndex, setSlideIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const playTimer = useRef<number | null>(null);

  useEffect(() => {
    setSlideIndex((prev) =>
      sortedItems.length ? Math.min(prev, sortedItems.length - 1) : 0
    );
    if (sortedItems.length <= 1) {
      setIsPlaying(false);
    }
  }, [sortedItems.length]);

  useEffect(() => {
    if (!isPlaying || sortedItems.length === 0) {
      if (playTimer.current !== null) {
        window.clearInterval(playTimer.current);
        playTimer.current = null;
      }
      return;
    }
    playTimer.current = window.setInterval(() => {
      setSlideIndex((prev) => (prev + 1) % sortedItems.length);
    }, 2500);
    return () => {
      if (playTimer.current !== null) {
        window.clearInterval(playTimer.current);
        playTimer.current = null;
      }
    };
  }, [isPlaying, sortedItems.length]);

  const [createSequence, createState] = useCreateImageSequenceMutation();
  const [addSequenceItem, addItemState] = useAddImageSequenceItemMutation();
  const [deleteSequence] = useDeleteImageSequenceMutation();
  const [deleteSequenceItem] = useDeleteImageSequenceItemMutation();
  const [deleteAsset] = useDeleteAssetMutation();
  const [reorderItems, reorderState] = useReorderImageSequenceItemsMutation();

  const [newSequenceName, setNewSequenceName] = useState('');
  const [newSequenceDescription, setNewSequenceDescription] = useState('');

  useEffect(() => {
    setSlideIndex(0);
    setIsPlaying(false);
  }, [selectedSequenceId]);

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
    setIsPlaying(false);
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
      setIsPlaying(false);
    }
  };

  const handleAppendGap = async () => {
    if (!selectedSequenceId) return;
    await addSequenceItem({ sequenceId: selectedSequenceId }).unwrap();
    setIsPlaying(false);
  };

  const handleAppendAsset = async (assetId: string, insertIndex: number | null) => {
    if (!selectedSequenceId) return;
    const addedItems = await addSequenceItem({
      sequenceId: selectedSequenceId,
      assetId,
    }).unwrap();
    const itemsSorted = [...addedItems].sort((a, b) => a.position - b.position);
    const newItemIndex = itemsSorted.findIndex(
      (item) => !item.is_gap && item.asset_id === assetId
    );
    if (newItemIndex === -1) return;
    const [newItem] = itemsSorted.splice(newItemIndex, 1);
    const targetIndex =
      insertIndex === null
        ? itemsSorted.length
        : Math.max(0, Math.min(insertIndex, itemsSorted.length));
    itemsSorted.splice(targetIndex, 0, newItem);
    await reorderItems({
      sequenceId: selectedSequenceId,
      items: mapItemsToPayload(itemsSorted),
    }).unwrap();
    setSlideIndex(targetIndex);
    setIsPlaying(false);
  };

  const handleReorder = async (sourceIndex: number, targetIndex: number | null) => {
    if (!selectedSequenceId) return;
    const itemsCopy = [...sortedItems];
    if (sourceIndex < 0 || sourceIndex >= itemsCopy.length) return;
    if (targetIndex !== null) {
      if (targetIndex < 0 || targetIndex > itemsCopy.length) return;
      if (targetIndex === sourceIndex || targetIndex === sourceIndex + 1) return;
    }
    const [moved] = itemsCopy.splice(sourceIndex, 1);
    if (!moved) return;
    let destination =
      targetIndex === null ? itemsCopy.length : Math.max(0, targetIndex);
    if (destination > itemsCopy.length) destination = itemsCopy.length;
    if (sourceIndex < destination) {
      destination = destination - 1;
    }
    itemsCopy.splice(destination, 0, moved);
    await reorderItems({
      sequenceId: selectedSequenceId,
      items: mapItemsToPayload(itemsCopy),
    }).unwrap();
    setSlideIndex(destination);
    setIsPlaying(false);
  };

  const handleDrop = async (
    event: React.DragEvent,
    targetIndex: number | null
  ) => {
    event.preventDefault();
    event.stopPropagation();
    if (dragAssetId) {
      await handleAppendAsset(dragAssetId, targetIndex);
    } else if (dragSourceIndex !== null) {
      await handleReorder(dragSourceIndex, targetIndex);
    }
    setDragAssetId(null);
    setDragSourceIndex(null);
  };

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = dragAssetId ? 'copy' : 'move';
  };

  const currentItem = sortedItems[slideIndex];
  const currentAsset =
    currentItem && !currentItem.is_gap && currentItem.asset_id
      ? assetLookup.get(currentItem.asset_id)
      : undefined;

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
          Created {formatDateTime(project?.created_at)} · Updated{' '}
          {formatDateTime(project?.updated_at)}
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
                onAssetDragStart={(asset) => {
                  setDragAssetId(asset.id);
                  setDragSourceIndex(null);
                }}
                onAssetDragEnd={() => setDragAssetId(null)}
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
                      onClick={() =>
                        setSelectedSequenceId(
                          sequencesQuery.data?.[0]?.id ?? null
                        )
                      }
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
                <div className="space-y-4">
                  <div className="border rounded-lg p-4 bg-white">
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="text-base font-semibold text-gray-900">
                        Sequence Preview
                      </h3>
                      <div className="flex items-center space-x-2">
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() =>
                            setSlideIndex((prev) =>
                              sortedItems.length
                                ? (prev - 1 + sortedItems.length) %
                                  sortedItems.length
                                : 0
                            )
                          }
                          disabled={sortedItems.length === 0}
                        >
                          ‹ Prev
                        </Button>
                        <Button
                          size="sm"
                          variant={isPlaying ? 'danger' : 'primary'}
                          onClick={() => setIsPlaying((prev) => !prev)}
                          disabled={sortedItems.length <= 1}
                        >
                          {isPlaying ? 'Stop' : 'Play'}
                        </Button>
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() =>
                            setSlideIndex((prev) =>
                              sortedItems.length ? (prev + 1) % sortedItems.length : 0
                            )
                          }
                          disabled={sortedItems.length === 0}
                        >
                          Next ›
                        </Button>
                      </div>
                    </div>
                    <div className="border rounded-md overflow-hidden bg-gray-50 flex items-center justify-center h-72">
                      {currentItem ? (
                        currentItem.is_gap ? (
                          <div className="text-gray-500 text-sm">Gap</div>
                        ) : currentAsset ? (
                          <img
                            src={currentAsset.src}
                            alt={currentAsset.name}
                            className="max-h-full max-w-full object-contain"
                          />
                        ) : (
                          <div className="text-gray-500 text-sm">
                            Asset not found
                          </div>
                        )
                      ) : (
                        <div className="text-gray-500 text-sm">
                          Sequence is empty
                        </div>
                      )}
                    </div>
                    {currentAsset && (
                      <div className="mt-3 text-sm text-gray-600">
                        {currentAsset.name} · {currentAsset.width} ×{' '}
                        {currentAsset.height}px
                      </div>
                    )}
                  </div>

                  <div className="border rounded-lg p-4 bg-white">
                    <div className="flex items-center justify-between mb-4">
                      <div>
                        <h3 className="font-semibold text-gray-900">
                          Sequence Items
                        </h3>
                        <p className="text-sm text-gray-500">
                          Drag assets from the left to add them, or drag items to
                          reorder.
                        </p>
                      </div>
                      <div className="flex space-x-2">
                        <Button
                          size="sm"
                          disabled={!selectedAssetId}
                          isLoading={addItemState.isLoading}
                          onClick={() => {
                            if (selectedAssetId) {
                              void handleAppendAsset(selectedAssetId, null);
                            }
                          }}
                        >
                          Add Selected Asset
                        </Button>
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() => void handleAppendGap()}
                          disabled={addItemState.isLoading}
                        >
                          Insert Gap
                        </Button>
                      </div>
                    </div>

                    <div
                      className="space-y-2"
                      onDragOver={handleDragOver}
                      onDrop={(event) => handleDrop(event, null)}
                    >
                      {sortedItems.length ? (
                        sortedItems.map((item, index) => {
                          const asset =
                            !item.is_gap && item.asset_id
                              ? assetLookup.get(item.asset_id)
                              : undefined;
                          const isActive = index === slideIndex;
                          return (
                            <div
                              key={`${item.position}-${item.asset_id ?? 'gap'}`}
                              className={`flex items-center justify-between border rounded-md px-3 py-2 bg-white transition-colors ${
                                isActive
                                  ? 'border-primary-500 ring-1 ring-primary-300'
                                  : 'border-gray-200 hover:border-primary-300'
                              }`}
                              draggable
                              onDragStart={(event) => {
                                event.dataTransfer.effectAllowed = 'move';
                                setDragSourceIndex(index);
                                setDragAssetId(null);
                              }}
                              onDragEnd={() => setDragSourceIndex(null)}
                              onDragOver={handleDragOver}
                              onDrop={(event) => handleDrop(event, index)}
                              onClick={() => setSlideIndex(index)}
                            >
                              <div className="flex items-center space-x-3">
                                <div className="w-12 h-12 rounded bg-gray-100 flex items-center justify-center overflow-hidden">
                                  {asset ? (
                                    <img
                                      src={asset.src}
                                      alt={asset.name}
                                      className="object-cover w-full h-full"
                                    />
                                  ) : (
                                    <span className="text-xs text-gray-500">Gap</span>
                                  )}
                                </div>
                                <div>
                                  <div className="text-sm font-medium text-gray-800">
                                    Position {index + 1}
                                  </div>
                                  <div className="text-xs text-gray-500">
                                    {item.is_gap
                                      ? 'Gap'
                                      : asset?.name ?? item.asset_id ?? 'Unknown asset'}
                                  </div>
                                </div>
                              </div>
                              <button
                                type="button"
                                className="text-gray-400 hover:text-red-600"
                                onClick={(event) => {
                                  event.stopPropagation();
                                  handleDeleteItem(item.position);
                                }}
                              >
                                Remove
                              </button>
                            </div>
                          );
                        })
                      ) : (
                        <p className="text-sm text-gray-500">
                          This sequence is empty. Drag assets here or add gaps.
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {(addItemState.isLoading || reorderState.isLoading) && (
                <div className="text-sm text-gray-500">Updating sequence…</div>
              )}
            </CardBody>
          </Card>
        </div>
      </div>
    </div>
  );
};
