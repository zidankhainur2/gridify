import { redirect } from 'next/navigation';
import { BadgeCheck, Mail, Shield, User as UserIcon } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { User } from '@/types/api';
import BasicInfoForm from '@/features/auth/components/basic-info-form';
import SecurityForm from '@/features/auth/components/security-form';
import { Badge } from '@/components/ui/badge';

export default async function ProfilePage() {
  let user: User;

  try {
    user = await apiClient.users.getMe();
  } catch {
    redirect('/login');
  }

  const fullName =
    user.first_name && user.last_name
      ? `${user.first_name} ${user.last_name}`
      : user.username;

  const initials =
    user.first_name && user.last_name
      ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
      : user.username.slice(0, 2).toUpperCase();

  return (
    <div className="max-w-5xl mx-auto space-y-6 animate-in fade-in duration-200 pb-12">
      {/* Header Profile Card */}
      <div className="p-6 md:p-8 rounded-xl bg-surface-raised border border-border-default flex flex-col md:flex-row items-center gap-6">
        {/* Avatar Capsule */}
        <div className="h-20 w-20 md:h-24 md:w-24 rounded-xl bg-surface-elevated border border-border-default flex items-center justify-center shrink-0 text-accent font-bold text-2xl font-mono shadow-xs">
          {initials}
        </div>

        {/* User Info */}
        <div className="text-center md:text-left flex-1">
          <div className="flex flex-col md:flex-row md:items-center gap-2 mb-1">
            <h1 className="text-2xl font-bold tracking-tight text-text-primary">
              {fullName}
            </h1>
            <Badge variant="purple" className="self-center md:self-auto">
              @{user.username}
            </Badge>
          </div>

          <div className="flex flex-wrap items-center justify-center md:justify-start gap-3 text-xs text-text-muted mt-2">
            <div className="flex items-center gap-1.5 px-3 py-1 rounded-md bg-surface-base border border-border-subtle">
              <Mail size={13} className="text-text-muted" />
              <span>{user.email}</span>
            </div>
            <div className="flex items-center gap-1.5 px-3 py-1 rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium">
              <BadgeCheck size={13} />
              <span>Verified Organizer</span>
            </div>
          </div>
        </div>
      </div>

      {/* Main Grid: Data Diri & Keamanan */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        <div className="lg:col-span-2">
          <BasicInfoForm user={user} />
        </div>
        <div className="lg:col-span-1">
          <SecurityForm user={user} />
        </div>
      </div>
    </div>
  );
}