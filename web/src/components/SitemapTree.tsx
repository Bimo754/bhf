import React, { useState, useEffect, useMemo } from 'react';
import { ChevronRight, ChevronDown, Folder, Globe, ExternalLink, Copy, Check, Search, RefreshCw } from 'lucide-react';
import { Endpoint } from '../types/endpoint';
import { Asset } from '../types/asset';
import { endpointService } from '../services/endpointService';

interface SitemapTreeProps {
  targetId: string;
  targetDomain: string;
  fallbackAssets?: Asset[];
}

interface TreeNode {
  name: string;
  fullPath: string;
  url?: string;
  statusCode?: number;
  method?: string;
  sourceTool?: string;
  children: Map<string, TreeNode>;
}

export const SitemapTree: React.FC<SitemapTreeProps> = ({
  targetId,
  targetDomain,
  fallbackAssets = [],
}) => {
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedHosts, setExpandedHosts] = useState<Record<string, boolean>>({});
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null);

  const fetchEndpoints = async () => {
    setIsLoading(true);
    try {
      const data = await endpointService.getAllForTarget(targetId);
      setEndpoints(data);
    } catch (err) {
      console.error('Failed to load sitemap endpoints', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (targetId) {
      fetchEndpoints();
    }
  }, [targetId]);

  // Combine dedicated endpoints with asset URLs if endpoints table is currently empty
  const allRoutes = useMemo(() => {
    if (endpoints.length > 0) {
      return endpoints;
    }
    // Fallback synthesis from discovered assets
    return fallbackAssets
      .filter((a) => a.url)
      .map((a) => {
        let path = '/';
        try {
          const u = new URL(a.url!);
          path = u.pathname + u.search;
        } catch {
          // ignore invalid
        }
        return {
          id: a.id,
          target_id: a.target_id,
          host: a.host,
          url: a.url!,
          path: path,
          method: 'GET',
          status_code: a.status_code,
          created_at: a.created_at,
        } as Endpoint;
      });
  }, [endpoints, fallbackAssets]);

  // Build hierarchical tree grouped by host
  const hostTrees = useMemo(() => {
    const trees: Record<string, TreeNode> = {};

    for (const ep of allRoutes) {
      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase();
        const matches =
          ep.path.toLowerCase().includes(query) ||
          ep.host.toLowerCase().includes(query) ||
          (ep.status_code && ep.status_code.toString().includes(query));
        if (!matches) continue;
      }

      if (!trees[ep.host]) {
        trees[ep.host] = {
          name: ep.host,
          fullPath: ep.host,
          children: new Map(),
        };
      }

      const root = trees[ep.host];
      const segments = ep.path.split('/').filter(Boolean);

      let current = root;
      let accumulated = '';

      if (segments.length === 0) {
        // Root path "/"
        const key = '/';
        if (!current.children.has(key)) {
          current.children.set(key, {
            name: '/',
            fullPath: '/',
            url: ep.url,
            statusCode: ep.status_code,
            method: ep.method || 'GET',
            sourceTool: ep.source_tool,
            children: new Map(),
          });
        }
      } else {
        for (let i = 0; i < segments.length; i++) {
          const seg = segments[i];
          accumulated += '/' + seg;
          if (!current.children.has(seg)) {
            const isLeaf = i === segments.length - 1;
            current.children.set(seg, {
              name: '/' + seg,
              fullPath: accumulated,
              url: isLeaf ? ep.url : undefined,
              statusCode: isLeaf ? ep.status_code : undefined,
              method: isLeaf ? ep.method || 'GET' : undefined,
              sourceTool: isLeaf ? ep.source_tool : undefined,
              children: new Map(),
            });
          }
          current = current.children.get(seg)!;
        }
      }
    }

    return trees;
  }, [allRoutes, searchQuery]);

  // Auto-expand hosts initially
  useEffect(() => {
    const keys = Object.keys(hostTrees);
    if (keys.length > 0 && Object.keys(expandedHosts).length === 0) {
      const initial: Record<string, boolean> = {};
      for (const k of keys) {
        initial[k] = true;
      }
      setExpandedHosts(initial);
    }
  }, [hostTrees]);

  const toggleHost = (host: string) => {
    setExpandedHosts((prev) => ({
      ...prev,
      [host]: !prev[host],
    }));
  };

  const handleCopy = (url: string) => {
    navigator.clipboard.writeText(url);
    setCopiedUrl(url);
    setTimeout(() => setCopiedUrl(null), 2000);
  };

  const getStatusBadge = (code?: number) => {
    if (!code || code === 0) return null;
    let color = 'text-zinc-400 bg-zinc-800/80 border-zinc-700';
    if (code >= 200 && code < 300) color = 'text-emerald-400 bg-emerald-950/60 border-emerald-800/80';
    else if (code >= 300 && code < 400) color = 'text-blue-400 bg-blue-950/60 border-blue-800/80';
    else if (code >= 400 && code < 500) color = 'text-amber-400 bg-amber-950/60 border-amber-800/80';
    else if (code >= 500) color = 'text-rose-400 bg-rose-950/60 border-rose-800/80';

    return (
      <span className={`px-1.5 py-0.2 rounded border text-[10px] font-mono font-bold ${color}`}>
        {code}
      </span>
    );
  };

  const renderNode = (node: TreeNode, depth = 1): React.ReactNode => {
    const children = Array.from(node.children.values());
    const isLeaf = children.length === 0;

    return (
      <div key={node.fullPath} className="font-mono text-xs select-none">
        <div
          className="flex items-center space-x-2 py-1 px-2 rounded hover:bg-[#21262d]/50 transition group"
          style={{ paddingLeft: `${depth * 14}px` }}
        >
          {isLeaf ? (
            <span className="w-3.5 h-3.5 text-zinc-600 block shrink-0 text-center">•</span>
          ) : (
            <Folder className="w-3.5 h-3.5 text-zinc-500 shrink-0" />
          )}

          <span className="text-zinc-300 group-hover:text-zinc-100 truncate">
            {node.name}
          </span>

          {node.method && (
            <span className="px-1 py-0.2 rounded bg-zinc-800 border border-zinc-700 text-zinc-400 text-[9px] uppercase font-bold shrink-0">
              {node.method}
            </span>
          )}

          {getStatusBadge(node.statusCode)}

          {node.sourceTool && (
            <span className="text-[10px] text-zinc-600 hidden md:inline shrink-0">
              [{node.sourceTool}]
            </span>
          )}

          {node.url && (
            <div className="ml-auto flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition shrink-0">
              <button
                type="button"
                onClick={() => handleCopy(node.url!)}
                title="Copy URL"
                className="p-1 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#30363d]"
              >
                {copiedUrl === node.url ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
              </button>
              <a
                href={node.url}
                target="_blank"
                rel="noopener noreferrer"
                title="Open in new tab"
                className="p-1 rounded text-zinc-400 hover:text-blue-400 hover:bg-[#30363d]"
              >
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          )}
        </div>

        {children.length > 0 && (
          <div className="border-l border-[#21262d] ml-4">
            {children.map((child) => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  const hostList = Object.keys(hostTrees);

  return (
    <div className="bg-[#161b22] border border-[#30363d] rounded-md overflow-hidden flex flex-col font-mono">
      {/* Sitemap Toolbar */}
      <div className="p-3 bg-[#161b22] border-b border-[#30363d] flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-2.5 top-2.5" />
          <input
            type="text"
            placeholder="Filter sitemap by route, extension (.json, .js), or status..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-[#0d1117] border border-[#30363d] rounded pl-8 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-500"
          />
        </div>

        <div className="flex items-center space-x-2 text-xs self-end sm:self-auto">
          <button
            onClick={fetchEndpoints}
            title="Refresh sitemap endpoints"
            className="flex items-center space-x-1 px-2.5 py-1.5 rounded bg-[#0d1117] border border-[#30363d] text-zinc-300 hover:text-zinc-100 hover:bg-[#21262d] transition"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin text-blue-400' : ''}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* Tree Content */}
      <div className="p-4 space-y-3 min-h-[400px] overflow-y-auto max-h-[70vh]">
        {hostList.length === 0 ? (
          <div className="py-12 text-center text-zinc-500 italic text-xs">
            {searchQuery
              ? 'No endpoints match the filter query.'
              : 'No web routes or endpoints cataloged for this target yet. Run Katana, Ffuf, or GAU to map routes.'}
          </div>
        ) : (
          hostList.map((host) => {
            const isExpanded = !!expandedHosts[host];
            const rootNode = hostTrees[host];
            const childCount = Array.from(rootNode.children.values()).length;

            return (
              <div
                key={host}
                className="border border-[#21262d] bg-[#0d1117]/80 rounded p-2"
              >
                {/* Host Node Header */}
                <button
                  type="button"
                  onClick={() => toggleHost(host)}
                  className="w-full flex items-center justify-between text-left py-1 px-1.5 rounded hover:bg-[#161b22] transition"
                >
                  <div className="flex items-center space-x-2 truncate">
                    {isExpanded ? (
                      <ChevronDown className="w-3.5 h-3.5 text-zinc-400 shrink-0" />
                    ) : (
                      <ChevronRight className="w-3.5 h-3.5 text-zinc-400 shrink-0" />
                    )}
                    <Globe className="w-3.5 h-3.5 text-blue-400 shrink-0" />
                    <span className="font-semibold text-zinc-200 text-xs truncate">
                      {host}
                    </span>
                  </div>

                  <span className="text-[11px] text-zinc-500 font-normal shrink-0">
                    {childCount} top route{childCount === 1 ? '' : 's'}
                  </span>
                </button>

                {/* Subdomain Route Children */}
                {isExpanded && (
                  <div className="mt-2 pl-2 border-l border-[#21262d] space-y-0.5">
                    {Array.from(rootNode.children.values()).map((child) =>
                      renderNode(child, 1)
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Footer Info */}
      <div className="p-2.5 bg-[#0d1117] border-t border-[#30363d] flex items-center justify-between text-[11px] text-zinc-500 font-mono">
        <span>
          Hierarchical Sitemap: <strong className="text-zinc-300">{hostList.length}</strong> hosts,{' '}
          <strong className="text-zinc-300">{allRoutes.length}</strong> endpoints mapped
        </span>
        <span className="text-zinc-600">Target: {targetDomain}</span>
      </div>
    </div>
  );
};
