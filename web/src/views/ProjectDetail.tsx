import React, { useMemo, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  useDeleteImageMutation,
  useGetImagesQuery,
  useReorderImagesMutation,
  useUploadImagesMutation
} from '../api'

const ImgCell: React.FC<{ src: string; alt: string; w: number; h: number }> = ({ src, alt, w, h }) => (
  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6 }}>
    <img src={src} alt={alt} style={{ maxWidth: 160, maxHeight: 120, objectFit: 'contain', border: '1px solid #ddd' }} />
    <small>{w}×{h}</small>
  </div>
)

export const ProjectDetail: React.FC = () => {
  const { id = '' } = useParams()
  const { data, isLoading, refetch } = useGetImagesQuery({ id })
  const [uploadImages, { isLoading: isUploading }] = useUploadImagesMutation()
  const [deleteImage] = useDeleteImageMutation()
  const [reorderImages, { isLoading: isReordering }] = useReorderImagesMutation()
  const fileRef = useRef<HTMLInputElement>(null)
  const [order, setOrder] = useState<string[] | null>(null)

  const currentOrder = order ?? data?.order ?? []
  const imagesById = useMemo(() => {
    const m = new Map<string, { id: string; name: string; width: number; height: number }>()
    data?.images?.forEach((im) => m.set(im.id, im))
    return m
  }, [data])
  const orderedImages = currentOrder.map((id) => imagesById.get(id)).filter(Boolean) as typeof data.images

  const onUpload = async (e: React.FormEvent) => {
    e.preventDefault()
    const files = fileRef.current?.files
    if (!files || files.length === 0) return
    await uploadImages({ id, files }).unwrap()
    fileRef.current!.value = ''
    setOrder(null)
    refetch()
  }

  const onDelete = async (imageId: string) => {
    await deleteImage({ id, imageId }).unwrap()
    setOrder(null)
    refetch()
  }

  const move = (idx: number, dir: -1 | 1) => {
    const next = [...currentOrder]
    const j = idx + dir
    if (j < 0 || j >= next.length) return
    ;[next[idx], next[j]] = [next[j], next[idx]]
    setOrder(next)
  }

  const saveOrder = async () => {
    if (!order) return
    await reorderImages({ id, order }).unwrap()
    setOrder(null)
    refetch()
  }

  return (
    <main>
      <p><Link to="/projects">← Back to Projects</Link></p>
      <h1>Project {id}</h1>

      <form onSubmit={onUpload} style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
        <input ref={fileRef} type="file" accept="image/png" multiple />
        <button type="submit" disabled={isUploading}>Upload</button>
      </form>

      {isLoading ? (
        <p>Loading images…</p>
      ) : (
        <div>
          <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr auto', gap: 12, alignItems: 'center' }}>
            {orderedImages.map((im, i) => (
              <React.Fragment key={im.id}>
                <div>{String(i + 1).padStart(2, '0')}</div>
                <ImgCell src={`/api/projects/${id}/images/${encodeURIComponent(im.id)}`} alt={im.name} w={im.width} h={im.height} />
                <div style={{ display: 'flex', gap: 4 }}>
                  <button onClick={() => move(i, -1)} disabled={i === 0}>↑</button>
                  <button onClick={() => move(i, +1)} disabled={i === orderedImages.length - 1}>↓</button>
                  <button onClick={() => onDelete(im.id)}>Delete</button>
                </div>
              </React.Fragment>
            ))}
          </div>
          <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
            <button onClick={() => refetch()}>Refresh</button>
            <button onClick={saveOrder} disabled={!order || isReordering}>Save Order</button>
          </div>
        </div>
      )}
    </main>
  )
}

