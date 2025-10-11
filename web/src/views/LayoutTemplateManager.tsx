import React, { useMemo, useState } from 'react';
import {
  useCreateImageLayoutTemplateMutation,
  useDeleteImageLayoutTemplateMutation,
  useGetImageLayoutTemplatesQuery,
  useUpdateImageLayoutTemplateMutation,
  type ImageLayoutTemplate,
} from '../api';
import { Button, Card, CardBody, CardHeader, Input } from '../components/ui';

interface LayoutTemplateManagerProps {
  projectId: string;
}

const prettyJSON = (value: unknown) => {
  if (!value) return '';
  try {
    return JSON.stringify(value, null, 2);
  } catch (err) {
    return '';
  }
};

const DEFAULT_TEMPLATE_JSON = `{
  "paper_width_in": 8.5,
  "paper_height_in": 11,
  "dpi": 300,
  "orientation": "portrait",
  "margin_top_in": 0.5,
  "margin_right_in": 0.5,
  "margin_bottom_in": 0.5,
  "margin_left_in": 0.5,
  "crop_to_fill": true,
  "user_scale": 1,
  "position_x": 0,
  "position_y": 0,
  "units": "normalized"
}`;

export const LayoutTemplateManager: React.FC<LayoutTemplateManagerProps> = ({ projectId }) => {
  const templatesQuery = useGetImageLayoutTemplatesQuery({ projectId }, { skip: !projectId });
  const [createTemplate, createState] = useCreateImageLayoutTemplateMutation();
  const [updateTemplate, updateState] = useUpdateImageLayoutTemplateMutation();
  const [deleteTemplate, deleteState] = useDeleteImageLayoutTemplateMutation();

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [createName, setCreateName] = useState('');
  const [createDescription, setCreateDescription] = useState('');
  const [createSettings, setCreateSettings] = useState(DEFAULT_TEMPLATE_JSON);

  const [editingTemplate, setEditingTemplate] = useState<ImageLayoutTemplate | null>(null);
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editSettings, setEditSettings] = useState('');

  const sortedTemplates = useMemo(() => {
    if (!templatesQuery.data) return [] as ImageLayoutTemplate[];
    return [...templatesQuery.data].sort((a, b) => a.name.localeCompare(b.name));
  }, [templatesQuery.data]);

  const resetCreateForm = () => {
    setCreateName('');
    setCreateDescription('');
    setCreateSettings(DEFAULT_TEMPLATE_JSON);
    setIsCreateOpen(false);
  };

  const handleSelectTemplate = (tpl: ImageLayoutTemplate) => {
    setEditingTemplate(tpl);
    setEditName(tpl.name);
    setEditDescription(tpl.description ?? '');
    setEditSettings(prettyJSON(tpl.settings));
  };

  const parseSettings = (value: string) => {
    try {
      const parsed = JSON.parse(value || '{}');
      if (typeof parsed !== 'object' || Array.isArray(parsed) || parsed === null) {
        throw new Error('Settings JSON must be an object');
      }
      return parsed as Record<string, unknown>;
    } catch (err) {
      throw new Error(`Invalid settings JSON: ${(err as Error).message}`);
    }
  };

  const handleCreateTemplate = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!projectId) return;
    try {
      const settings = parseSettings(createSettings);
      const template = await createTemplate({
        projectId,
        name: createName.trim() || 'Untitled Template',
        description: createDescription.trim() || undefined,
        settings,
      }).unwrap();
      setEditingTemplate(template);
      setEditName(template.name);
      setEditDescription(template.description ?? '');
      setEditSettings(prettyJSON(template.settings));
      resetCreateForm();
    } catch (err) {
      alert((err as Error).message);
    }
  };

  const handleUpdateTemplate = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!editingTemplate) return;
    try {
      const payload: {
        templateId: string;
        name?: string;
        description?: string;
        settings?: Record<string, unknown>;
      } = {
        templateId: editingTemplate.id,
      };
      const trimmedName = editName.trim();
      if (trimmedName) payload.name = trimmedName;
      payload.description = editDescription.trim() || undefined;
      if (editSettings.trim()) {
        payload.settings = parseSettings(editSettings);
      }
      const updated = await updateTemplate(payload).unwrap();
      setEditingTemplate(updated);
      setEditName(updated.name);
      setEditDescription(updated.description ?? '');
      setEditSettings(prettyJSON(updated.settings));
    } catch (err) {
      alert((err as Error).message);
    }
  };

  const handleDeleteTemplate = async (tpl: ImageLayoutTemplate) => {
    if (!window.confirm(`Delete template "${tpl.name}"? This action cannot be undone.`)) {
      return;
    }
    await deleteTemplate({ templateId: tpl.id, scopeKey: tpl.project_id ?? 'global' }).unwrap();
    if (editingTemplate?.id === tpl.id) {
      setEditingTemplate(null);
      setEditName('');
      setEditDescription('');
      setEditSettings('');
    }
  };

  return (
    <Card id="layout-templates" className="mt-8">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">Image Layout Templates</h2>
            <p className="text-sm text-gray-500">
              Create reusable layout presets and apply them when generating laid-out images.
            </p>
          </div>
          <Button onClick={() => setIsCreateOpen((prev) => !prev)}>
            {isCreateOpen ? 'Cancel' : 'New Template'}
          </Button>
        </div>
      </CardHeader>
      <CardBody className="space-y-8">
        {isCreateOpen && (
          <form onSubmit={handleCreateTemplate} className="space-y-4 p-4 border border-gray-200 rounded-lg">
            <h3 className="text-lg font-medium text-gray-800">Create Template</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Name</label>
                <Input value={createName} onChange={(e) => setCreateName(e.target.value)} placeholder="Full bleed portrait" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Description</label>
                <Input value={createDescription} onChange={(e) => setCreateDescription(e.target.value)} placeholder="Optional description" />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Settings (JSON)</label>
              <textarea
                value={createSettings}
                onChange={(e) => setCreateSettings(e.target.value)}
                rows={10}
                className="font-mono text-sm w-full border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
              />
              <p className="mt-1 text-xs text-gray-500">
                Settings map directly to the backend `imagelayout.ViewportSettings` structure. Ensure valid JSON before saving.
              </p>
            </div>
            <div className="flex justify-end space-x-3">
              <Button type="button" variant="secondary" onClick={resetCreateForm}>
                Cancel
              </Button>
              <Button type="submit" disabled={createState.isLoading}>
                {createState.isLoading ? 'Creating…' : 'Create Template'}
              </Button>
            </div>
          </form>
        )}

        <div className="grid gap-4 md:grid-cols-2">
          {sortedTemplates.map((tpl) => (
            <Card key={tpl.id} className={`border ${editingTemplate?.id === tpl.id ? 'border-primary-500' : 'border-gray-200'}`}>
              <CardBody className="space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm uppercase tracking-wide text-gray-400">{tpl.scope === 'project' ? 'Project template' : 'Global template'}</p>
                    <h3 className="text-lg font-semibold text-gray-900">{tpl.name}</h3>
                  </div>
                  <div className="flex space-x-2">
                    <Button size="sm" variant="secondary" onClick={() => handleSelectTemplate(tpl)}>
                      Edit
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => handleDeleteTemplate(tpl)} disabled={deleteState.isLoading}>
                      Delete
                    </Button>
                  </div>
                </div>
                {tpl.description && <p className="text-sm text-gray-600">{tpl.description}</p>}
                <p className="text-xs text-gray-500">
                  Updated {new Date(tpl.updated_at).toLocaleString()} • Template ID {tpl.id}
                </p>
              </CardBody>
            </Card>
          ))}
          {!templatesQuery.isLoading && sortedTemplates.length === 0 && (
            <div className="col-span-full text-sm text-gray-500">No templates yet. Create one to get started.</div>
          )}
        </div>

        {editingTemplate && (
          <form onSubmit={handleUpdateTemplate} className="space-y-4 p-4 border border-primary-200 rounded-lg">
            <h3 className="text-lg font-semibold text-gray-900">Edit Template</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Name</label>
                <Input value={editName} onChange={(e) => setEditName(e.target.value)} />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Description</label>
                <Input value={editDescription} onChange={(e) => setEditDescription(e.target.value)} />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Settings (JSON)</label>
              <textarea
                value={editSettings}
                onChange={(e) => setEditSettings(e.target.value)}
                rows={10}
                className="font-mono text-sm w-full border border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
              />
            </div>
            <div className="flex justify-end space-x-3">
              <Button type="button" variant="secondary" onClick={() => setEditingTemplate(null)}>
                Close
              </Button>
              <Button type="submit" disabled={updateState.isLoading}>
                {updateState.isLoading ? 'Saving…' : 'Save Changes'}
              </Button>
            </div>
          </form>
        )}
      </CardBody>
    </Card>
  );
};
