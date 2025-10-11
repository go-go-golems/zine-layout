import React, { useEffect, useMemo, useState } from 'react';
import {
  useAddLayoutSequenceItemMutation,
  useCreateLayoutSequenceMutation,
  useDeleteLayoutSequenceItemMutation,
  useDeleteLayoutSequenceMutation,
  useGetLaidOutImagesQuery,
  useGetLayoutSequenceDetailQuery,
  useGetLayoutSequencesQuery,
  useReorderLayoutSequenceItemsMutation,
  useUpdateLayoutSequenceMutation,
  type LaidOutImage,
  type LayoutSequence,
  type LayoutSequenceItem,
} from '../api';
import { Button, Card, CardBody, CardHeader, Input } from '../components/ui';

interface LayoutSequenceEditorProps {
  projectId: string;
}

export const LayoutSequenceEditor: React.FC<LayoutSequenceEditorProps> = ({ projectId }) => {
  const sequencesQuery = useGetLayoutSequencesQuery({ projectId }, { skip: !projectId });
  const [selectedSequenceId, setSelectedSequenceId] = useState<string | null>(null);

  const detailQuery = useGetLayoutSequenceDetailQuery(
    { sequenceId: selectedSequenceId ?? '' },
    { skip: !selectedSequenceId }
  );

  const laidOutImagesQuery = useGetLaidOutImagesQuery({ projectId }, { skip: !projectId });

  const [createSequence, createState] = useCreateLayoutSequenceMutation();
  const [updateSequence, updateState] = useUpdateLayoutSequenceMutation();
  const [deleteSequence, deleteState] = useDeleteLayoutSequenceMutation();
  const [addItem, addItemState] = useAddLayoutSequenceItemMutation();
  const [reorderItems, reorderState] = useReorderLayoutSequenceItemsMutation();
  const [deleteItem, deleteItemMutation] = useDeleteLayoutSequenceItemMutation();

  const [newSequenceName, setNewSequenceName] = useState('');
  const [newSequenceDescription, setNewSequenceDescription] = useState('');

  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');

  const [addItemId, setAddItemId] = useState('');

  const sequences = useMemo(() => sequencesQuery.data ?? [], [sequencesQuery.data]);
  const items = useMemo(() => detailQuery.data?.items ?? [], [detailQuery.data?.items]);
  const laidOutImages = useMemo(() => laidOutImagesQuery.data ?? [], [laidOutImagesQuery.data]);

  useEffect(() => {
    if (!selectedSequenceId && sequences.length) {
      setSelectedSequenceId(sequences[0]!.id);
    }
  }, [sequences, selectedSequenceId]);

  useEffect(() => {
    if (detailQuery.data?.layoutSequence) {
      setEditName(detailQuery.data.layoutSequence.name);
      setEditDescription(detailQuery.data.layoutSequence.description ?? '');
    }
  }, [detailQuery.data?.layoutSequence]);

  const layoutImageLookup = useMemo(() => {
    const map = new Map<string, LaidOutImage>();
    for (const image of laidOutImages) map.set(image.id, image);
    return map;
  }, [laidOutImages]);

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
    setSelectedSequenceId(sequence.id);
  };

  const handleUpdateSequence = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!selectedSequenceId) return;
    await updateSequence({
      sequenceId: selectedSequenceId,
      name: editName.trim() || undefined,
      description: editDescription.trim() || undefined,
    }).unwrap();
  };

  const handleDeleteSequence = async (sequence: LayoutSequence) => {
    if (!window.confirm(`Delete layout sequence "${sequence.name}"?`)) return;
    await deleteSequence({ sequenceId: sequence.id, projectId }).unwrap();
    if (selectedSequenceId === sequence.id) {
      setSelectedSequenceId(null);
      setEditName('');
      setEditDescription('');
    }
  };

  const handleAddItem = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!selectedSequenceId || !addItemId) return;
    await addItem({ sequenceId: selectedSequenceId, laidOutImageId: addItemId }).unwrap();
    setAddItemId('');
  };

  const handleRemoveItem = async (position: number) => {
    if (!selectedSequenceId) return;
    await deleteItem({ sequenceId: selectedSequenceId, position }).unwrap();
  };

  const handleMoveItem = async (item: LayoutSequenceItem, direction: -1 | 1) => {
    if (!selectedSequenceId) return;
    const index = items.findIndex((entry) => entry.position === item.position);
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= items.length) return;
    const reordered = [...items]
      .sort((a, b) => a.position - b.position)
      .map((entry) => entry.laid_out_image_id);
    const [removed] = reordered.splice(index, 1);
    reordered.splice(targetIndex, 0, removed);
    await reorderItems({ sequenceId: selectedSequenceId, laidOutImageIds: reordered }).unwrap();
  };

  return (
    <Card id="layout-sequences" className="mt-8">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Layout Sequences</h2>
            <p className="text-sm text-gray-500">
              Arrange laid-out images into ordering suitable for page composition.
            </p>
          </div>
        </div>
      </CardHeader>
      <CardBody className="space-y-8">
        <div className="grid gap-4 md:grid-cols-3">
          <Card className="md:col-span-1">
            <CardBody className="space-y-4">
              <form onSubmit={handleCreateSequence} className="space-y-3">
                <h3 className="text-sm font-semibold text-gray-800">Create Sequence</h3>
                <Input
                  value={newSequenceName}
                  onChange={(e) => setNewSequenceName(e.target.value)}
                  placeholder="Sequence name"
                />
                <Input
                  value={newSequenceDescription}
                  onChange={(e) => setNewSequenceDescription(e.target.value)}
                  placeholder="Description (optional)"
                />
                <Button type="submit" disabled={createState.isLoading || !newSequenceName.trim()}>
                  {createState.isLoading ? 'Creating…' : 'Create'}
                </Button>
              </form>

              <div className="space-y-2">
                <h3 className="text-sm font-semibold text-gray-800">Sequences</h3>
                <div className="space-y-2">
                  {sequences.map((seq) => (
                    <div
                      key={seq.id}
                      className={`p-3 border rounded-md flex items-center justify-between ${selectedSequenceId === seq.id ? 'border-primary-500 bg-primary-50' : 'border-gray-200'}`}
                    >
                      <button
                        type="button"
                        onClick={() => setSelectedSequenceId(seq.id)}
                        className="text-left flex-1"
                      >
                        <p className="text-sm font-medium text-gray-900">{seq.name}</p>
                        {seq.description && <p className="text-xs text-gray-500">{seq.description}</p>}
                      </button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDeleteSequence(seq)}
                        disabled={deleteState.isLoading}
                      >
                        Delete
                      </Button>
                    </div>
                  ))}
                  {!sequencesQuery.isLoading && sequences.length === 0 && (
                    <p className="text-xs text-gray-500">No layout sequences yet.</p>
                  )}
                </div>
              </div>
            </CardBody>
          </Card>

          <div className="md:col-span-2 space-y-6">
            {selectedSequenceId && detailQuery.data?.layoutSequence ? (
              <>
                <form onSubmit={handleUpdateSequence} className="grid gap-3 md:grid-cols-2 p-4 border border-gray-200 rounded-lg">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Name</label>
                    <Input value={editName} onChange={(e) => setEditName(e.target.value)} />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Description</label>
                    <Input value={editDescription} onChange={(e) => setEditDescription(e.target.value)} />
                  </div>
                  <div className="md:col-span-2 flex justify-end">
                    <Button type="submit" disabled={updateState.isLoading}>
                      {updateState.isLoading ? 'Saving…' : 'Save Changes'}
                    </Button>
                  </div>
                </form>

                <div className="p-4 border border-gray-200 rounded-lg space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-gray-800">Sequence Items</h3>
                    <form onSubmit={handleAddItem} className="flex items-center space-x-2">
                      <select
                        value={addItemId}
                        onChange={(e) => setAddItemId(e.target.value)}
                        className="border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                      >
                        <option value="">Add laid-out image…</option>
                        {laidOutImages.map((img) => (
                          <option key={img.id} value={img.id}>
                            {img.id}
                          </option>
                        ))}
                      </select>
                      <Button type="submit" size="sm" disabled={addItemState.isLoading || !addItemId}>
                        {addItemState.isLoading ? 'Adding…' : 'Add'}
                      </Button>
                    </form>
                  </div>

                  <div className="space-y-2">
                    {items
                      .slice()
                      .sort((a, b) => a.position - b.position)
                      .map((item, index) => {
                        const image = layoutImageLookup.get(item.laid_out_image_id);
                        const label = image ? `${image.id} · Template ${image.template_id}` : item.laid_out_image_id;
                        return (
                          <div
                            key={`${item.sequence_id}-${item.position}`}
                            className="p-3 border border-gray-200 rounded-md flex items-center justify-between"
                          >
                            <div>
                              <p className="text-sm font-medium text-gray-800">#{item.position + 1}</p>
                              <p className="text-xs text-gray-500">{label}</p>
                            </div>
                            <div className="flex items-center space-x-2">
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() => handleMoveItem(item, -1)}
                                disabled={index === 0 || reorderState.isLoading}
                              >
                                Up
                              </Button>
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={() => handleMoveItem(item, 1)}
                                disabled={index === items.length - 1 || reorderState.isLoading}
                              >
                                Down
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => handleRemoveItem(item.position)}
                                disabled={deleteItemMutation.isLoading}
                              >
                                Remove
                              </Button>
                            </div>
                          </div>
                        );
                      })}
                    {items.length === 0 && (
                      <p className="text-xs text-gray-500">No items in this sequence yet.</p>
                    )}
                  </div>
                </div>
              </>
            ) : (
              <p className="text-sm text-gray-500">Select a layout sequence to view details.</p>
            )}
          </div>
        </div>
      </CardBody>
    </Card>
  );
};
