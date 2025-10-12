import React, { useState } from 'react';
import { Button, Card, CardBody, CardHeader } from '../../components/ui';

interface ZineTabProps {
  projectId: string;
}

// Dummy data for demonstration
const DUMMY_ZINES = [
  {
    id: 'zne-001',
    name: 'Summer Photobook',
    description: 'Main photobook with 24 pages',
    pages: 24,
    created_at: '2025-10-11T09:00:00Z',
  },
  {
    id: 'zne-002',
    name: 'Draft v2',
    description: 'Alternative layout',
    pages: 18,
    created_at: '2025-10-11T11:30:00Z',
  },
];

const DUMMY_IMPOSITION_TEMPLATES = [
  {
    id: 'zimp-001',
    name: '8-Page Fold',
    description: 'Standard 8-page folded zine from single sheet',
    pages: 8,
    icon: '📑',
  },
  {
    id: 'zimp-002',
    name: '16-Page Booklet',
    description: 'Two-sheet booklet with staple binding',
    pages: 16,
    icon: '📕',
  },
  {
    id: 'zimp-003',
    name: 'Simple Stack',
    description: 'No imposition, just sequential pages',
    pages: 0,
    icon: '📚',
  },
];

const DUMMY_PAGES = [
  { id: 'lpg-001', name: 'Cover', type: 'cover' },
  { id: 'lpg-002', name: 'Title', type: 'title' },
  { id: 'lpg-003', name: 'Spread 1', type: 'spread' },
  { id: 'lpg-004', name: 'Spread 2', type: 'spread' },
  { id: 'lpg-005', name: 'Photo 1', type: 'photo' },
  { id: 'lpg-006', name: 'Photo 2', type: 'photo' },
];

export const ZineTab: React.FC<ZineTabProps> = ({ projectId }) => {
  const [selectedZine, setSelectedZine] = useState<string | null>(DUMMY_ZINES[0]?.id ?? null);
  const [isCreating, setIsCreating] = useState(false);
  const [selectedImposition, setSelectedImposition] = useState<string>('zimp-001');
  const [exportFormat, setExportFormat] = useState<'pdf' | 'png' | 'zip'>('pdf');
  const [includeCropMarks, setIncludeCropMarks] = useState(true);
  const [includeBleed, setIncludeBleed] = useState(true);

  const selectedZineData = DUMMY_ZINES.find((z) => z.id === selectedZine);
  const selectedImpositionData = DUMMY_IMPOSITION_TEMPLATES.find(
    (t) => t.id === selectedImposition
  );

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Zines</h2>
          <p className="text-sm text-gray-500 mt-1">
            Assemble pages into complete books and export for print
          </p>
        </div>
        <Button onClick={() => setIsCreating(!isCreating)} variant={isCreating ? 'secondary' : 'primary'}>
          {isCreating ? 'Cancel' : '+ Create Zine'}
        </Button>
      </div>

      {/* Phase 3/4 Notice */}
      <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
        <div className="flex items-start space-x-3">
          <div className="text-2xl">ℹ️</div>
          <div>
            <h3 className="font-semibold text-purple-900 mb-1">Phase 3 & 4 Preview</h3>
            <p className="text-sm text-purple-800">
              This is a preview of the Zine interface with dummy data. Full functionality including imposition and PDF export coming in Phase 3/4 when backend rendering and export systems are implemented.
            </p>
          </div>
        </div>
      </div>

      {/* Zine Selector */}
      <Card>
        <CardHeader>
          <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
            Your Zines
          </h3>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {DUMMY_ZINES.map((zine) => (
              <button
                key={zine.id}
                type="button"
                onClick={() => setSelectedZine(zine.id)}
                className={`text-left border-2 rounded-lg p-4 transition-all ${
                  selectedZine === zine.id
                    ? 'border-primary-600 bg-primary-50 shadow-md'
                    : 'border-gray-300 hover:border-primary-400 bg-white'
                }`}
              >
                <div className="flex items-start justify-between mb-2">
                  <h4 className="font-semibold text-gray-900">{zine.name}</h4>
                  {selectedZine === zine.id && <span className="text-primary-600 text-xl">●</span>}
                </div>
                {zine.description && (
                  <p className="text-sm text-gray-600 mb-2">{zine.description}</p>
                )}
                <p className="text-xs text-gray-500">{zine.pages} pages</p>
              </button>
            ))}
          </div>
        </CardBody>
      </Card>

      {/* Main Content: Three-Column Layout */}
      {selectedZineData && (
        <div className="grid lg:grid-cols-3 gap-6">
          {/* Left: Page Sequence */}
          <Card>
            <CardHeader>
              <h3 className="text-lg font-semibold text-gray-900">Zine Pages</h3>
              <p className="text-sm text-gray-500 mt-1">
                {DUMMY_PAGES.length} pages in this zine
              </p>
            </CardHeader>
            <CardBody>
              <div className="space-y-2">
                {DUMMY_PAGES.map((page, index) => (
                  <div
                    key={page.id}
                    className="flex items-center space-x-3 p-3 border border-gray-200 rounded-lg bg-white hover:border-primary-300 cursor-pointer"
                  >
                    <div className="w-12 h-12 bg-gray-100 rounded flex items-center justify-center text-xs font-medium text-gray-600">
                      P{index + 1}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900">{page.name}</p>
                      <p className="text-xs text-gray-500 capitalize">{page.type}</p>
                    </div>
                    <div className="flex flex-col space-y-1">
                      <button
                        type="button"
                        className="text-xs text-gray-400 hover:text-gray-600"
                        disabled={index === 0}
                      >
                        ↑
                      </button>
                      <button
                        type="button"
                        className="text-xs text-gray-400 hover:text-gray-600"
                        disabled={index === DUMMY_PAGES.length - 1}
                      >
                        ↓
                      </button>
                    </div>
                  </div>
                ))}
                <Button
                  size="sm"
                  variant="secondary"
                  className="w-full mt-4"
                  onClick={() => alert('Add page - Coming in Phase 3')}
                >
                  + Add Page
                </Button>
              </div>
            </CardBody>
          </Card>

          {/* Center: Preview */}
          <Card>
            <CardHeader>
              <h3 className="text-lg font-semibold text-gray-900">Preview</h3>
            </CardHeader>
            <CardBody>
              <div className="bg-gray-50 p-6 rounded-lg flex items-center justify-center min-h-[400px]">
                <div className="text-center">
                  <div className="w-64 h-80 bg-white border-2 border-gray-300 shadow-lg rounded-lg mb-4 flex items-center justify-center">
                    <div className="text-gray-400">
                      <div className="text-4xl mb-2">📄</div>
                      <p className="text-sm">Page preview</p>
                      <p className="text-xs mt-1">Click page to view</p>
                    </div>
                  </div>
                  <div className="text-sm text-gray-600">
                    Page 1 of {DUMMY_PAGES.length}
                  </div>
                  <div className="flex items-center justify-center space-x-2 mt-4">
                    <Button size="sm" variant="secondary">
                      ‹ Prev
                    </Button>
                    <Button size="sm" variant="secondary">
                      Next ›
                    </Button>
                  </div>
                </div>
              </div>
            </CardBody>
          </Card>

          {/* Right: Imposition & Export */}
          <Card>
            <CardHeader>
              <h3 className="text-lg font-semibold text-gray-900">Imposition & Export</h3>
            </CardHeader>
            <CardBody className="space-y-6">
              {/* Imposition Template */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Imposition Template
                </label>
                <div className="space-y-2">
                  {DUMMY_IMPOSITION_TEMPLATES.map((template) => (
                    <button
                      key={template.id}
                      type="button"
                      onClick={() => setSelectedImposition(template.id)}
                      className={`w-full text-left p-3 border-2 rounded-lg transition-all ${
                        selectedImposition === template.id
                          ? 'border-primary-600 bg-primary-50'
                          : 'border-gray-300 hover:border-primary-400 bg-white'
                      }`}
                    >
                      <div className="flex items-center space-x-3">
                        <div className="text-2xl">{template.icon}</div>
                        <div className="flex-1">
                          <p className="font-medium text-sm text-gray-900">{template.name}</p>
                          <p className="text-xs text-gray-500">{template.description}</p>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>

              {/* Imposition Preview */}
              {selectedImpositionData && (
                <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
                  <p className="text-xs font-medium text-gray-700 mb-3 uppercase">
                    Layout Preview
                  </p>
                  {selectedImposition === 'zimp-001' && (
                    <div className="space-y-2">
                      <div className="grid grid-cols-4 gap-1 text-xs text-center">
                        <div className="bg-gray-200 p-2 rounded">8</div>
                        <div className="bg-gray-200 p-2 rounded">1</div>
                        <div className="bg-gray-200 p-2 rounded">2</div>
                        <div className="bg-gray-200 p-2 rounded">7</div>
                      </div>
                      <div className="text-center text-xs text-gray-500">↓ fold →</div>
                      <div className="grid grid-cols-4 gap-1 text-xs text-center">
                        <div className="bg-gray-200 p-2 rounded">6</div>
                        <div className="bg-gray-200 p-2 rounded">3</div>
                        <div className="bg-gray-200 p-2 rounded">4</div>
                        <div className="bg-gray-200 p-2 rounded">5</div>
                      </div>
                      <p className="text-xs text-gray-500 mt-2">
                        Single sheet, folded twice
                      </p>
                    </div>
                  )}
                  {selectedImposition === 'zimp-002' && (
                    <div className="text-xs text-gray-500 text-center py-4">
                      16-page booklet layout (2 sheets)
                    </div>
                  )}
                  {selectedImposition === 'zimp-003' && (
                    <div className="text-xs text-gray-500 text-center py-4">
                      Sequential page order (no folding)
                    </div>
                  )}
                </div>
              )}

              <hr className="border-gray-200" />

              {/* Export Options */}
              <div className="space-y-4">
                <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
                  Export Format
                </h4>

                <div className="space-y-2">
                  <label className="flex items-center space-x-2">
                    <input
                      type="radio"
                      checked={exportFormat === 'pdf'}
                      onChange={() => setExportFormat('pdf')}
                      className="text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm text-gray-700">PDF (Print-Ready)</span>
                  </label>
                  <label className="flex items-center space-x-2">
                    <input
                      type="radio"
                      checked={exportFormat === 'png'}
                      onChange={() => setExportFormat('png')}
                      className="text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm text-gray-700">PNG Sequence</span>
                  </label>
                  <label className="flex items-center space-x-2">
                    <input
                      type="radio"
                      checked={exportFormat === 'zip'}
                      onChange={() => setExportFormat('zip')}
                      className="text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm text-gray-700">ZIP Archive</span>
                  </label>
                </div>

                <div className="space-y-2 pt-2">
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={includeCropMarks}
                      onChange={(e) => setIncludeCropMarks(e.target.checked)}
                      className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm text-gray-700">Include crop marks</span>
                  </label>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={includeBleed}
                      onChange={(e) => setIncludeBleed(e.target.checked)}
                      className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm text-gray-700">Include bleed area</span>
                  </label>
                </div>
              </div>

              <hr className="border-gray-200" />

              {/* Export Buttons */}
              <div className="space-y-2">
                <Button
                  className="w-full"
                  onClick={() =>
                    alert(
                      `Export as ${exportFormat.toUpperCase()} - Coming in Phase 4\n\nWould export: "${selectedZineData?.name}" with ${selectedImpositionData?.name}`
                    )
                  }
                >
                  Download Zine
                </Button>
                <Button
                  variant="secondary"
                  className="w-full"
                  onClick={() => alert('Preview export - Coming in Phase 4')}
                >
                  Preview Export
                </Button>
              </div>

              {/* Export Info */}
              <div className="bg-gray-50 p-3 rounded border border-gray-200">
                <p className="text-xs font-medium text-gray-700 mb-2">Export Preview:</p>
                <div className="space-y-1 text-xs text-gray-600">
                  <p>• Format: {exportFormat.toUpperCase()}</p>
                  <p>• Imposition: {selectedImpositionData?.name}</p>
                  <p>• Pages: {DUMMY_PAGES.length}</p>
                  <p>• Crop marks: {includeCropMarks ? 'Yes' : 'No'}</p>
                  <p>• Bleed: {includeBleed ? 'Yes' : 'No'}</p>
                </div>
              </div>
            </CardBody>
          </Card>
        </div>
      )}

      {/* Page Thumbnails */}
      {selectedZineData && (
        <Card>
          <CardHeader>
            <h3 className="text-lg font-semibold text-gray-900">Page Sequence</h3>
            <p className="text-sm text-gray-500 mt-1">
              Drag to reorder • Click to preview
            </p>
          </CardHeader>
          <CardBody>
            <div className="flex space-x-3 overflow-x-auto pb-2">
              {DUMMY_PAGES.map((page, index) => (
                <div
                  key={page.id}
                  className="flex-shrink-0 w-24 cursor-pointer group"
                >
                  <div className="w-24 h-32 bg-gray-100 border-2 border-gray-300 rounded-lg mb-2 flex items-center justify-center group-hover:border-primary-400 transition-colors">
                    <div className="text-center">
                      <div className="text-2xl mb-1">
                        {page.type === 'cover' ? '📕' : page.type === 'spread' ? '📄' : '🖼️'}
                      </div>
                      <p className="text-xs text-gray-600">P{index + 1}</p>
                    </div>
                  </div>
                  <p className="text-xs text-gray-700 text-center truncate">{page.name}</p>
                </div>
              ))}
            </div>
          </CardBody>
        </Card>
      )}

      {/* Features Coming Soon */}
      <Card>
        <CardHeader>
          <h3 className="text-lg font-semibold text-gray-900">Coming in Phase 3 & 4</h3>
        </CardHeader>
        <CardBody>
          <div className="grid md:grid-cols-2 gap-6">
            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-3">Phase 3 - Zine Assembly:</h4>
              <ul className="space-y-2 text-sm text-gray-600">
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Create zine from laid-out pages</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Drag-and-drop page reordering</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Page thumbnail previews</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Add/remove pages from zine</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Zine metadata (name, description)</span>
                </li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-3">Phase 4 - Print Export:</h4>
              <ul className="space-y-2 text-sm text-gray-600">
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Imposition algorithm (8-page fold, 16-page booklet)</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>PDF generation with embedded fonts</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Crop marks and bleed area rendering</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Color profile embedding (CMYK/RGB)</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Print-ready export validation</span>
                </li>
              </ul>
            </div>
          </div>

          <div className="mt-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
            <p className="text-sm text-gray-700">
              <strong>Backend Status:</strong> Partial implementation exists for zine persistence. See{' '}
              <code className="text-xs bg-gray-200 px-1 py-0.5 rounded">
                pkg/services/zines.go
              </code>{' '}
              and{' '}
              <code className="text-xs bg-gray-200 px-1 py-0.5 rounded">
                cmd/zine-layout/cmds/workflow/zines.go
              </code>
              . REST API and rendering pipeline are still needed.
            </p>
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

