import React, { useState, useEffect, useRef } from 'react';
import { Target } from './types/target';
import { Scan } from './types/scan';
import { ToolInfo } from './types/tool';
import { PaginatedAssets } from './types/asset';
import { PaginatedFindings } from './types/finding';
import { LogEntry, StreamMessage } from './types/log';
import { targetService } from './services/targetService';
import { scanService } from './services/scanService';
import { toolService } from './services/toolService';
import { assetService } from './services/assetService';
import { findingService } from './services/findingService';
import { ScanWebSocket } from './services/wsService';
import { Navbar } from './components/Navbar';
import { CreateTargetModal } from './components/CreateTargetModal';
import { DeleteTargetModal } from './components/DeleteTargetModal';
import { ToastContainer, ToastMessage } from './components/Toast';
import { FilterChip } from './components/AssetTable';
import { ToolsView, ActiveQueueState } from './views/ToolsView';
import { AssetsView } from './views/AssetsView';
import { FindingsView } from './views/FindingsView';
import { TerminalView } from './views/TerminalView';

export const App: React.FC = () => {
  const [targets, setTargets] = useState<Target[]>([]);
  const [activeTarget, setActiveTarget] = useState<Target | null>(null);
  const [tools, setTools] = useState<ToolInfo[]>([]);
  const [scans, setScans] = useState<Scan[]>([]);
  const [activeTab, setActiveTab] = useState<'tools' | 'assets' | 'findings' | 'terminal'>('tools');
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

  // When active target changes, refresh its scans, assets, and findings
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

  const handleSelectTab = (tab: 'tools' | 'assets' | 'findings' | 'terminal') => {
    setActiveTab(tab);
    if (tab === 'assets') {
      loadAssets(1, assetSearch, assetFilter, assetSort, assetOrder);
    } else if (tab === 'findings') {
      loadFindings(1, findingSearch, findingSeverity);
    }
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
      const current = scansRef.current.find((s) => s.id === activeScanId);
      if (current && current.status !== 'RUNNING') {
        setRunningScanId(null);
      }
    }).catch(console.error);

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
                message: 'Findings saved. Explore updated inventory in Discovered Assets.',
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
    setActiveScanId(null);
    addToast({
      type: 'info',
      title: 'Target Deleted',
      message: `Removed ${target.domain} and its associated asset records.`,
    });
  };

  return (
    <div className="min-h-screen bg-[#0d1117] flex flex-col font-sans text-zinc-100">
      <Navbar
        targets={targets}
        activeTarget={activeTarget}
        activeTab={activeTab}
        findingsCount={findingsData.stats?.total || findingsData.total || 0}
        isRunning={!!runningScanId}
        onSelectTarget={setActiveTarget}
        onOpenCreateTarget={() => setIsCreateModalOpen(true)}
        onOpenDeleteTarget={() => setIsDeleteModalOpen(true)}
        onSelectTab={handleSelectTab}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6">
        {activeTab === 'tools' && (
          <ToolsView
            tools={tools}
            scans={scans}
            activeTarget={activeTarget}
            currentRunningScanId={runningScanId}
            activeQueue={activeQueue}
            onExecuteTool={handleExecuteTool}
            onExecuteQueue={handleExecuteQueue}
            onStopQueue={handleStopQueue}
            onViewLogs={handleViewLogs}
            onStopScan={handleStopScan}
          />
        )}

        {activeTab === 'assets' && (
          <AssetsView
            activeTarget={activeTarget}
            assetsData={assetsData}
            currentPage={assetsData.page}
            searchQuery={assetSearch}
            activeFilter={assetFilter}
            sortField={assetSort}
            sortOrder={assetOrder}
            onPageChange={(p) => loadAssets(p, assetSearch, assetFilter, assetSort, assetOrder)}
            onSearchChange={(q) => {
              setAssetSearch(q);
              loadAssets(1, q, assetFilter, assetSort, assetOrder);
            }}
            onFilterChange={handleAssetFilterChange}
            onSortChange={handleAssetSortChange}
            onRefresh={() => loadAssets(assetsData.page, assetSearch, assetFilter, assetSort, assetOrder)}
            onLaunchScan={handleExecuteTool}
            isLoading={isAssetsLoading}
          />
        )}

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

        {activeTab === 'terminal' && (
          <TerminalView
            scans={scans}
            activeScanId={activeScanId}
            onSelectScan={setActiveScanId}
            logs={activeScanLogs}
            isRunning={!!runningScanId}
            onStop={() => runningScanId && handleStopScan(runningScanId)}
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
