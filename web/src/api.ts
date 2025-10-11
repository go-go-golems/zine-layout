import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export interface Project {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface Asset {
  id: string;
  project_id: string;
  filename: string;
  rel_path: string;
  content_type: string;
  bytes: number;
  width: number;
  height: number;
  uploaded_at: string;
  metadata?: Record<string, unknown>;
  url?: string;
}

export interface ImageSequence {
  id: string;
  project_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface ImageSequenceItem {
  sequence_id: string;
  position: number;
  asset_id?: string;
  is_gap: boolean;
}

export interface ImageLayoutTemplate {
  id: string;
  project_id?: string | null;
  scope: 'global' | 'project';
  name: string;
  description?: string;
  settings: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface LayoutComputation {
  settings: Record<string, unknown>;
  result: Record<string, unknown>;
  placement_trace?: Record<string, unknown>;
}

export interface LaidOutImage {
  id: string;
  project_id: string;
  asset_id: string;
  template_id: string;
  overrides?: Record<string, unknown>;
  result?: LayoutComputation;
  created_at: string;
  updated_at: string;
}

export interface LayoutSequence {
  id: string;
  project_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface LayoutSequenceItem {
  sequence_id: string;
  position: number;
  laid_out_image_id: string;
}

const baseQuery = fetchBaseQuery({ baseUrl: '/api' });

const asFileArray = (files: FileList | File[]) => Array.from(files as FileList);

export const api = createApi({
  reducerPath: 'api',
  baseQuery,
  tagTypes: [
    'Project',
    'Asset',
    'Sequence',
    'SequenceItems',
    'LayoutTemplate',
    'LaidOutImage',
    'LayoutSequence',
    'LayoutSequenceItems',
  ],
  endpoints: (builder) => ({
    getProjects: builder.query<Project[], void>({
      query: () => '/projects',
      transformResponse: (response: { projects: Project[] }) => response.projects ?? [],
      providesTags: (result) =>
        result
          ? [
              ...result.map((proj) => ({ type: 'Project' as const, id: proj.id })),
              { type: 'Project' as const, id: 'LIST' },
            ]
          : [{ type: 'Project', id: 'LIST' }],
    }),
    createProject: builder.mutation<Project, { name?: string; description?: string }>({
      query: (body) => ({ url: '/projects', method: 'POST', body }),
      transformResponse: (response: { project: Project }) => response.project,
      invalidatesTags: [{ type: 'Project', id: 'LIST' }],
    }),
    deleteProject: builder.mutation<void, { id: string }>({
      query: ({ id }) => ({ url: `/projects/${encodeURIComponent(id)}`, method: 'DELETE' }),
      invalidatesTags: (_result, _error, { id }) => [
        { type: 'Project', id },
        { type: 'Project', id: 'LIST' },
      ],
    }),

    getAssets: builder.query<Asset[], { projectId: string }>({
      query: ({ projectId }) => `/projects/${encodeURIComponent(projectId)}/assets`,
      transformResponse: (response: { assets: Asset[] }) => response.assets ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'Asset' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((asset) => ({ type: 'Asset' as const, id: asset.id })),
          ...base,
        ];
      },
    }),
    uploadAssets: builder.mutation<
      Asset[],
      { projectId: string; files: FileList | File[] }
    >({
      query: ({ projectId, files }) => {
        const fd = new FormData();
        for (const file of asFileArray(files)) {
          fd.append('images[]', file);
        }
        return {
          url: `/projects/${encodeURIComponent(projectId)}/images`,
          method: 'POST',
          body: fd,
        };
      },
      transformResponse: (response: { assets: Asset[] }) => response.assets ?? [],
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'Asset', id: `LIST-${projectId}` },
      ],
    }),
    deleteAsset: builder.mutation<void, { assetId: string; projectId: string }>({
      query: ({ assetId }) => ({ url: `/assets/${encodeURIComponent(assetId)}`, method: 'DELETE' }),
      invalidatesTags: (_result, _error, { assetId, projectId }) => [
        { type: 'Asset', id: assetId },
        { type: 'Asset', id: `LIST-${projectId}` },
      ],
    }),

    getImageSequences: builder.query<ImageSequence[], { projectId: string }>({
      query: ({ projectId }) => `/projects/${encodeURIComponent(projectId)}/image-sequences`,
      transformResponse: (response: { sequences: ImageSequence[] }) => response.sequences ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'Sequence' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((seq) => ({ type: 'Sequence' as const, id: seq.id })),
          ...base,
        ];
      },
    }),
    createImageSequence: builder.mutation<
      ImageSequence,
      { projectId: string; name: string; description?: string }
    >({
      query: ({ projectId, ...body }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/image-sequences`,
        method: 'POST',
        body,
      }),
      transformResponse: (response: { sequence: ImageSequence }) => response.sequence,
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'Sequence', id: `LIST-${projectId}` },
      ],
    }),
    updateImageSequence: builder.mutation<
      ImageSequence,
      { sequenceId: string; name?: string; description?: string }
    >({
      query: ({ sequenceId, ...body }) => ({
        url: `/image-sequences/${encodeURIComponent(sequenceId)}`,
        method: 'PATCH',
        body,
      }),
      transformResponse: (response: { sequence: ImageSequence }) => response.sequence,
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'Sequence', id: sequenceId },
        { type: 'SequenceItems', id: sequenceId },
      ],
    }),
    deleteImageSequence: builder.mutation<void, { sequenceId: string; projectId: string }>({
      query: ({ sequenceId }) => ({
        url: `/image-sequences/${encodeURIComponent(sequenceId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { sequenceId, projectId }) => [
        { type: 'Sequence', id: sequenceId },
        { type: 'Sequence', id: `LIST-${projectId}` },
        { type: 'SequenceItems', id: sequenceId },
      ],
    }),

    getImageSequenceDetail: builder.query<
      { sequence: ImageSequence; items: ImageSequenceItem[] },
      { sequenceId: string }
    >({
      query: ({ sequenceId }) => `/image-sequences/${encodeURIComponent(sequenceId)}`,
      transformResponse: (response: {
        sequence: ImageSequence;
        items: ImageSequenceItem[];
      }) => ({
        sequence: response.sequence,
        items: response.items ?? [],
      }),
      providesTags: (_result, _error, { sequenceId }) => [
        { type: 'Sequence', id: sequenceId },
        { type: 'SequenceItems', id: sequenceId },
      ],
    }),
    addImageSequenceItem: builder.mutation<
      ImageSequenceItem[],
      { sequenceId: string; assetId?: string }
    >({
      query: ({ sequenceId, assetId }) => ({
        url: `/image-sequences/${encodeURIComponent(sequenceId)}/items`,
        method: 'POST',
        body: { asset_id: assetId, is_gap: !assetId },
      }),
      transformResponse: (response: { items: ImageSequenceItem[] }) => response.items ?? [],
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'SequenceItems', id: sequenceId },
      ],
    }),
    reorderImageSequenceItems: builder.mutation<
      ImageSequenceItem[],
      { sequenceId: string; items: { assetId?: string; isGap?: boolean }[] }
    >({
      query: ({ sequenceId, items }) => ({
        url: `/image-sequences/${encodeURIComponent(sequenceId)}/items`,
        method: 'PUT',
        body: {
          items: items.map((item) => ({
            asset_id: item.assetId,
            is_gap: item.isGap ?? !item.assetId,
          })),
        },
      }),
      transformResponse: (response: { items: ImageSequenceItem[] }) => response.items ?? [],
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'SequenceItems', id: sequenceId },
      ],
    }),
    deleteImageSequenceItem: builder.mutation<
      void,
      { sequenceId: string; position: number }
    >({
      query: ({ sequenceId, position }) => ({
        url: `/image-sequences/${encodeURIComponent(sequenceId)}/items/${position}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'SequenceItems', id: sequenceId },
      ],
    }),

    getImageLayoutTemplates: builder.query<
      ImageLayoutTemplate[],
      { projectId: string }
    >({
      query: ({ projectId }) =>
        `/projects/${encodeURIComponent(projectId)}/image-layout-templates`,
      transformResponse: (response: { templates: ImageLayoutTemplate[] }) =>
        response.templates ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'LayoutTemplate' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((tpl) => ({ type: 'LayoutTemplate' as const, id: tpl.id })),
          ...base,
        ];
      },
    }),
    createImageLayoutTemplate: builder.mutation<
      ImageLayoutTemplate,
      { projectId: string; name: string; description?: string; settings: Record<string, unknown> }
    >({
      query: ({ projectId, ...body }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/image-layout-templates`,
        method: 'POST',
        body,
      }),
      transformResponse: (response: { template: ImageLayoutTemplate }) => response.template,
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'LayoutTemplate', id: `LIST-${projectId}` },
      ],
    }),
    updateImageLayoutTemplate: builder.mutation<
      ImageLayoutTemplate,
      {
        templateId: string;
        name?: string;
        description?: string;
        settings?: Record<string, unknown>;
      }
    >({
      query: ({ templateId, ...body }) => ({
        url: `/image-layout-templates/${encodeURIComponent(templateId)}`,
        method: 'PATCH',
        body,
      }),
      transformResponse: (response: { template: ImageLayoutTemplate }) => response.template,
      invalidatesTags: (result) => {
        if (!result) return [];
        const tags: { type: 'LayoutTemplate'; id: string }[] = [
          { type: 'LayoutTemplate', id: result.id },
        ];
        const listKey = result.project_id ?? 'global';
        tags.push({ type: 'LayoutTemplate', id: `LIST-${listKey}` });
        return tags;
      },
    }),
    deleteImageLayoutTemplate: builder.mutation<
      void,
      { templateId: string; scopeKey: string }
    >({
      query: ({ templateId }) => ({
        url: `/image-layout-templates/${encodeURIComponent(templateId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { templateId, scopeKey }) => [
        { type: 'LayoutTemplate', id: templateId },
        { type: 'LayoutTemplate', id: `LIST-${scopeKey}` },
      ],
    }),

    getLaidOutImages: builder.query<LaidOutImage[], { projectId: string }>({
      query: ({ projectId }) =>
        `/projects/${encodeURIComponent(projectId)}/laid-out-images`,
      transformResponse: (response: { laid_out_images: LaidOutImage[] }) =>
        response.laid_out_images ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'LaidOutImage' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((item) => ({ type: 'LaidOutImage' as const, id: item.id })),
          ...base,
        ];
      },
    }),
    getLaidOutImage: builder.query<LaidOutImage, { id: string }>({
      query: ({ id }) => `/laid-out-images/${encodeURIComponent(id)}`,
      transformResponse: (response: { laid_out_image: LaidOutImage }) =>
        response.laid_out_image,
      providesTags: (_result, _error, { id }) => [{ type: 'LaidOutImage', id }],
    }),
    createLaidOutImage: builder.mutation<
      LaidOutImage,
      {
        projectId: string;
        assetId: string;
        templateId: string;
        overrides?: Record<string, unknown>;
      }
    >({
      query: ({ projectId, assetId, templateId, overrides }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/laid-out-images`,
        method: 'POST',
        body: {
          asset_id: assetId,
          template_id: templateId,
          overrides,
        },
      }),
      transformResponse: (response: { laid_out_image: LaidOutImage }) => response.laid_out_image,
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'LaidOutImage', id: `LIST-${projectId}` },
      ],
    }),
    updateLaidOutImage: builder.mutation<
      LaidOutImage,
      {
        id: string;
        templateId?: string;
        overrides?: Record<string, unknown>;
      }
    >({
      query: ({ id, templateId, overrides }) => ({
        url: `/laid-out-images/${encodeURIComponent(id)}`,
        method: 'PATCH',
        body: {
          template_id: templateId,
          overrides,
        },
      }),
      transformResponse: (response: { laid_out_image: LaidOutImage }) => response.laid_out_image,
      invalidatesTags: (result) => {
        if (!result) return [];
        return [
          { type: 'LaidOutImage', id: result.id },
          { type: 'LaidOutImage', id: `LIST-${result.project_id}` },
        ];
      },
    }),
    deleteLaidOutImage: builder.mutation<
      void,
      { id: string; projectId: string }
    >({
      query: ({ id }) => ({
        url: `/laid-out-images/${encodeURIComponent(id)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { id, projectId }) => [
        { type: 'LaidOutImage', id },
        { type: 'LaidOutImage', id: `LIST-${projectId}` },
      ],
    }),
    previewLaidOutImage: builder.query<LayoutComputation, { id: string }>({
      query: ({ id }) => `/laid-out-images/${encodeURIComponent(id)}/preview`,
      transformResponse: (response: { result: LayoutComputation }) => response.result,
    }),

    getLayoutSequences: builder.query<LayoutSequence[], { projectId: string }>({
      query: ({ projectId }) => `/projects/${encodeURIComponent(projectId)}/layout-sequences`,
      transformResponse: (response: { layout_sequences: LayoutSequence[] }) =>
        response.layout_sequences ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'LayoutSequence' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((seq) => ({ type: 'LayoutSequence' as const, id: seq.id })),
          ...base,
        ];
      },
    }),
    createLayoutSequence: builder.mutation<
      LayoutSequence,
      { projectId: string; name: string; description?: string }
    >({
      query: ({ projectId, ...body }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/layout-sequences`,
        method: 'POST',
        body,
      }),
      transformResponse: (response: { layout_sequence: LayoutSequence }) =>
        response.layout_sequence,
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'LayoutSequence', id: `LIST-${projectId}` },
      ],
    }),
    updateLayoutSequence: builder.mutation<
      LayoutSequence,
      { sequenceId: string; name?: string; description?: string }
    >({
      query: ({ sequenceId, ...body }) => ({
        url: `/layout-sequences/${encodeURIComponent(sequenceId)}`,
        method: 'PATCH',
        body,
      }),
      transformResponse: (response: { layout_sequence: LayoutSequence }) =>
        response.layout_sequence,
      invalidatesTags: (result) =>
        result
          ? [
              { type: 'LayoutSequence', id: result.id },
              { type: 'LayoutSequence', id: `LIST-${result.project_id}` },
            ]
          : [],
    }),
    deleteLayoutSequence: builder.mutation<
      void,
      { sequenceId: string; projectId: string }
    >({
      query: ({ sequenceId }) => ({
        url: `/layout-sequences/${encodeURIComponent(sequenceId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { sequenceId, projectId }) => [
        { type: 'LayoutSequence', id: sequenceId },
        { type: 'LayoutSequence', id: `LIST-${projectId}` },
        { type: 'LayoutSequenceItems', id: sequenceId },
      ],
    }),
    getLayoutSequenceDetail: builder.query<
      { layoutSequence: LayoutSequence; items: LayoutSequenceItem[] },
      { sequenceId: string }
    >({
      query: ({ sequenceId }) => `/layout-sequences/${encodeURIComponent(sequenceId)}`,
      transformResponse: (response: {
        layout_sequence: LayoutSequence;
        items: LayoutSequenceItem[];
      }) => ({
        layoutSequence: response.layout_sequence,
        items: response.items ?? [],
      }),
      providesTags: (_result, _error, { sequenceId }) => [
        { type: 'LayoutSequence', id: sequenceId },
        { type: 'LayoutSequenceItems', id: sequenceId },
      ],
    }),
    addLayoutSequenceItem: builder.mutation<
      LayoutSequenceItem[],
      { sequenceId: string; laidOutImageId: string }
    >({
      query: ({ sequenceId, laidOutImageId }) => ({
        url: `/layout-sequences/${encodeURIComponent(sequenceId)}/items`,
        method: 'POST',
        body: { laid_out_image_id: laidOutImageId },
      }),
      transformResponse: (response: { items: LayoutSequenceItem[] }) => response.items ?? [],
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'LayoutSequenceItems', id: sequenceId },
      ],
    }),
    reorderLayoutSequenceItems: builder.mutation<
      LayoutSequenceItem[],
      { sequenceId: string; laidOutImageIds: string[] }
    >({
      query: ({ sequenceId, laidOutImageIds }) => ({
        url: `/layout-sequences/${encodeURIComponent(sequenceId)}/items`,
        method: 'PUT',
        body: {
          items: laidOutImageIds.map((id) => ({ laid_out_image_id: id })),
        },
      }),
      transformResponse: (response: { items: LayoutSequenceItem[] }) => response.items ?? [],
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'LayoutSequenceItems', id: sequenceId },
      ],
    }),
    deleteLayoutSequenceItem: builder.mutation<
      void,
      { sequenceId: string; position: number }
    >({
      query: ({ sequenceId, position }) => ({
        url: `/layout-sequences/${encodeURIComponent(sequenceId)}/items/${position}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'LayoutSequenceItems', id: sequenceId },
      ],
    }),
  }),
});

export const {
  useGetProjectsQuery,
  useCreateProjectMutation,
  useDeleteProjectMutation,
  useGetAssetsQuery,
  useUploadAssetsMutation,
  useDeleteAssetMutation,
  useGetImageSequencesQuery,
  useCreateImageSequenceMutation,
  useUpdateImageSequenceMutation,
  useDeleteImageSequenceMutation,
  useGetImageSequenceDetailQuery,
  useAddImageSequenceItemMutation,
  useReorderImageSequenceItemsMutation,
  useDeleteImageSequenceItemMutation,
  useGetImageLayoutTemplatesQuery,
  useCreateImageLayoutTemplateMutation,
  useUpdateImageLayoutTemplateMutation,
  useDeleteImageLayoutTemplateMutation,
  useGetLaidOutImagesQuery,
  useGetLaidOutImageQuery,
  useCreateLaidOutImageMutation,
  useUpdateLaidOutImageMutation,
  useDeleteLaidOutImageMutation,
  usePreviewLaidOutImageQuery,
  useGetLayoutSequencesQuery,
  useCreateLayoutSequenceMutation,
  useUpdateLayoutSequenceMutation,
  useDeleteLayoutSequenceMutation,
  useGetLayoutSequenceDetailQuery,
  useAddLayoutSequenceItemMutation,
  useReorderLayoutSequenceItemsMutation,
  useDeleteLayoutSequenceItemMutation,
} = api;
