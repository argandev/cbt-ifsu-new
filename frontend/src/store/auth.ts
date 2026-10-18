// useAuthStore — state auth global (Zustand, Tim 1).
// Sumber kebenaran: server (GET /api/v1/auth/me). Store hanya cache.
import { create } from "zustand";
import apiClient, { ApiError } from "../lib/api";

export interface AuthUser {
  id: string;
  nama: string;
  email: string | null;
  username: string | null;
  role: string;
  must_change_password: boolean;
  is_active: boolean;
}

interface MeResponse {
  user: AuthUser;
  permissions: string[];
}

interface AuthState {
  user: AuthUser | null;
  permissions: string[];
  loading: boolean;
  /** Ambil /auth/me. 401 -> user null (belum login). */
  fetchMe: () => Promise<void>;
  /** Login. Melempar ApiError; kode PASSWORD_CHANGE_REQUIRED ditangani pemanggil. */
  login: (login: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  hasPermission: (perm: string) => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  permissions: [],
  loading: true,

  fetchMe: async () => {
    set({ loading: true });
    try {
      const data = await apiClient.get<MeResponse>("/api/v1/auth/me");
      set({ user: data.user, permissions: data.permissions, loading: false });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        set({ user: null, permissions: [], loading: false });
        return;
      }
      set({ loading: false });
      throw err;
    }
  },

  login: async (login, password) => {
    const data = await apiClient.post<MeResponse>("/api/v1/auth/login", { login, password });
    set({ user: data.user, permissions: data.permissions, loading: false });
  },

  logout: async () => {
    try {
      await apiClient.post("/api/v1/auth/logout");
    } finally {
      set({ user: null, permissions: [], loading: false });
    }
  },

  hasPermission: (perm) => {
    const { user, permissions } = get();
    if (!user) return false;
    if (user.role === "admin") return true;
    return permissions.includes(perm);
  },
}));
