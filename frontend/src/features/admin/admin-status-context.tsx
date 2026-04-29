"use client";

import { createContext, useContext, useMemo } from "react";

import type { AdminStatus } from "@/types/admin";

type AdminStatusContextValue = {
  verifiedStatus: AdminStatus | null;
};

const AdminStatusContext = createContext<AdminStatusContextValue>({ verifiedStatus: null });

export function AdminStatusProvider({
  children,
  verifiedStatus
}: {
  children: React.ReactNode;
  verifiedStatus: AdminStatus | null;
}) {
  const value = useMemo(() => ({ verifiedStatus }), [verifiedStatus]);

  return (
    <AdminStatusContext.Provider value={value}>
      {children}
    </AdminStatusContext.Provider>
  );
}

export function useVerifiedAdminStatus() {
  return useContext(AdminStatusContext).verifiedStatus;
}
