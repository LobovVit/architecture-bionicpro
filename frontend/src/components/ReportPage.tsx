import React, { useEffect, useState } from 'react';

type MeResponse = {
  authenticated: boolean;
  username: string;
  roles: string[];
  session_id: string;
};

type ReportResponse = {
  owner: string;
  report_generated: string;
  prosthesis_ids: string[];
  summary: {
    daily_movements: number;
    battery_cycles: number;
    calibration_state: string;
  };
  security: {
    token_on_client: boolean;
    session_rotated: boolean;
    new_session_id?: string;
  };
};

type YandexProfile = {
  id: string;
  login: string;
  default_email: string;
  first_name?: string;
  last_name?: string;
  display_name?: string;
  real_name?: string;
  default_avatar_id?: string;
};

type YandexProfileResponse = {
  session_id: string;
  profile: YandexProfile;
};

type LogoutResponse = {
  logout_url: string;
};

const API_URL = process.env.REACT_APP_AUTH_URL || 'http://localhost:8081';

const ReportPage: React.FC = () => {
  const [me, setMe] = useState<MeResponse | null>(null);
  const [report, setReport] = useState<ReportResponse | null>(null);
  const [yandexProfile, setYandexProfile] = useState<YandexProfile | null>(null);

  const [loading, setLoading] = useState(true);
  const [reportLoading, setReportLoading] = useState(false);
  const [consentLoading, setConsentLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadSession = async () => {
    try {
      const response = await fetch(`${API_URL}/auth/me`, { credentials: 'include' });

      if (!response.ok) {
        setMe(null);
        return;
      }

      const data: MeResponse = await response.json();
      setMe(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load session');
    } finally {
      setLoading(false);
    }
  };

  const loadYandexProfile = async () => {
    try {
      const response = await fetch(`${API_URL}/auth/yandex/profile`, {
        credentials: 'include',
      });

      if (!response.ok) {
        setYandexProfile(null);
        return;
      }

      const data: YandexProfileResponse = await response.json();
      setYandexProfile(data.profile ?? null);
    } catch {
      setYandexProfile(null);
    }
  };

  useEffect(() => {
    void (async () => {
      await loadSession();
    })();
  }, []);

  useEffect(() => {
    if (!me) {
      setYandexProfile(null);
      return;
    }

    void loadYandexProfile();
  }, [me]);

  const login = () => {
    window.location.href = `${API_URL}/auth/login`;
  };

  const logout = async () => {
    try {
      setError(null);

      const response = await fetch(`${API_URL}/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error(`Logout failed with status ${response.status}`);
      }

      const data: LogoutResponse = await response.json();

      setMe(null);
      setReport(null);
      setYandexProfile(null);

      if (data.logout_url) {
        window.location.href = data.logout_url;
        return;
      }

      window.location.reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to logout');
    }
  };

  const downloadReport = async () => {
    try {
      setReportLoading(true);
      setError(null);

      const response = await fetch(`${API_URL}/api/reports`, {
        credentials: 'include',
      });

      if (response.status === 401) {
        setMe(null);
        setError('Session expired. Please sign in again.');
        return;
      }

      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      const data: ReportResponse = await response.json();
      setReport(data);

      await loadSession();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load report');
    } finally {
      setReportLoading(false);
    }
  };

  const approveYandexConsent = async () => {
    try {
      setConsentLoading(true);
      setError(null);

      const response = await fetch(`${API_URL}/auth/yandex/consent`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ approve: true }),
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error(text || `Consent failed with status ${response.status}`);
      }

      setYandexProfile(null);
      await loadSession();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save Yandex profile');
    } finally {
      setConsentLoading(false);
    }
  };

  const declineYandexConsent = () => {
    setYandexProfile(null);
  };

  if (loading) {
    return (
        <div className="flex items-center justify-center min-h-screen bg-gray-100">
          Loading...
        </div>
    );
  }

  if (!me) {
    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 gap-4">
          <div className="p-8 bg-white rounded-lg shadow-md text-center">
            <h1 className="text-2xl font-bold mb-2">Usage Reports</h1>
            <p className="text-gray-600 mb-6">
              Sign in to access your prosthesis usage report.
            </p>
            <button
                onClick={login}
                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
            >
              Sign in
            </button>
            {error && (
                <div className="mt-4 p-4 bg-red-100 text-red-700 rounded">
                  {error}
                </div>
            )}
          </div>
        </div>
    );
  }

  return (
      <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 p-6">
        <div className="w-full max-w-2xl p-8 bg-white rounded-lg shadow-md">
          <div className="flex items-start justify-between gap-4 mb-6">
            <div>
              <h1 className="text-2xl font-bold">Usage Reports</h1>
              <p className="text-gray-600 mt-1">
                Signed in as <span className="font-medium">{me.username}</span>
              </p>
              <p className="text-xs text-gray-500 mt-1">Session: {me.session_id}</p>
            </div>

            <button
                onClick={logout}
                className="px-3 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
            >
              Logout
            </button>
          </div>

          {yandexProfile && (
              <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded">
                <h2 className="text-lg font-semibold mb-2">Yandex profile consent</h2>
                <p className="text-sm text-gray-700 mb-4">
                  We received your Yandex profile data. Allow BionicPRO to use and save
                  this information?
                </p>

                <div className="text-sm bg-white border rounded p-3 mb-4">
                  <div><strong>ID:</strong> {yandexProfile.id}</div>
                  <div><strong>Login:</strong> {yandexProfile.login}</div>
                  <div><strong>Email:</strong> {yandexProfile.default_email}</div>
                  <div>
                    <strong>Name:</strong>{' '}
                    {yandexProfile.real_name ||
                        `${yandexProfile.first_name ?? ''} ${yandexProfile.last_name ?? ''}`.trim() ||
                        yandexProfile.display_name ||
                        '-'}
                  </div>
                </div>

                <div className="flex gap-3">
                  <button
                      onClick={approveYandexConsent}
                      disabled={consentLoading}
                      className={`px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 ${
                          consentLoading ? 'opacity-50 cursor-not-allowed' : ''
                      }`}
                  >
                    {consentLoading ? 'Saving...' : 'Allow and save'}
                  </button>

                  <button
                      onClick={declineYandexConsent}
                      disabled={consentLoading}
                      className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
                  >
                    Cancel
                  </button>
                </div>
              </div>
          )}

          <button
              onClick={downloadReport}
              disabled={reportLoading}
              className={`px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 ${
                  reportLoading ? 'opacity-50 cursor-not-allowed' : ''
              }`}
          >
            {reportLoading ? 'Generating report...' : 'Download report'}
          </button>

          {error && (
              <div className="mt-4 p-4 bg-red-100 text-red-700 rounded">
                {error}
              </div>
          )}

          {report && (
              <div className="mt-6 p-4 bg-gray-50 rounded border">
                <h2 className="text-lg font-semibold mb-3">Latest report</h2>
                <pre className="text-sm whitespace-pre-wrap break-words">
              {JSON.stringify(report, null, 2)}
            </pre>
              </div>
          )}
        </div>
      </div>
  );
};

export default ReportPage;