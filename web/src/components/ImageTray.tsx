import React, { useMemo, useRef, useState } from 'react'
import {
  useDeleteImageMutation,
  useGetImagesQuery,
  useReorderImagesMutation,
  useUploadImagesMutation
} from '../api'
import { ImgCell } from './ImgCell'

export const ImageTray: React.FC<{ id: string }> = ({ id }) => {
  const { data, isLoading, refetch } = useGetImagesQuery({ id })
  const [uploadImages, { isLoading: isUploading }] = useUploadImagesMutation()
  const [deleteImage] = useDeleteImageMutation()
  const [reorderImages, { isLoading: isReordering }] = useReorderImagesMutation()
  const fileRef = useRef<HTMLInputElement>(null)
  const [order, setOrder] = useState<string[] | null>(null)
  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const [isDragOverDropzone, setIsDragOverDropzone] = useState(false)

  const currentOrder = order ?? data?.order ?? []
  const imagesById = useMemo(() => {
    const m = new Map<string, { id: string; name: string; width: number; height: number }>()
    data?.images?.forEach((im) => m.set(im.id, im))
    return m
  }, [data])
  const orderedImages = currentOrder.map((i) => imagesById.get(i)).filter(Boolean) as typeof data.images

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

  const onItemDragStart = (idx: number) => (e: React.DragEvent) => {
    setDragIndex(idx)
    e.dataTransfer.effectAllowed = 'move'
  }
  const onItemDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }
  const onItemDrop = (targetIndex: number) => (e: React.DragEvent) => {
    e.preventDefault()
    if (dragIndex === null || dragIndex === targetIndex) return
    const next = [...currentOrder]
    const [moved] = next.splice(dragIndex, 1)
    next.splice(targetIndex, 0, moved)
    setOrder(next)
    setDragIndex(null)
  }

  const onDropzoneDragOver = (e: React.DragEvent) => { e.preventDefault(); setIsDragOverDropzone(true) }
  const onDropzoneDragLeave = () => setIsDragOverDropzone(false)
  const onDropzoneDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOverDropzone(false)
    const files = Array.from(e.dataTransfer.files || [])
    if (files.length === 0) return
    const pngs = files.filter((f) => f.type === 'image/png' || f.name.toLowerCase().endsWith('.png'))
    if (pngs.length === 0) return
    await uploadImages({ id, files: pngs }).unwrap()
    setOrder(null)
    refetch()
  }

  return (
    <section>
      <form onSubmit={onUpload} style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
        <input ref={fileRef} type="file" accept="image/png" multiple />
        <button type="submit" disabled={isUploading}>Upload</button>
      </form>
      <div
        onDragOver={onDropzoneDragOver}
        onDragLeave={onDropzoneDragLeave}
        onDrop={onDropzoneDrop}
        style={{ padding: 12, border: '2px dashed ' + (isDragOverDropzone ? '#007acc' : '#ccc'), background: isDragOverDropzone ? '#eef7ff' : 'transparent', marginBottom: 12 }}
      >
        Drop PNG files here to upload
      </div>
      {isLoading ? (
        <p>Loading images…</p>
      ) : (
        <div>
          <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr auto', gap: 12, alignItems: 'center' }}>
            {orderedImages.map((im, i) => (
              <React.Fragment key={im.id}>
                <div>{String(i + 1).padStart(2, '0')}</div>
                <div draggable onDragStart={onItemDragStart(i)} onDragOver={onItemDragOver} onDrop={onItemDrop(i)} style={{ padding: 6, border: '1px solid #eee' }} title="Drag to reorder">
                  <ImgCell src={`/api/projects/${id}/images/${encodeURIComponent(im.id)}`} alt={im.name} w={im.width} h={im.height} />
                </div>
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
    </section>
  )
}

