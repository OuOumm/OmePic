import { PageLayout } from "@/components/shared/PageLayout";
import { AnnouncementModal } from "@/features/announcements/AnnouncementModal";
import { UploadPageClient } from "@/features/upload/UploadPageClient";

export default function HomePage() {
  return (
    <PageLayout>
      <UploadPageClient />
      <AnnouncementModal />
    </PageLayout>
  );
}
