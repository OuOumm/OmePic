"use client";

import { createContext, useContext } from "react";
import type { AdminStatus } from "@/types";

const AdminStatusContext = createContext<AdminStatus | null>(null);

export function AdminStatusProvider({
  children,
  verifiedStatus,
}: {
  children: React.ReactNode;
  verifiedStatus: AdminStatus | null;
}) {
  return (
    <AdminStatusContext.Provider value={verifiedStatus}>
      {children}
    </AdminStatusContext.Provider>
  );
}

export function useAdminStatus() {
  return useContext(AdminStatusContext);
}
