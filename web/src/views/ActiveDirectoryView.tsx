import React, { useState } from 'react';
import {
  ShieldCheck,
  RefreshCw,
  Server,
  ArrowRight,
  AlertTriangle,
} from 'lucide-react';
import { Target } from '../types/target';
import { Host, NetworkService, ADAssessmentResult } from '../types/network';
import { networkService } from '../services/networkService';
import { Button } from '../components/Button';

interface ActiveDirectoryViewProps {
  activeTarget: Target | null;
  hosts: Host[];
  services: NetworkService[];
  onRefresh: () => void;
  onNavigateToFindings: () => void;
  isLoading?: boolean;
}

export const ActiveDirectoryView: React.FC<ActiveDirectoryViewProps> = ({
  activeTarget,
  hosts,
  services,
  onRefresh,
  onNavigateToFindings,
  isLoading = false,
}) => {
  const [isAuditing, setIsAuditing] = useState(false);
  const [adResults, setAdResults] = useState<ADAssessmentResult[]>([]);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  const adServices = services.filter((s) => [88, 389, 445, 636, 3268].includes(s.port));

  const handleRunAudit = async () => {
    if (!activeTarget) return;
    setIsAuditing(true);
    setStatusMessage('Auditing Active Directory domain controllers, LDAP transport, and authentication policies...');
    try {
      const results = await networkService.runADAudit(activeTarget.id);
      setAdResults(results);
      setStatusMessage(`Audit complete: Evaluated ${results.length} hosts for Domain Controller and identity configuration.`);
      onRefresh();
    } catch (err: any) {
      setStatusMessage(`Audit notice: ${err.message || err}`);
    } finally {
      setIsAuditing(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* Active Directory Assessment Control Toolbar */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-4 space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-white/[0.04] pb-2.5">
          <div className="flex items-center space-x-2">
            <ShieldCheck className="w-4 h-4 text-purple-400" />
            <h2 className="text-xs font-semibold text-slate-200">
              Active Directory & Identity Security Auditor
            </h2>
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="primary"
              size="sm"
              onClick={handleRunAudit}
              disabled={isAuditing || !activeTarget}
              className="text-xs font-medium"
            >
              <RefreshCw className={`w-3.5 h-3.5 mr-1.5 ${isAuditing ? 'animate-spin' : ''}`} />
              <span>{isAuditing ? 'Auditing AD...' : 'Run AD Security Audit'}</span>
            </Button>
            <button
              onClick={onRefresh}
              disabled={isLoading}
              className="p-1.5 rounded-lg text-slate-400 hover:text-white transition-colors"
              title="Refresh AD inventory"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        <div className="text-xs text-slate-400 flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center space-x-2">
            <span>Scope: <strong className="text-slate-200">{activeTarget?.domain || 'None'}</strong></span>
            <span className="text-slate-600">•</span>
            <span>Hosts: <strong className="text-slate-200">{hosts.length}</strong></span>
            <span className="text-slate-600">•</span>
            <span>AD Ports: <strong className="text-purple-400">{adServices.length}</strong></span>
          </div>
          <div className="text-[11px] text-slate-500 font-mono">
            Checks: Kerberos (88), LDAP (389), LDAPS (636), SMB (445), GC (3268)
          </div>
        </div>

        {statusMessage && (
          <div className="text-xs font-mono text-slate-300 bg-white/[0.02] border border-white/[0.04] px-3 py-1.5 rounded-lg">
            {statusMessage}
          </div>
        )}
      </div>

      {/* Discovered Domain Controllers & Directory Hosts */}
      <div className="space-y-3">
        <div className="text-xs font-semibold text-slate-300 flex items-center justify-between">
          <span>Domain Controllers & Directory Infrastructure</span>
          {adResults.length > 0 && (
            <button
              onClick={onNavigateToFindings}
              className="text-[11px] text-indigo-400 hover:text-indigo-300 font-medium flex items-center space-x-1"
            >
              <span>View Findings in Vulnerability Tracker</span>
              <ArrowRight className="w-3 h-3" />
            </button>
          )}
        </div>

        {adResults.length === 0 ? (
          <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-8 text-center text-slate-500 text-xs space-y-2">
            <Server className="w-7 h-7 text-slate-600 mx-auto" />
            <div>No Active Directory audits recorded yet.</div>
            <div className="text-[11px] text-slate-600 font-mono">
              Scan a network containing domain controllers, then click &quot;Run AD Security Audit&quot;.
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {adResults.map((res, idx) => (
              <div
                key={res.host_ip || idx}
                className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-4 space-y-3"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <span className="w-2 h-2 rounded-full bg-purple-400" />
                    <span className="text-xs font-mono font-bold text-slate-100">{res.host_ip}</span>
                    {res.domain_name && (
                      <span className="text-xs text-slate-400 font-mono">({res.domain_name})</span>
                    )}
                  </div>
                  {res.is_domain_controller && (
                    <span className="px-2 py-0.5 rounded text-[10px] bg-purple-500/10 text-purple-300 border border-purple-500/20 font-medium">
                      Domain Controller
                    </span>
                  )}
                </div>

                {/* Open Services */}
                {res.open_services && res.open_services.length > 0 && (
                  <div>
                    <div className="text-[10px] uppercase font-semibold text-slate-500 mb-1">
                      Active AD Services
                    </div>
                    <div className="flex flex-wrap gap-1.5">
                      {res.open_services.map((svc, i) => (
                        <span
                          key={i}
                          className="px-2 py-0.5 rounded text-[11px] font-mono bg-white/[0.04] text-slate-300 border border-white/[0.06]"
                        >
                          {svc}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {/* Security Observations */}
                {res.observations && res.observations.length > 0 && (
                  <div className="border-t border-white/[0.04] pt-2 space-y-1.5">
                    <div className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider">
                      Security Observations
                    </div>
                    {res.observations.map((obs, i) => (
                      <div
                        key={i}
                        className="text-[11px] text-slate-300 flex items-start space-x-1.5 bg-white/[0.02] p-2 rounded-lg border border-white/[0.04]"
                      >
                        <AlertTriangle className="w-3.5 h-3.5 text-amber-400 shrink-0 mt-0.5" />
                        <span>{obs}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
