import React, { useMemo, useState } from 'react';
import {
  useDeletePageMutation,
  useDeleteSpreadMutation,
  useGetPagesQuery,
  useGetSpreadsQuery,
  usePutPageMutation,
  usePutSpreadMutation,
  type PersistedPage,
  type PersistedSpread,
} from '../../api';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { buildSpreadSettingsFromState } from '../../utils/spreadRequestBuilder';
import { loadSettings } from '../../store/bookSpreadSlice';
import type { AssetSummary } from './ProjectAssetsPanel';

interface LayoutPersistencePanelProps {
  projectId: string | null;
  assets: AssetSummary[];
  selectedAssetId: string | null;
  onSelectAssetById: (assetId: string) => void;
}

const numberOrUndefined = (value: string): number | undefined => {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  return Number.isNaN(parsed) ? undefined : parsed;
};

export const LayoutPersistencePanel: React.FC<LayoutPersistencePanelProps> = ({
  projectId,
  assets,
  selectedAssetId,
  onSelectAssetById,
}) => {
  const dispatch = useAppDispatch();
  const spreadState = useAppSelector((state) => state.bookSpread);

  const [pageNumberInput, setPageNumberInput] = useState('1');
  const [spreadNumberInput, setSpreadNumberInput] = useState('1');
  const [spreadLeftInput, setSpreadLeftInput] = useState('');
  const [spreadRightInput, setSpreadRightInput] = useState('');
  const [feedback, setFeedback] = useState<string | null>(null);

  const pagesQuery = useGetPagesQuery({ id: projectId ?? '' }, { skip: !projectId });
  const spreadsQuery = useGetSpreadsQuery({ id: projectId ?? '' }, { skip: !projectId });

  const [putPage, putPageStatus] = usePutPageMutation();
  const [deletePage] = useDeletePageMutation();
  const [putSpread, putSpreadStatus] = usePutSpreadMutation();
  const [deleteSpread] = useDeleteSpreadMutation();

  const pagesByNumber = useMemo(() => {
    const map = new Map<number, PersistedPage>();
    const list = pagesQuery.data?.pages ?? [];
    for (const page of list) {
      map.set(page.page_number, page);
    }
    return map;
  }, [pagesQuery.data]);

  const spreadsByNumber = useMemo(() => {
    const map = new Map<number, PersistedSpread>();
    const list = spreadsQuery.data?.spreads ?? [];
    for (const spread of list) {
      map.set(spread.spread_number, spread);
    }
    return map;
  }, [spreadsQuery.data]);

  const assetById = useMemo(() => {
    const map = new Map<string, AssetSummary>();
    for (const asset of assets) {
      map.set(asset.id, asset);
    }
    return map;
  }, [assets]);

  const ensureAssetLoaded = (assetId: string | undefined | null) => {
    if (!assetId) return;
    const asset = assetById.get(assetId);
    if (!asset) {
      setFeedback(`Asset ${assetId} not found in project.`);
      return;
    }
    onSelectAssetById(asset.id);
  };

  const handleLoadPage = () => {
    if (!projectId) return;
    const pageNumber = Number(pageNumberInput);
    if (!Number.isFinite(pageNumber) || pageNumber <= 0) {
      setFeedback('Provide a positive page number.');
      return;
    }
    const persisted = pagesByNumber.get(pageNumber);
    if (!persisted) {
      setFeedback(`Page ${pageNumber} has no stored layout yet.`);
      return;
    }
    ensureAssetLoaded(persisted.asset_id);
    dispatch(loadSettings(persisted.settings));
    setFeedback(`Loaded layout for page ${pageNumber}.`);
  };

  const handleSavePage = async () => {
    if (!projectId) return;
    const pageNumber = Number(pageNumberInput);
    if (!Number.isFinite(pageNumber) || pageNumber <= 0) {
      setFeedback('Provide a positive page number.');
      return;
    }
    try {
      const settings = buildSpreadSettingsFromState(spreadState);
      await putPage({
        id: projectId,
        pageNumber,
        page: {
          asset_id: selectedAssetId ?? undefined,
          settings,
        },
      }).unwrap();
      await pagesQuery.refetch();
      setFeedback(`Saved layout to page ${pageNumber}.`);
    } catch (error) {
      console.error('Failed to save page layout', error);
      setFeedback('Failed to save page layout. Check console for details.');
    }
  };

  const handleDeletePage = async () => {
    if (!projectId) return;
    const pageNumber = Number(pageNumberInput);
    if (!Number.isFinite(pageNumber) || pageNumber <= 0) {
      setFeedback('Provide a positive page number.');
      return;
    }
    try {
      await deletePage({ id: projectId, pageNumber }).unwrap();
      await pagesQuery.refetch();
      setFeedback(`Deleted page ${pageNumber} layout.`);
    } catch (error) {
      console.error('Failed to delete page layout', error);
      setFeedback('Failed to delete page layout.');
    }
  };

  const handleLoadSpread = () => {
    if (!projectId) return;
    const spreadNumber = Number(spreadNumberInput);
    if (!Number.isFinite(spreadNumber) || spreadNumber <= 0) {
      setFeedback('Provide a positive spread number.');
      return;
    }
    const persisted = spreadsByNumber.get(spreadNumber);
    if (!persisted) {
      setFeedback(`Spread ${spreadNumber} has no stored layout yet.`);
      return;
    }
    setSpreadLeftInput(persisted.left_page_number?.toString() ?? '');
    setSpreadRightInput(persisted.right_page_number?.toString() ?? '');
    dispatch(loadSettings(persisted.settings));
    setFeedback(`Loaded spread ${spreadNumber}.`);
  };

  const handleSaveSpread = async () => {
    if (!projectId) return;
    const spreadNumber = Number(spreadNumberInput);
    if (!Number.isFinite(spreadNumber) || spreadNumber <= 0) {
      setFeedback('Provide a positive spread number.');
      return;
    }
    const leftNumber = numberOrUndefined(spreadLeftInput);
    const rightNumber = numberOrUndefined(spreadRightInput);

    try {
      const settings = buildSpreadSettingsFromState(spreadState);
      await putSpread({
        id: projectId,
        spreadNumber,
        spread: {
          left_page_number: leftNumber,
          right_page_number: rightNumber,
          settings,
        },
      }).unwrap();
      await spreadsQuery.refetch();
      setFeedback(`Saved spread ${spreadNumber}.`);
    } catch (error) {
      console.error('Failed to save spread layout', error);
      setFeedback('Failed to save spread layout.');
    }
  };

  const handleDeleteSpread = async () => {
    if (!projectId) return;
    const spreadNumber = Number(spreadNumberInput);
    if (!Number.isFinite(spreadNumber) || spreadNumber <= 0) {
      setFeedback('Provide a positive spread number.');
      return;
    }
    try {
      await deleteSpread({ id: projectId, spreadNumber }).unwrap();
      await spreadsQuery.refetch();
      setFeedback(`Deleted spread ${spreadNumber}.`);
    } catch (error) {
      console.error('Failed to delete spread layout', error);
      setFeedback('Failed to delete spread layout.');
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow space-y-6">
      <div>
        <h3 className="text-lg font-semibold">💾 Persisted Layouts</h3>
        <p className="text-sm text-gray-500">Save the current settings to a page or spread, or load a persisted layout back into the designer.</p>
      </div>

      <section className="space-y-3">
        <header className="flex items-center justify-between">
          <h4 className="text-base font-semibold text-gray-800">Pages</h4>
          <span className="text-xs text-gray-500">Existing: {pagesQuery.data?.pages?.map((p) => p.page_number).join(', ') || 'none'}</span>
        </header>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <label className="text-sm text-gray-600">
            Page number
            <input
              type="number"
              min={1}
              className="mt-1 w-full border rounded-md px-3 py-2 text-sm"
              value={pageNumberInput}
              onChange={(event) => setPageNumberInput(event.target.value)}
              disabled={!projectId}
            />
          </label>
          <label className="text-sm text-gray-600">
            Bound asset
            <select
              className="mt-1 w-full border rounded-md px-3 py-2 text-sm"
              value={selectedAssetId ?? ''}
              onChange={(event) => onSelectAssetById(event.target.value)}
              disabled={!projectId || assets.length === 0}
            >
              <option value="">(none)</option>
              {assets.map((asset) => (
                <option key={asset.id} value={asset.id}>
                  {asset.name}
                </option>
              ))}
            </select>
          </label>
        </div>

        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={handleLoadPage}
            disabled={!projectId || pagesQuery.isLoading}
            className="px-3 py-2 text-sm rounded-md bg-gray-200 hover:bg-gray-300"
          >
            Load
          </button>
          <button
            type="button"
            onClick={handleSavePage}
            disabled={!projectId || putPageStatus.isLoading}
            className="px-3 py-2 text-sm rounded-md bg-primary-600 text-white hover:bg-primary-700"
          >
            Save
          </button>
          <button
            type="button"
            onClick={handleDeletePage}
            disabled={!projectId}
            className="px-3 py-2 text-sm rounded-md bg-red-100 text-red-700 hover:bg-red-200"
          >
            Delete
          </button>
        </div>
      </section>

      <section className="space-y-3 border-t border-gray-100 pt-4">
        <header className="flex items-center justify-between">
          <h4 className="text-base font-semibold text-gray-800">Spreads</h4>
          <span className="text-xs text-gray-500">Existing: {spreadsQuery.data?.spreads?.map((s) => s.spread_number).join(', ') || 'none'}</span>
        </header>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <label className="text-sm text-gray-600">
            Spread number
            <input
              type="number"
              min={1}
              className="mt-1 w-full border rounded-md px-3 py-2 text-sm"
              value={spreadNumberInput}
              onChange={(event) => setSpreadNumberInput(event.target.value)}
              disabled={!projectId}
            />
          </label>
          <label className="text-sm text-gray-600">
            Left page #
            <input
              type="number"
              min={1}
              className="mt-1 w-full border rounded-md px-3 py-2 text-sm"
              value={spreadLeftInput}
              onChange={(event) => setSpreadLeftInput(event.target.value)}
              disabled={!projectId}
            />
          </label>
          <label className="text-sm text-gray-600">
            Right page #
            <input
              type="number"
              min={1}
              className="mt-1 w-full border rounded-md px-3 py-2 text-sm"
              value={spreadRightInput}
              onChange={(event) => setSpreadRightInput(event.target.value)}
              disabled={!projectId}
            />
          </label>
        </div>

        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={handleLoadSpread}
            disabled={!projectId || spreadsQuery.isLoading}
            className="px-3 py-2 text-sm rounded-md bg-gray-200 hover:bg-gray-300"
          >
            Load
          </button>
          <button
            type="button"
            onClick={handleSaveSpread}
            disabled={!projectId || putSpreadStatus.isLoading}
            className="px-3 py-2 text-sm rounded-md bg-primary-600 text-white hover:bg-primary-700"
          >
            Save
          </button>
          <button
            type="button"
            onClick={handleDeleteSpread}
            disabled={!projectId}
            className="px-3 py-2 text-sm rounded-md bg-red-100 text-red-700 hover:bg-red-200"
          >
            Delete
          </button>
        </div>
      </section>

      {feedback && <p className="text-sm text-gray-600">{feedback}</p>}
    </div>
  );
};
