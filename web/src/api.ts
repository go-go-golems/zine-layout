import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export interface Project {
  id: string
  name: string
  createdAt: string
  updatedAt: string
  presetId?: string
}

export interface ImageItem {
  id: string
  name: string
  width: number
  height: number
}

export interface PresetInfo {
  id: string
  name: string
  filename: string
}

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Project', 'Image', 'Preset'],
  endpoints: (b) => ({
    getProjects: b.query<{ projects: Project[] }, void>({
      query: () => '/projects',
      providesTags: ['Project']
    }),
    createProject: b.mutation<{ project: Project }, { name?: string; presetId?: string }>({
      query: (body) => ({ url: '/projects', method: 'POST', body }),
      invalidatesTags: ['Project']
    }),
    deleteProject: b.mutation<{ ok: boolean }, { id: string }>({
      query: ({ id }) => ({ url: `/projects/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Project']
    }),
    getImages: b.query<{ images: ImageItem[]; order: string[] }, { id: string }>({
      query: ({ id }) => `/projects/${id}/images`,
      providesTags: ['Image']
    }),
    uploadImages: b.mutation<{ images: ImageItem[] }, { id: string; files: FileList | File[] }>({
      query: ({ id, files }) => {
        const fd = new FormData()
        const list = Array.from(files as FileList)
        for (const f of list) fd.append('images[]', f)
        return { url: `/projects/${id}/images`, method: 'POST', body: fd }
      },
      invalidatesTags: ['Image']
    }),
    deleteImage: b.mutation<{ ok: boolean }, { id: string; imageId: string }>({
      query: ({ id, imageId }) => ({ url: `/projects/${id}/images/${encodeURIComponent(imageId)}`, method: 'DELETE' }),
      invalidatesTags: ['Image']
    }),
    reorderImages: b.mutation<{ ok: boolean }, { id: string; order: string[] }>({
      query: ({ id, order }) => ({ url: `/projects/${id}/images/reorder`, method: 'POST', body: { order } }),
      invalidatesTags: ['Image']
    }),
    getYaml: b.query<string, { id: string }>({
      query: ({ id }) => ({ url: `/projects/${id}/yaml` }),
      responseHandler: 'text'
    }),
    putYaml: b.mutation<{ ok: boolean }, { id: string; yaml: string }>({
      query: ({ id, yaml }) => ({ url: `/projects/${id}/yaml`, method: 'PUT', body: yaml, headers: { 'Content-Type': 'text/plain' } })
    }),
    getPresets: b.query<{ presets: PresetInfo[] }, void>({
      query: () => '/presets',
      providesTags: ['Preset']
    }),
    getPresetYaml: b.query<string, { id: string }>({
      query: ({ id }) => ({ url: `/presets/${encodeURIComponent(id)}` }),
      responseHandler: 'text',
      providesTags: (_r, _e, arg) => [{ type: 'Preset', id: arg.id } as any]
    }),
    applyPreset: b.mutation<{ ok: boolean }, { id: string; presetId: string }>({
      query: ({ id, presetId }) => ({ url: `/projects/${id}/preset`, method: 'POST', body: { presetId } }),
      invalidatesTags: ['Project']
    }),
    validateProject: b.query<{ ok: boolean; issues: string[]; details?: any }, { id: string }>({
      query: ({ id }) => ({ url: `/projects/${id}/validate`, method: 'POST', body: {} })
    })
  })
})

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
  useLazyValidateProjectQuery
} = api
