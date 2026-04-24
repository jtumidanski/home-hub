import { PackageSummaryWidget } from "@/components/features/packages/package-summary-widget";

export interface PackagesSummaryAdapterConfig {
  title?: string;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function PackagesSummaryAdapter({ config: _config }: { config: PackagesSummaryAdapterConfig }) {
  // PackageSummaryWidget manages its own title; the registry config
  // accepts an optional override for future use.
  return <PackageSummaryWidget />;
}
