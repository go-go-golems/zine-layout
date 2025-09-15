import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export interface Project {
  id: string
  name: string
  createdAt: string
  updatedAt: string
}

export interface ImageItem {
  id: string
  name: string
  width: number
  height: number
}

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Project', 'Image'],
  endpoints: (b) => ({
    getProjects: b.query<{ projects: Project[] }, void>({
      query: () => '/projects',
      providesTags: ['Project']
    }),
    createProject: b.mutation<{ project: Project }, { name?: string }>({
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
  useReorderImagesMutation
} = api
