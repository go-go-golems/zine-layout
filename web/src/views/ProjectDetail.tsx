import React from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import { useGetProjectsQuery } from '../api';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui';
import { AssetsTab } from './tabs/AssetsTab';
import { SequencesTab } from './tabs/SequencesTab';
import { LayoutTemplateManager } from './LayoutTemplateManager';
import { LaidOutImageViewer } from './LaidOutImageViewer';
import { LayoutSequenceEditor } from './LayoutSequenceEditor';

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
          <TabsTrigger value="templates" className="flex items-center space-x-2">
            <span>📐</span>
            <span>Templates</span>
          </TabsTrigger>
          <TabsTrigger value="layouts" className="flex items-center space-x-2">
            <span>🖼️</span>
            <span>Layouts</span>
          </TabsTrigger>
          <TabsTrigger value="output" className="flex items-center space-x-2">
            <span>📚</span>
            <span>Output</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="assets">
          {id && <AssetsTab projectId={id} />}
        </TabsContent>

        <TabsContent value="sequences">
          {id && <SequencesTab projectId={id} />}
        </TabsContent>

        <TabsContent value="templates">
          {id && <LayoutTemplateManager projectId={id} />}
        </TabsContent>

        <TabsContent value="layouts">
          {id && <LaidOutImageViewer projectId={id} />}
        </TabsContent>

        <TabsContent value="output">
          {id && <LayoutSequenceEditor projectId={id} />}
        </TabsContent>
      </Tabs>
    </div>
  );
};
