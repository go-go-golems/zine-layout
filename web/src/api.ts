import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export interface Project {
  id: string
  name: string
  createdAt: string
  updatedAt: string
}

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Project'],
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
    })
  })
})

export const {
  useGetProjectsQuery,
  useCreateProjectMutation,
  useDeleteProjectMutation
} = api

