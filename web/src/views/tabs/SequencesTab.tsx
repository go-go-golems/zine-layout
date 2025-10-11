import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  useAddImageSequenceItemMutation,
  useCreateImageSequenceMutation,
  useDeleteImageSequenceItemMutation,
  useDeleteImageSequenceMutation,
  useGetAssetsQuery,
  useGetImageSequenceDetailQuery,
  useGetImageSequencesQuery,
  useReorderImageSequenceItemsMutation,
  type Asset,
} from '../../api';
import { Button, Card, CardBody, CardHeader, Input } from '../../components/ui';
import type { AssetSummary } from '../../components/ProjectAssetsPanel';

interface SequencesTabProps {
  projectId: string;
}

const formatDateTime = (iso?: string) => {
  if (!iso) return '—';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '—';
  return date.toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
};

const mapItemsToPayload = (items: Array<{ asset_id?: string; is_gap: boolean }>) =>
  items.map((item) => ({
    assetId: item.is_gap ? undefined : item.asset_id,
    isGap: item.is_gap,
  }));

export const SequencesTab: React.FC<SequencesTabProps> = ({ projectId }) => {
  const assetsQuery = useGetAssetsQuery({ projectId }, { skip: !projectId });
  const sequencesQuery = useGetImageSequencesQuery({ projectId }, { skip: !projectId });
  const [selectedSequenceId, setSelectedSequenceId] = useState<string | null>(null);
  const [isCreatingSequence, setIsCreatingSequence] = useState(false);

  const sequenceDetailQuery = useGetImageSequenceDetailQuery(
    { sequenceId: selectedSequenceId ?? '' },
    { skip: !selectedSequenceId }
  );

  const [createSequence, createState] = useCreateImageSequenceMutation();
  const [addSequenceItem, addItemState] = useAddImageSequenceItemMutation();
  const [deleteSequence] = useDeleteImageSequenceMutation();
  const [deleteSequenceItem] = useDeleteImageSequenceItemMutation();
  const [reorderItems, reorderState] = useReorderImageSequenceItemsMutation();

  const [newSequenceName, setNewSequenceName] = useState('');
  const [newSequenceDescription, setNewSequenceDescription] = useState('');
  const [slideIndex, setSlideIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const playTimer = useRef<number | null>(null);

  const [dragSourceIndex, setDragSourceIndex] = useState<number | null>(null);

  const assets = useMemo<AssetSummary[]>(() => {
    if (!assetsQuery.data || !projectId) return [];
    return assetsQuery.data.map((asset) => {
      const base =
        asset.url ?? `/projects/${projectId}/images/${encodeURIComponent(asset.filename)}`;
      const bust = asset.uploaded_at ? new Date(asset.uploaded_at).getTime() : Date.now();
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

  const sortedItems = useMemo(() => {
    if (!sequenceDetailQuery.data?.items) return [];
    return [...sequenceDetailQuery.data.items].sort((a, b) => a.position - b.position);
  }, [sequenceDetailQuery.data?.items]);

  useEffect(() => {
    if (!isCreatingSequence && !selectedSequenceId && sequencesQuery.data?.length) {
      setSelectedSequenceId(sequencesQuery.data[0]!.id);
    }
  }, [sequencesQuery.data, selectedSequenceId, isCreatingSequence]);

  useEffect(() => {
    setSlideIndex((prev) => (sortedItems.length ? Math.min(prev, sortedItems.length - 1) : 0));
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

  useEffect(() => {
    setSlideIndex(0);
    setIsPlaying(false);
  }, [selectedSequenceId]);

  const handleCreateSequence = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!projectId || !newSequenceName.trim()) return;
    const sequence = await createSequence({
      projectId,
      name: newSequenceName.trim(),
      description: newSequenceDescription.trim() || undefined,
    }).unwrap();
    setNewSequenceName('');
    setNewSequenceDescription('');
    setIsCreatingSequence(false);
    setSelectedSequenceId(sequence.id);
  };

  const handleDeleteSequence = async (sequenceId: string) => {
    if (!projectId) return;
    const seq = sequencesQuery.data?.find((s) => s.id === sequenceId);
    if (
      window.confirm(`Delete sequence "${seq?.name ?? sequenceId}"? This cannot be undone.`)
    ) {
      await deleteSequence({ sequenceId, projectId }).unwrap();
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

  const handleAppendGap = async () => {
    if (!selectedSequenceId) return;
    await addSequenceItem({ sequenceId: selectedSequenceId }).unwrap();
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
    let destination = targetIndex === null ? itemsCopy.length : Math.max(0, targetIndex);
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

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  };

  const handleDrop = async (event: React.DragEvent, targetIndex: number | null) => {
    event.preventDefault();
    event.stopPropagation();
    if (dragSourceIndex !== null) {
      await handleReorder(dragSourceIndex, targetIndex);
    }
    setDragSourceIndex(null);
  };

  const currentItem = sortedItems[slideIndex];
  const currentAsset =
    currentItem && !currentItem.is_gap && currentItem.asset_id
      ? assetLookup.get(currentItem.asset_id)
      : undefined;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Image Sequences</h2>
          <p className="text-sm text-gray-500 mt-1">
            Organize your images into ordered collections
          </p>
        </div>
        <Button
          onClick={() => {
            setIsCreatingSequence(true);
            setSelectedSequenceId(null);
            setNewSequenceName('');
            setNewSequenceDescription('');
            setIsPlaying(false);
            setSlideIndex(0);
          }}
        >
          + New Sequence
        </Button>
      </div>

      {/* Sequence Selector */}
      {!isCreatingSequence && sequencesQuery.data && sequencesQuery.data.length > 0 && (
        <Card>
          <CardHeader>
            <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
              Select Sequence
            </h3>
          </CardHeader>
          <CardBody>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {sequencesQuery.data.map((sequence) => (
                <button
                  key={sequence.id}
                  type="button"
                  onClick={() => {
                    setIsCreatingSequence(false);
                    setSelectedSequenceId(sequence.id);
                    setSlideIndex(0);
                    setIsPlaying(false);
                  }}
                  className={`text-left border-2 rounded-lg p-4 transition-all ${
                    sequence.id === selectedSequenceId
                      ? 'border-primary-500 bg-primary-50 shadow-md'
                      : 'border-gray-200 hover:border-primary-300 hover:shadow-sm'
                  }`}
                >
                  <div className="flex items-start justify-between mb-2">
                    <h4 className="font-semibold text-gray-900 text-base">{sequence.name}</h4>
                    {sequence.id === selectedSequenceId && (
                      <span className="text-primary-600 text-xl">●</span>
                    )}
                  </div>
                  {sequence.description && (
                    <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                      {sequence.description}
                    </p>
                  )}
                  <p className="text-xs text-gray-500">
                    Updated {formatDateTime(sequence.updated_at)}
                  </p>
                </button>
              ))}
            </div>
          </CardBody>
        </Card>
      )}

      {/* Create Form */}
      {(!selectedSequenceId || isCreatingSequence) && (
        <Card>
          <CardHeader>
            <h3 className="text-lg font-semibold text-gray-900">Create New Sequence</h3>
          </CardHeader>
          <CardBody>
            <form onSubmit={handleCreateSequence} className="space-y-4">
              <Input
                label="Sequence Name"
                value={newSequenceName}
                onChange={(e) => setNewSequenceName(e.target.value)}
                placeholder="e.g., Best of Summer"
                required
              />
              <Input
                label="Description (optional)"
                value={newSequenceDescription}
                onChange={(e) => setNewSequenceDescription(e.target.value)}
                placeholder="e.g., Top 20 photos for the book"
              />
              <div className="flex justify-end space-x-3">
                <Button
                  variant="secondary"
                  type="button"
                  onClick={() => {
                    setIsCreatingSequence(false);
                    setSelectedSequenceId(sequencesQuery.data?.[0]?.id ?? null);
                  }}
                >
                  Cancel
                </Button>
                <Button type="submit" isLoading={createState.isLoading}>
                  Create Sequence
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>
      )}

      {/* Three-Column Layout: Assets + Preview + Sequence Builder */}
      {selectedSequenceId && sequenceDetailQuery.data && (
        <div className="grid lg:grid-cols-3 gap-6">
          {/* Left: Available Assets */}
          <Card>
            <CardHeader>
              <h3 className="text-lg font-semibold text-gray-900">Available Assets</h3>
              <p className="text-sm text-gray-500 mt-1">
                Click to add to sequence
              </p>
            </CardHeader>
            <CardBody>
              <div className="space-y-2 max-h-[600px] overflow-y-auto">
                {assets.map((asset) => (
                  <button
                    key={asset.id}
                    type="button"
                    onClick={async () => {
                      if (!selectedSequenceId) return;
                      const addedItems = await addSequenceItem({
                        sequenceId: selectedSequenceId,
                        assetId: asset.id,
                      }).unwrap();
                      setIsPlaying(false);
                    }}
                    className="w-full flex items-center space-x-3 p-2 border border-gray-200 rounded-lg bg-white hover:border-primary-300 hover:bg-primary-50 transition-colors"
                  >
                    <div className="w-16 h-16 bg-gray-100 rounded flex items-center justify-center overflow-hidden flex-shrink-0">
                      <img
                        src={asset.src}
                        alt={asset.name}
                        className="object-cover w-full h-full"
                      />
                    </div>
                    <div className="flex-1 min-w-0 text-left">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {asset.name}
                      </p>
                      <p className="text-xs text-gray-500">
                        {asset.width} × {asset.height}
                      </p>
                    </div>
                  </button>
                ))}
                {assets.length === 0 && (
                  <div className="text-center py-8 text-gray-500">
                    <p className="text-sm">No assets uploaded yet.</p>
                    <p className="text-xs mt-1">Go to Assets tab to upload images.</p>
                  </div>
                )}
              </div>
            </CardBody>
          </Card>

          {/* Center: Preview Pane */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-semibold text-gray-900">Preview</h3>
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => handleDeleteSequence(selectedSequenceId)}
                >
                  🗑️ Delete Sequence
                </Button>
              </div>
            </CardHeader>
            <CardBody className="space-y-4">
              {/* Main Preview */}
              <div className="border rounded-md overflow-hidden bg-gray-50 flex items-center justify-center aspect-[4/3]">
                {currentItem ? (
                  currentItem.is_gap ? (
                    <div className="text-gray-500 text-lg font-medium">Gap</div>
                  ) : currentAsset ? (
                    <img
                      src={currentAsset.src}
                      alt={currentAsset.name}
                      className="max-h-full max-w-full object-contain"
                    />
                  ) : (
                    <div className="text-gray-500 text-sm">Asset not found</div>
                  )
                ) : (
                  <div className="text-gray-500 text-sm">Sequence is empty</div>
                )}
              </div>

              {/* Controls */}
              <div className="flex items-center justify-between">
                <div className="text-sm text-gray-600">
                  {sortedItems.length > 0 ? (
                    <>
                      Item {slideIndex + 1} of {sortedItems.length}
                    </>
                  ) : (
                    'Empty sequence'
                  )}
                </div>
                <div className="flex items-center space-x-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() =>
                      setSlideIndex((prev) =>
                        sortedItems.length
                          ? (prev - 1 + sortedItems.length) % sortedItems.length
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

              {/* Asset Info */}
              {currentAsset && (
                <div className="pt-2 border-t text-sm text-gray-600">
                  <p className="font-medium text-gray-900">{currentAsset.name}</p>
                  <p>
                    {currentAsset.width} × {currentAsset.height} px
                  </p>
                </div>
              )}
            </CardBody>
          </Card>

          {/* Right: Sequence Builder */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-semibold text-gray-900">Sequence Items</h3>
                <Button size="sm" variant="secondary" onClick={() => void handleAppendGap()}>
                  + Insert Gap
                </Button>
              </div>
              <p className="text-sm text-gray-500 mt-1">
                Drag items to reorder or click to preview
              </p>
            </CardHeader>
            <CardBody>
              <div
                className="space-y-2 min-h-[300px]"
                onDragOver={handleDragOver}
                onDrop={(event) => handleDrop(event, null)}
              >
                {sortedItems.length ? (
                  sortedItems.map((item, index) => {
                    const asset =
                      !item.is_gap && item.asset_id ? assetLookup.get(item.asset_id) : undefined;
                    const isActive = index === slideIndex;
                    return (
                      <div
                        key={`${item.position}-${item.asset_id ?? 'gap'}`}
                        className={`flex items-center justify-between border rounded-md px-3 py-2 bg-white transition-all cursor-move ${
                          isActive
                            ? 'border-primary-500 ring-2 ring-primary-200 shadow-sm'
                            : 'border-gray-200 hover:border-primary-300'
                        }`}
                        draggable
                        onDragStart={(event) => {
                          event.dataTransfer.effectAllowed = 'move';
                          setDragSourceIndex(index);
                        }}
                        onDragEnd={() => setDragSourceIndex(null)}
                        onDragOver={handleDragOver}
                        onDrop={(event) => handleDrop(event, index)}
                        onClick={() => setSlideIndex(index)}
                      >
                        <div className="flex items-center space-x-3">
                          <div className="w-12 h-12 rounded bg-gray-100 flex items-center justify-center overflow-hidden flex-shrink-0">
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
                          <div className="min-w-0">
                            <div className="text-sm font-medium text-gray-800">
                              Position {index + 1}
                            </div>
                            <div className="text-xs text-gray-500 truncate">
                              {item.is_gap ? 'Gap' : asset?.name ?? item.asset_id ?? 'Unknown'}
                            </div>
                          </div>
                        </div>
                        <button
                          type="button"
                          className="text-gray-400 hover:text-red-600 ml-2 flex-shrink-0"
                          onClick={(event) => {
                            event.stopPropagation();
                            handleDeleteItem(item.position);
                          }}
                        >
                          ✕
                        </button>
                      </div>
                    );
                  })
                ) : (
                  <div className="text-center py-12 text-gray-500">
                    <p className="text-sm">This sequence is empty.</p>
                    <p className="text-xs mt-2">
                      Go to the Assets tab to add images, then drag them here.
                    </p>
                  </div>
                )}
              </div>

              {(addItemState.isLoading || reorderState.isLoading) && (
                <div className="text-sm text-gray-500 mt-4">Updating sequence…</div>
              )}
            </CardBody>
          </Card>
        </div>
      )}
    </div>
  );
};

