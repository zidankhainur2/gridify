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
            background: '#111620',
            borderColor: '#2A3447',
            color: '#F8FAFC',
          },
        }}
      />
    </>
  );
}
