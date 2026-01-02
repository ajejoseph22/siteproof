import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, Company, Project, Ticket } from '@/types';
import { authApi, projectsApi, ticketsApi, isDemoMode } from '@/lib/api';
import { mockProjects, mockTickets } from '@/data/mock';

// Get system preference
const getSystemPreference = (): boolean => {
  if (typeof window !== 'undefined') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  }
  return false;
};

// Apply dark mode to DOM
const applyDarkMode = (isDark: boolean) => {
  if (typeof document !== 'undefined') {
    const root = document.documentElement;
    if (isDark) {
      root.classList.add('dark');
      root.style.colorScheme = 'dark';
    } else {
      root.classList.remove('dark');
      root.style.colorScheme = 'light';
    }
  }
};

interface AppState {
  // Auth state
  isAuthenticated: boolean;
  user: User | null;
  company: Company | null;

  // Theme
  themePreference: 'light' | 'dark' | 'system';
  isDarkMode: boolean;

  // Data
  projects: Project[];
  tickets: Ticket[];
  isLoading: boolean;

  // Actions
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => void;
  toggleDarkMode: () => void;
  setThemePreference: (preference: 'light' | 'dark' | 'system') => void;
  initializeTheme: () => void;

  // Data actions
  fetchProjects: (status?: string) => Promise<void>;
  fetchTickets: (filters?: { projectId?: string; status?: string }) => Promise<void>;

  // Project actions
  setProjects: (projects: Project[]) => void;

  // Ticket actions
  setTickets: (tickets: Ticket[]) => void;
  addTicket: (ticket: Ticket) => void;
  updateTicket: (id: string, updates: Partial<Ticket>) => void;
  deleteTicket: (id: string) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      // Initial state
      isAuthenticated: false,
      user: null,
      company: null,
      themePreference: 'system',
      isDarkMode: getSystemPreference(),
      projects: [],
      tickets: [],
      isLoading: false,

      // Auth actions
      login: async (email: string, password: string) => {
        const result = await authApi.login(email, password);

        if (result) {
          // In demo mode, also load mock data
          const projectsData = isDemoMode() ? mockProjects : await projectsApi.listProjects();
          const ticketsData = isDemoMode() ? mockTickets : await ticketsApi.listTickets();

          set({
            isAuthenticated: true,
            user: result.user,
            company: result.company,
            projects: projectsData,
            tickets: ticketsData,
          });
          return true;
        }
        return false;
      },

      logout: async () => {
        await authApi.logout();
        set({
          isAuthenticated: false,
          user: null,
          company: null,
          projects: [],
          tickets: [],
        });
      },

      initializeTheme: () => {
        const { themePreference } = get();
        let isDark: boolean;

        if (themePreference === 'system') {
          isDark = getSystemPreference();
        } else {
          isDark = themePreference === 'dark';
        }

        set({ isDarkMode: isDark });
        applyDarkMode(isDark);
      },

      setThemePreference: (preference: 'light' | 'dark' | 'system') => {
        let isDark: boolean;

        if (preference === 'system') {
          isDark = getSystemPreference();
        } else {
          isDark = preference === 'dark';
        }

        set({ themePreference: preference, isDarkMode: isDark });
        applyDarkMode(isDark);
      },

      toggleDarkMode: () => {
        const { isDarkMode } = get();
        const newDarkMode = !isDarkMode;
        const newPreference: 'light' | 'dark' = newDarkMode ? 'dark' : 'light';

        set({
          isDarkMode: newDarkMode,
          themePreference: newPreference
        });
        applyDarkMode(newDarkMode);
      },

      // Data fetching
      fetchProjects: async (status?: string) => {
        set({ isLoading: true });
        try {
          const projects = await projectsApi.listProjects(status);
          set({ projects });
        } finally {
          set({ isLoading: false });
        }
      },

      fetchTickets: async (filters?: { projectId?: string; status?: string }) => {
        set({ isLoading: true });
        try {
          const tickets = await ticketsApi.listTickets(filters as Parameters<typeof ticketsApi.listTickets>[0]);
          set({ tickets });
        } finally {
          set({ isLoading: false });
        }
      },

      // Project actions
      setProjects: (projects) => set({ projects }),

      // Ticket actions
      setTickets: (tickets) => set({ tickets }),

      addTicket: (ticket) => set((state) => ({
        tickets: [ticket, ...state.tickets],
      })),

      updateTicket: (id, updates) => set((state) => ({
        tickets: state.tickets.map((t) =>
          t.id === id ? { ...t, ...updates, updatedAt: new Date().toISOString() } : t
        ),
      })),

      deleteTicket: (id) => set((state) => ({
        tickets: state.tickets.filter((t) => t.id !== id),
      })),
    }),
    {
      name: 'siteproof-storage',
      partialize: (state) => ({
        themePreference: state.themePreference,
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        company: state.company,
      }),
    }
  )
);

// Initialize theme on first load and listen for system preference changes
if (typeof window !== 'undefined') {
  // Apply initial theme
  const stored = localStorage.getItem('siteproof-storage');
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      const preference = parsed?.state?.themePreference || 'system';
      let isDark: boolean;

      if (preference === 'system') {
        isDark = getSystemPreference();
      } else {
        isDark = preference === 'dark';
      }

      applyDarkMode(isDark);
    } catch {
      applyDarkMode(getSystemPreference());
    }
  } else {
    applyDarkMode(getSystemPreference());
  }

  // Listen for system preference changes
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    const state = useAppStore.getState();
    if (state.themePreference === 'system') {
      const isDark = e.matches;
      useAppStore.setState({ isDarkMode: isDark });
      applyDarkMode(isDark);
    }
  });
}
