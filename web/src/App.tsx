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

  // Asset State
  const [assetsData, setAssetsData] = useState<PaginatedAssets>({
    assets: [],
    total: 0,
    page: 1,
    limit: 50,
  });
  const [assetSearch, setAssetSearch] = useState('');

  // Active Terminal & Streaming State
  const [activeScanId, setActiveScanId] = useState<string | null>(null);
  const [activeScanLogs, setActiveScanLogs] = useState<LogEntry[]>([]);
  const [runningScanId, setRunningScanId] = useState<string | null>(null);
  const wsRef = useRef<ScanWebSocket | null>(null);

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
    try {
      const data = await assetService.listByTarget(activeTarget.id, page, 50, search);
      setAssetsData(data);
    } catch (err) {
      console.error('Failed to load assets', err);
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
            loadAssets(assetsData.page, assetSearch);
          }
          if (msg.status !== 'RUNNING') {
            setRunningScanId(null);
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
          line: `[BHF] Initiating ${toolName} execution against ${activeTarget.domain}`,
          timestamp: new Date().toISOString(),
        },
      ]);
      // Navigate cleanly to console so user can see it running without jumping
      setActiveTab('terminal');
    } catch (err: any) {
      alert(`Failed to launch tool: ${err.message}`);
    }
  };

  const handleStopScan = async (scanId: string) => {
    try {
      await scanService.stop(scanId);
      setRunningScanId(null);
    } catch (err: any) {
      alert(`Failed to stop scan: ${err.message}`);
    }
  };

  const handleViewLogs = (scanId: string) => {
    setActiveScanId(scanId);
    setActiveTab('terminal');
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
        onSelectTab={setActiveTab}
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
        }}
      />
    </div>
  );
};

export default App;
