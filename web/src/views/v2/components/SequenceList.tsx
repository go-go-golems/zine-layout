import React, { useState } from 'react';
import {
  useGetImageSequencesQuery,
  useCreateImageSequenceMutation,
  type ImageSequence,
} from '../../../api';
import { Button, Card, CardBody, CardHeader, Input } from '../../../components/ui';

interface SequenceListProps {
  projectId: string;
  selectedSequenceId: string | null;
  onSelectSequence: (sequenceId: string | null) => void;
}

export const SequenceList: React.FC<SequenceListProps> = ({
  projectId,
  selectedSequenceId,
  onSelectSequence,
}) => {
  const { data: sequences = [], isLoading } = useGetImageSequencesQuery({ projectId });
  const [createSequence, { isLoading: isCreating }] = useCreateImageSequenceMutation();
  const [isCreatingNew, setIsCreatingNew] = useState(false);
  const [newSequenceName, setNewSequenceName] = useState('');
  const [newSequenceDescription, setNewSequenceDescription] = useState('');

  const handleCreateSequence = async () => {
    if (!newSequenceName.trim()) return;

    try {
      const result = await createSequence({
        projectId,
        name: newSequenceName.trim(),
        description: newSequenceDescription.trim() || undefined,
      }).unwrap();
      setNewSequenceName('');
      setNewSequenceDescription('');
      setIsCreatingNew(false);
      onSelectSequence(result.id);
    } catch (error) {
      console.error('Failed to create sequence:', error);
    }
  };

  const handleCancelCreate = () => {
    setNewSequenceName('');
    setNewSequenceDescription('');
    setIsCreatingNew(false);
  };

  return (
    <div className="p-4">
      <div className="mb-4">
        <h2 className="text-xl font-bold text-gray-900 mb-2">Sequences</h2>
        {!isCreatingNew ? (
          <Button
            onClick={() => setIsCreatingNew(true)}
            className="w-full"
            disabled={isCreating}
          >
            + Create Sequence
          </Button>
        ) : (
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium">New Sequence</h3>
            </CardHeader>
            <CardBody className="space-y-3">
              <Input
                placeholder="Sequence name"
                value={newSequenceName}
                onChange={(e) => setNewSequenceName(e.target.value)}
                autoFocus
              />
              <Input
                placeholder="Description (optional)"
                value={newSequenceDescription}
                onChange={(e) => setNewSequenceDescription(e.target.value)}
              />
              <div className="flex gap-2">
                <Button
                  onClick={handleCreateSequence}
                  disabled={!newSequenceName.trim() || isCreating}
                  className="flex-1"
                >
                  Create
                </Button>
                <Button
                  onClick={handleCancelCreate}
                  variant="secondary"
                  className="flex-1"
                  disabled={isCreating}
                >
                  Cancel
                </Button>
              </div>
            </CardBody>
          </Card>
        )}
      </div>

      {isLoading ? (
        <div className="text-center text-gray-500 py-8">Loading sequences...</div>
      ) : sequences.length === 0 ? (
        <div className="text-center text-gray-500 py-8">
          <p className="text-sm">No sequences yet</p>
          <p className="text-xs mt-1">Create your first sequence to get started</p>
        </div>
      ) : (
        <div className="space-y-2">
          {sequences.map((sequence) => (
            <SequenceListItem
              key={sequence.id}
              sequence={sequence}
              isSelected={sequence.id === selectedSequenceId}
              onClick={() => onSelectSequence(sequence.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
};

interface SequenceListItemProps {
  sequence: ImageSequence;
  isSelected: boolean;
  onClick: () => void;
}

const SequenceListItem: React.FC<SequenceListItemProps> = ({
  sequence,
  isSelected,
  onClick,
}) => {
  return (
    <Card
      className={`cursor-pointer transition-colors ${
        isSelected ? 'border-primary-500 bg-primary-50' : 'hover:bg-gray-50'
      }`}
      onClick={onClick}
    >
      <CardBody>
        <h3 className="font-medium text-gray-900">{sequence.name}</h3>
        {sequence.description && (
          <p className="text-sm text-gray-600 mt-1 line-clamp-2">{sequence.description}</p>
        )}
        <p className="text-xs text-gray-500 mt-2">
          Updated {new Date(sequence.updated_at).toLocaleDateString()}
        </p>
      </CardBody>
    </Card>
  );
};

