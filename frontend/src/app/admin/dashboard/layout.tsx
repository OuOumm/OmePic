import { PageLayout } from "@/components/shared/PageLayout";
import { AdminShell } from "@/features/admin/AdminShell";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <PageLayout>
      <AdminShell>{children}</AdminShell>
    </PageLayout>
  );
}
