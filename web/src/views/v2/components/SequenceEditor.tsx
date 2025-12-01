import React, { useMemo, useState, useRef, useCallback } from 'react';
import {
  useGetImageSequenceDetailQuery,
  useGetAssetsQuery,
  useReorderImageSequenceItemsMutation,
  useAddImageSequenceItemMutation,
  useDeleteImageSequenceItemMutation,
  type ImageSequenceItem,
} from '../../../api';
import { SequenceItem } from './SequenceItem';
import { AssetPicker } from './AssetPicker';
import { Button } from '../../../components/ui';

interface SequenceEditorProps {
  projectId: string;
  sequenceId: string;
}

// Simple debounce utility
function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;
  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      timeout = null;
      func(...args);
    };
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

export const SequenceEditor: React.FC<SequenceEditorProps> = ({
  projectId,
  sequenceId,
}) => {
  const { data: sequenceData, isLoading, error } = useGetImageSequenceDetailQuery({
    sequenceId,
  });
  const { data: assets = [] } = useGetAssetsQuery({ projectId });
  const [reorderItems, { isLoading: isReordering }] = useReorderImageSequenceItemsMutation();
  const [addItem] = useAddImageSequenceItemMutation();
  const [deleteItem] = useDeleteImageSequenceItemMutation();

  const [dragSourceIndex, setDragSourceIndex] = useState<number | null>(null);
  const [dragTargetIndex, setDragTargetIndex] = useState<number | null>(null);
  const [isAddingGap, setIsAddingGap] = useState(false);
  const [isAssetPickerOpen, setIsAssetPickerOpen] = useState(false);
  const [isAddingImages, setIsAddingImages] = useState(false);

  const sortedItems = useMemo(() => {
    if (!sequenceData?.items) return [];
    return [...sequenceData.items].sort((a, b) => a.position - b.position);
  }, [sequenceData?.items]);

  const assetLookup = useMemo(() => {
    const map = new Map(assets.map((asset) => [asset.id, asset]));
    return map;
  }, [assets]);

  // Debounced reorder function - use ref to persist across renders
  const reorderRef = useRef<ReturnType<typeof debounce>>();
  if (!reorderRef.current) {
    reorderRef.current = debounce((items: ImageSequenceItem[]) => {
      const payload = items.map((item) => ({
        assetId: item.asset_id ?? undefined,
        isGap: item.is_gap,
      }));
      reorderItems({ sequenceId, items: payload }).catch((err) => {
        console.error('Failed to reorder items:', err);
      });
    }, 300);
  }
  const debouncedReorder = reorderRef.current;

  const handleDragStart = (index: number) => {
    console.log('🟢 Drag start:', index);
    setDragSourceIndex(index);
    setDragTargetIndex(null);
  };

  const handleDragEnd = () => {
    console.log('🔴 Drag end');
    setDragSourceIndex(null);
    setDragTargetIndex(null);
  };

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  };

  const handleDragEnter = (index: number) => {
    console.log('🎯 Drag enter:', index, 'source:', dragSourceIndex);
    if (dragSourceIndex !== null && dragSourceIndex !== index) {
      setDragTargetIndex(index);
    }
  };

  const handleDrop = useCallback(
    (event: React.DragEvent, targetIndex: number | null) => {
      event.preventDefault();
      if (dragSourceIndex === null) return;

      const sourceIndex = dragSourceIndex;
      if (targetIndex === null || sourceIndex === targetIndex) {
        setDragSourceIndex(null);
        setDragTargetIndex(null);
        return;
      }

      // Calculate new order
      const newItems = [...sortedItems];
      const [removed] = newItems.splice(sourceIndex, 1);
      
      // After removing source, indices shift. When we show the indicator on the RIGHT
      // of an item (forward drag), we want to insert AFTER that item in the final array.
      // Since we removed an item before it, the target index naturally points to the
      // correct insertion position without adjustment.
      const insertIndex = targetIndex;
      
      console.log('💧 Drop: source', sourceIndex, '→ target', targetIndex, '→ insert', insertIndex);
      newItems.splice(insertIndex, 0, removed);

      // Update positions
      const reorderedItems = newItems.map((item, idx) => ({
        ...item,
        position: idx,
      }));

      // Optimistic update happens via RTK Query, but we also call debounced reorder
      debouncedReorder(reorderedItems);
      setDragSourceIndex(null);
      setDragTargetIndex(null);
    },
    [dragSourceIndex, sortedItems, debouncedReorder]
  );

  const handleAddGap = async () => {
    setIsAddingGap(true);
    try {
      await addItem({ sequenceId, assetId: undefined }).unwrap();
    } catch (error) {
      console.error('Failed to add gap:', error);
    } finally {
      setIsAddingGap(false);
    }
  };

  const handleDeleteItem = async (position: number) => {
    try {
      await deleteItem({ sequenceId, position }).unwrap();
    } catch (error) {
      console.error('Failed to delete item:', error);
    }
  };

  const handleAddImages = async (assetIds: string[]) => {
    if (assetIds.length === 0) return;
    
    setIsAddingImages(true);
    try {
      // Add images one by one (API supports single item addition)
      // Optimistic updates will handle UI updates
      for (const assetId of assetIds) {
        await addItem({ sequenceId, assetId }).unwrap();
      }
    } catch (error) {
      console.error('Failed to add images:', error);
      // Error handling: optimistic updates will rollback automatically
    } finally {
      setIsAddingImages(false);
      setIsAssetPickerOpen(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-500">Loading sequence...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-red-600">Error loading sequence</div>
      </div>
    );
  }

  if (!sequenceData) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-500">Sequence not found</div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">{sequenceData.sequence.name}</h1>
        {sequenceData.sequence.description && (
          <p className="text-gray-600 mt-1">{sequenceData.sequence.description}</p>
        )}
        <p className="text-xs text-purple-600 mt-2">🔄 TEST: Frontend v2.1 - Drag feedback enabled</p>
      </div>

      <div className="mb-4 flex gap-2 items-center">
        <Button
          onClick={() => setIsAssetPickerOpen(true)}
          disabled={isAddingImages}
          size="sm"
        >
          {isAddingImages ? 'Adding...' : '+ Add Image'}
        </Button>
        <Button onClick={handleAddGap} disabled={isAddingGap} size="sm" variant="secondary">
          {isAddingGap ? 'Adding...' : '+ Add Gap'}
        </Button>
        {isReordering && (
          <span className="text-sm text-gray-500 flex items-center">
            <svg className="w-4 h-4 mr-1 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" className="opacity-25" />
              <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" className="opacity-75" />
            </svg>
            Saving order...
          </span>
        )}
      </div>

      <AssetPicker
        projectId={projectId}
        isOpen={isAssetPickerOpen}
        onClose={() => setIsAssetPickerOpen(false)}
        onSelect={handleAddImages}
        multiple={true}
      />

      {sortedItems.length === 0 ? (
        <div className="text-center py-12 border-2 border-dashed border-gray-300 rounded-lg">
          <p className="text-gray-500">No items in this sequence</p>
          <p className="text-sm text-gray-400 mt-2">Add images or gaps to get started</p>
        </div>
      ) : (
        <div
          className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4"
          onDragOver={handleDragOver}
          onDrop={(e) => handleDrop(e, null)}
        >
          {sortedItems.map((item, index) => {
            const isDragging = dragSourceIndex === index;
            const isDropTarget = dragTargetIndex === index;
            
            // When dragging forward (source < target), we want to insert AFTER the target
            // So show indicator on the right side of the target
            // When dragging backward (source > target), insert BEFORE the target
            // So show indicator on the left side of the target
            const showDropIndicator = dragSourceIndex !== null && isDropTarget && dragSourceIndex !== index;
            const dropIndicatorOnRight = dragSourceIndex !== null && dragSourceIndex < index;

            if (showDropIndicator) {
              console.log('📍 Drop indicator at index:', index, 'side:', dropIndicatorOnRight ? 'right' : 'left');
            }

            return (
              <div
                key={`${item.sequence_id}-${item.position}`}
                className="relative"
              >
                {/* Drop indicator - shows where item will be inserted */}
                {showDropIndicator && (
                  <div
                    className={`absolute top-0 bottom-0 w-1 bg-primary-500 rounded-full z-10 shadow-lg ${
                      dropIndicatorOnRight ? '-right-2.5' : '-left-2.5'
                    }`}
                    style={{
                      animation: 'pulse 1s cubic-bezier(0.4, 0, 0.6, 1) infinite',
                    }}
                  />
                )}
                
                <div
                  draggable
                  onDragStart={() => handleDragStart(index)}
                  onDragEnd={handleDragEnd}
                  onDragOver={handleDragOver}
                  onDragEnter={() => handleDragEnter(index)}
                  onDrop={(e) => handleDrop(e, index)}
                  className={`transition-all duration-200 ${
                    isDragging ? 'opacity-30 scale-95' : ''
                  } ${
                    isDropTarget && !isDragging ? 'ring-2 ring-primary-400 ring-offset-2 rounded-lg' : ''
                  }`}
                >
                  <SequenceItem
                    item={item}
                    asset={item.asset_id ? assetLookup.get(item.asset_id) : undefined}
                    projectId={projectId}
                    index={index}
                    onDelete={() => handleDeleteItem(item.position)}
                    isDragging={isDragging}
                  />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

