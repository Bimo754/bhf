import React, { useState, useEffect, useRef } from 'react';
import { Target } from './types/target';
import { Scan } from './types/scan';
import { ToolInfo } from './types/tool';
import { PaginatedAssets } from './types/asset';
import { LogEntry, StreamMessage } from './types/log';
import { targetService } from './services/targetService';
import { scanService } from './services/scanService';
import { toolService } from './services/toolService';
import { assetService } from './services/assetService';
import { ScanWebSocket } from './services/wsService';
import { Navbar } from './components/Navbar';
import { CreateTargetModal } from './components/CreateTargetModal';
import { DeleteTargetModal } from './components/DeleteTargetModal';
import { ToastContainer, ToastMessage } from './components/Toast';
import { ToolsView } from './views/ToolsView';
import { AssetsView } from './views/AssetsView';
import { TerminalView } from './views/TerminalView';

export const App: React.FC = () => {
  const [targets, setTargets] = useState<Target[]>([]);
  const [activeTarget, setActiveTarget] = useState<Target | null>(null);
  const [tools, setTools] = useState<ToolInfo[]>([]);
  const [scans, setScans] = useState<Scan[]>([]);
  const [activeTab, setActiveTab] = useState<'tools' | 'assets' | 'terminal'>('tools');
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
  const [isAssetsLoading, setIsAssetsLoading] = useState(false);

  // Active Terminal & Streaming State
  const [activeScanId, setActiveScanId] = useState<string | null>(null);
  const [activeScanLogs, setActiveScanLogs] = useState<LogEntry[]>([]);
  const [runningScanId, setRunningScanId] = useState<string | null>(null);
  const wsRef = useRef<ScanWebSocket | null>(null);

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

  // When active target changes, refresh its scans and assets
  useEffect(() => {
    if (!activeTarget) return;

    scanService.listByTarget(activeTarget.id).then((scs) => {
      setScans(scs);
      if (scs.length > 0 && !activeScanId) {
        setActiveScanId(scs[0].id);
      }
    }).catch(console.error);

    loadAssets(1, assetSearch);
  }, [activeTarget]);

  const loadAssets = async (page: number, search: string) => {
    if (!activeTarget) return;
    setIsAssetsLoading(true);
    try {
      const data = await assetService.listByTarget(activeTarget.id, page, 50, search);
      setAssetsData(data);
    } catch (err) {
      console.error('Failed to load assets', err);
    } finally {
      setIsAssetsLoading(false);
    }
  };

  const handleSelectTab = (tab: 'tools' | 'assets' | 'terminal') => {
    setActiveTab(tab);
    if (tab === 'assets') {
      loadAssets(1, assetSearch);
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

    // Load historical logs
    scanService.getLogs(activeScanId).then(setActiveScanLogs).catch(console.error);

    // Stream real-time events
    const ws = new ScanWebSocket(
      activeScanId,
      (msg: StreamMessage) => {
        if (msg.type === 'log' && msg.line) {
          setActiveScanLogs((prev) => [
            ...prev,
            {
              scan_id: msg.scan_id,
              stream: msg.stream || 'stdout',
              line: msg.line!,
              timestamp: msg.timestamp,
            },
          ]);
        } else if (msg.type === 'complete' || msg.type === 'status') {
          if (activeTarget) {
            scanService.listByTarget(activeTarget.id).then(setScans);
            loadAssets(1, assetSearch);
          }
          if (msg.status !== 'RUNNING') {
            setRunningScanId(null);
          }
          if (msg.type === 'complete') {
            const scan = scans.find((s) => s.id === msg.scan_id);
            const toolTitle = scan ? scan.tool_name : 'Task';

            if (msg.status === 'COMPLETED') {
              addToast({
                type: 'success',
                title: `${toolTitle} Completed`,
                message: 'Findings saved. Explore updated inventory in Discovered Assets.',
              });
            } else if (msg.status === 'FAILED') {
              addToast({
                type: 'error',
                title: `${toolTitle} Failed`,
                message: msg.line || 'Check console logs for details.',
              });
            } else if (msg.status === 'STOPPED') {
              addToast({
                type: 'info',
                title: `${toolTitle} Stopped`,
                message: 'Process stopped by user. Existing findings preserved.',
              });
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
  }, [activeScanId, scans, activeTarget]);

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
            onExecuteTool={handleExecuteTool}
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
            onPageChange={(p) => loadAssets(p, assetSearch)}
            onSearchChange={(q) => {
              setAssetSearch(q);
              loadAssets(1, q);
            }}
            onRefresh={() => loadAssets(assetsData.page, assetSearch)}
            isLoading={isAssetsLoading}
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
