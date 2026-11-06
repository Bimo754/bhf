import React, { useState, useEffect } from 'react';
import {
  Search,
  RefreshCw,
  Copy,
  Check,
  ChevronLeft,
  ChevronRight,
  Code,
  Shield,
  Globe,
  Compass,
} from 'lucide-react';
import { TrafficRecord, PaginatedTraffic, TrafficStats } from '../types/traffic';
import { trafficService, ExplorerProgress } from '../services/trafficService';
import { Button } from '../components/Button';

interface TrafficViewProps {
  targetId?: string;
  data: PaginatedTraffic;
  stats?: TrafficStats | null;
  currentPage: number;
  searchQuery: string;
  categoryFilter: string;
  methodFilter: string;
  onPageChange: (page: number) => void;
  onSearchChange: (search: string) => void;
  onCategoryFilterChange: (category: string) => void;
  onMethodFilterChange: (method: string) => void;
  onRefresh: () => void;
  onAssessmentTriggered?: () => void;
  isLoading?: boolean;
}

export const TrafficView: React.FC<TrafficViewProps> = ({
  targetId,
  data,
  stats,
  currentPage,
  searchQuery,
  categoryFilter,
  methodFilter,
  onPageChange,
  onSearchChange,
  onCategoryFilterChange,
  onMethodFilterChange,
  onRefresh,
  onAssessmentTriggered,
  isLoading = false,
}) => {
  const [localSearch, setLocalSearch] = useState(searchQuery);
  const [selectedRecord, setSelectedRecord] = useState<TrafficRecord | null>(null);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [detailTab, setDetailTab] = useState<'request' | 'response' | 'params' | 'raw'>('request');
  const [isAssessing, setIsAssessing] = useState(false);
  const [isExploring, setIsExploring] = useState(false);
  const [explorerProgress, setExplorerProgress] = useState<ExplorerProgress | null>(null);
  const [proxyRunning, setProxyRunning] = useState(false);
  const [proxyAddr, setProxyAddr] = useState('');
  const [actionMessage, setActionMessage] = useState<string | null>(null);

  const totalPages = Math.ceil(data.total / data.limit) || 1;

  useEffect(() => {
    trafficService.getProxyStatus().then((st) => {
      setProxyRunning(st.running);
      if (st.addr) setProxyAddr(st.addr);
    }).catch(() => {});
  }, []);

  const handleAutoExplore = async () => {
    if (!targetId) return;
    setIsExploring(true);
    setActionMessage('Autonomous Explorer running: Navigating target DOM, clicking interactive elements, and intercepting network traffic...');
    try {
      const res = await trafficService.autoExplore(targetId, {
        prefer_chrome: true,
        max_pages: 30,
        max_depth: 3,
      });
      setExplorerProgress(res);
      setActionMessage(
        `Auto-exploration complete (${res.engine_used}): Visited ${res.pages_visited} pages, interacted with ${res.elements_clicked} elements, captured ${res.traffic_captured} transactions.`
      );
      onRefresh();
    } catch (err: any) {
      setActionMessage(`Auto-exploration notice: ${err.message || err}`);
    } finally {
      setIsExploring(false);
    }
  };

  const handleRunAssessment = async () => {
    if (!targetId) return;
    setIsAssessing(true);
    setActionMessage(null);
    try {
      const res = await trafficService.runAssessment(targetId);
      setActionMessage(`Assessment finished: ${res.findings_count} findings identified across ${res.traffic_analyzed} transactions.`);
      if (onAssessmentTriggered) onAssessmentTriggered();
    } catch (err: any) {
      setActionMessage(`Assessment notice: ${err.message || err}`);
    } finally {
      setIsAssessing(false);
    }
  };

  const handleToggleProxy = async () => {
    if (!targetId) return;
    try {
      if (proxyRunning) {
        await trafficService.stopProxy();
        setProxyRunning(false);
        setProxyAddr('');
        setActionMessage('Interceptor proxy stopped.');
      } else {
        const st = await trafficService.startProxy(targetId, 8888);
        setProxyRunning(st.running);
        if (st.addr) setProxyAddr(st.addr);
        setActionMessage(`Interceptor proxy active on ${st.addr}. Route HTTP traffic to capture records.`);
      }
    } catch (err: any) {
      setActionMessage(`Proxy error: ${err.message || err}`);
    }
  };

  const handleLaunchBrowser = async () => {
    if (!targetId) return;
    try {
      await trafficService.launchBrowser(targetId);
      setActionMessage('Isolated browser session launched.');
    } catch (err: any) {
      setActionMessage(`Browser error: ${err.message || err}`);
    }
  };

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearchChange(localSearch);
  };

  const handleCopy = (key: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const getMethodBadge = (method: string) => {
    const m = method.toUpperCase();
    return (
      <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-white/[0.05] text-slate-300 border border-white/[0.08]">
        {m}
      </span>
    );
  };

  const getStatusBadge = (code: number) => {
    if (code >= 200 && code < 300) {
      return <span className="px-1.5 py-0.2 rounded text-[10px] font-mono font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">{code}</span>;
    }
    if (code >= 300 && code < 400) {
      return <span className="px-1.5 py-0.2 rounded text-[10px] font-mono font-medium bg-sky-500/10 text-sky-400 border border-sky-500/20">{code}</span>;
    }
    if (code >= 400 && code < 500) {
      return <span className="px-1.5 py-0.2 rounded text-[10px] font-mono font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">{code}</span>;
    }
    if (code >= 500) {
      return <span className="px-1.5 py-0.2 rounded text-[10px] font-mono font-medium bg-rose-500/10 text-rose-400 border border-rose-500/20">{code}</span>;
    }
    return <span className="px-1.5 py-0.2 rounded text-[10px] font-mono text-slate-500">{code || '-'}</span>;
  };

  const getCategoryBadge = (category: string) => {
    return (
      <span className="px-1.5 py-0.2 rounded text-[10px] font-mono bg-white/[0.04] text-slate-400 border border-white/[0.06]">
        {category.toLowerCase()}
      </span>
    );
  };

  const buildCurl = (rec: TrafficRecord): string => {
    let curl = `curl -i -X ${rec.method} '${rec.url}'`;
    try {
      if (rec.request_headers) {
        const headers = JSON.parse(rec.request_headers);
        for (const [key, val] of Object.entries(headers)) {
          if (Array.isArray(val)) {
            for (const v of val) {
              curl += ` \\\n  -H '${key}: ${v}'`;
            }
          } else {
            curl += ` \\\n  -H '${key}: ${val}'`;
          }
        }
      }
    } catch {
      // ignore
    }
    if (rec.request_body) {
      curl += ` \\\n  --data-raw '${rec.request_body.replace(/'/g, "\\'")}'`;
    }
    return curl;
  };

  const parseParameters = (paramStr: string) => {
    try {
      if (!paramStr || paramStr === '[]') return [];
      return JSON.parse(paramStr);
    } catch {
      return [];
    }
  };

  const parseHeaders = (headerStr: string): Record<string, string> => {
    try {
      if (!headerStr) return {};
      const parsed = JSON.parse(headerStr);
      const result: Record<string, string> = {};
      for (const [k, v] of Object.entries(parsed)) {
        result[k] = Array.isArray(v) ? v.join(', ') : String(v);
      }
      return result;
    } catch {
      return {};
    }
  };

  const formatBody = (body: string): string => {
    if (!body) return '';
    try {
      return JSON.stringify(JSON.parse(body), null, 2);
    } catch {
      return body;
    }
  };

  const categories = [
    { key: 'all', label: 'All Categories' },
    { key: 'auth', label: 'Auth' },
    { key: 'api', label: 'API' },
    { key: 'admin', label: 'Admin' },
    { key: 'upload', label: 'Upload' },
    { key: 'search', label: 'Search' },
    { key: 'ugc', label: 'UGC' },
    { key: 'general', label: 'General' },
  ];

  const methods = ['all', 'GET', 'POST', 'PUT', 'DELETE', 'PATCH'];

  return (
    <div className="space-y-4">
      {/* Sleek Control Toolbar */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <Button
            variant="primary"
            size="sm"
            onClick={handleAutoExplore}
            disabled={isExploring || !targetId}
            className="text-xs font-medium"
          >
            <Compass className={`w-3.5 h-3.5 mr-1.5 ${isExploring ? 'animate-spin' : ''}`} />
            <span>{isExploring ? 'Exploring Site...' : 'Auto-Explore Site'}</span>
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={handleRunAssessment}
            disabled={isAssessing || !targetId || data.total === 0}
            className="text-xs font-medium"
          >
            <Shield className={`w-3.5 h-3.5 mr-1.5 ${isAssessing ? 'animate-spin' : ''}`} />
            <span>{isAssessing ? 'Auditing...' : 'Run Security Assessment'}</span>
          </Button>

          <Button
            variant={proxyRunning ? 'danger' : 'ghost'}
            size="sm"
            onClick={handleToggleProxy}
            disabled={!targetId}
            className="text-xs font-mono"
          >
            <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${proxyRunning ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`} />
            <span>{proxyRunning ? `Proxy Active: ${proxyAddr}` : 'Proxy Listener'}</span>
          </Button>

          <Button
            variant="ghost"
            size="sm"
            onClick={handleLaunchBrowser}
            disabled={!targetId}
            className="text-xs text-slate-400 hover:text-white"
          >
            <Globe className="w-3.5 h-3.5 mr-1" />
            <span>Browser</span>
          </Button>
        </div>

        {actionMessage && (
          <div className="text-xs font-mono text-slate-300 bg-white/[0.02] border border-white/[0.04] px-3 py-1 rounded-lg">
            {actionMessage}
          </div>
        )}
      </div>

      {/* Autonomous Explorer Summary Banner */}
      {explorerProgress && (
        <div className="bg-[#0e1420] border border-emerald-500/20 rounded-xl p-3 text-xs font-mono flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center space-x-2 text-emerald-400">
            <Compass className="w-4 h-4 text-emerald-400" />
            <span className="font-semibold">Auto-Explorer:</span>
            <span className="text-slate-200">{explorerProgress.pages_visited} pages</span>
            <span className="text-slate-600">•</span>
            <span className="text-slate-200">{explorerProgress.elements_clicked} elements clicked</span>
            <span className="text-slate-600">•</span>
            <span className="text-slate-200">{explorerProgress.traffic_captured} transactions captured</span>
          </div>
          <Button
            variant="secondary"
            size="xs"
            onClick={handleRunAssessment}
            disabled={isAssessing}
            className="text-slate-200"
          >
            Audit Traffic Now &rarr;
          </Button>
        </div>
      )}

      {/* Category Stats Overview Bar */}
      {stats && stats.categories && Object.keys(stats.categories).length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-7 gap-2">
          {Object.entries(stats.categories).map(([cat, count]) => (
            <div
              key={cat}
              onClick={() => onCategoryFilterChange(categoryFilter === cat ? 'all' : cat)}
              className={`p-2.5 rounded-xl border cursor-pointer transition-all ${
                categoryFilter === cat
                  ? 'bg-white/[0.08] border-white/20'
                  : 'bg-[#0e1420] border-white/[0.05] hover:bg-white/[0.03]'
              }`}
            >
              <div className="text-[10px] font-mono text-slate-400 uppercase tracking-wider">{cat}</div>
              <div className="text-base font-mono font-bold text-slate-100">{count}</div>
            </div>
          ))}
        </div>
      )}

      {/* Filter and Search Bar */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-3 flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div className="flex items-center space-x-2 w-full md:w-auto">
          {/* Category Dropdown */}
          <select
            value={categoryFilter || 'all'}
            onChange={(e) => onCategoryFilterChange(e.target.value)}
            className="bg-white/[0.03] border border-white/[0.08] text-xs font-mono text-slate-200 rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-indigo-500 cursor-pointer"
          >
            {categories.map((c) => (
              <option key={c.key} value={c.key} className="bg-[#0e1420] text-slate-200">
                {c.label}
              </option>
            ))}
          </select>

          {/* Method Dropdown */}
          <select
            value={methodFilter || 'all'}
            onChange={(e) => onMethodFilterChange(e.target.value)}
            className="bg-white/[0.03] border border-white/[0.08] text-xs font-mono text-slate-200 rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-indigo-500 cursor-pointer"
          >
            {methods.map((m) => (
              <option key={m} value={m} className="bg-[#0e1420] text-slate-200">
                {m === 'all' ? 'All Methods' : m}
              </option>
            ))}
          </select>
        </div>

        <div className="flex items-center space-x-2 flex-1 md:max-w-xs">
          <form onSubmit={handleSearchSubmit} className="relative flex-1">
            <input
              type="text"
              value={localSearch}
              onChange={(e) => setLocalSearch(e.target.value)}
              placeholder="Search path, params, body..."
              className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg px-3 py-1.5 pl-8 text-xs font-mono text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500"
            />
            <Search className="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-2.5" />
          </form>

          <button
            onClick={onRefresh}
            disabled={isLoading}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Main Split Layout: Records List & Inspector */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        {/* Traffic Records Table */}
        <div className={`${selectedRecord ? 'lg:col-span-6' : 'lg:col-span-12'} transition-all`}>
          <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs font-mono">
                <thead className="bg-white/[0.02] border-b border-white/[0.04] text-slate-400 text-[10px] uppercase">
                  <tr>
                    <th className="px-3 py-2">Method</th>
                    <th className="px-3 py-2">Status</th>
                    <th className="px-3 py-2">Endpoint</th>
                    <th className="px-3 py-2">Category</th>
                    <th className="px-3 py-2 hidden sm:table-cell">Auth</th>
                    <th className="px-3 py-2 text-right hidden sm:table-cell">Time</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/[0.03]">
                  {data.traffic.length === 0 ? (
                    <tr>
                      <td colSpan={6} className="px-4 py-8 text-center text-slate-500">
                        {isLoading ? 'Loading intercepted traffic...' : 'No intercepted traffic recorded yet. Run Auto-Explore to begin.'}
                      </td>
                    </tr>
                  ) : (
                    data.traffic.map((rec) => {
                      const isSelected = selectedRecord?.id === rec.id;
                      return (
                        <tr
                          key={rec.id}
                          onClick={() => setSelectedRecord(rec)}
                          className={`cursor-pointer transition-colors ${
                            isSelected
                              ? 'bg-white/[0.08] text-white'
                              : 'hover:bg-white/[0.02] text-slate-300'
                          }`}
                        >
                          <td className="px-3 py-2.5 whitespace-nowrap">{getMethodBadge(rec.method)}</td>
                          <td className="px-3 py-2.5 whitespace-nowrap">{getStatusBadge(rec.response_status_code)}</td>
                          <td className="px-3 py-2.5 font-medium truncate max-w-[240px] sm:max-w-[320px]" title={rec.url}>
                            {rec.path || rec.url}
                          </td>
                          <td className="px-3 py-2.5 whitespace-nowrap">{getCategoryBadge(rec.category)}</td>
                          <td className="px-3 py-2.5 whitespace-nowrap hidden sm:table-cell text-slate-500 text-[11px]">
                            {rec.auth_context || 'public'}
                          </td>
                          <td className="px-3 py-2.5 whitespace-nowrap text-right hidden sm:table-cell text-slate-500 text-[11px]">
                            {rec.duration_ms}ms
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            <div className="px-4 py-3 bg-white/[0.01] border-t border-white/[0.04] flex items-center justify-between text-xs font-mono text-slate-400">
              <div>
                Total: <span className="text-slate-200 font-semibold">{data.total}</span> records
              </div>
              <div className="flex items-center space-x-2">
                <span>
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  onClick={() => onPageChange(currentPage - 1)}
                  disabled={currentPage <= 1 || isLoading}
                  className="p-1 rounded text-slate-400 hover:text-white disabled:opacity-30"
                >
                  <ChevronLeft className="w-3.5 h-3.5" />
                </button>
                <button
                  onClick={() => onPageChange(currentPage + 1)}
                  disabled={currentPage >= totalPages || isLoading}
                  className="p-1 rounded text-slate-400 hover:text-white disabled:opacity-30"
                >
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Traffic Record Inspector Panel */}
        {selectedRecord && (
          <div className="lg:col-span-6 space-y-3">
            <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-4 space-y-4">
              {/* Header with Title and Actions */}
              <div className="flex items-start justify-between gap-2 border-b border-white/[0.04] pb-3">
                <div className="space-y-1 min-w-0">
                  <div className="flex items-center space-x-2">
                    {getMethodBadge(selectedRecord.method)}
                    {getStatusBadge(selectedRecord.response_status_code)}
                    <span className="text-xs font-mono font-semibold text-slate-100 truncate">
                      {selectedRecord.path || selectedRecord.url}
                    </span>
                  </div>
                  <div className="text-[11px] font-mono text-slate-500 truncate">
                    {selectedRecord.url}
                  </div>
                </div>
                <div className="flex items-center space-x-1 shrink-0">
                  <Button
                    variant="ghost"
                    size="xs"
                    onClick={() => handleCopy('curl', buildCurl(selectedRecord))}
                    title="Copy cURL Reproduction Command"
                  >
                    {copiedKey === 'curl' ? (
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                    ) : (
                      <span className="flex items-center space-x-1 text-[11px] font-mono">
                        <Copy className="w-3 h-3" />
                        <span>cURL</span>
                      </span>
                    )}
                  </Button>
                  <button
                    onClick={() => setSelectedRecord(null)}
                    className="p-1 rounded text-slate-500 hover:text-slate-200"
                  >
                    &times;
                  </button>
                </div>
              </div>

              {/* Inspector Sub-Tabs */}
              <div className="flex items-center space-x-1 border-b border-white/[0.04] text-xs font-mono pb-2">
                <button
                  onClick={() => setDetailTab('request')}
                  className={`px-2.5 py-1 rounded-lg transition-colors ${
                    detailTab === 'request'
                      ? 'bg-white/[0.08] text-white font-medium'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Request
                </button>
                <button
                  onClick={() => setDetailTab('response')}
                  className={`px-2.5 py-1 rounded-lg transition-colors ${
                    detailTab === 'response'
                      ? 'bg-white/[0.08] text-white font-medium'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Response ({selectedRecord.response_status_code})
                </button>
                <button
                  onClick={() => setDetailTab('params')}
                  className={`px-2.5 py-1 rounded-lg transition-colors ${
                    detailTab === 'params'
                      ? 'bg-white/[0.08] text-white font-medium'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Parameters
                </button>
                <button
                  onClick={() => setDetailTab('raw')}
                  className={`px-2.5 py-1 rounded-lg transition-colors ${
                    detailTab === 'raw'
                      ? 'bg-white/[0.08] text-white font-medium'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  Metadata
                </button>
              </div>

              {/* Sub-Tab Contents */}
              {detailTab === 'request' && (
                <div className="space-y-3 text-xs font-mono">
                  <div>
                    <div className="text-[10px] uppercase text-slate-500 font-semibold mb-1">Request Headers</div>
                    <div className="bg-[#080c14] border border-white/[0.05] rounded-lg p-2.5 max-h-48 overflow-y-auto space-y-1">
                      {Object.entries(parseHeaders(selectedRecord.request_headers)).map(([k, v]) => (
                        <div key={k} className="flex">
                          <span className="text-slate-500 font-semibold mr-2">{k}:</span>
                          <span className="text-slate-300 break-all">{v}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {selectedRecord.request_body && (
                    <div>
                      <div className="flex items-center justify-between text-[10px] uppercase text-slate-500 font-semibold mb-1">
                        <span>Request Body</span>
                        <button
                          onClick={() => handleCopy('req-body', selectedRecord.request_body)}
                          className="text-slate-500 hover:text-slate-300"
                        >
                          {copiedKey === 'req-body' ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                        </button>
                      </div>
                      <pre className="bg-[#080c14] border border-white/[0.05] rounded-lg p-2.5 max-h-56 overflow-y-auto text-slate-300 whitespace-pre-wrap break-all text-[11px]">
                        {formatBody(selectedRecord.request_body)}
                      </pre>
                    </div>
                  )}
                </div>
              )}

              {detailTab === 'response' && (
                <div className="space-y-3 text-xs font-mono">
                  <div>
                    <div className="text-[10px] uppercase text-slate-500 font-semibold mb-1">Response Headers</div>
                    <div className="bg-[#080c14] border border-white/[0.05] rounded-lg p-2.5 max-h-48 overflow-y-auto space-y-1">
                      {Object.entries(parseHeaders(selectedRecord.response_headers)).map(([k, v]) => (
                        <div key={k} className="flex">
                          <span className="text-slate-500 font-semibold mr-2">{k}:</span>
                          <span className="text-slate-300 break-all">{v}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {selectedRecord.response_body && (
                    <div>
                      <div className="flex items-center justify-between text-[10px] uppercase text-slate-500 font-semibold mb-1">
                        <span>Response Body</span>
                        <button
                          onClick={() => handleCopy('resp-body', selectedRecord.response_body)}
                          className="text-slate-500 hover:text-slate-300"
                        >
                          {copiedKey === 'resp-body' ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                        </button>
                      </div>
                      <pre className="bg-[#080c14] border border-white/[0.05] rounded-lg p-2.5 max-h-56 overflow-y-auto text-slate-300 whitespace-pre-wrap break-all text-[11px]">
                        {formatBody(selectedRecord.response_body)}
                      </pre>
                    </div>
                  )}
                </div>
              )}

              {detailTab === 'params' && (
                <div className="space-y-3 text-xs font-mono">
                  <div className="text-[10px] uppercase text-slate-500 font-semibold mb-1">Detected Parameters</div>
                  {parseParameters(selectedRecord.parameters).length === 0 ? (
                    <div className="bg-[#080c14] border border-white/[0.05] rounded-lg p-4 text-center text-slate-500">
                      No parameters detected in query or payload.
                    </div>
                  ) : (
                    <div className="bg-[#080c14] border border-white/[0.05] rounded-lg divide-y divide-white/[0.04] overflow-hidden">
                      {parseParameters(selectedRecord.parameters).map((p: any, idx: number) => (
                        <div key={idx} className="p-2.5 flex items-center justify-between">
                          <div className="flex items-center space-x-2">
                            <Code className="w-3.5 h-3.5 text-slate-500" />
                            <span className="font-semibold text-slate-200">{p.name}</span>
                          </div>
                          <span className="px-1.5 py-0.5 rounded text-[10px] bg-white/[0.04] text-slate-400 border border-white/[0.06]">
                            {p.type || 'unknown'}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {detailTab === 'raw' && (
                <div className="space-y-2 text-xs font-mono">
                  <div className="text-[10px] uppercase text-slate-500 font-semibold mb-1">Transaction Metadata</div>
                  <div className="bg-[#080c14] border border-white/[0.05] rounded-lg p-3 space-y-1.5 text-slate-300">
                    <div className="flex justify-between">
                      <span className="text-slate-500">Host:</span>
                      <span>{selectedRecord.host}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Content-Type:</span>
                      <span>{selectedRecord.content_type || 'none'}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Content-Length:</span>
                      <span>{selectedRecord.content_length} bytes</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Duration:</span>
                      <span>{selectedRecord.duration_ms} ms</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Auth Context:</span>
                      <span>{selectedRecord.auth_context}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Captured:</span>
                      <span>{new Date(selectedRecord.created_at).toLocaleString()}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
