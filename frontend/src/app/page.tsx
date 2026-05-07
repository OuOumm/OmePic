import { PageLayout } from "@/components/shared/PageLayout";
import { UploadPageClient } from "@/features/upload/UploadPageClient";

export default function HomePage() {
  return (
    <PageLayout>
      <UploadPageClient />
    </PageLayout>
  );
}
