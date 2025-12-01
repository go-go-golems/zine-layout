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

export interface ImageLayoutRect {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface ImageLayoutFocusPoint {
  source_x: number;
  source_y: number;
  target_x: number;
  target_y: number;
}

export interface ImageLayoutExportOptions {
  format: string;
  quality: number;
  background: string;
  filename_template: string;
  out_dir: string;
}

// Modern LayoutRequest-based API (Frame/Crop/Presentation split)
export interface ImageLayoutBoxSpacing {
  top: number;
  right: number;
  bottom: number;
  left: number;
}

export interface ImageLayoutPageFrame {
  width_in: number;
  height_in: number;
  dpi: number;
  orientation?: 'portrait' | 'landscape';
  margins_in: ImageLayoutBoxSpacing;
}

export interface ImageLayoutViewportFrame {
  width: number;
  height: number;
}

export interface ImageLayoutVec2 {
  x: number;
  y: number;
}

export interface ImageLayoutVec2Px {
  x: number;
  y: number;
}

export interface ImageLayoutFrameSpec {
  mode: 'ratio' | 'page' | 'viewport';
  ratio?: number | null;
  fill?: 'contain' | 'cover';
  page?: ImageLayoutPageFrame;
  viewport?: ImageLayoutViewportFrame;
  fit_axis?: 'width' | 'height' | 'auto';
}

export interface ImageLayoutCropSpec {
  strategy: 'auto' | 'focus' | 'anchor' | 'manual';
  ratio?: number | null;
  zoom?: number;
  extent?: number;
  anchor?: string;
  pan?: ImageLayoutVec2;
  focus?: ImageLayoutFocusPoint | null;
  units?: 'normalized' | 'px';
}

export interface ImageLayoutPresentationSpec {
  user_scale?: number;
  offset_px?: ImageLayoutVec2Px;
  clamp_to_canvas?: boolean;
}

export interface ImageLayoutRequest {
  frame: ImageLayoutFrameSpec;
  crop: ImageLayoutCropSpec;
  presentation: ImageLayoutPresentationSpec;
  export: ImageLayoutExportOptions;
}

export interface ImageLayoutTemplate {
  id: string;
  project_id?: string | null;
  scope: 'global' | 'project';
  name: string;
  description?: string;
  settings: ImageLayoutRequest;
  created_at: string;
  updated_at: string;
}

export interface ImageLayoutViewportResult {
  source_rect: ImageLayoutRect;
  target_rect: ImageLayoutRect;
  canvas_rect: ImageLayoutRect;
  scale: number;
  mode: 'cover' | 'contain' | string;
}

export interface ImageLayoutTraceStep {
  label: string;
  data: Record<string, unknown>;
}

export interface ImageLayoutTrace {
  inputs?: Record<string, unknown>;
  steps?: ImageLayoutTraceStep[];
}

export interface ImageLayoutComputation {
  layout?: ImageLayoutRequest;
  result: ImageLayoutViewportResult;
  trace?: ImageLayoutTrace;
}

export type LayoutComputation = ImageLayoutComputation;

export interface ImageLayoutPreviewPayload {
  layout: ImageLayoutRequest;
  assetId?: string;
  image?: {
    width: number;
    height: number;
  };
}

export interface ImageLayoutRenderPayload {
  layout: ImageLayoutRequest;
  assetId: string;
}

export interface LaidOutImage {
  id: string;
  project_id: string;
  asset_id: string;
  template_id: string;
  overrides?: Partial<ImageLayoutRequest> | null;
  result?: ImageLayoutComputation;
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

export type ImageLayoutSequence = LayoutSequence;

export interface LayoutSequenceItem {
  sequence_id: string;
  position: number;
  laid_out_image_id: string;
}

export type ImageLayoutSequenceItem = LayoutSequenceItem;

export interface PageTemplate {
  id: string;
  project_id?: string | null;
  scope: 'global' | 'project';
  name: string;
  description?: string;
  template: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface LaidOutPage {
  id: string;
  project_id: string;
  page_template_id: string;
  laid_out_image_id: string;
  result?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Zine {
  id: string;
  project_id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface ZinePage {
  position: number;
  laid_out_page_id: string;
}

const baseQuery = fetchBaseQuery({ baseUrl: '/api' });

const asFileArray = (files: FileList | File[]) => Array.from(files as FileList);

export const api = createApi({
  reducerPath: 'api',
  baseQuery,
  tagTypes: [
    'Project',
    'Asset',
    'ImageSequence',
    'ImageSequenceItems',
    'ImageLayoutTemplate',
    'LaidOutImage',
    'ImageLayoutSequence',
    'ImageLayoutSequenceItems',
    'PageTemplate',
    'LaidOutPage',
    'Zine',
    'ZinePages',
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
        const base = [{ type: 'ImageSequence' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((seq) => ({ type: 'ImageSequence' as const, id: seq.id })),
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
        { type: 'ImageSequence', id: `LIST-${projectId}` },
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
        { type: 'ImageSequence', id: sequenceId },
        { type: 'ImageSequenceItems', id: sequenceId },
      ],
    }),
    deleteImageSequence: builder.mutation<void, { sequenceId: string; projectId: string }>({
      query: ({ sequenceId }) => ({
        url: `/image-sequences/${encodeURIComponent(sequenceId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { sequenceId, projectId }) => [
        { type: 'ImageSequence', id: sequenceId },
        { type: 'ImageSequence', id: `LIST-${projectId}` },
        { type: 'ImageSequenceItems', id: sequenceId },
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
        { type: 'ImageSequence', id: sequenceId },
        { type: 'ImageSequenceItems', id: sequenceId },
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
      async onQueryStarted({ sequenceId, assetId }, { dispatch, queryFulfilled, getState }) {
        // Optimistic update: add item to the end of the sequence
        const patchResult = dispatch(
          api.util.updateQueryData('getImageSequenceDetail', { sequenceId }, (draft) => {
            const currentItems = draft.items ?? [];
            const maxPosition = currentItems.length > 0 
              ? Math.max(...currentItems.map(item => item.position))
              : -1;
            const newItem: ImageSequenceItem = {
              sequence_id: sequenceId,
              position: maxPosition + 1,
              asset_id: assetId,
              is_gap: !assetId,
            };
            draft.items = [...currentItems, newItem];
          })
        );
        try {
          await queryFulfilled;
        } catch (error) {
          patchResult.undo();
          // Error toast will be handled by component
        }
      },
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'ImageSequenceItems', id: sequenceId },
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
      async onQueryStarted({ sequenceId, items }, { dispatch, queryFulfilled }) {
        // Optimistic update: reorder items immediately
        const patchResult = dispatch(
          api.util.updateQueryData('getImageSequenceDetail', { sequenceId }, (draft) => {
            draft.items = items.map((item, idx) => ({
              sequence_id: sequenceId,
              position: idx,
              asset_id: item.assetId,
              is_gap: item.isGap ?? !item.assetId,
            }));
          })
        );
        try {
          await queryFulfilled;
        } catch (error) {
          patchResult.undo();
          // Error toast will be handled by component
        }
      },
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'ImageSequenceItems', id: sequenceId },
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
      async onQueryStarted({ sequenceId, position }, { dispatch, queryFulfilled }) {
        // Optimistic update: remove item immediately
        const patchResult = dispatch(
          api.util.updateQueryData('getImageSequenceDetail', { sequenceId }, (draft) => {
            draft.items = (draft.items ?? []).filter(item => item.position !== position);
            // Reindex remaining items
            draft.items = draft.items.map((item, idx) => ({
              ...item,
              position: idx,
            }));
          })
        );
        try {
          await queryFulfilled;
        } catch (error) {
          patchResult.undo();
          // Error toast will be handled by component
        }
      },
      invalidatesTags: (_result, _error, { sequenceId }) => [
        { type: 'ImageSequenceItems', id: sequenceId },
      ],
    }),

    getGlobalImageLayoutTemplates: builder.query<ImageLayoutTemplate[], void>({
      query: () => `/image-layout-templates`,
      transformResponse: (response: { templates: ImageLayoutTemplate[] }) =>
        response.templates ?? [],
      providesTags: (result) => {
        const base = [{ type: 'ImageLayoutTemplate' as const, id: 'LIST-global' }];
        if (!result) return base;
        return [
          ...result.map((tpl) => ({ type: 'ImageLayoutTemplate' as const, id: tpl.id })),
          ...base,
        ];
      },
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
        const base = [
          { type: 'ImageLayoutTemplate' as const, id: `LIST-${projectId}` },
          { type: 'ImageLayoutTemplate' as const, id: 'LIST-global' },
        ];
        if (!result) return base;
        return [
          ...result.map((tpl) => ({ type: 'ImageLayoutTemplate' as const, id: tpl.id })),
          ...base,
        ];
      },
    }),
    getImageLayoutTemplate: builder.query<
      ImageLayoutTemplate,
      { templateId: string }
    >({
      query: ({ templateId }) => `/image-layout-templates/${encodeURIComponent(templateId)}`,
      transformResponse: (response: { template: ImageLayoutTemplate }) => response.template,
      providesTags: (result) =>
        result ? [{ type: 'ImageLayoutTemplate', id: result.id }] : [],
    }),
    createGlobalImageLayoutTemplate: builder.mutation<
      ImageLayoutTemplate,
      {
        name: string;
        description?: string;
        settings: ImageLayoutTemplateSettingsPayload;
      }
    >({
      query: (body) => ({
        url: `/image-layout-templates`,
        method: 'POST',
        body,
      }),
      transformResponse: (response: { template: ImageLayoutTemplate }) => response.template,
      invalidatesTags: [{ type: 'ImageLayoutTemplate', id: 'LIST-global' }],
    }),
    createImageLayoutTemplate: builder.mutation<
      ImageLayoutTemplate,
      {
        projectId: string;
        name: string;
        description?: string;
        settings: ImageLayoutTemplateSettingsPayload;
      }
    >({
      query: ({ projectId, ...body }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/image-layout-templates`,
        method: 'POST',
        body,
      }),
      transformResponse: (response: { template: ImageLayoutTemplate }) => response.template,
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'ImageLayoutTemplate', id: `LIST-${projectId}` },
      ],
    }),
    updateImageLayoutTemplate: builder.mutation<
      ImageLayoutTemplate,
      {
        templateId: string;
        name?: string;
        description?: string;
        settings?: ImageLayoutTemplateSettingsPayload;
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
        const tags: { type: 'ImageLayoutTemplate'; id: string }[] = [
          { type: 'ImageLayoutTemplate', id: result.id },
        ];
        const listKey = result.project_id ?? 'global';
        tags.push({ type: 'ImageLayoutTemplate', id: `LIST-${listKey}` });
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
        { type: 'ImageLayoutTemplate', id: templateId },
        { type: 'ImageLayoutTemplate', id: `LIST-${scopeKey}` },
      ],
    }),

    getGlobalPageTemplates: builder.query<PageTemplate[], void>({
      query: () => `/page-templates`,
      transformResponse: (response: { page_templates: PageTemplate[] }) =>
        response.page_templates ?? [],
      providesTags: (result) => {
        const base = [{ type: 'PageTemplate' as const, id: 'LIST-global' }];
        if (!result) return base;
        return [
          ...result.map((tpl) => ({ type: 'PageTemplate' as const, id: tpl.id })),
          ...base,
        ];
      },
    }),
    getPageTemplates: builder.query<PageTemplate[], { projectId: string }>({
      query: ({ projectId }) =>
        `/projects/${encodeURIComponent(projectId)}/page-templates`,
      transformResponse: (response: { page_templates: PageTemplate[] }) =>
        response.page_templates ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [
          { type: 'PageTemplate' as const, id: `LIST-${projectId}` },
          { type: 'PageTemplate' as const, id: 'LIST-global' },
        ];
        if (!result) return base;
        return [
          ...result.map((tpl) => ({ type: 'PageTemplate' as const, id: tpl.id })),
          ...base,
        ];
      },
    }),
    getPageTemplate: builder.query<PageTemplate, { templateId: string }>({
      query: ({ templateId }) => `/page-templates/${encodeURIComponent(templateId)}`,
      transformResponse: (response: { page_template: PageTemplate }) => response.page_template,
      providesTags: (result) =>
        result ? [{ type: 'PageTemplate', id: result.id }] : [],
    }),
    createGlobalPageTemplate: builder.mutation<
      PageTemplate,
      { name: string; description?: string; template: Record<string, unknown> }
    >({
      query: (body) => ({
        url: `/page-templates`,
        method: 'POST',
        body,
      }),
      transformResponse: (response: { page_template: PageTemplate }) =>
        response.page_template,
      invalidatesTags: [{ type: 'PageTemplate', id: 'LIST-global' }],
    }),
    createPageTemplate: builder.mutation<
      PageTemplate,
      {
        projectId: string;
        name: string;
        description?: string;
        template: Record<string, unknown>;
      }
    >({
      query: ({ projectId, ...body }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/page-templates`,
        method: 'POST',
        body,
      }),
      transformResponse: (response: { page_template: PageTemplate }) =>
        response.page_template,
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'PageTemplate', id: `LIST-${projectId}` },
      ],
    }),
    updatePageTemplate: builder.mutation<
      PageTemplate,
      {
        templateId: string;
        name?: string;
        description?: string;
        template?: Record<string, unknown>;
        projectId?: string | null;
      }
    >({
      query: ({ templateId, ...body }) => ({
        url: `/page-templates/${encodeURIComponent(templateId)}`,
        method: 'PATCH',
        body,
      }),
      transformResponse: (response: { page_template: PageTemplate }) =>
        response.page_template,
      invalidatesTags: (result) => {
        if (!result) return [];
        const tags: { type: 'PageTemplate'; id: string }[] = [
          { type: 'PageTemplate', id: result.id },
        ];
        const scopeKey = result.project_id ?? 'global';
        tags.push({ type: 'PageTemplate', id: `LIST-${scopeKey}` });
        if (scopeKey !== 'global') {
          tags.push({ type: 'PageTemplate', id: 'LIST-global' });
        }
        return tags;
      },
    }),
    deletePageTemplate: builder.mutation<
      void,
      { templateId: string; scopeKey: string }
    >({
      query: ({ templateId }) => ({
        url: `/page-templates/${encodeURIComponent(templateId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { templateId, scopeKey }) => [
        { type: 'PageTemplate', id: templateId },
        { type: 'PageTemplate', id: `LIST-${scopeKey}` },
        { type: 'PageTemplate', id: 'LIST-global' },
      ],
    }),

    previewLayoutRequest: builder.mutation<
      ImageLayoutComputation,
      { projectId: string; layout: ImageLayoutRequest; assetId?: string; image?: { width: number; height: number } }
    >({
      query: ({ projectId, layout, assetId, image }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/image-layout/preview`,
        method: 'POST',
        body: {
          layout,
          asset_id: assetId,
          image,
        },
      }),
      transformResponse: (response: { result: ImageLayoutComputation }) => response.result,
    }),

    renderLayoutRequest: builder.mutation<
      Blob,
      { projectId: string; layout: ImageLayoutRequest; assetId: string }
    >({
      query: ({ projectId, layout, assetId }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/image-layout/render`,
        method: 'POST',
        body: { layout, asset_id: assetId },
      }),
      responseHandler: async (response) => response.blob(),
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
        overrides?: Partial<ImageLayoutRequest>;
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
        overrides?: Partial<ImageLayoutRequest>;
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
    previewLaidOutImage: builder.query<ImageLayoutComputation, { id: string }>({
      query: ({ id }) => `/laid-out-images/${encodeURIComponent(id)}/preview`,
      transformResponse: (response: { result: ImageLayoutComputation }) => response.result,
    }),

    getLaidOutPages: builder.query<LaidOutPage[], { projectId: string }>({
      query: ({ projectId }) =>
        `/projects/${encodeURIComponent(projectId)}/laid-out-pages`,
      transformResponse: (response: { laid_out_pages: LaidOutPage[] }) =>
        response.laid_out_pages ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'LaidOutPage' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((page) => ({ type: 'LaidOutPage' as const, id: page.id })),
          ...base,
        ];
      },
    }),
    getLaidOutPage: builder.query<LaidOutPage, { id: string }>({
      query: ({ id }) => `/laid-out-pages/${encodeURIComponent(id)}`,
      transformResponse: (response: { laid_out_page: LaidOutPage }) => response.laid_out_page,
      providesTags: (_result, _error, { id }) => [{ type: 'LaidOutPage', id }],
    }),
    createLaidOutPage: builder.mutation<
      LaidOutPage,
      { projectId: string; pageTemplateId: string; laidOutImageId: string }
    >({
      query: ({ projectId, pageTemplateId, laidOutImageId }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/laid-out-pages`,
        method: 'POST',
        body: {
          page_template_id: pageTemplateId,
          laid_out_image_id: laidOutImageId,
        },
      }),
      transformResponse: (response: { laid_out_page: LaidOutPage }) => response.laid_out_page,
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'LaidOutPage', id: `LIST-${projectId}` },
      ],
    }),
    updateLaidOutPage: builder.mutation<
      LaidOutPage,
      { pageId: string; laidOutImageId: string }
    >({
      query: ({ pageId, laidOutImageId }) => ({
        url: `/laid-out-pages/${encodeURIComponent(pageId)}`,
        method: 'PATCH',
        body: { laid_out_image_id: laidOutImageId },
      }),
      transformResponse: (response: { laid_out_page: LaidOutPage }) => response.laid_out_page,
      invalidatesTags: (result) =>
        result
          ? [
              { type: 'LaidOutPage', id: result.id },
              { type: 'LaidOutPage', id: `LIST-${result.project_id}` },
            ]
          : [],
    }),
    deleteLaidOutPage: builder.mutation<
      void,
      { pageId: string; projectId: string }
    >({
      query: ({ pageId }) => ({
        url: `/laid-out-pages/${encodeURIComponent(pageId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { pageId, projectId }) => [
        { type: 'LaidOutPage', id: pageId },
        { type: 'LaidOutPage', id: `LIST-${projectId}` },
      ],
    }),

    getLayoutSequences: builder.query<LayoutSequence[], { projectId: string }>({
      query: ({ projectId }) => `/projects/${encodeURIComponent(projectId)}/layout-sequences`,
      transformResponse: (response: { layout_sequences: LayoutSequence[] }) =>
        response.layout_sequences ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'ImageLayoutSequence' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((seq) => ({ type: 'ImageLayoutSequence' as const, id: seq.id })),
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
        { type: 'ImageLayoutSequence', id: `LIST-${projectId}` },
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
              { type: 'ImageLayoutSequence', id: result.id },
              { type: 'ImageLayoutSequence', id: `LIST-${result.project_id}` },
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
        { type: 'ImageLayoutSequence', id: sequenceId },
        { type: 'ImageLayoutSequence', id: `LIST-${projectId}` },
        { type: 'ImageLayoutSequenceItems', id: sequenceId },
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
        { type: 'ImageLayoutSequence', id: sequenceId },
        { type: 'ImageLayoutSequenceItems', id: sequenceId },
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
        { type: 'ImageLayoutSequenceItems', id: sequenceId },
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
        { type: 'ImageLayoutSequenceItems', id: sequenceId },
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
        { type: 'ImageLayoutSequenceItems', id: sequenceId },
      ],
    }),
    getZines: builder.query<Zine[], { projectId: string }>({
      query: ({ projectId }) => `/projects/${encodeURIComponent(projectId)}/zines`,
      transformResponse: (response: { zines: Zine[] }) => response.zines ?? [],
      providesTags: (result, _error, { projectId }) => {
        const base = [{ type: 'Zine' as const, id: `LIST-${projectId}` }];
        if (!result) return base;
        return [
          ...result.map((zine) => ({ type: 'Zine' as const, id: zine.id })),
          ...base,
        ];
      },
    }),
    createZine: builder.mutation<
      { zine: Zine; pages: ZinePage[] },
      { projectId: string; name?: string; description?: string; laidOutPageIds?: string[] }
    >({
      query: ({ projectId, name, description, laidOutPageIds }) => ({
        url: `/projects/${encodeURIComponent(projectId)}/zines`,
        method: 'POST',
        body: {
          name,
          description,
          laid_out_page_ids: laidOutPageIds ?? [],
        },
      }),
      transformResponse: (response: { zine: Zine; pages: ZinePage[] }) => ({
        zine: response.zine,
        pages: response.pages ?? [],
      }),
      invalidatesTags: (_result, _error, { projectId }) => [
        { type: 'Zine', id: `LIST-${projectId}` },
      ],
    }),
    getZine: builder.query<{ zine: Zine; pages: ZinePage[] }, { zineId: string }>({
      query: ({ zineId }) => `/zines/${encodeURIComponent(zineId)}`,
      transformResponse: (response: { zine: Zine; pages: ZinePage[] }) => ({
        zine: response.zine,
        pages: response.pages ?? [],
      }),
      providesTags: (result) =>
        result
          ? [
              { type: 'Zine' as const, id: result.zine.id },
              { type: 'ZinePages' as const, id: result.zine.id },
            ]
          : [],
    }),
    updateZine: builder.mutation<
      Zine,
      { zineId: string; name?: string; description?: string }
    >({
      query: ({ zineId, ...body }) => ({
        url: `/zines/${encodeURIComponent(zineId)}`,
        method: 'PATCH',
        body,
      }),
      transformResponse: (response: { zine: Zine }) => response.zine,
      invalidatesTags: (result) =>
        result
          ? [
              { type: 'Zine', id: result.id },
              { type: 'Zine', id: `LIST-${result.project_id}` },
            ]
          : [],
    }),
    deleteZine: builder.mutation<void, { zineId: string; projectId: string }>({
      query: ({ zineId }) => ({
        url: `/zines/${encodeURIComponent(zineId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { zineId, projectId }) => [
        { type: 'Zine', id: zineId },
        { type: 'Zine', id: `LIST-${projectId}` },
        { type: 'ZinePages', id: zineId },
      ],
    }),
    getZinePages: builder.query<ZinePage[], { zineId: string }>({
      query: ({ zineId }) => `/zines/${encodeURIComponent(zineId)}/pages`,
      transformResponse: (response: { pages: ZinePage[] }) => response.pages ?? [],
      providesTags: (_result, _error, { zineId }) => [
        { type: 'ZinePages', id: zineId },
      ],
    }),
    setZinePages: builder.mutation<
      ZinePage[],
      { zineId: string; laidOutPageIds: string[] }
    >({
      query: ({ zineId, laidOutPageIds }) => ({
        url: `/zines/${encodeURIComponent(zineId)}/pages`,
        method: 'PUT',
        body: { laid_out_page_ids: laidOutPageIds },
      }),
      transformResponse: (response: { pages: ZinePage[] }) => response.pages ?? [],
      invalidatesTags: (_result, _error, { zineId }) => [
        { type: 'ZinePages', id: zineId },
        { type: 'Zine', id: zineId },
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
  useGetGlobalImageLayoutTemplatesQuery,
  useGetImageLayoutTemplatesQuery,
  useGetImageLayoutTemplateQuery,
  useCreateGlobalImageLayoutTemplateMutation,
  useCreateImageLayoutTemplateMutation,
  useUpdateImageLayoutTemplateMutation,
  useDeleteImageLayoutTemplateMutation,
  useGetGlobalPageTemplatesQuery,
  useGetPageTemplatesQuery,
  useGetPageTemplateQuery,
  useCreateGlobalPageTemplateMutation,
  useCreatePageTemplateMutation,
  useUpdatePageTemplateMutation,
  useDeletePageTemplateMutation,
  usePreviewLayoutRequestMutation,
  useRenderLayoutRequestMutation,
  useGetLaidOutImagesQuery,
  useGetLaidOutImageQuery,
  useCreateLaidOutImageMutation,
  useUpdateLaidOutImageMutation,
  useDeleteLaidOutImageMutation,
  usePreviewLaidOutImageQuery,
  useGetLaidOutPagesQuery,
  useGetLaidOutPageQuery,
  useCreateLaidOutPageMutation,
  useUpdateLaidOutPageMutation,
  useDeleteLaidOutPageMutation,
  useGetLayoutSequencesQuery,
  useCreateLayoutSequenceMutation,
  useUpdateLayoutSequenceMutation,
  useDeleteLayoutSequenceMutation,
  useGetLayoutSequenceDetailQuery,
  useAddLayoutSequenceItemMutation,
  useReorderLayoutSequenceItemsMutation,
  useDeleteLayoutSequenceItemMutation,
  useGetZinesQuery,
  useCreateZineMutation,
  useGetZineQuery,
  useUpdateZineMutation,
  useDeleteZineMutation,
  useGetZinePagesQuery,
  useSetZinePagesMutation,
} = api;
