import React from 'react';
import { useGetRendersQuery, useRenderProjectMutation } from '../api';

export const ProjectRenderPanel: React.FC<{ id: string }> = ({ id }) => {
  const { data, refetch, isFetching } = useGetRendersQuery({ id });
  const [renderProject, { isLoading }] = useRenderProjectMutation();
  const [test, setTest] = React.useState(false);
  const [testBW, setTestBW] = React.useState(false);
  const [testDimensions, setTestDimensions] = React.useState('600px,800px');

  const onRender = async () => {
    await renderProject({ id, test, test_bw: testBW, test_dimensions: testDimensions }).unwrap();
    refetch();
  };

  return (
    <section>
      <h2>Render</h2>
      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <label>
          <input type="checkbox" checked={test} onChange={(e) => setTest(e.target.checked)} /> Test
        </label>
        <label>
          <input type="checkbox" checked={testBW} onChange={(e) => setTestBW(e.target.checked)} />{' '}
          BW
        </label>
        <input
          value={testDimensions}
          onChange={(e) => setTestDimensions(e.target.value)}
          placeholder="WIDTH,HEIGHT"
          style={{ width: 160 }}
        />
        <button type="button" disabled={isLoading} onClick={onRender}>
          Render
        </button>
        <button type="button" disabled={isFetching} onClick={() => refetch()}>
          Refresh
        </button>
      </div>
      <div style={{ marginTop: 8 }}>
        {data?.renders?.length ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {data.renders.map((r) => (
              <div key={r.id} style={{ border: '1px solid #eee', padding: 8 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <strong>{r.id}</strong>
                  <a href={`/api/projects/${id}/renders/${r.id}/download.zip`}>Download ZIP</a>
                </div>
                <div style={{ display: 'flex', gap: 8, marginTop: 8, overflowX: 'auto' }}>
                  {r.files.map((f) => (
                    <img
                      key={f}
                      src={`/api/projects/${id}/renders/${r.id}/files/${encodeURIComponent(f)}`}
                      alt={f}
                      style={{ height: 120, border: '1px solid #ddd' }}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p>No renders yet.</p>
        )}
      </div>
    </section>
  );
};
