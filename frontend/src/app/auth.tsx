import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState
} from "react";

import { api } from "../lib/api";
import type { AppSession, AuthTokens, UserProfile } from "../lib/types";
import { safeParseJson } from "../lib/utils";

type AuthContextValue = {
  session: AppSession | null;
  isBooting: boolean;
  login: (email: string, password: string) => Promise<AppSession>;
  register: (token: string, password: string) => Promise<AppSession>;
  logout: () => void;
  refreshProfile: () => Promise<void>;
  updateProfile: (input: Partial<Pick<UserProfile, "photoUrl" | "about">>) => Promise<void>;
};

const STORAGE_KEY = "profdnk.session.v1";
const AuthContext = createContext<AuthContextValue | null>(null);

function readStoredSession() {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);
  const parsed = safeParseJson<{ tokens: AuthTokens; profile: UserProfile } | null>(raw, null);

  if (!parsed && raw) {
    window.localStorage.removeItem(STORAGE_KEY);
  }

  return parsed;
}

function persistSession(nextSession: AppSession | null) {
  if (typeof window === "undefined") {
    return;
  }

  if (!nextSession) {
    window.localStorage.removeItem(STORAGE_KEY);
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextSession));
}

async function hydrateSession(tokens: AuthTokens) {
  const profile = await api.getProfile(tokens.accessToken);
  return {
    tokens,
    profile
  };
}

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<AppSession | null>(null);
  const [isBooting, setIsBooting] = useState(true);

  useEffect(() => {
    const stored = readStoredSession();

    if (!stored) {
      setIsBooting(false);
      return;
    }

    hydrateSession(stored.tokens)
      .then((nextSession) => {
        setSession(nextSession);
        persistSession(nextSession);
      })
      .catch(async () => {
        try {
          const nextTokens = await api.refreshToken(stored.tokens.refreshToken);
          const nextSession = await hydrateSession(nextTokens);
          setSession(nextSession);
          persistSession(nextSession);
        } catch {
          setSession(null);
          persistSession(null);
        }
      })
      .finally(() => {
        setIsBooting(false);
      });
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      isBooting,
      async login(email: string, password: string) {
        const tokens = await api.login(email, password);
        const nextSession = await hydrateSession(tokens);
        setSession(nextSession);
        persistSession(nextSession);
        return nextSession;
      },
      async register(token: string, password: string) {
        const tokens = await api.register(token, password);
        const nextSession = await hydrateSession(tokens);
        setSession(nextSession);
        persistSession(nextSession);
        return nextSession;
      },
      logout() {
        setSession(null);
        persistSession(null);
      },
      async refreshProfile() {
        if (!session) {
          return;
        }

        const profile = await api.getProfile(session.tokens.accessToken);
        const nextSession = {
          ...session,
          profile
        };
        setSession(nextSession);
        persistSession(nextSession);
      },
      async updateProfile(input) {
        if (!session) {
          return;
        }

        const profile = await api.updateProfile(session.tokens.accessToken, input);
        const nextSession = {
          ...session,
          profile
        };
        setSession(nextSession);
        persistSession(nextSession);
      }
    }),
    [isBooting, session]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);

  if (!context) {
    throw new Error("useAuth must be used inside AuthProvider");
  }

  return context;
}
