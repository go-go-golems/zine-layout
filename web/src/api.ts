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

const baseQuery = fetchBaseQuery({ baseUrl: '/api' });

const asFileArray = (files: FileList | File[]) => Array.from(files as FileList);

export const api = createApi({
  reducerPath: 'api',
  baseQuery,
  tagTypes: ['Project', 'Asset', 'Sequence', 'SequenceItems'],
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
} = api;
