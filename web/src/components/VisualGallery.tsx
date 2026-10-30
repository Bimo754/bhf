import React, { useState, useMemo } from 'react';
import { ExternalLink, Copy, Check, Search, Camera, X, Play } from 'lucide-react';
import { Asset } from '../types/asset';

interface VisualGalleryProps {
  assets: Asset[];
  onLaunchScan?: (toolName: string, customArgs: string[]) => void;
}

export const VisualGallery: React.FC<VisualGalleryProps> = ({
  assets,
  onLaunchScan,
}) => {
  const [search, setSearch] = useState('');
  const [lightboxImage, setLightboxImage] = useState<{ src: string; title: string; url: string } | null>(null);
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null);

  // Filter assets to those with live URLs or screenshots
  const galleryAssets = useMemo(() => {
    return assets.filter((a) => {
      const liveURL = a.url || (a.status_code ? `https://${a.host}` : null);
      if (!liveURL) return false;

      if (!search.trim()) return true;
      const query = search.toLowerCase();
      return (
        a.host.toLowerCase().includes(query) ||
        (a.title && a.title.toLowerCase().includes(query)) ||
        (a.technologies && a.technologies.some((t) => t.toLowerCase().includes(query)))
      );
    });
  }, [assets, search]);

  const handleCopy = (url: string) => {
    navigator.clipboard.writeText(url);
    setCopiedUrl(url);
    setTimeout(() => setCopiedUrl(null), 2000);
  };

  const getScreenshotUrl = (asset: Asset): string | null => {
    const sTag = asset.technologies?.find((t) => t.startsWith('Screenshot:'));
    return sTag ? sTag.replace('Screenshot:', '') : null;
  };

  const getStatusPill = (code?: number) => {
    if (!code || code === 0) return null;
    let style = 'text-zinc-400 bg-zinc-800/80 border-zinc-700';
    if (code >= 200 && code < 300) style = 'text-emerald-400 bg-emerald-950/60 border-emerald-800/80';
    else if (code >= 300 && code < 400) style = 'text-blue-400 bg-blue-950/60 border-blue-800/80';
    else if (code >= 400 && code < 500) style = 'text-amber-400 bg-amber-950/60 border-amber-800/80';
    else if (code >= 500) style = 'text-rose-400 bg-rose-950/60 border-rose-800/80';

    return (
      <span className={`px-1.5 py-0.2 rounded border text-[10px] font-mono font-bold ${style}`}>
        {code}
      </span>
    );
  };

  return (
    <div className="bg-[#161b22] border border-[#30363d] rounded-md overflow-hidden flex flex-col font-mono">
      {/* Gallery Toolbar */}
      <div className="p-3 bg-[#161b22] border-b border-[#30363d] flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-2.5 top-2.5" />
          <input
            type="text"
            placeholder="Filter gallery by host, title, or technology..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-[#0d1117] border border-[#30363d] rounded pl-8 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-500"
          />
        </div>

        <div className="text-xs text-zinc-400 self-end sm:self-auto">
          Showing <span className="text-zinc-200 font-semibold">{galleryAssets.length}</span> live endpoints
        </div>
      </div>

      {/* Cards Grid */}
      <div className="p-4 overflow-y-auto max-h-[75vh]">
        {galleryAssets.length === 0 ? (
          <div className="py-16 text-center text-zinc-500 italic text-xs">
            {search
              ? 'No live services match the gallery filter.'
              : 'No live HTTP/S web endpoints discovered yet. Run Httpx or Gowitness to populate visual screenshots.'}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {galleryAssets.map((asset) => {
              const liveURL = asset.url || (asset.status_code ? `https://${asset.host}` : `http://${asset.host}`);
              const screenshot = getScreenshotUrl(asset);
              const cleanTechs = (asset.technologies || []).filter((t) => !t.startsWith('Screenshot:'));

              return (
                <div
                  key={asset.id}
                  className="bg-[#0d1117] border border-[#30363d] rounded-md overflow-hidden flex flex-col hover:border-zinc-500 transition group shadow-sm"
                >
                  {/* Card Header */}
                  <div className="p-2.5 border-b border-[#21262d] flex items-center justify-between gap-1.5 bg-[#161b22]">
                    <span className="text-xs font-semibold text-zinc-200 truncate" title={asset.host}>
                      {asset.host}
                    </span>
                    {getStatusPill(asset.status_code)}
                  </div>

                  {/* Thumbnail Container */}
                  <div
                    className="relative aspect-video bg-[#0d1117] flex items-center justify-center overflow-hidden border-b border-[#21262d] cursor-pointer"
                    onClick={() => {
                      if (screenshot) {
                        setLightboxImage({
                          src: screenshot,
                          title: asset.title || asset.host,
                          url: liveURL,
                        });
                      }
                    }}
                  >
                    {screenshot ? (
                      <img
                        src={screenshot}
                        alt={`Screenshot of ${asset.host}`}
                        className="w-full h-full object-cover object-top hover:scale-105 transition-transform duration-200"
                        loading="lazy"
                      />
                    ) : (
                      <div className="flex flex-col items-center justify-center p-4 text-center text-zinc-600">
                        <Camera className="w-6 h-6 mb-1 text-zinc-700" />
                        <span className="text-[10px]">No screenshot</span>
                      </div>
                    )}

                    {asset.priority_score && asset.priority_score >= 30 ? (
                      <span className="absolute top-1.5 left-1.5 px-1.5 py-0.2 rounded text-[9px] font-bold bg-amber-950/80 text-amber-300 border border-amber-700/80 shadow">
                        ★ {asset.priority_score}
                      </span>
                    ) : null}
                  </div>

                  {/* Card Body */}
                  <div className="p-2.5 flex-1 flex flex-col justify-between space-y-2">
                    <div>
                      <p className="text-[11px] text-zinc-300 truncate font-sans" title={asset.title || asset.host}>
                        {asset.title || <span className="text-zinc-600 italic">No Title</span>}
                      </p>

                      {cleanTechs.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1.5">
                          {cleanTechs.slice(0, 3).map((t, idx) => (
                            <span
                              key={idx}
                              className="px-1.5 py-0.2 rounded text-[9px] bg-[#161b22] border border-[#30363d] text-zinc-400"
                            >
                              {t}
                            </span>
                          ))}
                          {cleanTechs.length > 3 && (
                            <span className="px-1 py-0.2 rounded text-[9px] text-zinc-500">
                              +{cleanTechs.length - 3}
                            </span>
                          )}
                        </div>
                      )}
                    </div>

                    {/* Quick Action Toolbar */}
                    <div className="pt-2 border-t border-[#21262d] flex items-center justify-between text-xs text-zinc-400">
                      <div className="flex items-center space-x-1">
                        <button
                          type="button"
                          onClick={() => handleCopy(liveURL)}
                          className="p-1 rounded hover:bg-[#21262d] hover:text-zinc-200 transition"
                          title="Copy URL"
                        >
                          {copiedUrl === liveURL ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                        </button>
                        <a
                          href={liveURL}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="p-1 rounded hover:bg-[#21262d] hover:text-blue-400 transition"
                          title="Open in new tab"
                        >
                          <ExternalLink className="w-3.5 h-3.5" />
                        </a>
                      </div>

                      {onLaunchScan && (
                        <button
                          type="button"
                          onClick={() => onLaunchScan('nuclei', ['-u', liveURL])}
                          className="flex items-center space-x-1 px-1.5 py-0.5 rounded text-[10px] bg-[#21262d] border border-[#30363d] hover:border-zinc-500 text-zinc-300 hover:text-zinc-100 transition"
                          title="Scan this target with Nuclei"
                        >
                          <Play className="w-2.5 h-2.5 fill-current text-blue-400" />
                          <span>Scan</span>
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Lightbox Preview Modal */}
      {lightboxImage && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-sm p-4 animate-in fade-in duration-150">
          <div className="relative bg-[#161b22] border border-[#30363d] rounded-lg max-w-5xl w-full p-4 shadow-2xl flex flex-col space-y-3 font-mono">
            <div className="flex items-center justify-between border-b border-[#30363d] pb-2">
              <div className="flex items-center space-x-2 truncate">
                <Camera className="w-4 h-4 text-emerald-400 shrink-0" />
                <span className="text-xs font-semibold text-zinc-200 truncate">
                  {lightboxImage.title}
                </span>
                <a
                  href={lightboxImage.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs text-blue-400 hover:underline shrink-0"
                >
                  ({lightboxImage.url})
                </a>
              </div>
              <button
                onClick={() => setLightboxImage(null)}
                className="p-1 rounded text-zinc-400 hover:text-white hover:bg-[#21262d] transition"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className="overflow-auto max-h-[80vh] flex items-center justify-center bg-[#0d1117] rounded border border-[#30363d] p-2">
              <img
                src={lightboxImage.src}
                alt="Visual Reconnaissance Fullscreen"
                className="max-w-full h-auto object-contain rounded"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
