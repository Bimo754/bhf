import React, { useState } from 'react';
import {
  Server,
  Radio,
  RefreshCw,
  ArrowRight,
  Globe,
  Lock,
  Layers,
} from 'lucide-react';
import { Target } from '../types/target';
import { Host, NetworkService, NetworkScanResponse } from '../types/network';
import { networkService } from '../services/networkService';
import { Button } from '../components/Button';

interface NetworkSecurityViewProps {
  activeTarget: Target | null;
  hosts: Host[];
  services: NetworkService[];
  onRefresh: () => void;
  onHandoffToWebSecurity: (webHost: string, port: number) => void;
  onHandoffToADSecurity: (adHost: string) => void;
  isLoading?: boolean;
}

export const NetworkSecurityView: React.FC<NetworkSecurityViewProps> = ({
  activeTarget,
  hosts,
  services,
  onRefresh,
  onHandoffToWebSecurity,
  onHandoffToADSecurity,
  isLoading = false,
}) => {
  const [cidrInput, setCidrInput] = useState('');
  const [portProfile, setPortProfile] = useState<'core' | 'web' | 'ad' | 'full'>('core');
  const [isScanning, setIsScanning] = useState(false);
  const [scanResult, setScanResult] = useState<NetworkScanResponse | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  const getPortsForProfile = () => {
    switch (portProfile) {
      case 'web':
        return [80, 443, 8000, 8080, 8443, 8888, 9000, 9443];
      case 'ad':
        return [88, 389, 445, 636, 3268];
      case 'full':
        return [21, 22, 23, 25, 53, 80, 88, 110, 135, 139, 143, 389, 443, 445, 636, 1433, 3306, 3389, 5432, 6379, 8080, 8443, 9200, 27017];
      case 'core':
      default:
        return [21, 22, 23, 80, 88, 389, 443, 445, 3389, 6379, 8080, 8443];
    }
  };

  const handleLaunchScan = async () => {
    if (!activeTarget) return;
    setIsScanning(true);
    setStatusMessage('Scanning network range, probing ports, and fingerprinting services...');
    try {
      const ports = getPortsForProfile();
      const res = await networkService.scanSubnet(activeTarget.id, cidrInput || undefined, ports);
      setScanResult(res);
      setStatusMessage(`Scan complete: Discovered ${res.hosts_discovered} live hosts and ${res.services_open} open services.`);
      onRefresh();
    } catch (err: any) {
      setStatusMessage(`Scan notice: ${err.message || err}`);
    } finally {
      setIsScanning(false);
    }
  };

  // Identify web services for dynamic handoff
  const webServices = services.filter((s) =>
    ['http', 'https'].includes(s.name.toLowerCase()) || [80, 443, 8000, 8080, 8443, 8888, 9000].includes(s.port)
  );

  // Identify Active Directory services for dynamic handoff
  const adServices = services.filter((s) =>
    ['ldap', 'kerberos', 'smb'].includes(s.name.toLowerCase()) || [88, 389, 445, 636, 3268].includes(s.port)
  );

  return (
    <div className="space-y-4">
      {/* Unified Scanner Control Bar */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-4 space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-white/[0.04] pb-2.5">
          <div className="flex items-center space-x-2">
            <Server className="w-4 h-4 text-slate-400" />
            <h2 className="text-xs font-semibold text-slate-200">
              Network Discovery & Port Prober
            </h2>
          </div>
          <button
            onClick={onRefresh}
            disabled={isLoading}
            className="p-1 rounded text-slate-400 hover:text-white transition-colors self-end sm:self-auto"
            title="Refresh network inventory"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>

        {/* Search & Scan Unified Row */}
        <div className="flex flex-col md:flex-row items-center gap-2.5">
          <div className="flex-1 w-full">
            <input
              type="text"
              value={cidrInput}
              onChange={(e) => setCidrInput(e.target.value)}
              placeholder={`Target CIDR or IP (e.g. 192.168.1.0/24 or ${activeTarget?.domain || '10.0.0.1'})`}
              className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg px-3.5 py-1.5 text-xs font-mono text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div className="flex items-center space-x-2 w-full md:w-auto">
            {/* Integrated Profile Selector */}
            <select
              value={portProfile}
              onChange={(e) => setPortProfile(e.target.value as any)}
              className="w-full md:w-auto bg-white/[0.03] border border-white/[0.08] rounded-lg px-3 py-1.5 text-xs font-mono text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer"
            >
              <option value="core" className="bg-[#0e1420] text-slate-200">Profile: Core (12 ports)</option>
              <option value="web" className="bg-[#0e1420] text-slate-200">Profile: Web Ports</option>
              <option value="ad" className="bg-[#0e1420] text-slate-200">Profile: Active Directory</option>
              <option value="full" className="bg-[#0e1420] text-slate-200">Profile: Full (25 ports)</option>
            </select>

            <Button
              variant="primary"
              size="md"
              onClick={handleLaunchScan}
              disabled={isScanning || !activeTarget}
              className="text-xs font-medium shrink-0"
            >
              <Radio className={`w-3.5 h-3.5 mr-1.5 ${isScanning ? 'animate-pulse text-sky-400' : ''}`} />
              <span>{isScanning ? 'Probing...' : 'Scan Range'}</span>
            </Button>
          </div>
        </div>

        {statusMessage && (
          <div className="text-xs font-mono text-slate-300 bg-white/[0.02] border border-white/[0.04] px-3 py-1.5 rounded-lg flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <span>{statusMessage}</span>
            {scanResult && (
              <span className="text-[11px] text-slate-500 font-mono">
                Scan: {scanResult.scan_id.slice(0, 8)} • Discovered: {scanResult.hosts_discovered} hosts / {scanResult.services_open} services
              </span>
            )}
          </div>
        )}
      </div>

      {/* Sleek, Discreet Dynamic Handoff Banner (Only if actionable services found) */}
      {(webServices.length > 0 || adServices.length > 0) && (
        <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
          <div className="flex items-center space-x-3 text-slate-300">
            <span className="text-slate-400 font-medium">Discovered Attack Surfaces:</span>
            {webServices.length > 0 && (
              <span className="flex items-center space-x-1.5 text-slate-300">
                <Globe className="w-3.5 h-3.5 text-indigo-400" />
                <span>{webServices.length} Web Services</span>
              </span>
            )}
            {adServices.length > 0 && (
              <span className="flex items-center space-x-1.5 text-slate-300">
                <Lock className="w-3.5 h-3.5 text-purple-400" />
                <span>{adServices.length} AD Ports</span>
              </span>
            )}
          </div>

          <div className="flex items-center space-x-2 shrink-0">
            {webServices.length > 0 && (
              <Button
                variant="secondary"
                size="xs"
                onClick={() => onHandoffToWebSecurity(activeTarget?.domain || '127.0.0.1', webServices[0].port)}
                className="text-xs"
              >
                <span>Handoff to Web Recon</span>
                <ArrowRight className="w-3 h-3 ml-1" />
              </Button>
            )}
            {adServices.length > 0 && (
              <Button
                variant="secondary"
                size="xs"
                onClick={() => onHandoffToADSecurity(activeTarget?.domain || '127.0.0.1')}
                className="text-xs"
              >
                <span>Handoff to AD Audit</span>
                <ArrowRight className="w-3 h-3 ml-1" />
              </Button>
            )}
          </div>
        </div>
      )}

      {/* Discovered Hosts & Services Overview */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl overflow-hidden">
        <div className="px-4 py-3 border-b border-white/[0.04] flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Layers className="w-3.5 h-3.5 text-slate-400" />
            <h3 className="text-xs font-semibold text-slate-200">
              Discovered Network Inventory ({hosts.length} hosts, {services.length} services)
            </h3>
          </div>
        </div>

        {hosts.length === 0 ? (
          <div className="p-8 text-center text-xs text-slate-500 font-mono">
            No active network hosts discovered yet. Enter a CIDR subnet above and run a scan.
          </div>
        ) : (
          <div className="divide-y divide-white/[0.04]">
            {hosts.map((host) => {
              const hostServices = services.filter((s) => s.host_id === host.id);
              return (
                <div key={host.id} className="p-4 space-y-2 hover:bg-white/[0.01] transition-colors">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2.5">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                      <span className="font-mono text-xs font-semibold text-slate-200">{host.ip}</span>
                      {host.hostname && (
                        <span className="text-xs text-slate-400 font-mono">({host.hostname})</span>
                      )}
                    </div>
                    <span className="text-[11px] text-slate-500 font-mono">
                      {hostServices.length} {hostServices.length === 1 ? 'service' : 'services'}
                    </span>
                  </div>

                  {hostServices.length > 0 && (
                    <div className="flex flex-wrap gap-1.5 pt-1">
                      {hostServices.map((svc) => (
                        <span
                          key={svc.id}
                          className="px-2 py-0.5 rounded text-[11px] font-mono bg-white/[0.04] text-slate-300 border border-white/[0.06]"
                        >
                          <strong className="text-slate-100">{svc.port}</strong>/{svc.protocol || 'tcp'} {svc.name}
                          {svc.product ? ` (${svc.product})` : ''}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};
