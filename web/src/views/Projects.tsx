import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { useCreateProjectMutation, useDeleteProjectMutation, useGetProjectsQuery } from '../api'

export const Projects: React.FC = () => {
  const { data, isLoading, refetch } = useGetProjectsQuery()
  const [createProject, { isLoading: isCreating }] = useCreateProjectMutation()
  const [deleteProject] = useDeleteProjectMutation()
  const [name, setName] = useState('')

  const onCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    await createProject({ name: name.trim() || undefined }).unwrap()
    setName('')
    refetch()
  }

  return (
    <main>
      <h1>Projects</h1>
      <form onSubmit={onCreate} style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="New project name" />
        <button type="submit" disabled={isCreating}>Create</button>
      </form>
      {isLoading ? (
        <p>Loading…</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Updated</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {data?.projects?.map((p) => (
              <tr key={p.id}>
                <td><Link to={`/projects/${p.id}`}>{p.name}</Link></td>
                <td>{new Date(p.updatedAt).toLocaleString()}</td>
                <td>
                  <button onClick={() => deleteProject({ id: p.id }).unwrap().then(() => refetch())}>Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </main>
  )
}
