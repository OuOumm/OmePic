import { AdminShell } from "@/features/admin/AdminShell";

export default function AdminDashboardLayout({
  children
}: Readonly<{ children: React.ReactNode }>) {
  return <AdminShell>{children}</AdminShell>;
}
