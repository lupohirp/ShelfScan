export const getApiUrl = (): string => {
  const envUrl = import.meta.env.VITE_API_URL;
  if (envUrl) {
    return envUrl;
  }
  const hostname = window.location.hostname;
  if (hostname === 'localhost' || hostname === '127.0.0.1' || hostname.startsWith('192.168.') || hostname.startsWith('10.')) {
    return `http://${hostname}:8080`;
  }
  if (import.meta.env.PROD) {
    const baseHost = hostname.replace(/^admin-/, '');
    return `https://api-${baseHost}`;
  }
  return `http://${hostname}:8080`;
};
