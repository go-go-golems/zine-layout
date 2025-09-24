import React, { CSSProperties, useEffect, useState } from 'react';
import { ImageUploadSection } from '../components/bookSpread/ImageUploadSection';
import { PaperSettingsPanel } from '../components/bookSpread/PaperSettingsPanel';
import { MarginControlsPanel } from '../components/bookSpread/MarginControlsPanel';
import { ImageControlsPanel } from '../components/bookSpread/ImageControlsPanel';
import { ImageInformationPanel } from '../components/bookSpread/ImageInformationPanel';
import { ExportPanel } from '../components/bookSpread/ExportPanel';
import { AlgorithmDebugPanel } from '../components/bookSpread/AlgorithmDebugPanel';
import { useAppSelector } from '../hooks/redux';
import {
  useGetPreviewSpreadQuery,
  useGetProjectsQuery,
  useLazyExportBookYamlQuery,
} from '../api';
import { usePreviewRequest } from '../utils/spreadRequestBuilder';
import { getCurrentDimensions } from '../utils/bookSpreadUtils';

const MAX_EDGE_SINGLE = 520;
const MAX_EDGE_PANEL = 420;

const computeFrameDimensions = (widthIn: number, heightIn: number, maxEdge: number): CSSProperties => {
  if (widthIn <= 0 || heightIn <= 0) {
    return { width: maxEdge, height: maxEdge * 1.2 };
  }
  if (widthIn >= heightIn) {
    const frameWidth = maxEdge;
    const frameHeight = Math.max(maxEdge * (heightIn / widthIn), maxEdge * 0.55);
    return { width: frameWidth, height: frameHeight };
  }
  const frameHeight = maxEdge;
  const frameWidth = Math.max(maxEdge * (widthIn / heightIn), maxEdge * 0.55);
  return { width: frameWidth, height: frameHeight };
};

const frameClassName = 'relative bg-white border-8 border-gray-200 rounded-3xl shadow-lg flex items-center justify-center overflow-hidden transition-all duration-300';

const renderFramedPreview = (label: string, url: string | undefined | null, frameStyle: CSSProperties) => (
  <div className="text-center">
    <div className="text-xs text-gray-500 mb-2">{label}</div>
    <div className={frameClassName} style={frameStyle}>
      {url ? (
        <img
          src={url}
          alt={`${label} preview`}
          className="w-full h-full object-contain"
        />
      ) : (
        <div className="flex flex-col items-center justify-center text-gray-400 gap-2">
          <span className="text-3xl">⏳</span>
          <span className="text-xs">Rendering…</span>
        </div>
      )}
    </div>
  </div>
);

export const BookSpreadDesigner: React.FC = () => {
  const { image, isSpread, paperSize, orientation } = useAppSelector((state) => state.bookSpread);

  const { data: projectsData, isLoading: projectsLoading } = useGetProjectsQuery();
  const [selectedProjectId, setSelectedProjectId] = useState('');
  const [baseDir, setBaseDir] = useState('');
  const [exportFeedback, setExportFeedback] = useState<{ text: string; tone: 'success' | 'error' | 'info' } | null>(null);
  const [triggerExportYaml, { isFetching: exportingYaml }] = useLazyExportBookYamlQuery();

  useEffect(() => {
    if (!selectedProjectId && projectsData?.projects?.length) {
      setSelectedProjectId(projectsData.projects[0].id);
    }
  }, [projectsData, selectedProjectId]);

  const projectOptions = projectsData?.projects ?? [];

  // Get preview requests for different panels
  const leftRequest = usePreviewRequest(600, 'left');
  const rightRequest = usePreviewRequest(600, 'right');
  const singleRequest = usePreviewRequest(600, 'single');
  
  // Get preview images from backend
  const { data: leftPreviewUrl, isLoading: leftLoading, error: leftError } = useGetPreviewSpreadQuery(
    leftRequest!,
    { skip: !leftRequest || !isSpread }
  );
  
  const { data: rightPreviewUrl, isLoading: rightLoading, error: rightError } = useGetPreviewSpreadQuery(
    rightRequest!,
    { skip: !rightRequest || !isSpread }
  );
  
  const { data: singlePreviewUrl, isLoading: singleLoading, error: singleError } = useGetPreviewSpreadQuery(
    singleRequest!,
    { skip: !singleRequest || isSpread }
  );
  
  const isLoading = isSpread ? (leftLoading || rightLoading) : singleLoading;
  const error = isSpread ? (leftError || rightError) : singleError;

  const pageDimensions = getCurrentDimensions(paperSize, orientation, false);
  const singleFrameStyle = computeFrameDimensions(pageDimensions.width, pageDimensions.height, MAX_EDGE_SINGLE);
  const panelFrameStyle = computeFrameDimensions(pageDimensions.width, pageDimensions.height, MAX_EDGE_PANEL);

  const handleExportYaml = async () => {
    if (!selectedProjectId) {
      setExportFeedback({ text: 'Select a project before exporting.', tone: 'error' });
      return;
    }
    setExportFeedback(null);
    try {
      const yaml = await triggerExportYaml({
        id: selectedProjectId,
        baseDir: baseDir.trim() || undefined,
      }).unwrap();

      const fileName = `${selectedProjectId}-book.yaml`;
      const blob = new Blob([yaml], { type: 'text/yaml;charset=utf-8' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = fileName;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);

      try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
          await navigator.clipboard.writeText(yaml);
          setExportFeedback({ text: 'Book YAML downloaded and copied to clipboard.', tone: 'success' });
        } else {
          setExportFeedback({ text: 'Book YAML downloaded. Copy it from the downloaded file if needed.', tone: 'success' });
        }
      } catch (_err) {
        setExportFeedback({ text: 'Book YAML downloaded. Clipboard copy is unavailable in this browser.', tone: 'success' });
      }
    } catch (_err) {
      setExportFeedback({ text: 'Failed to export book YAML.', tone: 'error' });
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">📖 Photobook Spread Designer</h1>
        
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Controls Panel */}
          <div className="space-y-6">
            <ImageUploadSection />
            <PaperSettingsPanel />
            <MarginControlsPanel />
            <div className="bg-white p-4 rounded-lg shadow border border-gray-100">
              <h3 className="text-base font-semibold text-gray-800 mb-3">Export Book YAML</h3>
              <label className="block text-sm font-medium text-gray-700 mb-1" htmlFor="project-select">
                Project
              </label>
              <select
                id="project-select"
                className="w-full border rounded-md px-3 py-2 text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-300"
                value={selectedProjectId}
                onChange={(e) => setSelectedProjectId(e.target.value)}
                disabled={projectsLoading || projectOptions.length === 0}
              >
                {projectsLoading ? (
                  <option value="">Loading projects…</option>
                ) : projectOptions.length === 0 ? (
                  <option value="">No projects available</option>
                ) : (
                  projectOptions.map((project) => (
                    <option key={project.id} value={project.id}>
                      {project.name || project.id}
                    </option>
                  ))
                )}
              </select>

              <label className="block text-sm font-medium text-gray-700 mt-3 mb-1" htmlFor="base-dir-input">
                Base directory (optional)
              </label>
              <input
                id="base-dir-input"
                type="text"
                value={baseDir}
                onChange={(e) => setBaseDir(e.target.value)}
                placeholder="Defaults to server data root"
                className="w-full border rounded-md px-3 py-2 text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-300"
              />

              <button
                type="button"
                onClick={handleExportYaml}
                disabled={!selectedProjectId || exportingYaml || projectOptions.length === 0}
                className={`mt-4 inline-flex items-center justify-center w-full rounded-md px-3 py-2 text-sm font-medium text-white shadow-sm transition-colors duration-150 ${
                  !selectedProjectId || exportingYaml || projectOptions.length === 0
                    ? 'bg-gray-300 cursor-not-allowed'
                    : 'bg-primary-600 hover:bg-primary-700'
                }`}
              >
                {exportingYaml ? 'Exporting…' : 'Export YAML'}
              </button>

              {exportFeedback && (
                <p
                  className={`mt-2 text-sm ${
                    exportFeedback.tone === 'success'
                      ? 'text-green-600'
                      : exportFeedback.tone === 'error'
                      ? 'text-red-600'
                      : 'text-gray-600'
                  }`}
                >
                  {exportFeedback.text}
                </p>
              )}
            </div>
          </div>

          {/* Preview Panel */}
          <div className="lg:col-span-2">
            <div className="bg-white p-6 rounded-lg shadow">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold">🖼️ Preview</h3>
                <ExportPanel />
              </div>

              <div className="space-y-6">
                {/* Image Controls */}
                <ImageControlsPanel />

                {/* Preview */}
                <div className="flex justify-center">
                  {image ? (
                    <div className="text-center">
                      {isLoading && (
                        <div className="text-gray-500 py-12">
                          <div className="text-4xl mb-4">⏳</div>
                          <p>Generating preview...</p>
                        </div>
                      )}
                      {error && (
                        <div className="text-red-500 py-12">
                          <div className="text-4xl mb-4">❌</div>
                          <p>Error generating preview</p>
                        </div>
                      )}
                      {!isLoading && !error && (
                        <div className="preview-container">
                          {isSpread ? (
                            <div className="flex flex-wrap gap-6 items-start justify-center">
                              {renderFramedPreview('Left Page', leftPreviewUrl, panelFrameStyle)}
                              {renderFramedPreview('Right Page', rightPreviewUrl, panelFrameStyle)}
                            </div>
                          ) : (
                            renderFramedPreview('Single Page', singlePreviewUrl, singleFrameStyle)
                          )}
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="text-center text-gray-500 py-12">
                      <div className="text-6xl mb-4">🖼️</div>
                      <p>Upload an image to start designing your photobook spread</p>
                    </div>
                  )}
                </div>

                {/* Image Information */}
                <ImageInformationPanel />

                {/* Debug Panel */}
                <AlgorithmDebugPanel />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
