import { BrandMark } from "@/components/common/brand-mark";
import { SidebarTrigger } from "@/components/ui/sidebar";

export function MobileHeader() {
  return (
    <header className="flex h-14 items-center gap-2 border-b bg-sidebar px-3 md:hidden">
      <SidebarTrigger aria-label="Open navigation menu" />
      <BrandMark className="size-7 shrink-0" />
      <span className="text-lg font-semibold leading-none">Home Hub</span>
    </header>
  );
}
