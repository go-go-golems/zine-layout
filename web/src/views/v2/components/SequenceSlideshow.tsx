import React, { useState, useEffect, useCallback, useMemo } from 'react';
import type { ImageSequenceItem, Asset } from '../../../api';
import { Button } from '../../../components/ui';

type Slide = {
  item: ImageSequenceItem;
  asset?: Asset;
};

type SpreadSlide = {
  left?: Slide;
  right?: Slide;
};

type ViewMode = 'single' | 'spread';

interface SequenceSlideshowProps {
  items: ImageSequenceItem[];
  assets: Map<string, Asset>;
  projectId: string;
  isOpen: boolean;
  onClose: () => void;
  initialIndex?: number;
  initialViewMode?: ViewMode;
}

export const SequenceSlideshow: React.FC<SequenceSlideshowProps> = ({
  items,
  assets,
  projectId,
  isOpen,
  onClose,
  initialIndex = 0,
  initialViewMode = 'single',
}) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>(initialViewMode);
  const [anchorPosition, setAnchorPosition] = useState(initialIndex);

  const singleSlides = useMemo<Slide[]>(() => {
    return items
      .filter((item) => !item.is_gap && item.asset_id)
      .map((item) => ({
        item,
        asset: item.asset_id ? assets.get(item.asset_id) : undefined,
      }));
  }, [items, assets]);

  const spreadSlides = useMemo<SpreadSlide[]>(() => {
    const pages: Slide[] = items.map((item) => ({
      item,
      asset: item.asset_id ? assets.get(item.asset_id) : undefined,
    }));

    const spreads: SpreadSlide[] = [];
    for (let i = 0; i < pages.length; i += 2) {
      spreads.push({
        left: pages[i],
        right: pages[i + 1],
      });
    }
    return spreads;
  }, [items, assets]);

  const activeLength = viewMode === 'spread' ? spreadSlides.length : singleSlides.length;

  // Reset to initial state when opening
  useEffect(() => {
    if (isOpen) {
      setViewMode(initialViewMode);
      setAnchorPosition(initialIndex);
    }
  }, [isOpen, initialViewMode, initialIndex]);

  // Ensure current index stays within bounds and mirrors anchor position
  useEffect(() => {
    if (!isOpen) return;
    const targetSlides = viewMode === 'spread' ? spreadSlides : singleSlides;
    if (targetSlides.length === 0) {
      setCurrentIndex(0);
      return;
    }

    if (viewMode === 'spread') {
      const spreadIndex = targetSlides.findIndex((slide) => {
        const leftPos = slide.left?.item.position;
        const rightPos = slide.right?.item.position;
        return leftPos === anchorPosition || rightPos === anchorPosition;
      });
      if (spreadIndex >= 0) {
        setCurrentIndex(spreadIndex);
      } else {
        const fallback = Math.floor(anchorPosition / 2);
        setCurrentIndex(Math.min(Math.max(fallback, 0), targetSlides.length - 1));
      }
    } else {
      const slideIndex = targetSlides.findIndex(
        (slide) => slide.item.position === anchorPosition
      );
      if (slideIndex >= 0) {
        setCurrentIndex(slideIndex);
      } else {
        setCurrentIndex(0);
      }
    }
  }, [isOpen, viewMode, anchorPosition, singleSlides, spreadSlides]);

  useEffect(() => {
    const targetSlides = viewMode === 'spread' ? spreadSlides : singleSlides;
    if (currentIndex >= targetSlides.length && targetSlides.length > 0) {
      setCurrentIndex(targetSlides.length - 1);
    }
  }, [viewMode, currentIndex, singleSlides, spreadSlides]);

  // Track the sequence position the user is viewing to keep context when switching modes
  useEffect(() => {
    if (!isOpen) return;
    if (viewMode === 'spread') {
      const spread = spreadSlides[currentIndex];
      const page = spread?.left ?? spread?.right;
      if (page && page.item.position !== anchorPosition) {
        setAnchorPosition(page.item.position);
      }
    } else {
      const slide = singleSlides[currentIndex];
      if (slide && slide.item.position !== anchorPosition) {
        setAnchorPosition(slide.item.position);
      }
    }
  }, [isOpen, viewMode, currentIndex, spreadSlides, singleSlides, anchorPosition]);

  // Keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (isFullscreen) {
          exitFullscreen();
        } else {
          onClose();
        }
      } else if (e.key === 'ArrowLeft') {
        goToPrevious();
      } else if (e.key === 'ArrowRight') {
        goToNext();
      } else if (e.key === 'f' || e.key === 'F') {
        if (!isFullscreen) {
          enterFullscreen();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, isFullscreen, goToPrevious, goToNext, enterFullscreen, exitFullscreen, onClose]);

  // Fullscreen change listener
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange);
  }, []);

  const goToPrevious = useCallback(() => {
    if (activeLength === 0) return;
    setCurrentIndex((prev) => (prev > 0 ? prev - 1 : activeLength - 1));
  }, [activeLength]);

  const goToNext = useCallback(() => {
    if (activeLength === 0) return;
    setCurrentIndex((prev) => (prev < activeLength - 1 ? prev + 1 : 0));
  }, [activeLength]);

  const enterFullscreen = useCallback(async () => {
    const element = document.documentElement;
    try {
      if (element.requestFullscreen) {
        await element.requestFullscreen();
      }
    } catch (err) {
      console.error('Failed to enter fullscreen:', err);
    }
  }, []);

  const exitFullscreen = useCallback(async () => {
    try {
      if (document.exitFullscreen) {
        await document.exitFullscreen();
      }
    } catch (err) {
      console.error('Failed to exit fullscreen:', err);
    }
  }, []);

  if (!isOpen) return null;

  const currentSingleSlide = viewMode === 'single' ? singleSlides[currentIndex] : undefined;
  const currentSpreadSlide = viewMode === 'spread' ? spreadSlides[currentIndex] : undefined;
  const imageUrl =
    viewMode === 'single' && currentSingleSlide?.asset
      ? currentSingleSlide.asset.url ??
        `/projects/${projectId}/images/${encodeURIComponent(currentSingleSlide.asset.filename)}`
      : null;

  const renderPage = (page: Slide | undefined, side: 'left' | 'right') => {
    const isBlank = !page || page.item.is_gap || !page.asset;
    const label = side === 'left' ? 'Left Page' : 'Right Page';

    return (
      <div className="flex-1 flex flex-col items-center gap-2" key={side}>
        <div className="relative w-full aspect-[3/4] bg-white rounded-2xl shadow-2xl overflow-hidden flex items-center justify-center border border-gray-200">
          {isBlank ? (
            <div className="text-center text-gray-400">
              <div className="text-4xl mb-2">⏸</div>
              <p className="text-base font-medium">Blank page</p>
              {page?.item && (
                <p className="text-xs mt-2 text-gray-500">Position {page.item.position + 1}</p>
              )}
            </div>
          ) : (
            <img
              src={`${(page.asset?.url ??
                `/projects/${projectId}/images/${encodeURIComponent(
                  page.asset!.filename
                )}`)}?t=${
                page.asset?.uploaded_at ? new Date(page.asset.uploaded_at).getTime() : Date.now()
              }`}
              alt={page.asset?.filename}
              className="w-full h-full object-contain"
            />
          )}
          <div className="absolute bottom-3 left-4 text-[11px] uppercase tracking-wide text-gray-500">
            {label}
            {page?.item ? ` · Position ${page.item.position + 1}` : ''}
          </div>
        </div>
        {!isBlank && page?.asset?.filename && (
          <p className="text-xs text-white/80">{page.asset.filename}</p>
        )}
      </div>
    );
  };

  const noSlidesAvailable = activeLength === 0;

  return (
    <div
      className={`fixed inset-0 z-50 bg-black ${
        isFullscreen ? '' : 'bg-opacity-90'
      } flex flex-col`}
    >
      {/* Header controls */}
      <div className="absolute top-0 left-0 right-0 z-10 bg-black/50 backdrop-blur-sm p-4 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            onClick={onClose}
            variant="secondary"
            size="sm"
            className="bg-white/10 hover:bg-white/20 text-white border-white/20"
          >
            ✕ Close
          </Button>
          <span className="text-white text-sm">
            {activeLength ? `${currentIndex + 1} / ${activeLength}` : 'No slides in this view'}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => setViewMode((prev) => (prev === 'spread' ? 'single' : 'spread'))}
            variant="secondary"
            size="sm"
            className="bg-white/10 hover:bg-white/20 text-white border-white/20"
          >
            {viewMode === 'spread' ? 'Single Page View' : '📖 View as Book Spread'}
          </Button>
          <Button
            onClick={isFullscreen ? exitFullscreen : enterFullscreen}
            variant="secondary"
            size="sm"
            className="bg-white/10 hover:bg-white/20 text-white border-white/20"
          >
            {isFullscreen ? '⤓ Exit Fullscreen' : '⤢ Fullscreen'}
          </Button>
        </div>
      </div>

      {/* Main image area */}
      <div className="flex-1 flex items-center justify-center p-8 relative">
        {/* Previous button */}
        {activeLength > 1 && (
          <button
            onClick={goToPrevious}
            className="absolute left-4 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white p-3 rounded-full transition-all z-10"
            aria-label="Previous image"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 19l-7-7 7-7"
              />
            </svg>
          </button>
        )}

        {/* Image */}
        {noSlidesAvailable ? (
          <div className="text-white/80 text-center space-y-4 max-w-md">
            <p className="text-lg font-semibold">Nothing to preview in this mode.</p>
            <p className="text-sm text-white/60">
              {viewMode === 'single'
                ? 'This sequence only contains gaps. Switch to the book spread view to see how blank pages will look.'
                : 'Add images or gaps to your sequence to preview this spread.'}
            </p>
            {viewMode === 'single' && (
              <Button
                onClick={() => setViewMode('spread')}
                variant="secondary"
                size="sm"
                className="bg-white/10 hover:bg-white/20 text-white border-white/20"
              >
                📖 Switch to Book Spread
              </Button>
            )}
          </div>
        ) : viewMode === 'spread' && currentSpreadSlide ? (
          <div className="w-full max-w-6xl flex gap-6 items-center">
            {renderPage(currentSpreadSlide.left, 'left')}
            {renderPage(currentSpreadSlide.right, 'right')}
          </div>
        ) : currentSingleSlide && imageUrl ? (
          <img
            src={`${imageUrl}?t=${
              currentSingleSlide.asset?.uploaded_at
                ? new Date(currentSingleSlide.asset.uploaded_at).getTime()
                : Date.now()
            }`}
            alt={currentSingleSlide.asset?.filename || 'Slide'}
            className="max-w-full max-h-full object-contain"
          />
        ) : (
          <div className="text-white text-center">
            <p className="text-lg">Image not found</p>
          </div>
        )}

        {/* Next button */}
        {activeLength > 1 && (
          <button
            onClick={goToNext}
            className="absolute right-4 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white p-3 rounded-full transition-all z-10"
            aria-label="Next image"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>
        )}
      </div>

      {/* Bottom controls */}
      <div className="absolute bottom-0 left-0 right-0 z-10 bg-black/50 backdrop-blur-sm p-4">
        <div className="flex items-center justify-center gap-2">
          {activeLength > 1 && (
            <>
              <Button
                onClick={goToPrevious}
                variant="secondary"
                size="sm"
                className="bg-white/10 hover:bg-white/20 text-white border-white/20"
              >
                ← Previous
              </Button>
              <div className="flex gap-1">
                {Array.from({ length: activeLength }).map((_, idx) => (
                  <button
                    key={idx}
                    onClick={() => setCurrentIndex(idx)}
                    className={`w-2 h-2 rounded-full transition-all ${
                      idx === currentIndex
                        ? 'bg-white w-8'
                        : 'bg-white/40 hover:bg-white/60'
                    }`}
                    aria-label={`Go to slide ${idx + 1}`}
                  />
                ))}
              </div>
              <Button
                onClick={goToNext}
                variant="secondary"
                size="sm"
                className="bg-white/10 hover:bg-white/20 text-white border-white/20"
              >
                Next →
              </Button>
            </>
          )}
        </div>
        {viewMode === 'single' && currentSingleSlide?.asset?.filename && (
          <p className="text-white/70 text-xs text-center mt-2">
            {currentSingleSlide.asset.filename}
          </p>
        )}
      </div>
    </div>
  );
};

