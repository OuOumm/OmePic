import { PageLayout } from "@/components/shared/PageLayout";
import { HistoryPageClient } from "@/features/history/HistoryPageClient";

export default function HistoryPage() {
  return (
    <PageLayout>
      <HistoryPageClient />
    </PageLayout>
  );
}
