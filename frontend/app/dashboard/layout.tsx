'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import {
  LayoutDashboard,
  Trophy,
  User as UserIcon,
  LogOut,
  Plus,
  Menu,
  X,
  Gamepad2,
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { User } from '@/types/api';
import { toast } from 'sonner';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    let mounted = true;
    apiClient.users
      .getMe()
      .then((data) => {
        if (mounted) setUser(data);
      })
      .catch(() => {
        // jika gagal getMe, redirect proxy akan menangani atau biarkan login
      });
    return () => {
      mounted = false;
    };
  }, []);

  const handleLogout = async () => {
    try {
      await apiClient.auth.logout();
      toast.success('Berhasil keluar');
      router.push('/login');
      router.refresh();
    } catch {
      router.push('/login');
    }
  };

  const menuItems = [
    { name: 'Ringkasan', icon: LayoutDashboard, href: '/dashboard' },
    { name: 'Turnamen Saya', icon: Trophy, href: '/dashboard/tournaments' },
    { name: 'Profil', icon: UserIcon, href: '/dashboard/profile' },
  ];

  const userInitial = user?.first_name
    ? user.first_name.charAt(0).toUpperCase()
    : user?.username?.charAt(0).toUpperCase() || 'U';

  const userDisplayName =
    user?.first_name && user?.last_name
      ? `${user.first_name} ${user.last_name}`
      : user?.username || 'Organizer';

  return (
    <div className="flex h-screen bg-surface-base text-text-primary overflow-hidden font-sans">
      {/* Mobile Header Bar */}
      <div className="md:hidden fixed top-0 left-0 right-0 z-50 bg-surface-raised border-b border-border-default px-4 py-3 flex items-center justify-between">
        <Link href="/dashboard" className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
            <Gamepad2 size={18} />
          </div>
          <span className="font-bold tracking-tight text-text-primary">Gridify</span>
        </Link>
        <button
          onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
          className="p-2 text-text-muted hover:text-text-primary transition-colors"
          aria-label="Toggle Menu"
        >
          {isMobileMenuOpen ? <X size={22} /> : <Menu size={22} />}
        </button>
      </div>

      {/* Sidebar */}
      <aside
        className={`
        fixed md:static inset-y-0 left-0 z-40 w-64 bg-surface-raised border-r border-border-default
        transform transition-transform duration-200 ease-in-out shrink-0 flex flex-col justify-between
        ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
      `}
      >
        <div className="flex flex-col flex-1 overflow-y-auto">
          {/* Logo Area */}
          <div className="p-6 border-b border-border-subtle flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
              <Gamepad2 size={18} />
            </div>
            <div>
              <span className="text-lg font-bold tracking-tight text-text-primary block leading-none">
                Gridify
              </span>
              <span className="text-[11px] font-mono text-text-muted">Tournament OS</span>
            </div>
          </div>

          {/* User Info Capsule */}
          <div className="p-4 mx-3 my-3 rounded-lg bg-surface-base border border-border-subtle flex items-center gap-3">
            <div className="w-9 h-9 rounded-md bg-surface-elevated border border-border-default flex items-center justify-center text-accent font-bold text-sm shrink-0">
              {userInitial}
            </div>
            <div className="overflow-hidden">
              <p className="text-sm font-medium text-text-primary truncate leading-tight">
                {userDisplayName}
              </p>
              <p className="text-xs text-text-muted truncate">
                @{user?.username || 'user'}
              </p>
            </div>
          </div>

          {/* Navigation Links */}
          <nav className="px-3 space-y-1 mt-2">
            {menuItems.map((item) => {
              const isActive =
                item.href === '/dashboard'
                  ? pathname === '/dashboard'
                  : pathname?.startsWith(item.href);

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className={`flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-surface-elevated text-accent border border-border-default'
                      : 'text-text-muted hover:text-text-primary hover:bg-surface-base'
                  }`}
                >
                  <item.icon size={18} className={isActive ? 'text-accent' : 'text-text-muted'} />
                  <span>{item.name}</span>
                </Link>
              );
            })}
          </nav>
        </div>

        {/* Action Buttons: Buat Turnamen & Keluar */}
        <div className="p-4 border-t border-border-subtle space-y-2">
          <Link
            href="/dashboard/create"
            onClick={() => setIsMobileMenuOpen(false)}
            className="flex items-center justify-center gap-2 w-full py-2.5 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-hover transition-colors shadow-sm"
          >
            <Plus size={16} />
            <span>Buat Turnamen</span>
          </Link>

          <button
            onClick={handleLogout}
            className="flex items-center gap-2.5 px-3.5 py-2 w-full rounded-lg text-text-muted hover:text-danger hover:bg-danger/10 text-sm font-medium transition-colors"
          >
            <LogOut size={16} />
            <span>Keluar</span>
          </button>
        </div>
      </aside>

      {/* Backdrop for Mobile Menu */}
      {isMobileMenuOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/60 backdrop-blur-xs md:hidden"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* Main Content Area */}
      <main className="flex-1 relative overflow-y-auto pt-16 md:pt-0">
        <div className="max-w-7xl mx-auto p-4 md:p-8">{children}</div>
      </main>
    </div>
  );
}