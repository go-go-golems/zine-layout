import type React from 'react';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  useCreateProjectMutation,
  useDeleteProjectMutation,
  useGetPresetsQuery,
  useGetProjectsQuery,
} from '../api';
import { Button, Input, Card, CardBody } from '../components/ui';

export const Projects: React.FC = () => {
  const { data, isLoading, refetch } = useGetProjectsQuery();
  const [createProject, { isLoading: isCreating }] = useCreateProjectMutation();
  const [deleteProject] = useDeleteProjectMutation();
  const [name, setName] = useState('');
  const [presetId, setPresetId] = useState('');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const { data: presetsData } = useGetPresetsQuery();

  const onCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    await createProject({
      name: name.trim() || undefined,
      presetId: presetId || undefined,
    }).unwrap();
    setName('');
    setPresetId('');
    setShowCreateForm(false);
    refetch();
  };

  const onDelete = async (projectId: string, projectName: string) => {
    if (window.confirm(`Delete project "${projectName}"? This action cannot be undone.`)) {
      await deleteProject({ id: projectId }).unwrap();
      refetch();
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-6xl mx-auto">
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <div className="w-8 h-8 border-4 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <p className="text-gray-600">Loading projects...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Projects</h1>
          <p className="text-gray-600 mt-1">Manage your zine layout projects</p>
        </div>
        <Button onClick={() => setShowCreateForm(!showCreateForm)}>
          {showCreateForm ? 'Cancel' : 'New Project'}
        </Button>
      </div>

      {/* Create Project Form */}
      {showCreateForm && (
        <Card className="mb-8">
          <CardBody>
            <form onSubmit={onCreate} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="Project Name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="My Zine Project"
                  required
                />
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Preset (Optional)
                  </label>
                  <select 
                    value={presetId} 
                    onChange={(e) => setPresetId(e.target.value)}
                    className="input"
                  >
                    <option value="">No preset</option>
                    {presetsData?.presets?.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="flex justify-end space-x-3">
                <Button 
                  type="button" 
                  variant="secondary" 
                  onClick={() => setShowCreateForm(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" isLoading={isCreating}>
                  Create Project
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>
      )}

      {/* Projects Grid */}
      {data?.projects?.length === 0 ? (
        <Card>
          <CardBody className="text-center py-12">
            <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No projects yet</h3>
            <p className="text-gray-600 mb-4">Get started by creating your first zine project</p>
            <Button onClick={() => setShowCreateForm(true)}>
              Create Your First Project
            </Button>
          </CardBody>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {data?.projects?.map((project) => (
            <Card key={project.id} className="hover:shadow-medium transition-shadow duration-200">
              <CardBody>
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-1">
                      <Link 
                        to={`/projects/${project.id}`}
                        className="hover:text-primary-600 transition-colors duration-200"
                      >
                        {project.name}
                      </Link>
                    </h3>
                    <p className="text-sm text-gray-500">
                      Updated {new Date(project.updatedAt).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex-shrink-0">
                    <button
                      onClick={() => onDelete(project.id, project.name)}
                      className="text-gray-400 hover:text-red-600 transition-colors duration-200"
                      title="Delete project"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </div>
                
                {project.presetId && (
                  <div className="mb-4">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                      {presetsData?.presets?.find(p => p.id === project.presetId)?.name || project.presetId}
                    </span>
                  </div>
                )}

                <div className="flex justify-between items-center">
                  <Link to={`/projects/${project.id}`}>
                    <Button size="sm">Open Project</Button>
                  </Link>
                  <div className="flex space-x-2">
                    <Link 
                      to={`/projects/${project.id}/yaml`}
                      className="text-sm text-gray-500 hover:text-gray-700 transition-colors duration-200"
                    >
                      Edit YAML
                    </Link>
                  </div>
                </div>
              </CardBody>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};
