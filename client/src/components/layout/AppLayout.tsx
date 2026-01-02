import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  LayoutDashboard,
  FolderKanban,
  ClipboardList,
  Settings,
  LogOut,
  Plus,
  Moon,
  Sun,
  Menu,
  X,
  HardHat,
  ChevronRight,
} from 'lucide-react';
import { useState, useEffect } from 'react';
import { useAppStore } from '@/stores/app';
import { Avatar, Button } from '@/components/ui';
import { cn } from '@/lib/utils';

const navItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/projects', icon: FolderKanban, label: 'Projects' },
  { to: '/tickets', icon: ClipboardList, label: 'Tickets' },
  { to: '/settings', icon: Settings, label: 'Settings' },
];

export function AppLayout() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const [isDesktop, setIsDesktop] = useState(false);
  const { user, company, isDarkMode, toggleDarkMode, logout } = useAppStore();
  const navigate = useNavigate();

  // Track screen size for proper sidebar behavior
  useEffect(() => {
    const checkDesktop = () => setIsDesktop(window.innerWidth >= 1024);
    checkDesktop();
    window.addEventListener('resize', checkDesktop);
    return () => window.removeEventListener('resize', checkDesktop);
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  // Desktop: sidebar always visible. Mobile: controlled by state
  const showSidebar = isDesktop || isSidebarOpen;

  return (
    <div className="min-h-screen bg-charcoal-50 dark:bg-charcoal-950">
      {/* ════════════════════════════════════════════════════════════════════════
          MOBILE HEADER - Fixed with solid background
          ════════════════════════════════════════════════════════════════════════ */}
      <header className="lg:hidden fixed top-0 left-0 right-0 z-40 bg-white dark:bg-charcoal-900 border-b-2 border-charcoal-200 dark:border-charcoal-800 shadow-md">
        <div className="flex items-center justify-between h-16 px-4">
          <button
            onClick={() => setIsSidebarOpen(true)}
            className="p-2.5 rounded-lg bg-charcoal-100 dark:bg-charcoal-800 hover:bg-charcoal-200 dark:hover:bg-charcoal-700 transition-colors active:scale-95"
            aria-label="Open menu"
          >
            <Menu className="w-5 h-5 text-charcoal-700 dark:text-charcoal-200" />
          </button>

          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-md bg-gradient-to-br from-safety-500 to-safety-600 flex items-center justify-center shadow-md">
              <HardHat className="w-5 h-5 text-white" />
            </div>
            <span className="font-display text-xl tracking-wide text-charcoal-900 dark:text-white">
              SITEPROOF
            </span>
          </div>

          <button
            onClick={toggleDarkMode}
            className="p-2.5 rounded-lg bg-charcoal-100 dark:bg-charcoal-800 hover:bg-charcoal-200 dark:hover:bg-charcoal-700 transition-colors active:scale-95"
            aria-label="Toggle theme"
          >
            {isDarkMode ? (
              <Sun className="w-5 h-5 text-safety-500" />
            ) : (
              <Moon className="w-5 h-5 text-charcoal-600" />
            )}
          </button>
        </div>
        {/* Construction tape accent */}
        <div className="h-1 bg-gradient-to-r from-safety-500 via-charcoal-900 to-safety-500 dark:via-charcoal-950" />
      </header>

      {/* ════════════════════════════════════════════════════════════════════════
          MOBILE OVERLAY
          ════════════════════════════════════════════════════════════════════════ */}
      <AnimatePresence>
        {isSidebarOpen && !isDesktop && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 z-50 bg-charcoal-950/70 backdrop-blur-sm lg:hidden"
            onClick={() => setIsSidebarOpen(false)}
          />
        )}
      </AnimatePresence>

      {/* ════════════════════════════════════════════════════════════════════════
          SIDEBAR - Desktop: always visible, Mobile: slide-in sheet
          ════════════════════════════════════════════════════════════════════════ */}
      <AnimatePresence mode="wait">
        {showSidebar && (
          <motion.aside
            initial={isDesktop ? false : { x: '-100%' }}
            animate={{ x: 0 }}
            exit={{ x: '-100%' }}
            transition={{ type: 'spring', damping: 30, stiffness: 300 }}
            className={cn(
              'fixed top-0 left-0 bottom-0 z-50 w-72',
              'bg-white dark:bg-charcoal-900',
              'border-r-2 border-charcoal-200 dark:border-charcoal-800',
              'shadow-2xl lg:shadow-lg',
              'lg:z-30'
            )}
          >
            <div className="flex flex-col h-full">
              {/* ─────────────────────────────────────────────────────────────────
                  Logo Section
                  ───────────────────────────────────────────────────────────────── */}
              <div className="p-5 border-b-2 border-charcoal-100 dark:border-charcoal-800">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-safety-500 to-safety-600 flex items-center justify-center shadow-lg shadow-safety-500/25">
                      <HardHat className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <span className="font-display text-xl tracking-wide block text-charcoal-900 dark:text-white">
                        SITEPROOF
                      </span>
                      <span className="text-xs font-medium text-charcoal-500 dark:text-charcoal-400 truncate block max-w-[140px]">
                        {company?.name || 'Construction Co.'}
                      </span>
                    </div>
                  </div>
                  {/* Mobile close button */}
                  <button
                    onClick={() => setIsSidebarOpen(false)}
                    className="lg:hidden p-2 rounded-lg bg-charcoal-100 dark:bg-charcoal-800 hover:bg-charcoal-200 dark:hover:bg-charcoal-700 transition-colors"
                    aria-label="Close menu"
                  >
                    <X className="w-5 h-5 text-charcoal-600 dark:text-charcoal-300" />
                  </button>
                </div>
              </div>

              {/* Construction tape */}
              <div className="construction-tape-sm" />

              {/* ─────────────────────────────────────────────────────────────────
                  Navigation Links
                  ───────────────────────────────────────────────────────────────── */}
              <nav className="flex-1 p-4 space-y-1.5 overflow-y-auto">
                {navItems.map((item, index) => (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    onClick={() => setIsSidebarOpen(false)}
                    className={({ isActive }) =>
                      cn(
                        'group flex items-center gap-3 px-4 py-3.5 rounded-xl font-medium transition-all duration-200',
                        isActive
                          ? 'bg-gradient-to-r from-safety-500 to-safety-600 text-white shadow-lg shadow-safety-500/30'
                          : 'text-charcoal-600 dark:text-charcoal-400 hover:bg-charcoal-100 dark:hover:bg-charcoal-800 hover:text-charcoal-900 dark:hover:text-white'
                      )
                    }
                    style={{ animationDelay: `${index * 50}ms` }}
                  >
                    <item.icon className="w-5 h-5 flex-shrink-0" />
                    <span className="flex-1">{item.label}</span>
                    <ChevronRight className="w-4 h-4 opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all" />
                  </NavLink>
                ))}
              </nav>

              {/* ─────────────────────────────────────────────────────────────────
                  Quick Action Button
                  ───────────────────────────────────────────────────────────────── */}
              <div className="p-4 border-t-2 border-charcoal-100 dark:border-charcoal-800">
                <Button
                  onClick={() => {
                    setIsSidebarOpen(false);
                    navigate('/tickets/new');
                  }}
                  className="w-full"
                  size="lg"
                  leftIcon={<Plus className="w-5 h-5" />}
                >
                  New Ticket
                </Button>
              </div>

              {/* ─────────────────────────────────────────────────────────────────
                  User Section
                  ───────────────────────────────────────────────────────────────── */}
              <div className="p-4 border-t-2 border-charcoal-100 dark:border-charcoal-800 bg-charcoal-50 dark:bg-charcoal-950/50">
                <div className="flex items-center gap-3 mb-4">
                  <Avatar
                    name={`${user?.firstName} ${user?.lastName}`}
                    src={user?.avatarUrl}
                    size="md"
                  />
                  <div className="flex-1 min-w-0">
                    <p className="font-semibold text-charcoal-900 dark:text-white truncate">
                      {user?.firstName} {user?.lastName}
                    </p>
                    <p className="text-sm text-charcoal-500 dark:text-charcoal-400 truncate">
                      {user?.email}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={toggleDarkMode}
                    className="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-lg text-sm font-medium bg-charcoal-100 dark:bg-charcoal-800 text-charcoal-700 dark:text-charcoal-300 hover:bg-charcoal-200 dark:hover:bg-charcoal-700 transition-colors"
                  >
                    {isDarkMode ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
                    <span>{isDarkMode ? 'Light' : 'Dark'}</span>
                  </button>
                  <button
                    onClick={handleLogout}
                    className="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-lg text-sm font-medium bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors"
                  >
                    <LogOut className="w-4 h-4" />
                    <span>Logout</span>
                  </button>
                </div>
              </div>
            </div>
          </motion.aside>
        )}
      </AnimatePresence>

      {/* ════════════════════════════════════════════════════════════════════════
          MAIN CONTENT AREA
          ════════════════════════════════════════════════════════════════════════ */}
      <main className={cn(
        'min-h-screen transition-all duration-300',
        'pt-[68px] lg:pt-0', // Account for mobile header
        'lg:ml-72' // Offset for desktop sidebar
      )}>
        {/* Desktop construction tape */}
        <div className="hidden lg:block construction-tape" />

        {/* Page content */}
        <div className="p-4 sm:p-6 lg:p-8">
          <Outlet />
        </div>
      </main>

      {/* ════════════════════════════════════════════════════════════════════════
          MOBILE FAB - Quick ticket creation
          ════════════════════════════════════════════════════════════════════════ */}
      <motion.button
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.95 }}
        onClick={() => navigate('/tickets/new')}
        className={cn(
          'lg:hidden fixed right-4 bottom-6 z-30',
          'w-14 h-14 rounded-2xl',
          'bg-gradient-to-br from-safety-500 to-safety-600',
          'text-white shadow-xl shadow-safety-500/40',
          'flex items-center justify-center',
          'border-2 border-safety-400/50'
        )}
        aria-label="Create new ticket"
      >
        <Plus className="w-6 h-6" />
      </motion.button>
    </div>
  );
}
