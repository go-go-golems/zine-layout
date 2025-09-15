import type React from 'react';

export const ImgCell: React.FC<{ src: string; alt: string; w: number; h: number }> = ({
  src,
  alt,
  w,
  h,
}) => (
  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6 }}>
    <img
      src={src}
      alt={alt}
      style={{ maxWidth: 160, maxHeight: 120, objectFit: 'contain', border: '1px solid #ddd' }}
    />
    <small>
      {w}×{h}
    </small>
  </div>
);
