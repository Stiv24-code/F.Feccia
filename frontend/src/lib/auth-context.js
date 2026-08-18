import { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import {
  login as loginApi,
  logout as logoutApi,
  refreshSession,
  getMe,
  registerClient as registerClientApi,
  verifyEmail as verifyEmailApi,
  setAccessToken,
  setOnAuthFailure,
} from './api';
import { logger } from './logger';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  // Al mount: tenta un refresh usando il cookie httpOnly. Se va bene
  // otteniamo un access token fresh + profilo user, senza chiedere credenziali.
  // Se fallisce (nessun cookie / scaduto) restiamo non autenticati.
  const bootstrap = useCallback(async () => {
    try {
      const refreshed = await refreshSession();
      if (refreshed?.data?.access_token) {
        setAccessToken(refreshed.data.access_token);
        if (refreshed.data.user) {
          setUser(refreshed.data.user);
        } else {
          const me = await getMe();
          setUser(me.data);
        }
      }
    } catch (err) {
      // Nessuna sessione valida: stato normale al primo accesso.
      logger.debug('No active session on bootstrap:', err?.response?.status);
      setAccessToken(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // Se il refresh interceptor non riesce, vogliamo sloggarci subito
    // invece di rimanere in stato incoerente.
    setOnAuthFailure(() => {
      setAccessToken(null);
      setUser(null);
    });
    bootstrap();
  }, [bootstrap]);

  const login = useCallback(async (email, password) => {
    const res = await loginApi({ email, password });
    setAccessToken(res.data.access_token);
    setUser(res.data.user);
    return res.data.user;
  }, []);

  // Autoregistrazione cliente: due esiti possibili (vedi lib/api.js).
  // Con verifica email attiva la risposta non contiene un access token —
  // non c'è nessuna sessione da aprire, solo un messaggio da mostrare.
  // Senza verifica (SMTP non configurato) la risposta è identica a login:
  // auto-login immediato, come prima di questa funzionalità.
  const registerClient = useCallback(async (payload) => {
    const res = await registerClientApi(payload);
    if (res.data.verified === false) {
      return { verified: false, message: res.data.message, email: res.data.email };
    }
    setAccessToken(res.data.access_token);
    setUser(res.data.user);
    return { verified: true, user: res.data.user };
  }, []);

  // Completa la verifica email (link ricevuto da registerClient) — successo
  // apre la sessione esattamente come login.
  const verifyEmail = useCallback(async (token) => {
    const res = await verifyEmailApi(token);
    setAccessToken(res.data.access_token);
    setUser(res.data.user);
    return res.data.user;
  }, []);

  const logout = useCallback(async () => {
    try {
      await logoutApi();
    } catch (err) {
      logger.error('Logout endpoint failed:', err?.message);
    }
    setAccessToken(null);
    setUser(null);
  }, []);

  const contextValue = useMemo(
    () => ({ user, login, registerClient, verifyEmail, logout, loading }),
    [user, login, registerClient, verifyEmail, logout, loading],
  );

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
