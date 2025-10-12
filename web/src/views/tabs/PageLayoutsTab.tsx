import React, { useState } from 'react';
import { Button, Card, CardBody, CardHeader } from '../../components/ui';

interface PageLayoutsTabProps {
  projectId: string;
}

// Dummy data for demonstration
const DUMMY_PAGE_TEMPLATES = [
  {
    id: 'ptpl-001',
    name: 'Single Image',
    description: 'One image per page',
    slots: 1,
    icon: '📄',
    layout: 'single',
  },
  {
    id: 'ptpl-002',
    name: '2-Up Spread',
    description: 'Two images side-by-side',
    slots: 2,
    icon: '📄📄',
    layout: 'spread',
  },
  {
    id: 'ptpl-003',
    name: '4-Up Grid',
    description: 'Four images in a 2×2 grid',
    slots: 4,
    icon: '⊞',
    layout: 'grid-2x2',
  },
  {
    id: 'ptpl-004',
    name: '3-Up Vertical',
    description: 'Three images stacked vertically',
    slots: 3,
    icon: '☰',
    layout: 'vertical',
  },
];

const DUMMY_PAGES = [
  {
    id: 'lpg-001',
    name: 'Cover Page',
    template: 'Single Image',
    images: ['img01.png'],
    created_at: '2025-10-11T10:30:00Z',
  },
  {
    id: 'lpg-002',
    name: 'Title Spread',
    template: '2-Up Spread',
    images: ['img02.png', 'img03.png'],
    created_at: '2025-10-11T10:32:00Z',
  },
  {
    id: 'lpg-003',
    name: 'Photo Spread 1',
    template: '2-Up Spread',
    images: ['img04.png', 'img05.png'],
    created_at: '2025-10-11T10:35:00Z',
  },
  {
    id: 'lpg-004',
    name: 'Gallery Grid',
    template: '4-Up Grid',
    images: ['img06.png', 'img07.png', 'img08.png', 'img09.png'],
    created_at: '2025-10-11T10:40:00Z',
  },
];

export const PageLayoutsTab: React.FC<PageLayoutsTabProps> = ({ projectId }) => {
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);
  const [selectedPage, setSelectedPage] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [composerSlots, setComposerSlots] = useState<(string | null)[]>([null]);

  const selectedTemplateData = DUMMY_PAGE_TEMPLATES.find((t) => t.id === selectedTemplate);

  const handleSelectTemplate = (templateId: string) => {
    setSelectedTemplate(templateId);
    const template = DUMMY_PAGE_TEMPLATES.find((t) => t.id === templateId);
    if (template) {
      setComposerSlots(Array(template.slots).fill(null));
      setIsCreating(true);
    }
  };

  const handleSlotClick = (index: number) => {
    // In real implementation, this would open a modal to select a laid-out image
    alert(`Slot ${index + 1}: In production, this would open a selector for laid-out images.`);
  };

  const renderComposerLayout = () => {
    if (!selectedTemplateData) return null;

    const { layout, slots } = selectedTemplateData;

    if (layout === 'single') {
      return (
        <div className="w-full aspect-[4/5] border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center hover:border-primary-400 cursor-pointer transition-colors"
          onClick={() => handleSlotClick(0)}
        >
          {composerSlots[0] ? (
            <div className="text-gray-700">{composerSlots[0]}</div>
          ) : (
            <div className="text-gray-400 text-center">
              <div className="text-4xl mb-2">+</div>
              <div className="text-sm">Click to add image</div>
            </div>
          )}
        </div>
      );
    }

    if (layout === 'spread') {
      return (
        <div className="grid grid-cols-2 gap-2 w-full">
          {[0, 1].map((index) => (
            <div
              key={index}
              className="aspect-[4/5] border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center hover:border-primary-400 cursor-pointer transition-colors"
              onClick={() => handleSlotClick(index)}
            >
              {composerSlots[index] ? (
                <div className="text-gray-700">{composerSlots[index]}</div>
              ) : (
                <div className="text-gray-400 text-center">
                  <div className="text-2xl mb-1">+</div>
                  <div className="text-xs">Slot {index + 1}</div>
                </div>
              )}
            </div>
          ))}
        </div>
      );
    }

    if (layout === 'grid-2x2') {
      return (
        <div className="grid grid-cols-2 gap-2 w-full">
          {[0, 1, 2, 3].map((index) => (
            <div
              key={index}
              className="aspect-square border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center hover:border-primary-400 cursor-pointer transition-colors"
              onClick={() => handleSlotClick(index)}
            >
              {composerSlots[index] ? (
                <div className="text-gray-700 text-sm">{composerSlots[index]}</div>
              ) : (
                <div className="text-gray-400 text-center">
                  <div className="text-xl mb-1">+</div>
                  <div className="text-xs">Slot {index + 1}</div>
                </div>
              )}
            </div>
          ))}
        </div>
      );
    }

    if (layout === 'vertical') {
      return (
        <div className="flex flex-col gap-2 w-full">
          {[0, 1, 2].map((index) => (
            <div
              key={index}
              className="aspect-[16/9] border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center hover:border-primary-400 cursor-pointer transition-colors"
              onClick={() => handleSlotClick(index)}
            >
              {composerSlots[index] ? (
                <div className="text-gray-700">{composerSlots[index]}</div>
              ) : (
                <div className="text-gray-400 text-center">
                  <div className="text-2xl mb-1">+</div>
                  <div className="text-xs">Slot {index + 1}</div>
                </div>
              )}
            </div>
          ))}
        </div>
      );
    }

    return null;
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Page Layouts</h2>
          <p className="text-sm text-gray-500 mt-1">
            Compose multi-image pages using layout templates
          </p>
        </div>
        <Button onClick={() => setIsCreating(!isCreating)} variant={isCreating ? 'secondary' : 'primary'}>
          {isCreating ? 'Cancel' : '+ Create Page'}
        </Button>
      </div>

      {/* Phase 3 Notice */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex items-start space-x-3">
          <div className="text-2xl">ℹ️</div>
          <div>
            <h3 className="font-semibold text-blue-900 mb-1">Phase 3 Preview</h3>
            <p className="text-sm text-blue-800">
              This is a preview of the Page Layouts interface with dummy data. Full functionality coming in Phase 3 when backend page composition is implemented.
            </p>
          </div>
        </div>
      </div>

      {/* Page Composer (when creating) */}
      {isCreating && (
        <div className="grid lg:grid-cols-2 gap-6">
          {/* Left: Template Selection + Composer */}
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">Select Page Template</h3>
              </CardHeader>
              <CardBody>
                <div className="grid grid-cols-2 gap-3">
                  {DUMMY_PAGE_TEMPLATES.map((template) => (
                    <button
                      key={template.id}
                      type="button"
                      onClick={() => handleSelectTemplate(template.id)}
                      className={`p-4 border-2 rounded-lg text-left transition-all ${
                        selectedTemplate === template.id
                          ? 'border-primary-600 bg-primary-50 shadow-md'
                          : 'border-gray-300 hover:border-primary-400 bg-white'
                      }`}
                    >
                      <div className="text-3xl mb-2 text-center">{template.icon}</div>
                      <h4 className="font-semibold text-gray-900 text-sm">{template.name}</h4>
                      <p className="text-xs text-gray-500 mt-1">{template.description}</p>
                      <p className="text-xs text-gray-400 mt-2">{template.slots} slot{template.slots > 1 ? 's' : ''}</p>
                    </button>
                  ))}
                </div>
              </CardBody>
            </Card>

            {selectedTemplateData && (
              <Card>
                <CardHeader>
                  <h3 className="text-lg font-semibold text-gray-900">Page Composer</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {selectedTemplateData.name} - {selectedTemplateData.slots} image slots
                  </p>
                </CardHeader>
                <CardBody>
                  <div className="bg-gray-50 p-6 rounded-lg">
                    {renderComposerLayout()}
                  </div>
                  <div className="mt-4 text-sm text-gray-600">
                    <p className="mb-2">In production, you will:</p>
                    <ul className="list-disc list-inside space-y-1 text-xs text-gray-500">
                      <li>Click slots to assign laid-out images</li>
                      <li>Drag laid-out images from Image Layouts tab</li>
                      <li>Reorder images within the page</li>
                      <li>See live preview of final page composition</li>
                    </ul>
                  </div>
                  <div className="mt-4 flex justify-end space-x-3">
                    <Button variant="secondary" onClick={() => setIsCreating(false)}>
                      Cancel
                    </Button>
                    <Button disabled>
                      Save Page (Not Yet Implemented)
                    </Button>
                  </div>
                </CardBody>
              </Card>
            )}
          </div>

          {/* Right: Available Laid-Out Images */}
          <Card>
            <CardHeader>
              <h3 className="text-lg font-semibold text-gray-900">Available Laid-Out Images</h3>
              <p className="text-sm text-gray-500 mt-1">
                Drag images to page slots (Phase 3)
              </p>
            </CardHeader>
            <CardBody>
              <div className="space-y-2">
                {['loi-001', 'loi-002', 'loi-003', 'loi-004', 'loi-005'].map((id, index) => (
                  <div
                    key={id}
                    className="flex items-center space-x-3 p-3 border border-gray-200 rounded-lg bg-white hover:border-primary-300 cursor-move"
                  >
                    <div className="w-16 h-16 bg-gray-200 rounded flex items-center justify-center text-xs text-gray-400">
                      Preview
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        img{String(index + 1).padStart(2, '0')}.png
                      </p>
                      <p className="text-xs text-gray-500">8×10 Portrait template</p>
                    </div>
                  </div>
                ))}
                <p className="text-xs text-gray-400 text-center pt-2">
                  (Dummy data - will load from Image Layouts tab)
                </p>
              </div>
            </CardBody>
          </Card>
        </div>
      )}

      {/* Composed Pages Grid */}
      <div>
        <h3 className="text-xl font-semibold text-gray-900 mb-4">Composed Pages</h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {DUMMY_PAGES.map((page) => (
            <div key={page.id} onClick={() => setSelectedPage(page.id)}>
              <Card
                className={`cursor-pointer border-2 transition-all ${
                  selectedPage === page.id
                    ? 'border-primary-500 shadow-md'
                    : 'border-gray-200 hover:border-primary-300'
                }`}
              >
                <CardBody className="space-y-3">
                  {/* Page Preview */}
                  <div className="aspect-[4/5] bg-gray-100 rounded overflow-hidden flex items-center justify-center border border-gray-200">
                    {page.template === 'Single Image' && (
                      <div className="w-full h-full flex items-center justify-center">
                        <div className="w-3/4 h-3/4 bg-gray-300 rounded flex items-center justify-center text-xs text-gray-500">
                          {page.images[0]}
                        </div>
                      </div>
                    )}
                    {page.template === '2-Up Spread' && (
                      <div className="w-full h-full grid grid-cols-2 gap-1 p-2">
                        {page.images.map((img, idx) => (
                          <div key={idx} className="bg-gray-300 rounded flex items-center justify-center text-xs text-gray-500">
                            {img}
                          </div>
                        ))}
                      </div>
                    )}
                    {page.template === '4-Up Grid' && (
                      <div className="w-full h-full grid grid-cols-2 gap-1 p-2">
                        {page.images.map((img, idx) => (
                          <div key={idx} className="bg-gray-300 rounded flex items-center justify-center text-xs text-gray-500">
                            {img}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>

                  <div>
                    <p className="text-sm font-semibold text-gray-900">{page.name}</p>
                    <p className="text-xs text-gray-500">
                      {page.template} • {page.images.length} image{page.images.length > 1 ? 's' : ''}
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      {new Date(page.created_at).toLocaleDateString()}
                    </p>
                  </div>

                  <div className="flex space-x-2">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={(e) => {
                        e.stopPropagation();
                        alert('Edit page - Coming in Phase 3');
                      }}
                    >
                      Edit
                    </Button>
                    <Button
                      size="sm"
                      variant="danger"
                      onClick={(e) => {
                        e.stopPropagation();
                        if (window.confirm(`Delete page "${page.name}"?`)) {
                          alert('Delete would happen here');
                        }
                      }}
                    >
                      Delete
                    </Button>
                  </div>
                </CardBody>
              </Card>
            </div>
          ))}

          {/* Empty State */}
          {!isCreating && DUMMY_PAGES.length === 0 && (
            <div className="col-span-full text-center py-16 border-2 border-dashed border-gray-300 rounded-lg bg-gray-50">
              <div className="text-6xl mb-4">📄</div>
              <h4 className="text-lg font-semibold text-gray-900 mb-2">No pages yet</h4>
              <p className="text-sm text-gray-600 mb-4">
                Create your first page by selecting a template above
              </p>
              <Button onClick={() => setIsCreating(true)}>+ Create Page</Button>
            </div>
          )}
        </div>
      </div>

      {/* Selected Page Detail */}
      {selectedPage && !isCreating && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900">Page Details</h3>
              <Button variant="secondary" size="sm" onClick={() => setSelectedPage(null)}>
                ✕ Close
              </Button>
            </div>
          </CardHeader>
          <CardBody>
            <div className="grid lg:grid-cols-2 gap-6">
              {/* Preview */}
              <div className="bg-gray-50 p-6 rounded-lg flex items-center justify-center">
                <div className="w-full max-w-sm aspect-[4/5] bg-white border-2 border-gray-300 shadow-lg rounded">
                  {(() => {
                    const page = DUMMY_PAGES.find((p) => p.id === selectedPage);
                    if (!page) return null;

                    if (page.template === 'Single Image') {
                      return (
                        <div className="w-full h-full flex items-center justify-center p-4">
                          <div className="w-full h-full bg-gray-200 rounded flex items-center justify-center text-sm text-gray-600">
                            {page.images[0]}
                          </div>
                        </div>
                      );
                    }

                    if (page.template === '2-Up Spread') {
                      return (
                        <div className="w-full h-full grid grid-cols-2 gap-2 p-3">
                          {page.images.map((img, idx) => (
                            <div key={idx} className="bg-gray-200 rounded flex items-center justify-center text-xs text-gray-600">
                              {img}
                            </div>
                          ))}
                        </div>
                      );
                    }

                    if (page.template === '4-Up Grid') {
                      return (
                        <div className="w-full h-full grid grid-cols-2 gap-2 p-3">
                          {page.images.map((img, idx) => (
                            <div key={idx} className="bg-gray-200 rounded flex items-center justify-center text-xs text-gray-600">
                              {img}
                            </div>
                          ))}
                        </div>
                      );
                    }

                    return null;
                  })()}
                </div>
              </div>

              {/* Info */}
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium text-gray-700">Page Name</label>
                  <p className="text-base text-gray-900">
                    {DUMMY_PAGES.find((p) => p.id === selectedPage)?.name}
                  </p>
                </div>

                <div>
                  <label className="text-sm font-medium text-gray-700">Template</label>
                  <p className="text-base text-gray-900">
                    {DUMMY_PAGES.find((p) => p.id === selectedPage)?.template}
                  </p>
                </div>

                <div>
                  <label className="text-sm font-medium text-gray-700">Images</label>
                  <div className="mt-2 space-y-1">
                    {DUMMY_PAGES.find((p) => p.id === selectedPage)?.images.map((img, idx) => (
                      <div key={idx} className="text-sm text-gray-600">
                        {idx + 1}. {img}
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-gray-700">Created</label>
                  <p className="text-sm text-gray-600">
                    {new Date(
                      DUMMY_PAGES.find((p) => p.id === selectedPage)?.created_at ?? ''
                    ).toLocaleString()}
                  </p>
                </div>

                <div className="pt-4 space-y-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    className="w-full"
                    onClick={() => alert('Export page - Coming in Phase 3')}
                  >
                    Export Page
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    className="w-full"
                    onClick={() => alert('Add to Zine - Coming in Phase 3')}
                  >
                    Add to Zine
                  </Button>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>
      )}

      {/* Page Templates Library */}
      {!isCreating && (
        <div>
          <h3 className="text-xl font-semibold text-gray-900 mb-4">Page Templates</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {DUMMY_PAGE_TEMPLATES.map((template) => (
              <Card key={template.id}>
                <CardBody>
                  <div className="text-center">
                    <div className="text-5xl mb-3">{template.icon}</div>
                    <h4 className="font-semibold text-gray-900 mb-1">{template.name}</h4>
                    <p className="text-xs text-gray-500 mb-3">{template.description}</p>
                    <Button
                      size="sm"
                      variant="secondary"
                      className="w-full"
                      onClick={() => handleSelectTemplate(template.id)}
                    >
                      Use Template
                    </Button>
                  </div>
                </CardBody>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Statistics */}
      <Card>
        <CardHeader>
          <h3 className="text-lg font-semibold text-gray-900">Coming in Phase 3</h3>
        </CardHeader>
        <CardBody>
          <div className="grid md:grid-cols-2 gap-6">
            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-3">Features to Implement:</h4>
              <ul className="space-y-2 text-sm text-gray-600">
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Backend page template repository and API</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Laid-out page creation with image slot assignment</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Drag-and-drop from Image Layouts tab</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Page rendering service (server-side composition)</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Custom grid builder for template creation</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-gray-400">○</span>
                  <span>Page preview export (PNG/PDF)</span>
                </li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-3">Current Capabilities:</h4>
              <ul className="space-y-2 text-sm text-gray-600">
                <li className="flex items-start space-x-2">
                  <span className="text-green-500">✓</span>
                  <span>Database schema for page templates</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-green-500">✓</span>
                  <span>Repository layer for CRUD operations</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-green-500">✓</span>
                  <span>Service layer for page composition</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-green-500">✓</span>
                  <span>CLI workflow commands for testing</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-yellow-500">⧗</span>
                  <span>REST API endpoints (pending)</span>
                </li>
                <li className="flex items-start space-x-2">
                  <span className="text-yellow-500">⧗</span>
                  <span>Page renderer (stubbed)</span>
                </li>
              </ul>
            </div>
          </div>

          <div className="mt-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
            <p className="text-sm text-gray-700">
              <strong>Note:</strong> The UI mockup above demonstrates the planned workflow. Backend persistence (Phase 3) is partially complete - see{' '}
              <code className="text-xs bg-gray-200 px-1 py-0.5 rounded">
                pkg/services/pages.go
              </code>{' '}
              and{' '}
              <code className="text-xs bg-gray-200 px-1 py-0.5 rounded">
                cmd/zine-layout/cmds/workflow/page-templates.go
              </code>{' '}
              for current implementation status.
            </p>
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

