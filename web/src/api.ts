import type { FetchBaseQueryError } from '@reduxjs/toolkit/query';
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export interface Project {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
  presetId?: string;
}

export interface ImageItem {
  id: string;
  name: string;
  width: number;
  height: number;
}

export interface PresetInfo {
  id: string;
  name: string;
  filename: string;
}

export interface ValidationDetails {
  count: number;
  width: number;
  height: number;
  rows: number;
  columns: number;
  pages: number;
  multiple: number;
}

export interface UploadResponse {
  name: string;
  url: string;
  bytes: number;
}

export interface SpreadSettings {
  paper_width_in: number;
  paper_height_in: number;
  dpi: number;
  orientation: string;
  margin_top_in: number;
  margin_right_in: number;
  margin_bottom_in: number;
  margin_left_in: number;
  is_spread: boolean;
  gutter_in: number;
  crop_ratio?: number;
  crop_to_fill: boolean;
  user_scale: number;
  position_x: number;
  position_y: number;
  units: string;
  export: {
    format: string;
    quality: number;
    background: string;
    out_dir: string;
    filename_template: string;
  };
}

export interface ComputeRequest {
  image_path?: string;
  meta?: { width: number; height: number };
  name?: string;
  settings: SpreadSettings;
}

export interface ComputeResult {
  result: any;
  trace?: string[];
}

export interface YamlRenderRequest {
  yaml: string;
  base_dir?: string;
}

export interface PanelPreview {
  panel: string;
  mime_type: string;
  data_url: string;
  width: number;
  height: number;
}

export interface YamlRenderSpread {
  name: string;
  image_path: string;
  result: any;
  trace?: string[];
  panels: PanelPreview[];
}

export interface YamlRenderResponse {
  spreads: YamlRenderSpread[];
}

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Project', 'Image', 'Preset'],
  endpoints: (b) => ({
    getProjects: b.query<{ projects: Project[] }, void>({
      query: () => '/projects',
      providesTags: ['Project'],
    }),
    createProject: b.mutation<{ project: Project }, { name?: string; presetId?: string }>({
      query: (body) => ({ url: '/projects', method: 'POST', body }),
      invalidatesTags: ['Project'],
    }),
    deleteProject: b.mutation<{ ok: boolean }, { id: string }>({
      query: ({ id }) => ({ url: `/projects/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Project'],
    }),
    getImages: b.query<{ images: ImageItem[]; order: string[] }, { id: string }>({
      query: ({ id }) => `/projects/${id}/images`,
      providesTags: ['Image'],
    }),
    uploadImages: b.mutation<{ images: ImageItem[] }, { id: string; files: FileList | File[] }>({
      query: ({ id, files }) => {
        const fd = new FormData();
        const list = Array.from(files as FileList);
        for (const f of list) fd.append('images[]', f);
        return { url: `/projects/${id}/images`, method: 'POST', body: fd };
      },
      invalidatesTags: ['Image'],
    }),
    deleteImage: b.mutation<{ ok: boolean }, { id: string; imageId: string }>({
      query: ({ id, imageId }) => ({
        url: `/projects/${id}/images/${encodeURIComponent(imageId)}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Image'],
    }),
    reorderImages: b.mutation<{ ok: boolean }, { id: string; order: string[] }>({
      query: ({ id, order }) => ({
        url: `/projects/${id}/images/reorder`,
        method: 'POST',
        body: { order },
      }),
      invalidatesTags: ['Image'],
    }),
    getYaml: b.query<string, { id: string }>({
      // fetch raw text via queryFn
      async queryFn({ id }) {
        try {
          const resp = await fetch(`/api/projects/${id}/yaml`);
          if (!resp.ok) {
            const err: FetchBaseQueryError = {
              status: resp.status,
              data: await resp.text(),
            } as unknown as FetchBaseQueryError;
            return { error: err };
          }
          const text = await resp.text();
          return { data: text };
        } catch (e) {
          const err: FetchBaseQueryError = {
            status: 'FETCH_ERROR',
            data: String(e),
          } as unknown as FetchBaseQueryError;
          return { error: err };
        }
      },
    }),
    putYaml: b.mutation<{ ok: boolean }, { id: string; yaml: string }>({
      query: ({ id, yaml }) => ({
        url: `/projects/${id}/yaml`,
        method: 'PUT',
        body: yaml,
        headers: { 'Content-Type': 'text/plain' },
      }),
    }),
    getPresets: b.query<{ presets: PresetInfo[] }, void>({
      query: () => '/presets',
      providesTags: ['Preset'],
    }),
    getPresetYaml: b.query<string, { id: string }>({
      async queryFn({ id }) {
        try {
          const resp = await fetch(`/api/presets/${encodeURIComponent(id)}`);
          if (!resp.ok) {
            const err: FetchBaseQueryError = {
              status: resp.status,
              data: await resp.text(),
            } as unknown as FetchBaseQueryError;
            return { error: err };
          }
          const text = await resp.text();
          return { data: text };
        } catch (e) {
          const err: FetchBaseQueryError = {
            status: 'FETCH_ERROR',
            data: String(e),
          } as unknown as FetchBaseQueryError;
          return { error: err };
        }
      },
      providesTags: (_r, _e, arg) => [{ type: 'Preset' as const, id: arg.id }],
    }),
    applyPreset: b.mutation<{ ok: boolean }, { id: string; presetId: string }>({
      query: ({ id, presetId }) => ({
        url: `/projects/${id}/preset`,
        method: 'POST',
        body: { presetId },
      }),
      invalidatesTags: ['Project'],
    }),
    validateProject: b.query<
      { ok: boolean; issues: string[]; details?: ValidationDetails },
      { id: string }
    >({
      query: ({ id }) => ({ url: `/projects/${id}/validate`, method: 'POST', body: {} }),
    }),
    renderProject: b.mutation<
      { renderId: string; files: string[] },
      { id: string; test?: boolean; test_bw?: boolean; test_dimensions?: string }
    >({
      query: ({ id, ...body }) => ({ url: `/projects/${id}/render`, method: 'POST', body }),
    }),
    getRenders: b.query<{ renders: { id: string; files: string[] }[] }, { id: string }>({
      query: ({ id }) => `/projects/${id}/renders`,
    }),
    uploadFile: b.mutation<UploadResponse, File>({
      query: (file) => {
        const fd = new FormData();
        fd.append('file', file);
        return { url: '/uploads', method: 'POST', body: fd };
      },
    }),
    computeSpread: b.mutation<ComputeResult, ComputeRequest>({
      query: (body) => ({ url: '/v1/compute', method: 'POST', body }),
    }),
    previewSpread: b.mutation<Blob, ComputeRequest>({
      query: (body) => ({
        url: '/v1/preview',
        method: 'POST',
        body,
        responseHandler: (response) => response.blob(),
      }),
    }),
    buildYaml: b.mutation<string, ComputeRequest>({
      query: (body) => ({
        url: '/v1/yaml',
        method: 'POST',
        body,
        responseHandler: async (response) => response.text(),
      }),
    }),
    renderYaml: b.mutation<YamlRenderResponse, YamlRenderRequest>({
      query: (body) => ({
        url: '/v1/yaml/render',
        method: 'POST',
        body,
      }),
    }),
    exportBookYaml: b.query<string, { id: string; baseDir?: string }>({
      query: ({ id, baseDir }) => ({
        url: `/projects/${id}/yaml/book`,
        method: 'GET',
        params: baseDir ? { base_dir: baseDir } : undefined,
        responseHandler: async (response) => response.text(),
      }),
    }),
    getPreviewSpread: b.query<string, ComputeRequest>({
      query: (body) => ({
        url: '/v1/preview',
        method: 'POST',
        body,
        responseHandler: async (response) => {
          const blob = await response.blob();
          return URL.createObjectURL(blob);
        },
      }),
      keepUnusedDataFor: 30, // Cache preview for 30 seconds
    }),
  }),
});

export const {
  useGetProjectsQuery,
  useCreateProjectMutation,
  useDeleteProjectMutation,
  useGetImagesQuery,
  useUploadImagesMutation,
  useDeleteImageMutation,
  useReorderImagesMutation,
  useGetPresetsQuery,
  useGetPresetYamlQuery,
  useApplyPresetMutation,
  useGetYamlQuery,
  usePutYamlMutation,
  useValidateProjectQuery,
  useLazyValidateProjectQuery,
  useRenderProjectMutation,
  useGetRendersQuery,
  useUploadFileMutation,
  useComputeSpreadMutation,
  usePreviewSpreadMutation,
  useGetPreviewSpreadQuery,
  useBuildYamlMutation,
  useRenderYamlMutation,
  useExportBookYamlQuery,
  useLazyExportBookYamlQuery,
} = api;
