'use client';

import { Toaster } from 'sonner';

export default function AppProviders({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      {children}
      <Toaster
        position="top-right"
        richColors
        closeButton
        theme="dark"
        toastOptions={{
          style: {
            background: '#18181B',
            borderColor: '#27272A',
            color: '#F4F4F5',
          },
        }}
      />
    </>
  );
}
