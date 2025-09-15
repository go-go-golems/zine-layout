import yaml from 'js-yaml';
import React from 'react';
import { useGetImagesQuery, useSpecFromUIMutation } from '../api';

type Cell = { row: number; col: number };

export const ProjectLayoutEditor: React.FC<{ id: string }> = ({ id }) => {
  const { data } = useGetImagesQuery({ id });
  const [rows, setRows] = React.useState(2);
  const [cols, setCols] = React.useState(2);
  const [assignments, setAssignments] = React.useState<Map<string, number>>(new Map());
  const [specFromUI, { isLoading: isSaving }] = useSpecFromUIMutation();

  const order = data?.order ?? [];
  const indexOf = (imageId: string) => {
    const idx = order.indexOf(imageId);
    return idx >= 0 ? idx + 1 : 0; // input_index is 1-based; 0 means unassigned
  };

  const onDropCell = (cell: Cell) => (e: React.DragEvent) => {
    e.preventDefault();
    const im =
      e.dataTransfer.getData('application/x-zine-image-id') || e.dataTransfer.getData('text/plain');
    if (!im) return;
    const inputIndex = indexOf(im);
    if (!inputIndex) return;
    const key = `${cell.row},${cell.col}`;
    const next = new Map(assignments);
    next.set(key, inputIndex);
    setAssignments(next);
  };
  const onDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };
  const onClear = (cell: Cell) => {
    const key = `${cell.row},${cell.col}`;
    const next = new Map(assignments);
    next.delete(key);
    setAssignments(next);
  };

  const toYaml = () => {
    const layout: {
      input_index: number;
      position: { row: number; column: number };
      rotation: number;
    }[] = [];
    assignments.forEach((inputIndex, key) => {
      const [r, c] = key.split(',').map((x) => Number.parseInt(x, 10));
      layout.push({ input_index: inputIndex, position: { row: r, column: c }, rotation: 0 });
    });
    const doc = {
      global: { ppi: 300 },
      page_setup: { grid_size: { rows, columns: cols }, orientation: 'portrait' },
      output_pages: [{ id: 'page1', layout }],
    };
    return yaml.dump(doc, { noRefs: true });
  };

  const onSave = async () => {
    const y = toYaml();
    await specFromUI({ id, yaml: y }).unwrap();
  };

  return (
    <section>
      <h2>Layout</h2>
      <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 8 }}>
        <label>
          Rows
          <input
            type="number"
            min={1}
            max={12}
            value={rows}
            onChange={(e) => setRows(Number.parseInt(e.target.value || '1', 10))}
            style={{ width: 60, marginLeft: 6 }}
          />
        </label>
        <label>
          Columns
          <input
            type="number"
            min={1}
            max={12}
            value={cols}
            onChange={(e) => setCols(Number.parseInt(e.target.value || '1', 10))}
            style={{ width: 60, marginLeft: 6 }}
          />
        </label>
        <button type="button" disabled={isSaving} onClick={onSave}>
          Save Layout
        </button>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: `repeat(${cols}, 120px)`, gap: 8 }}>
        {Array.from({ length: rows }).map((_, r) =>
          Array.from({ length: cols }).map((__, c) => {
            const key = `${r},${c}`;
            const idx = assignments.get(key) || 0;
            return (
              <div
                key={key}
                onDragOver={onDragOver}
                onDrop={onDropCell({ row: r, col: c })}
                style={{
                  width: 120,
                  height: 90,
                  border: '2px dashed #ccc',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  position: 'relative',
                }}
                title="Drop an image here"
              >
                {idx ? (
                  <span style={{ fontSize: 24, fontWeight: 700 }}>{idx}</span>
                ) : (
                  <span style={{ color: '#aaa' }}>Empty</span>
                )}
                {idx ? (
                  <button
                    type="button"
                    onClick={() => onClear({ row: r, col: c })}
                    style={{ position: 'absolute', top: 2, right: 2 }}
                  >
                    ×
                  </button>
                ) : null}
              </div>
            );
          }),
        )}
      </div>
      <div style={{ marginTop: 12 }}>
        <h3>Generated YAML</h3>
        <pre style={{ background: '#f7f7f7', padding: 8, maxHeight: 240, overflow: 'auto' }}>
          {toYaml()}
        </pre>
      </div>
      <p style={{ color: '#666' }}>
        Tip: Drag an image from the tray onto a cell to assign input_index.
      </p>
    </section>
  );
};
