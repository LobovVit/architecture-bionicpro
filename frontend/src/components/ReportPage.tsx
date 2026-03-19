import React, { useEffect, useState } from 'react';

type MeResponse = { authenticated: boolean; username: string; roles: string[]; session_id: string; };
type ReportResponse = { owner: string; report_generated: string; prosthesis_ids: string[]; summary: { daily_movements: number; battery_cycles: number; calibration_state: string; }; security: { token_on_client: boolean; session_rotated: boolean; }; };
const API_URL = process.env.REACT_APP_AUTH_URL || 'http://localhost:8081';
const ReportPage: React.FC = () => {
  const [me, setMe] = useState<MeResponse | null>(null);
  const [report, setReport] = useState<ReportResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [reportLoading, setReportLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const loadSession = async () => {
    try {
      const response = await fetch(`${API_URL}/auth/me`, { credentials: 'include' });
      if (!response.ok) { setMe(null); return; }
      setMe(await response.json());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load session');
    } finally { setLoading(false); }
  };
  useEffect(() => { void loadSession(); }, []);
  const login = () => { window.location.href = `${API_URL}/auth/login`; };
  const logout = async () => { await fetch(`${API_URL}/auth/logout`, { method: 'POST', credentials: 'include' }); setMe(null); setReport(null); };
  const downloadReport = async () => {
    try {
      setReportLoading(true); setError(null);
      const response = await fetch(`${API_URL}/api/reports`, { credentials: 'include' });
      if (response.status === 401) { setMe(null); setError('Session expired. Please sign in again.'); return; }
      if (!response.ok) throw new Error(`Request failed with status ${response.status}`);
      setReport(await response.json());
      await loadSession();
    } catch (err) { setError(err instanceof Error ? err.message : 'Failed to load report'); } finally { setReportLoading(false); }
  };
  if (loading) return <div className="flex items-center justify-center min-h-screen bg-gray-100">Loading...</div>;
  if (!me) return <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 gap-4"><div className="p-8 bg-white rounded-lg shadow-md text-center"><h1 className="text-2xl font-bold mb-2">Usage Reports</h1><p className="text-gray-600 mb-6">Sign in to access your prosthesis usage report.</p><button onClick={login} className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">Sign in</button>{error && <div className="mt-4 p-4 bg-red-100 text-red-700 rounded">{error}</div>}</div></div>;
  return <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 p-6"><div className="w-full max-w-2xl p-8 bg-white rounded-lg shadow-md"><div className="flex items-start justify-between gap-4 mb-6"><div><h1 className="text-2xl font-bold">Usage Reports</h1><p className="text-gray-600 mt-1">Signed in as <span className="font-medium">{me.username}</span></p><p className="text-xs text-gray-500 mt-1">Session: {me.session_id}</p></div><button onClick={logout} className="px-3 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300">Logout</button></div><button onClick={downloadReport} disabled={reportLoading} className={`px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 ${reportLoading ? 'opacity-50 cursor-not-allowed' : ''}`}>{reportLoading ? 'Generating report...' : 'Download report'}</button>{error && <div className="mt-4 p-4 bg-red-100 text-red-700 rounded">{error}</div>}{report && <div className="mt-6 p-4 bg-gray-50 rounded border"><h2 className="text-lg font-semibold mb-3">Latest report</h2><pre className="text-sm whitespace-pre-wrap break-words">{JSON.stringify(report, null, 2)}</pre></div>}</div></div>;
};
export default ReportPage;
