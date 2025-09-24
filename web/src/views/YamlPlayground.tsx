import React, { useMemo, useState } from 'react';
import type { SerializedError } from '@reduxjs/toolkit';
import { useRenderYamlMutation, type YamlRenderSpread } from '../api';

const SAMPLE_YAML = `version: "0.2"

defaults:
  paper:
    width_in: 8.0
    height_in: 10.0
    orientation: portrait
    dpi: 300
  margins:
    top_in: 0.25
    right_in: 0.25
    bottom_in: 0.25
    left_in: 0.25
  spread:
    is_spread: false
    gutter_in: 0.0
  crop:
    ratio: original
    to_fill: false
  scale:
    user_scale: 1.0
  position:
    x: 0.0
    y: 0.0
    units: normalized
  export:
    format: png
    quality: 90
    background: white
    out_dir: ./out
    filename_template: "{name}-{panel}.{ext}"

spreads:
  - name: page-0001
    image: ./projects/PRJ/images/0001.png
  - name: page-0002
    image: ./projects/PRJ/images/0002.png
    spread:
      is_spread: true
      gutter_in: 0.25
`;

const formatError = (error: unknown): string => {
  if (!error) return '';
  if (typeof error === 'string') return error;
  if ((error as SerializedError).message) {
    return (error as SerializedError).message as string;
  }
  if (typeof error === 'object') {
    const maybe = error as Record<string, unknown>;
    const status = maybe['status'];
    const data = maybe['data'];
    if (typeof status !== 'undefined' && typeof data !== 'undefined') {
      return `${String(status)}: ${String(data)}`;
    }
  }
  try {
    return JSON.stringify(error);
  } catch {
    return String(error);
  }
};

const SpreadPreviewView: React.FC<{ spread: YamlRenderSpread }> = ({ spread }) => {
  return (
    <div className="mt-6">
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-lg font-semibold text-gray-800">Simple Algorithm</h4>
        <span className="text-xs text-gray-500">{spread.panels.length} panel(s)</span>
      </div>
      <div className="flex flex-wrap gap-4 mb-4">
        {spread.panels.map((panel) => (
          <div key={`simple-${panel.panel}`} className="bg-gray-50 border border-gray-200 rounded-lg p-3 shadow-sm max-w-xs">
            <div className="text-sm font-medium text-gray-700 mb-2">
              {panel.panel.toUpperCase()} · {panel.width}×{panel.height}
            </div>
            <img
              src={panel.data_url}
              alt={`Simple ${panel.panel}`}
              className="rounded border border-gray-200 max-h-64 object-contain"
            />
            <div className="mt-2 text-[11px] text-gray-500">{panel.mime_type}</div>
          </div>
        ))}
      </div>
      <details className="mb-2">
        <summary className="cursor-pointer text-sm font-medium text-gray-700">Result JSON</summary>
        <pre className="bg-gray-900 text-gray-100 text-xs p-3 rounded mt-2 overflow-auto max-h-64">
          {JSON.stringify(spread.result ?? {}, null, 2)}
        </pre>
      </details>
      {spread.trace && spread.trace.length > 0 && (
        <details>
          <summary className="cursor-pointer text-sm font-medium text-gray-700">Trace</summary>
          <pre className="bg-gray-900 text-gray-100 text-xs p-3 rounded mt-2 overflow-auto max-h-48 whitespace-pre-wrap">
            {spread.trace.join('\n')}
          </pre>
        </details>
      )}
    </div>
  );
};

export const YamlPlayground: React.FC = () => {
  const [yamlInput, setYamlInput] = useState('');
  const [baseDir, setBaseDir] = useState('');
  const [renderYaml, { data, isLoading, error }] = useRenderYamlMutation();

  const errorMessage = useMemo(() => formatError(error), [error]);

  const handleRun = () => {
    if (!yamlInput.trim()) {
      alert('Please paste YAML before rendering.');
      return;
    }
    renderYaml({ yaml: yamlInput, base_dir: baseDir.trim() || undefined });
  };

  const handleLoadSample = () => {
    setYamlInput(SAMPLE_YAML);
  };

  return (
    <div className="space-y-6">
      <div className="bg-white border border-gray-200 rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-semibold text-gray-900">YAML Playground</h2>
          <button
            onClick={handleLoadSample}
            className="px-3 py-1 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded"
          >
            Load Sample
          </button>
        </div>
        <p className="text-sm text-gray-600 mb-4">
          Paste a Simple YAML configuration below and render it with the simple algorithm.
          The server will return metadata and preview panels so you can inspect the output.
        </p>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">YAML</label>
            <textarea
              className="w-full h-64 font-mono text-sm border border-gray-300 rounded-lg shadow-sm focus:ring-primary-500 focus:border-primary-500 p-3"
              value={yamlInput}
              onChange={(e) => setYamlInput(e.target.value)}
              placeholder="Paste Simple YAML here"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">Base directory (optional)</label>
              <input
                className="w-full border border-gray-300 rounded-lg shadow-sm focus:ring-primary-500 focus:border-primary-500 px-3 py-2 text-sm"
                value={baseDir}
                onChange={(e) => setBaseDir(e.target.value)}
                placeholder="Defaults to server data root"
              />
            </div>
            <div className="flex md:justify-end">
              <button
                onClick={handleRun}
                disabled={isLoading}
                className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg shadow disabled:opacity-50"
              >
                {isLoading ? 'Rendering…' : 'Render YAML'}
              </button>
            </div>
          </div>

          {errorMessage && (
            <div className="bg-red-50 border border-red-200 text-red-700 text-sm px-3 py-2 rounded">
              {errorMessage}
            </div>
          )}
        </div>
      </div>

      {data && data.spreads.length > 0 && (
        <div className="space-y-6">
          {data.spreads.map((spread) => (
            <div key={spread.name} className="bg-white border border-gray-200 rounded-lg shadow-sm p-6">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h3 className="text-xl font-semibold text-gray-900">Spread: {spread.name}</h3>
                  <p className="text-sm text-gray-500">Image: {spread.image_path}</p>
                </div>
              </div>
              <SpreadPreviewView spread={spread} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
