import React from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import { useGetProjectsQuery } from '../api';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui';
import { AssetsTab } from './tabs/AssetsTab';
import { SequencesTab } from './tabs/SequencesTab';
import { ImageLayoutsTab } from './tabs/ImageLayoutsTab';

const formatDateTime = (iso?: string) => {
  if (!iso) return '—';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '—';
  return date.toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
};

export const ProjectDetail: React.FC = () => {
  const { id = '' } = useParams();
  const { data: projects } = useGetProjectsQuery();
  const project = projects?.find((p) => p.id === id);

  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get('tab') || 'assets';

  const handleTabChange = (newTab: string) => {
    setSearchParams({ tab: newTab });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <nav className="flex items-center space-x-2 text-sm text-gray-600 mb-2">
          <Link to="/projects" className="hover:text-gray-900">
            Projects
          </Link>
          <span>/</span>
          <span className="text-gray-900 font-medium">{project?.name ?? id}</span>
        </nav>
        <h1 className="text-3xl font-bold text-gray-900">
          {project?.name ?? 'Project'}
        </h1>
        <p className="text-sm text-gray-500">
          Created {formatDateTime(project?.created_at)} · Updated{' '}
          {formatDateTime(project?.updated_at)}
        </p>
        {project?.description && (
          <p className="mt-4 max-w-2xl text-gray-700">{project.description}</p>
        )}
      </div>

      {/* Tabbed Interface */}
      <Tabs value={activeTab} onValueChange={handleTabChange} className="space-y-6">
        <TabsList className="w-full justify-start">
          <TabsTrigger value="assets" className="flex items-center space-x-2">
            <span>📁</span>
            <span>Assets</span>
          </TabsTrigger>
          <TabsTrigger value="sequences" className="flex items-center space-x-2">
            <span>🔢</span>
            <span>Sequences</span>
          </TabsTrigger>
          <TabsTrigger value="image-layouts" className="flex items-center space-x-2">
            <span>🖼️</span>
            <span>Image Layouts</span>
          </TabsTrigger>
          <TabsTrigger value="page-layouts" className="flex items-center space-x-2">
            <span>📄</span>
            <span>Page Layouts</span>
          </TabsTrigger>
          <TabsTrigger value="zine" className="flex items-center space-x-2">
            <span>📚</span>
            <span>Zine</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="assets">
          {id && <AssetsTab projectId={id} />}
        </TabsContent>

        <TabsContent value="sequences">
          {id && <SequencesTab projectId={id} />}
        </TabsContent>

        <TabsContent value="image-layouts">
          {id && <ImageLayoutsTab projectId={id} />}
        </TabsContent>

        <TabsContent value="page-layouts">
          {id && (
            <div className="p-12 text-center border-2 border-dashed border-gray-300 rounded-lg bg-gray-50">
              <div className="text-4xl mb-4">📄</div>
              <h3 className="text-xl font-semibold text-gray-900 mb-2">Page Layouts</h3>
              <p className="text-gray-600 max-w-md mx-auto">
                Multi-image page composition coming in Phase 3. This will allow you to create pages with multiple laid-out images using grid templates.
              </p>
                </div>
              )}
        </TabsContent>

        <TabsContent value="zine">
          {id && (
            <div className="p-12 text-center border-2 border-dashed border-gray-300 rounded-lg bg-gray-50">
              <div className="text-4xl mb-4">📚</div>
              <h3 className="text-xl font-semibold text-gray-900 mb-2">Zine Assembly & Export</h3>
              <p className="text-gray-600 max-w-md mx-auto">
                Zine assembly, imposition templates, and print export coming in Phase 3/4. This will allow you to create complete books with proper page ordering for printing and folding.
                        </p>
                      </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
};
