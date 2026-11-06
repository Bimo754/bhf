import React, { useState, useEffect, useRef } from 'react';
import { Target } from './types/target';
import { Scan } from './types/scan';
import { ToolInfo } from './types/tool';
import { PaginatedAssets } from './types/asset';
import { PaginatedFindings } from './types/finding';
import { PaginatedTraffic, TrafficStats } from './types/traffic';
import { LogEntry, StreamMessage } from './types/log';
import { Host, NetworkService as NetService } from './types/network';
import { targetService } from './services/targetService';
import { scanService } from './services/scanService';
import { toolService } from './services/toolService';
import { assetService } from './services/assetService';
import { findingService } from './services/findingService';
import { trafficService } from './services/trafficService';
import { networkService } from './services/networkService';
import { ScanWebSocket } from './services/wsService';
import { Navbar, MainTab } from './components/Navbar';
import { CreateTargetModal } from './components/CreateTargetModal';
import { DeleteTargetModal } from './components/DeleteTargetModal';
import { ToastContainer, ToastMessage } from './components/Toast';
import { FilterChip } from './components/AssetTable';
import { ActiveQueueState } from './views/ToolsView';
import { WebSecurityView, WebSecuritySubTab } from './views/WebSecurityView';
import { NetworkSecurityView } from './views/NetworkSecurityView';
import { ActiveDirectoryView } from './views/ActiveDirectoryView';
import { FindingsView } from './views/FindingsView';
import { TerminalView } from './views/TerminalView';

export const App: React.FC = () => {
  const [targets, setTargets] = useState<Target[]>([]);
  const [activeTarget, setActiveTarget] = useState<Target | null>(null);
  const [tools, setTools] = useState<ToolInfo[]>([]);
  const [scans, setScans] = useState<Scan[]>([]);
  const [activeTab, setActiveTab] = useState<MainTab>('web');
  const [webSubTab, setWebSubTab] = useState<WebSecuritySubTab>('methodology');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  // Asset State
  const [assetsData, setAssetsData] = useState<PaginatedAssets>({
    assets: [],
    total: 0,
    page: 1,
    limit: 50,
  });
  const [assetSearch, setAssetSearch] = useState('');
  const [assetFilter, setAssetFilter] = useState<FilterChip>('all');
  const [assetSort, setAssetSort] = useState<string>('priority');
  const [assetOrder, setAssetOrder] = useState<'asc' | 'desc'>('desc');
  const [isAssetsLoading, setIsAssetsLoading] = useState(false);

  // Traffic / App Intel State
  const [trafficData, setTrafficData] = useState<PaginatedTraffic>({
    traffic: [],
    total: 0,
    page: 1,
    limit: 50,
  });
  const [trafficStats, setTrafficStats] = useState<TrafficStats | null>(null);
  const [trafficSearch, setTrafficSearch] = useState('');
  const [trafficCategory, setTrafficCategory] = useState('');
  const [trafficMethod, setTrafficMethod] = useState('');
  const [isTrafficLoading, setIsTrafficLoading] = useState(false);

  // Network State
  const [networkHosts, setNetworkHosts] = useState<Host[]>([]);
  const [networkServices, setNetworkServices] = useState<NetService[]>([]);
  const [isNetworkLoading, setIsNetworkLoading] = useState(false);

  // Finding State
  const [findingsData, setFindingsData] = useState<PaginatedFindings>({
    findings: [],
    total: 0,
    page: 1,
    limit: 50,
    stats: { total: 0, critical: 0, high: 0, medium: 0, low: 0, info: 0 },
  });
  const [findingSearch, setFindingSearch] = useState('');
  const [findingSeverity, setFindingSeverity] = useState('all');
  const [isFindingsLoading, setIsFindingsLoading] = useState(false);

  // Active Terminal & Streaming State
  const [activeScanId, setActiveScanId] = useState<string | null>(null);
  const [activeScanLogs, setActiveScanLogs] = useState<LogEntry[]>([]);
  const [runningScanId, setRunningScanId] = useState<string | null>(null);
  const wsRef = useRef<ScanWebSocket | null>(null);

  // Scan Queue State (Sequential Pipeline Runner)
  const [activeQueue, setActiveQueue] = useState<ActiveQueueState | null>(null);
  const activeQueueRef = useRef<ActiveQueueState | null>(null);
  activeQueueRef.current = activeQueue;
  const skippedScanIdsRef = useRef<Set<string>>(new Set());

  const addToast = (toast: Omit<ToastMessage, 'id'>) => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev, { ...toast, id }]);
  };

  const handleDismissToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  // Initial load
  useEffect(() => {
    targetService.list().then((list) => {
      setTargets(list);
      if (list.length > 0) {
        setActiveTarget(list[0]);
      }
    }).catch(console.error);

    toolService.list().then(setTools).catch(console.error);
  }, []);

  // When active target changes, refresh its scans, assets, traffic, network, and findings
  useEffect(() => {
    if (!activeTarget) return;

    scanService.listByTarget(activeTarget.id).then((scs) => {
      setScans(scs);
      if (scs.length > 0) {
        if (!activeScanId || !scs.some((s) => s.id === activeScanId)) {
          setActiveScanId(scs[0].id);
        }
      } else {
        setActiveScanId(null);
      }
    }).catch(console.error);

    loadAssets(1, assetSearch, assetFilter, assetSort, assetOrder);
    loadFindings(1, findingSearch, findingSeverity);
    loadTraffic(1, trafficSearch, trafficCategory, trafficMethod);
    loadNetwork();
  }, [activeTarget]);

  // Synchronize latest values to refs so WebSocket stream events and callbacks
  // do not recreate connections or cause infinite state loops.
  const scansRef = useRef<Scan[]>(scans);
  scansRef.current = scans;

  const activeTargetRef = useRef<Target | null>(activeTarget);
  activeTargetRef.current = activeTarget;

  const assetSearchRef = useRef(assetSearch);
  assetSearchRef.current = assetSearch;

  const assetFilterRef = useRef<FilterChip>(assetFilter);
  assetFilterRef.current = assetFilter;

  const assetSortRef = useRef<string>(assetSort);
  assetSortRef.current = assetSort;

  const assetOrderRef = useRef<'asc' | 'desc'>(assetOrder);
  assetOrderRef.current = assetOrder;

  const findingSearchRef = useRef(findingSearch);
  findingSearchRef.current = findingSearch;

  const findingSeverityRef = useRef(findingSeverity);
  findingSeverityRef.current = findingSeverity;

  const loadAssets = async (
    page: number,
    search: string,
    filter = assetFilterRef.current,
    sort = assetSortRef.current,
    order = assetOrderRef.current
  ) => {
    const target = activeTargetRef.current;
    if (!target) return;
    setIsAssetsLoading(true);
    try {
      const data = await assetService.listByTarget(target.id, page, 50, search, filter, sort, order);
      setAssetsData(data);
    } catch (err) {
      console.error('Failed to load assets', err);
    } finally {
      setIsAssetsLoading(false);
    }
  };

  const handleAssetFilterChange = (filter: FilterChip) => {
    setAssetFilter(filter);
    loadAssets(1, assetSearch, filter, assetSort, assetOrder);
  };

  const handleAssetSortChange = (field: string) => {
    let nextOrder: 'asc' | 'desc' = 'desc';
    if (assetSort === field) {
      nextOrder = assetOrder === 'desc' ? 'asc' : 'desc';
    }
    setAssetSort(field);
    setAssetOrder(nextOrder);
    loadAssets(1, assetSearch, assetFilter, field, nextOrder);
  };

  const loadFindings = async (page: number, search: string, severity: string) => {
    const target = activeTargetRef.current;
    if (!target) return;
    setIsFindingsLoading(true);
    try {
      const data = await findingService.listByTarget(target.id, page, 50, search, severity);
      setFindingsData(data);
    } catch (err) {
      console.error('Failed to load findings', err);
    } finally {
      setIsFindingsLoading(false);
    }
  };

  const loadTraffic = async (
    page: number,
    search = trafficSearch,
    category = trafficCategory,
    method = trafficMethod
  ) => {
    const target = activeTargetRef.current;
    if (!target) return;
    setIsTrafficLoading(true);
    try {
      const [data, stats] = await Promise.all([
        trafficService.listByTarget(target.id, page, 50, category, method, search),
        trafficService.getStats(target.id).catch(() => null),
      ]);
      setTrafficData(data);
      if (stats) setTrafficStats(stats);
    } catch (err) {
      console.error('Failed to load traffic', err);
    } finally {
      setIsTrafficLoading(false);
    }
  };

  const loadNetwork = async () => {
    const target = activeTargetRef.current;
    if (!target) return;
    setIsNetworkLoading(true);
    try {
      const [hList, sList] = await Promise.all([
        networkService.listHosts(target.id).catch(() => []),
        networkService.listServices(target.id).catch(() => []),
      ]);
      setNetworkHosts(hList);
      setNetworkServices(sList);
    } catch (err) {
      console.error('Failed to load network entities', err);
    } finally {
      setIsNetworkLoading(false);
    }
  };

  const handleSelectTab = (tab: MainTab) => {
    setActiveTab(tab);
    if (tab === 'web') {
      if (webSubTab === 'assets') {
        loadAssets(1, assetSearch, assetFilter, assetSort, assetOrder);
      } else if (webSubTab === 'appintel') {
        loadTraffic(1, trafficSearch, trafficCategory, trafficMethod);
      }
    } else if (tab === 'network' || tab === 'ad') {
      loadNetwork();
    } else if (tab === 'findings') {
      loadFindings(1, findingSearch, findingSeverity);
    }
  };

  // Cross-discipline dynamic handoffs
  const handleHandoffToWebSecurity = (webHost: string, port: number) => {
    setActiveTab('web');
    setWebSubTab('methodology');
    addToast({
      type: 'info',
      title: 'Discipline Handoff: Web Security',
      message: `Focused on discovered web service at ${webHost}:${port}. Ready for HTTP reconnaissance.`,
    });
  };

  const handleHandoffToADSecurity = (adHost: string) => {
    setActiveTab('ad');
    addToast({
      type: 'info',
      title: 'Discipline Handoff: Active Directory',
      message: `Focused on discovered Domain Controller at ${adHost}. Ready for AD audit.`,
    });
  };

  // Connect WebSocket when activeScanId changes
  useEffect(() => {
    if (wsRef.current) {
      wsRef.current.disconnect();
      wsRef.current = null;
    }

    if (!activeScanId) {
      setActiveScanLogs([]);
      return;
    }

    // Load historical logs for the selected scan
    scanService.getLogs(activeScanId).then((loadedLogs) => {
      setActiveScanLogs(loadedLogs.length > 2000 ? loadedLogs.slice(-2000) : loadedLogs);
    }).catch(console.error);

    // Reconcile status directly from server to avoid race conditions with fast-finishing scans
    if (activeTargetRef.current) {
      scanService.listByTarget(activeTargetRef.current.id).then((freshScans) => {
        setScans(freshScans);
        scansRef.current = freshScans;
        const current = freshScans.find((s) => s.id === activeScanId);
        if (current && current.status !== 'RUNNING') {
          setRunningScanId(null);
        }
      }).catch(console.error);
    }

    // Stream real-time events
    const ws = new ScanWebSocket(
      activeScanId,
      (msg: StreamMessage) => {
        if (msg.type === 'log' && msg.line) {
          setActiveScanLogs((prev) => {
            const next = [
              ...prev,
              {
                scan_id: msg.scan_id,
                stream: msg.stream || 'stdout',
                line: msg.line!,
                timestamp: msg.timestamp,
              },
            ];
            return next.length > 2000 ? next.slice(-2000) : next;
          });
        } else if (msg.type === 'complete' || msg.type === 'status') {
          if (activeTargetRef.current) {
            scanService.listByTarget(activeTargetRef.current.id).then(setScans);
            loadAssets(1, assetSearchRef.current, assetFilterRef.current, assetSortRef.current, assetOrderRef.current);
            loadFindings(1, findingSearchRef.current, findingSeverityRef.current);
            loadNetwork();
          }
          if (msg.status !== 'RUNNING') {
            setRunningScanId(null);
          }
          if (msg.type === 'complete') {
            const scan = scansRef.current.find((s) => s.id === msg.scan_id);
            const toolTitle = scan ? scan.tool_name : 'Task';

            if (msg.status === 'COMPLETED') {
              addToast({
                type: 'success',
                title: `${toolTitle} Completed`,
                message: 'Findings saved. Explore updated inventory across security disciplines.',
              });

              // Check if an automated sequence is running
              const currentQ = activeQueueRef.current;
              if (currentQ) {
                const nextIndex = currentQ.currentIndex + 1;
                if (nextIndex < currentQ.tools.length) {
                  const nextTool = currentQ.tools[nextIndex];
                  const updatedQueue = {
                    ...currentQ,
                    currentIndex: nextIndex,
                  };
                  setActiveQueue(updatedQueue);
                  activeQueueRef.current = updatedQueue;

                  // Small delay to ensure DB writes are committed before next tool starts
                  setTimeout(() => {
                    handleExecuteTool(nextTool, []);
                  }, 600);
                } else {
                  addToast({
                    type: 'success',
                    title: `${currentQ.title} Finished`,
                    message: `All ${currentQ.tools.length} tools in sequence completed successfully.`,
                  });
                  setActiveQueue(null);
                  activeQueueRef.current = null;
                }
              }
            } else if (msg.status === 'FAILED') {
              if (skippedScanIdsRef.current.has(msg.scan_id)) {
                skippedScanIdsRef.current.delete(msg.scan_id);
                return;
              }
              addToast({
                type: 'error',
                title: `${toolTitle} Failed`,
                message: msg.line || 'Check console logs for details.',
              });
              if (activeQueueRef.current) {
                addToast({
                  type: 'error',
                  title: 'Pipeline Halted',
                  message: `Sequence "${activeQueueRef.current.title}" halted because ${toolTitle} failed.`,
                });
                setActiveQueue(null);
                activeQueueRef.current = null;
              }
            } else if (msg.status === 'STOPPED') {
              if (skippedScanIdsRef.current.has(msg.scan_id)) {
                skippedScanIdsRef.current.delete(msg.scan_id);
                return;
              }
              addToast({
                type: 'info',
                title: `${toolTitle} Stopped`,
                message: 'Process stopped by user. Existing findings preserved.',
              });
              if (activeQueueRef.current) {
                addToast({
                  type: 'info',
                  title: 'Pipeline Stopped',
                  message: `Sequence "${activeQueueRef.current.title}" stopped by operator.`,
                });
                setActiveQueue(null);
                activeQueueRef.current = null;
              }
            }
          }
        }
      }
    );

    ws.connect();
    wsRef.current = ws;

    return () => {
      ws.disconnect();
    };
  }, [activeScanId]);

  // Periodic safety reconciliation while a scan is running, preventing UI desync if WS drops
  useEffect(() => {
    if (!runningScanId || !activeTarget) return;

    const interval = setInterval(() => {
      scanService.listByTarget(activeTarget.id).then((freshScans) => {
        setScans(freshScans);
        scansRef.current = freshScans;
        const current = freshScans.find((s) => s.id === runningScanId);
        if (current && current.status !== 'RUNNING') {
          setRunningScanId(null);
        }
      }).catch(console.error);
    }, 4000);

    return () => clearInterval(interval);
  }, [runningScanId, activeTarget]);

  const handleExecuteTool = async (toolName: string, customArgs: string[]) => {
    if (!activeTarget) return;

    try {
      const scan = await scanService.start({
        target_id: activeTarget.id,
        tool_name: toolName,
        custom_args: customArgs.length > 0 ? customArgs : undefined,
      });

      setScans((prev) => [scan, ...prev]);
      setActiveScanId(scan.id);
      setRunningScanId(scan.id);
      setActiveScanLogs([
        {
          scan_id: scan.id,
          stream: 'system',
          line: `[BHF] Launching ${toolName} against ${activeTarget.domain}`,
          timestamp: new Date().toISOString(),
        },
      ]);
      // Switch cleanly to console
      setActiveTab('terminal');
    } catch (err: any) {
      addToast({
        type: 'error',
        title: 'Launch Failed',
        message: err.message,
      });
    }
  };

  const handleExecuteQueue = async (toolNames: string[], title: string) => {
    if (!activeTarget || toolNames.length === 0) return;
    if (runningScanId) {
      addToast({
        type: 'error',
        title: 'Scan in Progress',
        message: 'Please wait for the current scan to finish or stop it before starting a sequence.',
      });
      return;
    }

    const queue: ActiveQueueState = {
      title,
      tools: toolNames,
      currentIndex: 0,
    };
    setActiveQueue(queue);
    activeQueueRef.current = queue;

    addToast({
      type: 'info',
      title: `Starting ${title}`,
      message: `Executing ${toolNames.length} tools in sequence: ${toolNames.join(' -> ')}`,
    });

    await handleExecuteTool(toolNames[0], []);
  };

  const handleStopQueue = async () => {
    setActiveQueue(null);
    activeQueueRef.current = null;
    if (runningScanId) {
      await handleStopScan(runningScanId);
    }
    addToast({
      type: 'info',
      title: 'Sequence Aborted',
      message: 'Automated scan sequence stopped by operator.',
    });
  };

  const handleSkipCurrentTool = async () => {
    const currentQ = activeQueueRef.current;
    if (!currentQ) return;

    const currentTool = currentQ.tools[currentQ.currentIndex];
    const currentScanId = runningScanId;

    if (currentScanId) {
      skippedScanIdsRef.current.add(currentScanId);
      try {
        await scanService.stop(currentScanId);
      } catch (err) {
        console.warn('Failed to stop scan during skip:', err);
      }
      setRunningScanId(null);
    }

    const nextIndex = currentQ.currentIndex + 1;
    if (nextIndex < currentQ.tools.length) {
      const nextTool = currentQ.tools[nextIndex];
      const updatedQueue: ActiveQueueState = {
        ...currentQ,
        currentIndex: nextIndex,
      };
      setActiveQueue(updatedQueue);
      activeQueueRef.current = updatedQueue;

      addToast({
        type: 'info',
        title: `Skipped ${currentTool}`,
        message: `Advancing to ${nextTool} (${nextIndex + 1}/${currentQ.tools.length})`,
      });

      setTimeout(() => {
        handleExecuteTool(nextTool, []);
      }, 400);
    } else {
      addToast({
        type: 'success',
        title: `${currentQ.title} Finished`,
        message: `Pipeline completed after skipping ${currentTool}.`,
      });
      setActiveQueue(null);
      activeQueueRef.current = null;
    }
  };

  const handleStopScan = async (scanId: string) => {
    try {
      await scanService.stop(scanId);
      setRunningScanId(null);
    } catch (err: any) {
      addToast({
        type: 'error',
        title: 'Stop Failed',
        message: err.message,
      });
    }
  };

  const handleViewLogs = (scanId: string) => {
    setActiveScanId(scanId);
    setActiveTab('terminal');
  };

  const handleDeleteTarget = async (target: Target) => {
    await targetService.delete(target.id);
    const remaining = targets.filter((t) => t.id !== target.id);
    setTargets(remaining);
    setActiveTarget(remaining.length > 0 ? remaining[0] : null);
    setScans([]);
    setAssetsData({ assets: [], total: 0, page: 1, limit: 50 });
    setAssetFilter('all');
    setAssetSort('priority');
    setAssetOrder('desc');
    setNetworkHosts([]);
    setNetworkServices([]);
    setActiveScanId(null);
    addToast({
      type: 'info',
      title: 'Target Deleted',
      message: `Removed ${target.domain} and its associated asset records.`,
    });
  };

  return (
    <div className="min-h-screen bg-[#080c14] flex flex-col font-sans text-slate-100 antialiased">
      <Navbar
        targets={targets}
        activeTarget={activeTarget}
        activeTab={activeTab}
        webCount={assetsData.total}
        networkCount={networkHosts.length}
        adCount={networkServices.filter((s) => [88, 389, 445, 636, 3268].includes(s.port)).length}
        findingsCount={findingsData.stats?.total || findingsData.total || 0}
        isRunning={!!runningScanId}
        onSelectTarget={setActiveTarget}
        onOpenCreateTarget={() => setIsCreateModalOpen(true)}
        onOpenDeleteTarget={() => setIsDeleteModalOpen(true)}
        onSelectTab={handleSelectTab}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6">
        {/* 1. Web Security Discipline */}
        {activeTab === 'web' && (
          <WebSecurityView
            activeSubTab={webSubTab}
            onSelectSubTab={(sub) => {
              setWebSubTab(sub);
              if (sub === 'assets') {
                loadAssets(1, assetSearch, assetFilter, assetSort, assetOrder);
              } else if (sub === 'appintel') {
                loadTraffic(1, trafficSearch, trafficCategory, trafficMethod);
              }
            }}
            activeTarget={activeTarget}
            tools={tools}
            scans={scans}
            currentRunningScanId={runningScanId}
            activeQueue={activeQueue}
            onExecuteTool={handleExecuteTool}
            onExecuteQueue={handleExecuteQueue}
            onSkipTool={handleSkipCurrentTool}
            onStopQueue={handleStopQueue}
            onViewLogs={handleViewLogs}
            onStopScan={handleStopScan}
            assetsData={assetsData}
            assetSearch={assetSearch}
            assetFilter={assetFilter}
            assetSort={assetSort}
            assetOrder={assetOrder}
            isAssetsLoading={isAssetsLoading}
            onPageChangeAssets={(p) => loadAssets(p, assetSearch, assetFilter, assetSort, assetOrder)}
            onSearchChangeAssets={(q) => {
              setAssetSearch(q);
              loadAssets(1, q, assetFilter, assetSort, assetOrder);
            }}
            onFilterChangeAssets={handleAssetFilterChange}
            onSortChangeAssets={handleAssetSortChange}
            onRefreshAssets={() => loadAssets(assetsData.page, assetSearch, assetFilter, assetSort, assetOrder)}
            trafficData={trafficData}
            trafficStats={trafficStats}
            trafficSearch={trafficSearch}
            trafficCategory={trafficCategory}
            trafficMethod={trafficMethod}
            isTrafficLoading={isTrafficLoading}
            onPageChangeTraffic={(p) => loadTraffic(p, trafficSearch, trafficCategory, trafficMethod)}
            onSearchChangeTraffic={(q) => {
              setTrafficSearch(q);
              loadTraffic(1, q, trafficCategory, trafficMethod);
            }}
            onCategoryFilterChangeTraffic={(cat) => {
              setTrafficCategory(cat);
              loadTraffic(1, trafficSearch, cat, trafficMethod);
            }}
            onMethodFilterChangeTraffic={(m) => {
              setTrafficMethod(m);
              loadTraffic(1, trafficSearch, trafficCategory, m);
            }}
            onRefreshTraffic={() => loadTraffic(trafficData.page, trafficSearch, trafficCategory, trafficMethod)}
            onAssessmentTriggered={() => loadFindings(1, findingSearch, findingSeverity)}
          />
        )}

        {/* 2. Network Security Discipline */}
        {activeTab === 'network' && (
          <NetworkSecurityView
            activeTarget={activeTarget}
            hosts={networkHosts}
            services={networkServices}
            onRefresh={loadNetwork}
            onHandoffToWebSecurity={handleHandoffToWebSecurity}
            onHandoffToADSecurity={handleHandoffToADSecurity}
            isLoading={isNetworkLoading}
          />
        )}

        {/* 3. Active Directory Discipline */}
        {activeTab === 'ad' && (
          <ActiveDirectoryView
            activeTarget={activeTarget}
            hosts={networkHosts}
            services={networkServices}
            onRefresh={loadNetwork}
            onNavigateToFindings={() => setActiveTab('findings')}
            isLoading={isNetworkLoading}
          />
        )}

        {/* 4. Consolidated Findings & Evidence */}
        {activeTab === 'findings' && (
          <FindingsView
            data={findingsData}
            currentPage={findingsData.page}
            searchQuery={findingSearch}
            severityFilter={findingSeverity}
            onPageChange={(p) => loadFindings(p, findingSearch, findingSeverity)}
            onSearchChange={(q) => {
              setFindingSearch(q);
              loadFindings(1, q, findingSeverity);
            }}
            onSeverityFilterChange={(sev) => {
              setFindingSeverity(sev);
              loadFindings(1, findingSearch, sev);
            }}
            onRefresh={() => loadFindings(findingsData.page, findingSearch, findingSeverity)}
            isLoading={isFindingsLoading}
          />
        )}

        {/* 5. Live Terminal & Console Stream */}
        {activeTab === 'terminal' && (
          <TerminalView
            scans={scans}
            activeScanId={activeScanId}
            onSelectScan={setActiveScanId}
            logs={activeScanLogs}
            isRunning={!!runningScanId}
            onStop={() => runningScanId && handleStopScan(runningScanId)}
            onSkipTool={activeQueue ? handleSkipCurrentTool : undefined}
            onClear={() => setActiveScanLogs([])}
          />
        )}
      </main>

      <CreateTargetModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onCreated={(target) => {
          setTargets((prev) => [target, ...prev]);
          setActiveTarget(target);
          addToast({
            type: 'success',
            title: 'Scope Registered',
            message: `Target ${target.domain} created and active.`,
          });
        }}
      />

      <DeleteTargetModal
        isOpen={isDeleteModalOpen}
        target={activeTarget}
        onClose={() => setIsDeleteModalOpen(false)}
        onConfirmDelete={handleDeleteTarget}
      />

      <ToastContainer toasts={toasts} onDismiss={handleDismissToast} />
    </div>
  );
};

export default App;
